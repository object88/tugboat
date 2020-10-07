package cache

import (
	"context"
	"fmt"
	"sync"

	"helm.sh/helm/v3/pkg/repo"
)

// RepoCache contains the in-memory cache for a particular repository
type RepoCache struct {
	*Config

	repository *repo.Entry

	contents map[string]*countedrepo
	lock     sync.RWMutex
}

type countedrepo struct {
	contents map[string]*countedversion
	count    uint
}

type countedversion struct {
	cv    *repo.ChartVersion
	count uint
}

// NewRepoCache returns a new instance of a RepoCache
func NewRepoCache(config *Config, repository *repo.Entry) (*RepoCache, error) {
	if config == nil {
		return nil, fmt.Errorf("No Config provided to NewRepoCache")
	}

	rc := &RepoCache{
		Config:     config,
		repository: repository,

		contents: map[string]*countedrepo{},
	}
	return rc, nil
}

// Get returns a pointer to the chart.ChartVersion. If the chart is not cached,
// it will refresh the repo index and attempt to download it.
func (rc *RepoCache) Get(name, version string) (*repo.ChartVersion, bool) {
	get := func(name, version string) (*repo.ChartVersion, bool) {
		rc.lock.RLock()
		defer rc.lock.RUnlock()

		r, ok := rc.contents[name]
		if !ok {
			rc.Logger.Info("repocache does not have chart", "repository", rc.repository, "chart", name)
			return nil, false
		}

		v, ok := r.contents[version]
		if !ok {
			rc.Logger.Info("repocache does not have chart and version", "repository", rc.repository, "chart", name, "version", version)
			return nil, false
		}

		return v.cv, true
	}

	metadata, ok := get(name, version)
	if ok {
		return metadata, true
	}

	rc.Logger.Info("cache miss; loading", "repository", rc.repository, "chart", name, "version", version)

	// We have a cache miss; start a request for the missing repo contents.  This
	// is a blocking operation.  Once it returns, the cache should be repopulated
	// with a call to `Load`.
	err := rc.Loader.Load(context.Background(), rc.repository.Name)
	if err != nil {
		return nil, false
	}

	metadata, ok = get(name, version)
	if !ok {
		rc.Logger.Info("second cache miss; failed", "repository", rc.repository, "chart", name, "version", version)
		return nil, false
	}

	return metadata, true
}

// Load is the callback from the Loader
func (rc *RepoCache) Load(index *repo.IndexFile) error {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	// TODO: retain state and augment instead of wholesale replacement
	rc.contents = map[string]*countedrepo{}

	rc.Logger.Info("repocache contents have been reset", "repository", rc.repository)

	repocount := 0
	versioncount := 0
	for name, versions := range index.Entries {
		v, ok := rc.contents[name]
		if !ok {
			v = &countedrepo{
				contents: map[string]*countedversion{},
			}
			rc.contents[name] = v
			repocount++
		}

		for _, version := range versions {
			v.contents[version.Version] = &countedversion{
				cv: version,
			}
			versioncount++
		}
	}

	for _, r := range rc.contents {
		for _, v := range r.contents {
			rc.Logger.Info("item", "chart", v.cv.Name, "version", v.cv.Metadata.Version)
		}
	}

	rc.Logger.Info("repocache loaded with charts", "repository", rc.repository, "repository-count", repocount, "version-count", versioncount)

	return nil
}

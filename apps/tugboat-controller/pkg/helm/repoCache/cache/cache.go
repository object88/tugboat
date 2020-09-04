package cache

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/repoCache/loader"
	"github.com/object88/tugboat/pkg/errs"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
)

const (
	// ErrMissingRepository is returned when a function is called with an unknown
	// repository
	ErrMissingRepository = errs.ConstError("Repository is not in cache")
)

// Cache contains chart metadata in memory
type Cache struct {
	Config

	contents     map[string]*RepoCache
	contentsLock sync.RWMutex
}

// Config contains configuration shared between an instance of a Cache and the
// RepoCache instances it collects
type Config struct {
	Loader loader.Loader
	Logger logr.Logger
}

func (c *Cache) Initialize(repoLoader loader.RepoLoader) {
	c.contents = map[string]*RepoCache{}
	c.Loader.CacheLoader = c
	c.Loader.RepoLoader = repoLoader
	c.Loader.Queue.Initialize(c.Loader.Work)
}

// AddRepository is used to add a new repository to the cache.  This will
// grab a write-lock, so should be used before processing begins.
func (c *Cache) AddRepository(repository string) {
	c.contentsLock.Lock()
	defer c.contentsLock.Unlock()

	if _, ok := c.contents[repository]; !ok {
		rc, err := NewRepoCache(&c.Config, repository)
		if err != nil {
			// TODO: Not sure what to do with an error here
		}
		c.contents[repository] = rc
	}
}

// Get retrieves the chart metadata from in-memory cache.  Upon a cache miss,
// it will use the Loader to attempt to find the chart.
func (c *Cache) Get(ctx context.Context, repository, name string, version string) (*chart.Metadata, bool, error) {
	c.contentsLock.RLock()
	defer c.contentsLock.RUnlock()

	rc, ok := c.contents[repository]
	if !ok {
		return nil, false, ErrMissingRepository
	}
	metadata, ok := rc.Get(name, version)
	return metadata, ok, nil
}

// Load populates a particular RepoCache from the provided index.  It aquires a
// read-lock on the local RepoCache map, and a write lock on the particular
// RepoCache.
func (c *Cache) Load(repository string, index *repo.IndexFile) error {
	c.contentsLock.RLock()
	defer c.contentsLock.RUnlock()

	c.Logger.Info("Loading repository contents", "repository", repository)

	rc, ok := c.contents[repository]
	if !ok {
		return ErrMissingRepository
	}

	return rc.Load(index)
}

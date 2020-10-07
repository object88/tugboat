package repos

import (
	"context"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos/cache"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

// Cache manages helm charts
type Cache struct {
	// cache is not a pointer; it is embedded
	cache cache.Cache

	logger logr.Logger

	settings *cli.EnvSettings

	f *repo.File
}

// New is provided for consistency.  The `Helm` struct is properly configured
// with the `Connect` func.
func New() *Cache {
	h := &Cache{}
	h.cache.Initialize(h)
	return h
}

func (h *Cache) Connect(opts ...OptionFunc) error {
	ho := HelmOptions{}
	for _, opt := range opts {
		if err := opt(&ho); err != nil {
			return fmt.Errorf("failed to configure options: %w", err)
		}
	}

	if ho.logger == nil {
		z, err := zap.NewProduction()
		if err != nil {
			return fmt.Errorf("no logger provided and failed to create production looger: %w", err)
		}
		ho.logger = zapr.NewLogger(z)
	}

	h.settings = ho.settings
	h.logger = ho.logger
	h.cache.Logger = h.logger
	if ho.cooldown != 0 {
		h.cache.Loader.Queue.Cooldown = ho.cooldown
	}
	if ho.timeout != 0 {
		h.cache.Loader.Queue.Timeout = ho.timeout
	}

	var f *repo.File
	if _, err := os.Stat(h.settings.RepositoryConfig); err != nil {
		// Did not find the file; creating new one
		h.logger.Info("did not find repo file, creating new file", "location", h.settings.RepositoryConfig)
		f = repo.NewFile()
	} else {
		// Have file; loading.
		h.logger.Info("have repo file, loading", "location", h.settings.RepositoryConfig)
		f, err = repo.LoadFile(h.settings.RepositoryConfig)
		if err != nil {
			return fmt.Errorf("failed to open repository config '%s': %w", h.settings.RepositoryConfig, err)
		}
	}

	h.f = f

	return nil
}

// UpsertRepo adds or updates a repository to the repository config file
func (h *Cache) UpsertRepo(repo *repo.Entry) error {
	// Ensure that we have the helm repo.
	h.logger.Info("checking for museum", "museum", repo.Name, "url", repo.URL)

	// TODO: determine action if partial match per name or URL
	found := false
	for _, entry := range h.f.Repositories {
		if entry.URL == repo.URL {
			found = true
			break
		}
	}
	if !found {
		h.logger.Info("Did not find repo; adding...", "repository", repo.Name, "url", repo.URL)
		h.f.Add(repo)
		h.logger.Info("Added.")
		err := h.f.WriteFile(h.settings.RepositoryConfig, 0644)
		if err != nil {
			return fmt.Errorf("failed to write repository config '%s': %w", h.settings.RepositoryConfig, err)
		}
		h.logger.Info("Saved repository file", "file", h.settings.RepositoryConfig)

		h.cache.AddRepository(repo)
	}

	return nil
}

// UpdateRepositories reads the index files of each repository.
func (h *Cache) UpdateRepositories() error {
	if h.f == nil {
		return fmt.Errorf("must call `Connect` before")
	}

	// // Use a little concurrency to update many repositories
	// var wg sync.WaitGroup
	// wg.Add(len(h.f.Repositories))

	for _, entry := range h.f.Repositories {
		h.cache.AddRepository(entry)
	}

	return nil
}

// GetChartVersion retrieves the metadata for a particular chart and version
func (h *Cache) GetChartVersion(chartrepo, name string, version *semver.Version) (*repo.ChartVersion, error) {
	v := version.String()
	cv, ok, err := h.cache.Get(context.Background(), chartrepo, name, v)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("Failed to find '%s/%s:%s'", chartrepo, name, version)
	}
	return cv, nil
}

func (h *Cache) GetChartRepository(chartrepo string) (*repo.Entry, error) {
	return h.cache.GetRepository(chartrepo)
}

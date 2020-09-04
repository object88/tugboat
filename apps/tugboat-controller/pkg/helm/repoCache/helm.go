package repoCache

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/repoCache/cache"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

// Helm manages helm charts
type Helm struct {
	// cache is not a pointer; it is not shared properties
	cache cache.Cache

	logger logr.Logger

	settings *cli.EnvSettings

	f *repo.File
}

// New is provided for consistency.  The `Helm` struct is properly configured
// with the `Connect` func.
func New() *Helm {
	h := &Helm{}
	h.cache.Initialize(h)
	return h
}

func (h *Helm) Connect(opts ...OptionFunc) error {
	ho := HelmOptions{}
	for _, opt := range opts {
		if err := opt(&ho); err != nil {
			return fmt.Errorf("failed to configure options: %w", err)
		}
	}

	h.settings = ho.settings
	h.logger = ho.logger
	h.cache.Logger = h.logger

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

// EnsureRepo adds a repository to the repository config file
func (h *Helm) EnsureRepo(repo *repo.Entry) error {
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

		h.cache.AddRepository(repo.Name)
	}

	return nil
}

// UpdateRepositories reads the index files of each repository.
func (h *Helm) UpdateRepositories() error {
	if h.f == nil {
		return fmt.Errorf("must call `Connect` before")
	}

	// Use a little concurrency to update many repositories
	var wg sync.WaitGroup
	wg.Add(len(h.f.Repositories))

	for _, entry := range h.f.Repositories {
		h.cache.AddRepository(entry.Name)
	}

	return nil
}

func (h *Helm) GetChartMetadata(chartrepo, name string, version *semver.Version) (*chart.Metadata, error) {
	v := version.String()
	metadata, ok, err := h.cache.Get(context.Background(), chartrepo, name, v)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("Failed to find '%s/%s:%s'", chartrepo, name, version)
	}
	return metadata, nil
}

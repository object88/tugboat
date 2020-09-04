package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

// GetRepoIndexFile satisfies `loader.GetRepoIndexFile`
func (h *Helm) GetRepoIndexFile(name string) (*repo.IndexFile, error) {
	// h.logger.Infof("Getting chart metadata for '%s/%s@%s'\n", chartrepo, name, version.String())
	var e *repo.Entry
	for _, entry := range h.f.Repositories {
		// h.logger.Infof("Checking repo '%s'\n", entry.Name)
		if entry.Name == name {
			e = entry
			break
		}
	}
	if e == nil {
		// Failed to find the repository
		// h.logger.Infof("No entries of %d matched\n", len(h.f.Repositories))
		return nil, fmt.Errorf("failed to find repository '%s'", name)
	}

	cr, err := repo.NewChartRepository(e, getter.All(h.settings))
	if err != nil {
		return nil, fmt.Errorf("failed to create new chart repository '%s': %w", name, err)
	}

	h.logger.Info("downloading index file", "repository", name, "url", cr.Config.URL)

	indexFilePath, err := cr.DownloadIndexFile()
	if err != nil {
		return nil, fmt.Errorf("failed to download index for chart repository '%s' at '%s': %w", cr.Config.Name, cr.Config.URL, err)
	}

	h.logger.Info("loading index file", "repository", name)

	index, err := repo.LoadIndexFile(indexFilePath)

	h.logger.Info("loaded index file", "repository", name)

	return index, nil
}

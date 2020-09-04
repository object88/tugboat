package loader

import (
	"context"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/repoCache/queue"
	"helm.sh/helm/v3/pkg/repo"
)

type CacheLoader interface {
	Load(repository string, index *repo.IndexFile) error
}

type RepoLoader interface {
	GetRepoIndexFile(repo string) (*repo.IndexFile, error)
}

// Loader is the connection between `cache.RepoCache` and `helm.Helm`
type Loader struct {
	CacheLoader CacheLoader
	RepoLoader  RepoLoader

	Queue queue.Queue
}

// Load uses the queue to ensure a single invocation of `Work`, which uses the
// native helm packages to download an index file and populate the cache.
func (l *Loader) Load(ctx context.Context, repo string) error {
	return l.Queue.Work(ctx, repo)
}

// Work implements `queue.Worker`
func (l *Loader) Work(ctx context.Context, repo string) error {
	if l == nil {
		return ErrNilPointer
	}

	index, err := l.RepoLoader.GetRepoIndexFile(repo)
	if err != nil {
		return err
	}

	// Call back into cache
	return l.CacheLoader.Load(repo, index)
}

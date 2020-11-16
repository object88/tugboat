package watchers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/pkg/http/probes"
	"k8s.io/client-go/tools/cache"
)

type WatchManager struct {
	log logr.Logger
}

func New(log logr.Logger) *WatchManager {
	return &WatchManager{
		log: log,
	}
}

func (wm *WatchManager) Run(ctx context.Context, r probes.Reporter, watchers ...Watcher) error {
	stopper := make(chan struct{})
	defer func() {
		r.NotReady()
		close(stopper)
		wm.log.Info("closed watchmanager stopper")
	}()

	go func() {
		for _, w := range watchers {
			ssi := wm.startInformer(w.GetInformer(), stopper)
			if ssi == nil {
				// TODO: if there _is_ an error, it goes unnoticed and we never get to
				// ready.  This may functionally work out correctly, but it would be
				// better to get out of `Run` entirely.
				return
			}
		}
		r.Ready()
	}()

	for {
		select {
		case <-ctx.Done():
			wm.log.Info("watchmanager context complete")
			return ctx.Err()
		}
	}
}

func (wm *WatchManager) startInformer(informer cache.SharedIndexInformer, stopper <-chan struct{}) cache.SharedIndexInformer {
	go func() {
		informer.Run(stopper)
		wm.log.Info("informer complete")
	}()

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		wm.log.Info("watchmanager failed to sync cache")
		return nil
	}

	wm.log.Info("watchmanager informer running")
	return informer
}

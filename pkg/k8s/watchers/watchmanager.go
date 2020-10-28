package watchers

import (
	"context"

	"github.com/go-logr/logr"
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

func (wm *WatchManager) Run(ctx context.Context, watchers ...Watcher) error {
	stopper := make(chan struct{})
	defer func() {
		close(stopper)
		wm.log.Info("closed watchmanager stopper")
	}()

	go func() {
		for _, w := range watchers {
			wm.startInformer(w.GetInformer(), stopper)
		}
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

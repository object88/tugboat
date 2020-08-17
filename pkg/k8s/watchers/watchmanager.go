package watchers

import (
	"context"

	"k8s.io/client-go/tools/cache"
)

type WatchManager struct {
}

func New() *WatchManager {
	return &WatchManager{}
}

func (wm *WatchManager) Run(ctx context.Context, watchers ...Watcher) error {
	stopper := make(chan struct{})
	defer close(stopper)

	go func() {
		for _, w := range watchers {
			wm.startInformer(w.GetInformer(), stopper)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (wm *WatchManager) startInformer(informer cache.SharedIndexInformer, stopper <-chan struct{}) cache.SharedIndexInformer {
	go informer.Run(stopper)

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		return nil
	}

	return informer
}

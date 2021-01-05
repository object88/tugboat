package informermanager

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/pkg/http/probes"
	"k8s.io/client-go/tools/cache"
)

type Manager struct {
	log logr.Logger
}

func New(log logr.Logger) *Manager {
	return &Manager{
		log: log,
	}
}

func (m *Manager) Run(ctx context.Context, r probes.Reporter, informers ...cache.SharedIndexInformer) error {
	stopper := make(chan struct{})
	defer func() {
		r.NotReady()
		close(stopper)
		m.log.Info("closed watchmanager stopper")
	}()

	go func() {
		for _, i := range informers {
			m.startInformer(i, stopper)
		}
		r.Ready()
	}()

	for {
		select {
		case <-ctx.Done():
			m.log.Info("watchmanager context complete")
			return ctx.Err()
		}
	}
}

func (m *Manager) startInformer(informer cache.SharedIndexInformer, stopper <-chan struct{}) {
	// TODO: if there _is_ an error, it goes unnoticed and we never get to
	// ready.  This may functionally work out correctly, but it would be
	// better to get out of `Run` entirely.

	go func() {
		informer.Run(stopper)
		m.log.Info("informer complete")
	}()

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		m.log.Info("watchmanager failed to sync cache")
		return
	}

	m.log.Info("watchmanager informer running")
}

package watcher

import (
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Watcher struct {
	log       logr.Logger
	clientset *kubernetes.Clientset
}

func New(log logr.Logger, clientset *kubernetes.Clientset) *Watcher {
	return &Watcher{
		clientset: clientset,
		log:       log,
	}
}

func (w *Watcher) GetInformer() cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactory(w.clientset, 10*time.Second)

	// factory.Discovery().V1beta1().EndpointSlices().Informer()
	informer := factory.Core().V1().Secrets().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.added,
		UpdateFunc: w.updated,
		DeleteFunc: w.deleted,
	})

	return informer
}

func (w *Watcher) added(obj interface{}) {
	w.log.Info("added", "obj", obj)
}

func (w *Watcher) updated(oldObj interface{}, newObj interface{}) {
	w.log.Info("updated", "oldObj", oldObj, "newObj", newObj)
}

func (w *Watcher) deleted(obj interface{}) {
	w.log.Info("deleted", "obj", obj)
}

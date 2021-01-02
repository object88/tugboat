package watcher

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/internal/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Watcher struct {
	log       logr.Logger
	clientset *kubernetes.Clientset
	dyn       dynamic.Interface
}

func New(log logr.Logger, clientset *kubernetes.Clientset) *Watcher {
	return &Watcher{
		clientset: clientset,
		log:       log,
	}
}

func (w *Watcher) GetInformer() cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactory(w.clientset, 10*time.Second)
	// factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(w.dyn, 0, v1.NamespaceAll, nil)

	informer := factory.Core().V1().Secrets().Informer()
	// informer := factory.ForResource().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.added,
		UpdateFunc: w.updated,
		DeleteFunc: w.deleted,
	})

	return informer
}

func (w *Watcher) added(obj interface{}) {
	newScrt, ok := obj.(*v1.Secret)
	if !ok || newScrt.Type != constants.HelmSecretType {
		return
	}

	w.log.Info("added", "name", newScrt.Name, "namespace", newScrt.Namespace)
}

func (w *Watcher) updated(oldObj interface{}, newObj interface{}) {
	// w.log.Info("updated", "oldObj", oldObj, "newObj", newObj)
}

func (w *Watcher) deleted(obj interface{}) {
	oldScrt, ok := obj.(*v1.Secret)
	if !ok || oldScrt.Type != constants.HelmSecretType {
		return
	}

	w.log.Info("deleted", "name", oldScrt.Name, "namespace", oldScrt.Namespace)
}

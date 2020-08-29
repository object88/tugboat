package watcher

import (
	"fmt"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Watcher struct {
	clientset *kubernetes.Clientset
}

func New(clientset *kubernetes.Clientset) *Watcher {
	return &Watcher{
		clientset: clientset,
	}
}

func (w *Watcher) GetInformer() cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactory(w.clientset, 10*time.Second)

	// factory.Discovery().V1beta1().EndpointSlices().Informer()
	informer := factory.Core().V1().Pods().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.added,
		UpdateFunc: w.updated,
		DeleteFunc: w.deleted,
	})

	return informer
}

func (w *Watcher) added(obj interface{}) {
	fmt.Printf("added: %#v\n", obj)
}

func (w *Watcher) updated(oldObj interface{}, newObj interface{}) {
	fmt.Printf("updated: %$#v\n", oldObj)
}

func (w *Watcher) deleted(obj interface{}) {
	fmt.Printf("deleted: %#v\n", obj)
}

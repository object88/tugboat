package watcher

import (
	"fmt"
	"time"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/informers/externalversions"
	"k8s.io/client-go/tools/cache"
)

type LaunchWatcher struct {
	clientset *versioned.Clientset
}

func New(clientset *versioned.Clientset) *LaunchWatcher {
	return &LaunchWatcher{
		clientset: clientset,
	}
}

func (lw *LaunchWatcher) GetInformer() cache.SharedIndexInformer {
	factory := externalversions.NewSharedInformerFactory(lw.clientset, 10*time.Second)
	informer := factory.Tugboat().V1alpha1().Launches().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    lw.added,
		UpdateFunc: lw.updated,
		DeleteFunc: lw.deleted,
	})

	return informer
}

func (lw *LaunchWatcher) added(obj interface{}) {
	lnch, ok := obj.(*v1alpha1.Launch)
	if !ok {
		return
	}

	fmt.Printf("added: %s\n", lnch.GetName())
}

func (lw *LaunchWatcher) updated(oldObj interface{}, newObj interface{}) {
	oldLnch, ok := oldObj.(*v1alpha1.Launch)
	if !ok {
		return
	}
	newLnch, ok := newObj.(*v1alpha1.Launch)
	if !ok {
		return
	}
	if oldLnch.UID == newLnch.UID {
		// Actually unchanged
		return
	}
	fmt.Printf("updated: %s\n", oldLnch.GetName())
}

func (lw *LaunchWatcher) deleted(obj interface{}) {
	lnch, ok := obj.(*v1alpha1.Launch)
	if !ok {
		return
	}

	fmt.Printf("deleted: %s\n", lnch.GetName())
}

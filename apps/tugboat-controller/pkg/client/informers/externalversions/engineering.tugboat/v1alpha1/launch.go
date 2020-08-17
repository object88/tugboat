/*
LICENSE
*/
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	engineeringtugboatv1alpha1 "github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	versioned "github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned"
	internalinterfaces "github.com/object88/tugboat/apps/tugboat-controller/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/object88/tugboat/apps/tugboat-controller/pkg/client/listers/engineering.tugboat/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// LaunchInformer provides access to a shared informer and lister for
// Launches.
type LaunchInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.LaunchLister
}

type launchInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewLaunchInformer constructs a new informer for Launch type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewLaunchInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredLaunchInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredLaunchInformer constructs a new informer for Launch type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredLaunchInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.TugboatV1alpha1().Launches(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.TugboatV1alpha1().Launches(namespace).Watch(context.TODO(), options)
			},
		},
		&engineeringtugboatv1alpha1.Launch{},
		resyncPeriod,
		indexers,
	)
}

func (f *launchInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredLaunchInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *launchInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&engineeringtugboatv1alpha1.Launch{}, f.defaultInformer)
}

func (f *launchInformer) Lister() v1alpha1.LaunchLister {
	return v1alpha1.NewLaunchLister(f.Informer().GetIndexer())
}

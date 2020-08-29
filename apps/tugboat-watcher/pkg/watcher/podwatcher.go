package watcher

import (
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type PodWatcher struct {
	log       logr.Logger
	clientset *kubernetes.Clientset

	encoder runtime.Encoder
}

func NewPodWatcher(log logr.Logger, clientset *kubernetes.Clientset) *PodWatcher {
	s := runtime.NewScheme()
	v1.AddToScheme(s)

	codecs := serializer.NewCodecFactory(s)
	enc := unstructured.NewJSONFallbackEncoder(codecs.LegacyCodec(s.PrioritizedVersionsAllGroups()...))
	return &PodWatcher{
		log:       log,
		clientset: clientset,
		encoder:   enc,
	}
}

func (w *PodWatcher) GetInformer() cache.SharedIndexInformer {
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

func (w *PodWatcher) added(obj interface{}) {
	if p, ok := obj.(*v1.Pod); ok {
		w.log.Info("added pod", "name", p.Name)
	}
}

func (w *PodWatcher) updated(oldObj interface{}, newObj interface{}) {
	oldP, ok0 := w.castToPod(oldObj)
	newP, ok1 := w.castToPod(newObj)
	if ok0 && ok1 {
		// Reference material for comparison
		// https://stackoverflow.com/questions/47100389/what-is-the-difference-between-a-resourceversion-and-a-generation
		if oldP.ObjectMeta.ResourceVersion == newP.ObjectMeta.ResourceVersion {
			return
		}

		before, err := runtime.Encode(w.encoder, oldP)
		if err != nil {
			w.log.Error(err, "failed to encode old pod")
			return
		}
		after, err := runtime.Encode(w.encoder, newP)
		if err != nil {
			w.log.Error(err, "failed to encode new pod")
			return
		}
		buf, err := strategicpatch.CreateTwoWayMergePatch(before, after, newP)
		if err != nil {
			w.log.Error(err, "failed to generate patch")
			return
		}

		w.log.Info("updated pod", "name", newP.Name, "patch", string(buf))
	}
}

func (w *PodWatcher) deleted(obj interface{}) {
	if p, ok := obj.(*v1.Pod); ok {
		w.log.Info("deleted pod", "name", p.Name)
	}
}

func (w *PodWatcher) castToPod(obj interface{}) (*v1.Pod, bool) {
	p, ok := obj.(*v1.Pod)
	if !ok {
		w.log.Error(fmt.Errorf("PodWatcher received unexpected type"), "blerg", "type", reflect.TypeOf(obj).String())
		return nil, false
	}
	return p, true
}

package watcher

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	helmSecretType v1.SecretType = "helm.sh/release.v1"
)

type Watcher struct {
	log                logr.Logger
	clientset          *kubernetes.Clientset
	versionedclientset *versioned.Clientset
}

func New(log logr.Logger, clientset *kubernetes.Clientset, versionedclientset *versioned.Clientset) *Watcher {
	return &Watcher{
		clientset:          clientset,
		log:                log,
		versionedclientset: versionedclientset,
	}
}

func (w *Watcher) GetInformer() cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactory(w.clientset, 10*time.Second)

	informer := factory.Core().V1().Secrets().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.added,
		UpdateFunc: w.updated,
		DeleteFunc: w.deleted,
	})

	return informer
}

func (w *Watcher) added(obj interface{}) {
	newScrt, ok := obj.(*v1.Secret)
	if !ok || newScrt.Type != helmSecretType {
		// Either the added object wasn't a secret or it was not a helm release
		// secret.  Either way, ignore it.
		return
	}

	name := newScrt.GetName()
	namespace := newScrt.GetNamespace()
	uid := newScrt.GetUID()

	w.log.Info("added", "name", name, "namespace", namespace, "uid", uid)
	rh := &v1alpha1.ReleaseHistory{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ReleaseHistorySpec{
			ReleaseName:      name,
			ReleaseNamespace: namespace,
			ReleaseUID:       uid,
		},
	}
	_, err := w.versionedclientset.TugboatV1alpha1().ReleaseHistories(namespace).Create(context.Background(), rh, metav1.CreateOptions{})
	if err != nil {
		w.log.Error(err, "failed to create v1alpha1.VersionHistory")
	}
}

func (w *Watcher) updated(oldObj interface{}, newObj interface{}) {
	newScrt, ok := newObj.(*v1.Secret)
	if !ok || newScrt.Type != helmSecretType {
		return
	}

	oldScrt, ok := oldObj.(*v1.Secret)
	if !ok || oldScrt.Type != helmSecretType {
		return
	}

	if newScrt.ObjectMeta.GetGeneration() == oldScrt.ObjectMeta.GetGeneration() {
		return
	}

	w.log.Info("updated", "name", oldScrt.Name, "namespace", oldScrt.Namespace, "uid", oldScrt.UID)
}

func (w *Watcher) deleted(obj interface{}) {
	oldScrt, ok := obj.(*v1.Secret)
	if !ok || oldScrt.Type != helmSecretType {
		return
	}

	w.log.Info("deleted", "name", oldScrt.Name, "namespace", oldScrt.Namespace, "uid", oldScrt.UID)
}

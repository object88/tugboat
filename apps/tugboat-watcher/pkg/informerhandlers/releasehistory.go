package informerhandlers

import (
	"fmt"
	"regexp"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/internal/constants"
	"github.com/object88/tugboat/pkg/k8s/apis/engineering.tugboat/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

const (
	helmSecretNameRegex string = `^sh\.helm\.release\.v1\.(?P<releasename>.+)\.v[1-9][0-9]*$`
)

type ReleaseHistoryInformerHandler struct {
	log                    logr.Logger
	releaseSecretName      *regexp.Regexp
	releaseSecretNameIndex int
}

func NewReleaseHistory(log logr.Logger) (*ReleaseHistoryInformerHandler, error) {
	r, err := regexp.Compile(helmSecretNameRegex)
	if err != nil {
		return nil, fmt.Errorf("internal error; helm secret name regex failed to compile: %w", err)
	}
	index := r.SubexpIndex("releasename")
	if index == -1 {
		return nil, fmt.Errorf("internal error; failed to find 'releasename' subexp in helm secret name regex")
	}

	w := &ReleaseHistoryInformerHandler{
		log:                    log,
		releaseSecretName:      r,
		releaseSecretNameIndex: index,
	}
	return w, nil
}

func (w *ReleaseHistoryInformerHandler) OnAdd(obj interface{}) {
	newRH, ok := obj.(*v1alpha1.ReleaseHistory)
	if !ok {
		return
	}

	w.log.Info("added", "name", newRH.Name, "namespace", newRH.Namespace)
	// newScrt, ok := obj.(*v1.Secret)
	// if !ok || newScrt.Type != HelmSecretType {
	// 	// Either the added object wasn't a secret or it was not a helm release
	// 	// secret.  Either way, ignore it.
	// 	return
	// }

	// name, ok := newScrt.GetLabels()["name"]
	// if !ok {
	// 	// No label, so let's try to extract the release name from the secret name.
	// 	// The secret looks like "sh.helm.release.v1.[RELEASE].v[REVISION]".
	// 	secretname := newScrt.GetName()
	// 	submatches := w.releaseSecretName.FindStringSubmatch(secretname)
	// 	if submatches == nil {
	// 		// The secret name doesn't match the anticipated shape.  Log the problem
	// 		// and get out.
	// 		w.log.Error(fmt.Errorf("failed to get release name"), "secret does not have 'name' secret and secret name does not match regexp", "secretname", secretname)
	// 		return
	// 	}
	// 	name = submatches[w.releaseSecretNameIndex]
	// }
	// namespace := newScrt.GetNamespace()
	// uid := newScrt.GetUID()

	// // TODO: It's _possible_ that helm may create, destroy, and recreate a
	// // release within 1 second.  If that happens, newScret.CreationTimestamp
	// // will be reused, and we will fail to create the ReleaseHistory.  It appears
	// // that the CreationTimestamp does not have milliseconds.  This may need to
	// // be revisited.
	// historyname := fmt.Sprintf("%s-%s", name, newScrt.CreationTimestamp.Format("2006-01-02-15-04-05"))

	// w.log.Info("added", "name", name, "namespace", namespace, "uid", uid)
	// rh := &v1alpha1.ReleaseHistory{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      historyname,
	// 		Namespace: namespace,
	// 		Labels: map[string]string{
	// 			"tugboat.engineering/release-name":      name,
	// 			"tugboat.engineering/release-namespace": namespace,
	// 		},
	// 	},
	// 	Spec: v1alpha1.ReleaseHistorySpec{
	// 		ReleaseName:      name,
	// 		ReleaseNamespace: namespace,
	// 		ReleaseUID:       uid,
	// 	},
	// }
	// _, err := w.versionedclientset.TugboatV1alpha1().ReleaseHistories(namespace).Create(context.Background(), rh, metav1.CreateOptions{})
	// if err != nil {
	// 	w.log.Error(err, "failed to create v1alpha1.ReleaseHistory")
	// }
}

func (w *ReleaseHistoryInformerHandler) OnUpdate(oldObj interface{}, newObj interface{}) {
	newScrt, ok := newObj.(*v1.Secret)
	if !ok || newScrt.Type != constants.HelmSecretType {
		return
	}

	oldScrt, ok := oldObj.(*v1.Secret)
	if !ok || oldScrt.Type != constants.HelmSecretType {
		return
	}

	if newScrt.ObjectMeta.GetGeneration() == oldScrt.ObjectMeta.GetGeneration() {
		return
	}

	w.log.Info("updated", "name", oldScrt.Name, "namespace", oldScrt.Namespace, "uid", oldScrt.UID)
}

func (w *ReleaseHistoryInformerHandler) OnDelete(obj interface{}) {
	oldScrt, ok := obj.(*v1.Secret)
	if !ok || oldScrt.Type != constants.HelmSecretType {
		return
	}

	w.log.Info("deleted", "name", oldScrt.Name, "namespace", oldScrt.Namespace, "uid", oldScrt.UID)
}

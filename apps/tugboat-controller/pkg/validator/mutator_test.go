package validator

import (
	"fmt"
	"testing"
	"time"

	"github.com/object88/tugboat/pkg/k8s/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/pkg/logging/testlogger"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func Test_Mutator_FindDeployingRevision(t *testing.T) {
	l := testlogger.TestLogger{T: t}

	s0 := createSecret("test", "testns", 1, release.StatusSuperseded)
	s1 := createSecret("test", "testns", 2, release.StatusSuperseded)
	s2 := createSecret("test", "testns", 3, release.StatusSuperseded)
	s3 := createSecret("test", "testns", 4, release.StatusPendingInstall)

	// findDeployingRevision requires the secretlister
	stopCh := make(chan struct{})
	defer close(stopCh)
	fakeclientset := fake.NewSimpleClientset(s0, s1, s2, s3)
	factory := informers.NewSharedInformerFactory(fakeclientset, time.Second*1)
	secretinformer := factory.Core().V1().Secrets().Informer()
	go secretinformer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, secretinformer.HasSynced) {
		t.Fatal("watchmanager failed to sync cache")
	}
	secretlister := listercorev1.NewSecretLister(secretinformer.GetIndexer())

	m := M{
		Webhook: Webhook{
			Log: l,
		},
		secretlister: secretlister,
	}

	revs := []v1alpha1.ReleaseHistoryRevision{
		{Revision: v1alpha1.Revision(1)},
		{Revision: v1alpha1.Revision(2)},
		{Revision: v1alpha1.Revision(3)},
		{Revision: v1alpha1.Revision(4)},
	}
	foundrev := m.findDeployingRevision("testns", "test", revs)
	if foundrev == v1alpha1.Revision(0) {
		t.Error("unexpected did not find revision")
	}
	if foundrev != v1alpha1.Revision(4) {
		t.Errorf("found incorrect revision '%d'", foundrev)
	}
}

func createSecret(name string, namespace string, revision int, status release.Status) *v1.Secret {
	now := time.Now()
	createdtime := metav1.Time{Time: now.Add(-60 * time.Minute)}
	// var deletedtime *metav1.Time
	// if deleted {
	// 	deletedtime = &metav1.Time{Time: now.Add(-5 * time.Minute)}
	// }
	secretname := fmt.Sprintf("sh.helm.release.v1.%s.v%d", name, revision)

	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: createdtime,
			// DeletionTimestamp: deletedtime,
			Labels: map[string]string{
				"name":    name,
				"owner":   "helm",
				"status":  status.String(),
				"version": fmt.Sprintf("%d", revision),
			},
			Name:      secretname,
			Namespace: namespace,
		},
	}

	return s
}

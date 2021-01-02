package secret

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned/fake"
	"github.com/object88/tugboat/internal/constants"
	"github.com/object88/tugboat/pkg/logging/testlogger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func Test_markReleaseHistoryUninstalled(t *testing.T) {
	now := metav1.Time{Time: time.Now().Add(-1 * time.Hour)}

	rel := &v1alpha1.ReleaseHistory{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"tugboat.engineering/state": "active",
			},
			Name:      "test",
			Namespace: "testns",
		},
		Spec: v1alpha1.ReleaseHistorySpec{
			ReleaseName: "test",
		},
	}

	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: now,
			Labels: map[string]string{
				"name": "test",
			},
			Namespace: "testns",
		},
	}
	rs := &ReconcileSecret{
		VersionedClient: fake.NewSimpleClientset(rel),
	}
	err := rs.markReleaseHistoryUninstalled(s)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	actual, err := rs.VersionedClient.TugboatV1alpha1().ReleaseHistories("testns").Get(context.TODO(), "test", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Unexpected error getting rel: %s", err.Error())
	}
	if val, ok := actual.Labels["tugboat.engineering/state"]; !ok {
		t.Errorf("releasehistory does not have state label")
	} else if val != "uninstalled" {
		t.Errorf("releasehistory has incorrect state '%s'", val)
	}
}

func Test_Reconcile_SecretWithoutFinalizer(t *testing.T) {
	req, s, rel, _, _ := createSecretAndReleaseHistory("test", "testns", false, false)
	rs := &ReconcileSecret{
		Client:          fakeclient.NewFakeClient(s),
		Log:             testlogger.TestLogger{T: t},
		VersionedClient: fake.NewSimpleClientset(rel),
	}

	if result, err := rs.Reconcile(req); err != nil {
		t.Fatalf("Unexpected error while reconciling: %s", err.Error())
	} else if result.Requeue {
		t.Errorf("Unexpectedly set to requeue")
	}

	if !hasFinalizer(getSecretFromFakeClient(t, rs.Client, s)) {
		t.Errorf("Secret does not have expected finalizer")
	}

	if !hasState(getReleaseHistoryFromFakeClient(t, rs.VersionedClient, rel), "active") {
		t.Errorf("Release history does not have 'active' state")
	}
}

func Test_Reconcile_SecretWithFinalizer(t *testing.T) {
	req, s, rel, _, _ := createSecretAndReleaseHistory("test", "testns", true, false)
	rs := &ReconcileSecret{
		Client:          fakeclient.NewFakeClient(s),
		Log:             testlogger.TestLogger{T: t},
		VersionedClient: fake.NewSimpleClientset(rel),
	}

	if result, err := rs.Reconcile(req); err != nil {
		t.Fatalf("Unexpected error while reconciling: %s", err.Error())
	} else if result.Requeue {
		t.Errorf("Unexpectedly set to requeue")
	}

	if !hasFinalizer(getSecretFromFakeClient(t, rs.Client, s)) {
		t.Errorf("Secret does not have expected finalizer")
	}

	if !hasState(getReleaseHistoryFromFakeClient(t, rs.VersionedClient, rel), "active") {
		t.Errorf("Release history does not have 'active' state")
	}
}

func Test_Reconcile_DeletedSecretWithFinalizer(t *testing.T) {
	req, s, rel, _, _ := createSecretAndReleaseHistory("test", "testns", true, true)
	rs := &ReconcileSecret{
		Client:          fakeclient.NewFakeClient(s),
		Log:             testlogger.TestLogger{T: t},
		VersionedClient: fake.NewSimpleClientset(rel),
	}

	result, err := rs.Reconcile(req)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if result.Requeue {
		t.Errorf("Unexpectedly set to requeue")
	}

	// Ensure that the finalizer has been removed.
	if hasFinalizer(getSecretFromFakeClient(t, rs.Client, s)) {
		t.Errorf("Deleting secret still has finalizer")
	}

	// Ensure that the release history has been marked uninstalled.
	if !hasState(getReleaseHistoryFromFakeClient(t, rs.VersionedClient, rel), "uninstalled") {
		t.Error("Deleting secret does not have uninstalled state")
	}
}

func createSecretAndReleaseHistory(name string, namespace string, withfinalizer bool, deleted bool) (reconcile.Request, *v1.Secret, *v1alpha1.ReleaseHistory, metav1.Time, *metav1.Time) {
	now := time.Now()
	createdtime := metav1.Time{Time: now.Add(-60 * time.Minute)}
	var deletedtime *metav1.Time
	if deleted {
		deletedtime = &metav1.Time{Time: now.Add(-5 * time.Minute)}
	}
	secretname := fmt.Sprintf("sh.helm.release.v1.%s.v1", name)

	rel := &v1alpha1.ReleaseHistory{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"tugboat.engineering/state": "active",
			},
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ReleaseHistorySpec{
			ReleaseName: name,
		},
	}

	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: createdtime,
			DeletionTimestamp: deletedtime,
			Labels: map[string]string{
				"name": name,
			},
			Name:      secretname,
			Namespace: namespace,
		},
	}

	if withfinalizer {
		s.ObjectMeta.Finalizers = []string{constants.HelmSecretFinalizer}
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      secretname,
			Namespace: namespace,
		},
	}

	return req, s, rel, createdtime, deletedtime
}

func getReleaseHistoryFromFakeClient(t *testing.T, c versioned.Interface, original *v1alpha1.ReleaseHistory) *v1alpha1.ReleaseHistory {
	// Ensure that the release history has been marked uninstalled.
	actual, err := c.TugboatV1alpha1().ReleaseHistories(original.Namespace).Get(context.TODO(), original.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Unexpected error getting releasehistory: %s", err.Error())
	}
	return actual
}

func getSecretFromFakeClient(t *testing.T, c client.Client, original *v1.Secret) *v1.Secret {
	actual := &v1.Secret{}
	key, err := client.ObjectKeyFromObject(original)
	if err != nil {
		t.Fatalf("Unexpected error getting key for secret: %s", err.Error())
	}
	if err = c.Get(context.TODO(), key, actual); err != nil {
		t.Fatalf("Unexpected error getting secret back after test: %s", err.Error())
	}
	return actual
}

func hasFinalizer(obj metav1.ObjectMetaAccessor) bool {
	fs := obj.GetObjectMeta().GetFinalizers()
	for _, f := range fs {
		if f == constants.HelmSecretFinalizer {
			return true
		}
	}
	return false
}

func hasState(obj metav1.ObjectMetaAccessor, expected string) bool {
	lbls := obj.GetObjectMeta().GetLabels()
	state, ok := lbls["tugboat.engineering/state"]
	return ok && state == expected
}

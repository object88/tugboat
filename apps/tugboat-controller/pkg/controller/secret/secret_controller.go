package secret

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/predicates"
	"github.com/object88/tugboat/internal/constants"
	"github.com/object88/tugboat/internal/util/slice"
	"github.com/object88/tugboat/pkg/k8s/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// blank assignment to verify that ReconcileSecret implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSecret{}

// ReconcileSecret reconciles a Secret object
type ReconcileSecret struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client          client.Client
	VersionedClient versioned.Interface
	Log             logr.Logger
}

func (r *ReconcileSecret) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		WithLogger(r.Log).
		For(&v1.Secret{}).
		WithEventFilter(predicates.HelmSecretFilterPredicate()).
		// WithEventFilter(predicates.ResourceGenerationOrFinalizerChangedPredicate{}).
		Complete(r)
	if err != nil {
		return err
	}

	return nil
}

// Reconcile will retrieve a Secret and ensure that it has the
// HelmSecretFinalizer if the secret is alive, and removes it if the secret is
// being deleted. If the secret is being deleted, Reconcile will find the
// matching ReleaseHistory and mark it as "uninstalled"
// Reconcile implements reconcile.Reconciler.
func (r *ReconcileSecret) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	recLogger := r.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	recLogger.Info("Reconciling Secret")

	instance := &v1.Secret{}
	err := r.Client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if filteredErr := client.IgnoreNotFound(err); filteredErr != nil {
			recLogger.Error(err, "Error requesting secret")
			return reconcile.Result{}, filteredErr
		}
		// There was an error processing the request; requeue
		return reconcile.Result{}, nil
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Secret is not being deleted; ensure that the finalizer is attached.

		recLogger.Info("helm secret is alive")

		hasFinalizer := false
		for _, f := range instance.ObjectMeta.Finalizers {
			if f == constants.HelmSecretFinalizer {
				hasFinalizer = true
			}
		}
		if !hasFinalizer {
			newInstance := instance.DeepCopy()
			newInstance.ObjectMeta.Finalizers = append(newInstance.ObjectMeta.Finalizers, constants.HelmSecretFinalizer)
			if err := r.Client.Update(ctx, newInstance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Secret is being deleted.  Reset the label on the associated
		// ReleaseHistory, then remove the finalizer

		recLogger.Info("helm secret is being deleted")

		// TODO: change this from just setting a label to migrating to some kind of
		// "archived" release history.
		if r.markReleaseHistoryUninstalled(ctx, instance) {
			// Didn't go well; log and retry
			recLogger.Error(err, "failed to update label on releasehistory")
			return ctrl.Result{Requeue: true}, nil
		}

		index := -1
		for k, f := range instance.ObjectMeta.Finalizers {
			if f == constants.HelmSecretFinalizer {
				index = k
				break
			}
		}
		if index != -1 {
			newInstance := instance.DeepCopy()
			newInstance.Finalizers = slice.RemoveString(newInstance.Finalizers, constants.HelmSecretFinalizer)

			if err := r.Client.Update(ctx, newInstance); err != nil {
				recLogger.Info("failed to update remove finalizer from secret", "err", err.Error())
				return ctrl.Result{Requeue: true}, nil
			}

			recLogger.Info("removed finalizer from helm secret")
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileSecret) markReleaseHistoryUninstalled(ctx context.Context, s *v1.Secret) bool {
	lbls := s.Labels
	chartname, ok := lbls["name"]
	if !ok {
		// Odd.
		r.Log.Info("secret does not have a 'name' label", "name", chartname, "namespace", s.Namespace)
		return false
	}

	// TODO: A helm chart may get "deleted", but end up simply "uninstalling".
	// If the user then deletes the `releasehistory`, and the tugboat controller
	// is restarted, then during the startup process, the reconcile in the
	// startup will attempt to patch a non-existant release history.
	// The lesson: don't be surprised if this patch fails.
	// The TODO: change this to a Get & Update, and ensure that it retries if
	// the Get works and Update doesn't.

	rh, err := r.VersionedClient.TugboatV1alpha1().ReleaseHistories(s.Namespace).Get(ctx, chartname, metav1.GetOptions{})
	if err != nil {
		r.Log.Info("did not get release history", "name", s.Name, "namespace", s.Namespace, "err", err.Error())
		return false
	}

	newrh := rh.DeepCopy()
	if _, ok := newrh.Labels["tugboat.engineering/state"]; ok {
		newrh.Labels["tugboat.engineering/state"] = "uninstalled"
		_, err = r.VersionedClient.TugboatV1alpha1().ReleaseHistories(s.Namespace).Update(ctx, newrh, metav1.UpdateOptions{})
		if err != nil {
			r.Log.Info("failed to update; retrying", "name", chartname, "namespace", s.Namespace, "err", err.Error())
			return true
		}
	}

	return false
}

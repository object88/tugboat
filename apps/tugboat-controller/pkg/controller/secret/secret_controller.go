package secret

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/predicates"
	"github.com/object88/tugboat/internal/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
		WithEventFilter(predicates.HelmSecretFilterPredicate{}).
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
func (r *ReconcileSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	recLogger := r.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	recLogger.Info("Reconciling Secret")

	instance := &v1.Secret{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
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
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, constants.HelmSecretFinalizer)
			if err := r.Client.Update(context.TODO(), instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Secret is being deleted.  Reset the label on the associated
		// ReleaseHistory, then remove the finalizer

		recLogger.Info("helm secret is being deleted")

		if err := r.markReleaseHistoryUninstalled(instance); err != nil {
			// Didn't go well; log and continue.
			recLogger.Error(err, "failed to update label on releasehistory")
		}

		index := -1
		for k, f := range instance.ObjectMeta.Finalizers {
			if f == constants.HelmSecretFinalizer {
				index = k
				break
			}
		}
		if index != -1 {
			pos := []patchOperation{
				{
					Op:   "remove",
					Path: fmt.Sprintf("/metadata/finalizers/%d", index),
				},
			}
			buf, err := json.Marshal(pos)
			if err != nil {
				recLogger.Error(err, "internal error: failed to marshal patch", "err", err.Error())
				return ctrl.Result{}, err
			}

			if err := r.Client.Patch(context.TODO(), instance, client.ConstantPatch(types.JSONPatchType, buf), &client.PatchOptions{}); err != nil {
				recLogger.Error(err, "failed to patch secret", "err", err.Error())
				return ctrl.Result{}, err
			}

			recLogger.Info("removed finalizer from helm secret")

		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileSecret) markReleaseHistoryUninstalled(s *v1.Secret) error {
	lbls := s.Labels
	chartname := lbls["name"]

	pos := []patchOperation{
		{
			Op:    "replace",
			Path:  "/metadata/labels/tugboat.engineering~1state",
			Value: "uninstalled",
		},
	}

	buf, err := json.Marshal(pos)
	if err != nil {
		return err
	}

	_, err = r.VersionedClient.
		TugboatV1alpha1().
		ReleaseHistories(s.Namespace).
		Patch(context.TODO(), chartname, types.JSONPatchType, buf, metav1.PatchOptions{})
	return err
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

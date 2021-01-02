package releasehistory

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/predicates"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// blank assignment to verify that ReconcileReleaseHistory implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileReleaseHistory{}

// ReconcileReleaseHistory reconciles a ReleaseHistory object
type ReconcileReleaseHistory struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ReconcileReleaseHistory) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		WithLogger(r.Log).
		For(&v1alpha1.ReleaseHistory{}).
		WithEventFilter(predicates.ResourceGenerationOrFinalizerChangedPredicate{}).
		Complete(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileReleaseHistory) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	recLogger := r.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	recLogger.Info("Reconciling ReleaseHistory")

	instance := &v1alpha1.ReleaseHistory{}
	err := r.Client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		// There was an error processing the request; requeue
		if !errors.IsNotFound(err) {
			recLogger.Error(err, "Error requesting release history")
			return reconcile.Result{}, err
		}
	}

	// c := repos.New()
	// c.Connect(repos.WithHelmEnvSettings(r.HelmSettings), repos.WithLogger(recLogger))
	// // c.EnsureRepo()

	return reconcile.Result{}, nil
}

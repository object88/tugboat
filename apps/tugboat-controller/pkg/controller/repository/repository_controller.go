package repository

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/predicates"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// blank assignment to verify that ReconcileRepository implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRepository{}

// ReconcileRepository reconciles a Repository object
type ReconcileRepository struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client       client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	HelmSettings *cli.EnvSettings
}

func (r *ReconcileRepository) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithLogger(r.Log).
		For(&v1alpha1.Repository{}).
		WithEventFilter(predicates.ResourceGenerationOrFinalizerChangedPredicate{}).
		Complete(r)
}

func (r *ReconcileRepository) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	recLogger := r.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	recLogger.Info("Reconciling Repository")

	instance := &v1alpha1.Repository{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		// There was an error processing the request; requeue
		if !errors.IsNotFound(err) {
			recLogger.Error(err, "Error requesting launch")
			return reconcile.Result{}, err
		}
	}

	c := repos.New()
	c.Connect(repos.WithHelmEnvSettings(r.HelmSettings), repos.WithLogger(recLogger))
	// c.EnsureRepo()

	return reconcile.Result{}, nil
}

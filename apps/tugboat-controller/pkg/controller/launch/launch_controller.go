package launch

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/charts"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/predicates"
	notificationsclient "github.com/object88/tugboat/internal/notifications/client"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// blank assignment to verify that ReconcileLaunch implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLaunch{}

// ReconcileLaunch reconciles a Launch object
type ReconcileLaunch struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	// Getter       genericclioptions.RESTClientGetter
	Client       client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	HelmSettings *cli.EnvSettings
	Cache        *charts.Cache
	Notifier     *notificationsclient.Client
}

func (r *ReconcileLaunch) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithLogger(r.Log).
		For(&v1alpha1.Launch{}).
		WithEventFilter(predicates.ResourceGenerationOrFinalizerChangedPredicate{}).
		Complete(r)
}

// Reconcile reads that state of the cluster for a Launch object and makes changes based on the state read
// and what is in the Launch.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLaunch) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	recLogger := r.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	recLogger.Info("Reconciling Launch")

	h := helm.New(recLogger, r.HelmSettings, r.Cache)
	name := request.Name

	// Fetch the Launch instance
	instance := &v1alpha1.Launch{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		// There was an error processing the request; requeue
		if !errors.IsNotFound(err) {
			recLogger.Error(err, "Error requesting launch")
			return reconcile.Result{}, err
		}

		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		// Return and don't requeue
		recLogger.Info("Did not find launch")
		if ok, err := h.IsDeployed(name); err != nil {
			recLogger.Error(err, "Check for deployment failed")
			return reconcile.Result{}, err
		} else if ok {
			// The chart is still deployed; remove it now.
			if err := h.Delete(name); err != nil {
				recLogger.Info("Failed to delete chart")
				return reconcile.Result{}, err
			}
		}
	} else {
		// The Launch resource does exist; ensure that the helm deployment is
		// aligned

		if ok, err := h.IsDeployed(name); err != nil {
			recLogger.Error(err, "Check for deployment failed")
			return reconcile.Result{}, err
		} else if ok {
			instance.Status.State = "UPDATING"
			if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, err
			}

			if err := h.Update(name, &instance.Spec); err != nil {
				recLogger.Error(err, "Update failed")
				return reconcile.Result{}, err
			}
		} else {

			if err = r.Notifier.DeploymentStarted(); err != nil {
				recLogger.Info("Failed to notify one or more listeners of started deployment", "error", err)
			}

			instance.Status.State = "INSTALLING"
			if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, err
			}

			if err := h.Deploy(name, &instance.Spec); err != nil {
				recLogger.Error(err, "Install failed")
				return reconcile.Result{}, err
			}
		}

		instance.Status.State = "INSTALLED"
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			recLogger.Error(err, "Update to installed status failed")
			return reconcile.Result{}, err
		}
	}

	recLogger.Info("Complete")

	// Reconciliation complete.
	return reconcile.Result{}, nil
}

package launch

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	launchv1alpha1 "github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm"
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
}

func (r *ReconcileLaunch) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithLogger(r.Log).
		For(&v1alpha1.Launch{}).
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
	reqLogger := r.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Launch")

	// Fetch the Launch instance
	instance := &launchv1alpha1.Launch{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Did not find launch")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Error requesting launch")
		return reconcile.Result{}, err
	}

	name := instance.Name
	l := instance.Spec

	h := helm.New(r.Log, r.HelmSettings)
	if ok, err := h.IsDeployed(name); err != nil {
		reqLogger.Error(err, "Check for deployment failed")
		return reconcile.Result{}, err
	} else if !ok {
		if err := h.Deploy(name, &l); err != nil {
			reqLogger.Error(err, "Install failed")
			return reconcile.Result{}, err
		}
	} else {
		if err := h.Update(name, &l); err != nil {
			reqLogger.Error(err, "Update failed")
			return reconcile.Result{}, err
		}
	}

	// // Define a new Pod object
	// pod := newPodForCR(instance)

	// // Set Launch instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Check if this Pod already exists
	// found := &corev1.Pod{}
	// err = r.Client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	// 	err = r.Client.Create(context.TODO(), pod)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Pod created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// // newPodForCR returns a busybox pod with the same name/namespace as the cr
// func newPodForCR(cr *launchv1alpha1.Launch) *corev1.Pod {
// 	labels := map[string]string{
// 		"app": cr.Name,
// 	}
// 	return &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      cr.Name + "-pod",
// 			Namespace: cr.Namespace,
// 			Labels:    labels,
// 		},
// 		Spec: corev1.PodSpec{
// 			Containers: []corev1.Container{
// 				{
// 					Name:    "busybox",
// 					Image:   "busybox",
// 					Command: []string{"sleep", "3600"},
// 				},
// 			},
// 		},
// 	}
// }

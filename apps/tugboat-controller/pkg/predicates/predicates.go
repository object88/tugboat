package predicates

import (
	"reflect"

	"github.com/object88/tugboat/internal/constants"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// This code is based on an OpenShift article:
// https://www.openshift.com/blog/kubernetes-operators-best-practices

type ResourceGenerationOrFinalizerChangedPredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating resource version change
func (ResourceGenerationOrFinalizerChangedPredicate) UpdateFunc(e event.UpdateEvent) bool {
	if e.ObjectNew.GetGeneration() == e.ObjectNew.GetGeneration() && reflect.DeepEqual(e.ObjectNew.GetFinalizers(), e.ObjectNew.GetFinalizers()) {
		return false
	}
	return true
}

func HelmSecretFilterPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			s, ok := e.Object.(*v1.Secret)
			if !ok {
				return false
			}
			if s.Namespace == "tugboat" || s.Namespace == "kube-system" {
				return false
			}
			if s.Type != constants.HelmSecretType {
				return false
			}

			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			s, ok := e.Object.(*v1.Secret)
			if !ok {
				return false
			}
			if s.Namespace == "tugboat" || s.Namespace == "kube-system" {
				return false
			}
			if s.Type != constants.HelmSecretType {
				return false
			}

			return true
		},
		GenericFunc: func(event.GenericEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			s, ok := e.ObjectOld.(*v1.Secret)
			if !ok {
				return false
			}
			if s.Namespace == "tugboat" || s.Namespace == "kube-system" {
				return false
			}
			if s.Type != constants.HelmSecretType {
				return false
			}
			for _, x := range e.ObjectOld.GetFinalizers() {
				if x == constants.HelmSecretFinalizer {
					return false
				}
			}
			return true
		},
	}
}

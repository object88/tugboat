package predicates

import (
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// This code is based on an OpenShift article:
// https://www.openshift.com/blog/kubernetes-operators-best-practices

type ResourceGenerationOrFinalizerChangedPredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating resource version change
func (ResourceGenerationOrFinalizerChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.MetaNew.GetGeneration() == e.MetaOld.GetGeneration() && reflect.DeepEqual(e.MetaNew.GetFinalizers(), e.MetaOld.GetFinalizers()) {
		return false
	}
	return true
}

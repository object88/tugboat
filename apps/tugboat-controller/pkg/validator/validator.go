package validator

import (
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/pkg/k8s/apis/engineering.tugboat/v1alpha1"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// V ensures that an incoming ReleaseHistory is properly shaped
type V struct {
	Webhook
	scheme *runtime.Scheme
}

func New(log logr.Logger, scheme *runtime.Scheme) *V {
	v := V{
		Webhook: NewWebhook(log),
		scheme:  scheme,
	}
	v.WebhookProcessor = &v
	return &v
}

func (v *V) Process(req *v1.AdmissionRequest) *v1.AdmissionResponse {
	// Validator should be very careful about what it does not allow through.  If
	// there is a problem unmarshaling the object, then we can reject it.  If the
	// object is not a `releasehistory`, let it through; we SHOULD never get
	// them, but just in case something else slips through, let's not be the
	// arbiter.
	// Put another way, we should only reject objects that are not unmarshalable
	// or are a `releasehistory` with invalid properties.

	var obj *v1alpha1.ReleaseHistory
	if err := json.Unmarshal(req.Object.Raw, &obj); err != nil {
		v.Log.Error(err, "Could not unmarshal raw object", "name", req.Name, "namespace", req.Namespace)
		return &v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: err.Error(),
			},
		}
	} else if obj.TypeMeta.APIVersion != "tugboat.engineering/v1alpha1" || obj.TypeMeta.Kind != "ReleaseHistory" {
		err = fmt.Errorf("Unexpected type")
		v.Log.Error(err, "Unmarshalled object was not a releasehistory", "name", req.Name, "namespace", req.Namespace, "actualapiversion", obj.TypeMeta.APIVersion, "actualkind", obj.TypeMeta.Kind)
		return &v1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: err.Error(),
			},
		}
	}

	v.Log.Info("Validated incoming releasehistory", "name", req.Name, "namespace", req.Namespace)
	return &v1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
	}
}

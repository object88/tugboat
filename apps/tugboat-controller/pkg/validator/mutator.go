package validator

import (
	"encoding/json"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type M struct {
	Webhook
	scheme *runtime.Scheme
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func NewMutator(log logr.Logger, scheme *runtime.Scheme) *M {
	m := M{
		Webhook: NewWebhook(log),
		scheme:  scheme,
	}
	m.WebhookProcessor = &m
	return &m
}

// Process implements WebhookProcessor
func (m *M) Process(req *v1.AdmissionRequest) *v1.AdmissionResponse {
	var obj runtime.Object
	if err := json.Unmarshal(req.Object.Raw, &obj); err != nil {
		m.Log.Error(err, "Could not unmarshal raw object", "name", req.Name, "namespace", req.Namespace)
		return &v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: err.Error(),
			},
		}
	}

	// m.Log.Info("Let it through", "name", req.Name, "namespace", req.Namespace)
	// return &v1.AdmissionResponse{
	// 	Allowed: true,
	// }

	switch x := obj.(type) {
	case *corev1.Pod:
		// Proceed
		m.Log.Info("Have pod", "name", req.Name, "namespace", req.Namespace)
		x.Labels["tugboat.engineering/"] = "foo"
	default:
		m.Log.Info("Not pod", "name", req.Name, "namespace", req.Namespace)
		return &v1.AdmissionResponse{
			Allowed: true,
			UID:     req.UID,
		}
	}

	m.Log.Info("Creating patch", "name", req.Name, "namespace", req.Namespace)

	pos := []patchOperation{
		{
			Op:    "add",
			Path:  "/metadata/labels/tugboat.engineering/bar",
			Value: "foo",
		},
	}
	buf, err := json.Marshal(pos)
	if err != nil {
		m.Log.Info("Failed to marshal patch", "name", req.Name, "namespace", req.Namespace)
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	m.Log.Info("Returning response with patch", "name", req.Name, "namespace", req.Namespace)

	return &v1.AdmissionResponse{
		Allowed: true,
		Patch:   buf,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
		UID: req.UID,
	}
}

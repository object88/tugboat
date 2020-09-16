package validator

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
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
	ar := &v1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
	}
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(req.Object.Raw), 4096)

	ext := runtime.RawExtension{}
	if err := d.Decode(&ext); err != nil {
		if err != io.EOF {
			m.Log.Error(err, "")
		}
		return ar
	}

	unstruct := unstructured.Unstructured{
		Object: map[string]interface{}{},
	}
	if err := json.Unmarshal(ext.Raw, &unstruct.Object); err != nil {
		m.Log.Error(err, "Incoming object could not be unmarshaled; ignoring", "name", req.Name, "namespace", req.Namespace)
		return ar
	}

	labels := unstruct.GetLabels()
	labels["engineering.tugboat/foo"] = "bar"
	unstruct.SetLabels(labels)

	m.Log.Info("Creating patch", "name", req.Name, "namespace", req.Namespace)

	pos := []patchOperation{
		{
			Op:    "add",
			Path:  "/metadata/labels/tugboat.engineering~1bar",
			Value: "foo",
		},
	}
	buf, err := json.Marshal(pos)
	if err != nil {
		m.Log.Error(err, "Failed to marshal patch", "name", req.Name, "namespace", req.Namespace)
		return ar
	}

	m.Log.Info("Returning response with patch", "name", req.Name, "namespace", req.Namespace)

	pt := v1.PatchTypeJSONPatch
	ar.Patch = buf
	ar.PatchType = &pt
	return ar
}

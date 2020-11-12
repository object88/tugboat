package validator

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/listers/engineering.tugboat/v1alpha1"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type M struct {
	Webhook
	scheme *runtime.Scheme
	lister v1alpha1.ReleaseHistoryLister
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func NewMutator(log logr.Logger, scheme *runtime.Scheme, lister v1alpha1.ReleaseHistoryLister) *M {
	m := M{
		Webhook: NewWebhook(log),
		scheme:  scheme,
		lister:  lister,
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

	kind := unstruct.GetKind()

	annotations := unstruct.GetAnnotations()
	lbls := unstruct.GetLabels()

	if managedBy, ok := lbls["app.kubernetes.io/managed-by"]; !ok {
		// No managed-by label; ignore this object
		return ar
	} else if managedBy != "Helm" {
		// There is a managed-by label, and it's not helm.  Ignore this also.
		return ar
	}

	helmReleaseName := annotations["meta.helm.sh/release-name"]
	helmReleaseNamespace := annotations["meta.helm.sh/release-namespace"]

	r0, err0 := labels.NewRequirement("tugboat.engineering/release-name", selection.Equals, []string{helmReleaseName})
	r1, err1 := labels.NewRequirement("tugboat.engineering/release-namespace", selection.Equals, []string{helmReleaseNamespace})
	if err0 != nil || err1 != nil {
		m.Log.Info("failed to create requirement", "err0", err0.Error(), "err1", err1.Error())
		return ar
	}
	rhs, err := m.lister.List(labels.NewSelector().Add(*r0, *r1))
	if err != nil {
		m.Log.Info("failed to list", "err", err.Error())
		return ar
	}

	// We are not getting any matched release histories here.  It would seem that
	// they are created too late, if they are created in the watcher.  Will need
	// to reshuffle them into the validator, most likely.
	m.Log.Info("Completed lister search", "count", len(rhs))

	for _, rh := range rhs {
		m.Log.Info("Have matching releasehistory", "name", rh.ObjectMeta.Name)
	}

	lbls["engineering.tugboat/foo"] = "bar"
	unstruct.SetLabels(lbls)

	m.Log.Info("Creating patch", "name", req.Name, "namespace", req.Namespace, "kind", kind)

	pos := []patchOperation{
		{
			Op:    "add",
			Path:  "/metadata/labels/tugboat.engineering~1bar",
			Value: "foo",
		},
	}
	buf, err := json.Marshal(pos)
	if err != nil {
		m.Log.Error(err, "Failed to marshal patch", "name", req.Name, "namespace", req.Namespace, "kind", kind)
		return ar
	}

	m.Log.Info("Returning response with patch", "name", req.Name, "namespace", req.Namespace, "kind", kind)

	pt := v1.PatchTypeJSONPatch
	ar.Patch = buf
	ar.PatchType = &pt
	return ar
}

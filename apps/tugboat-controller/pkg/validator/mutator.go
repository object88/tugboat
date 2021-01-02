package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/listers/engineering.tugboat/v1alpha1"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

type M struct {
	Webhook

	dyn    dynamic.Interface
	mapper *restmapper.DeferredDiscoveryRESTMapper
	scheme *runtime.Scheme
	lister v1alpha1.ReleaseHistoryLister
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func NewMutator(log logr.Logger, scheme *runtime.Scheme, lister v1alpha1.ReleaseHistoryLister, dynamicclient dynamic.Interface, mapper *restmapper.DeferredDiscoveryRESTMapper) *M {
	m := M{
		Webhook: NewWebhook(log),
		scheme:  scheme,
		lister:  lister,
		dyn:     dynamicclient,
		mapper:  mapper,
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

	unstruct := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}
	if err := json.Unmarshal(ext.Raw, &unstruct.Object); err != nil {
		m.Log.Error(err, "Incoming object could not be unmarshaled; ignoring", "name", req.Name, "namespace", req.Namespace)
		return ar
	}

	log := m.Log.WithValues("kind", unstruct.GetKind(), "name", unstruct.GetName(), "namespace", unstruct.GetNamespace())

	unstruct = m.findOwner(log, unstruct)
	if unstruct == nil {
		log.Info("Incoming object does not originate from a tracked helm release")
		return ar
	}

	annotations := unstruct.GetAnnotations()
	helmReleaseName := annotations["meta.helm.sh/release-name"]
	pos := []patchOperation{
		{
			Op:    "add",
			Path:  "/metadata/labels/tugboat.engineering~1releasehistory",
			Value: helmReleaseName,
		},
	}
	buf, err := json.Marshal(pos)
	if err != nil {
		log.Info("Failed to marshal patch", "err", err)
		return ar
	}

	log.Info("patching object")

	pt := v1.PatchTypeJSONPatch
	ar.Patch = buf
	ar.PatchType = &pt
	return ar
}

func (m *M) findOwner(log logr.Logger, unstruct *unstructured.Unstructured) *unstructured.Unstructured {
	if m.checkUnstruct(log, unstruct) {
		log.Info("Found matching owner", "find-name", unstruct.GetName())
		return unstruct
	}

	refs := unstruct.GetOwnerReferences()
	for _, ref := range refs {
		gvk := schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind)
		mapping, err := m.mapper.RESTMapping(schema.GroupKind{
			Group: gvk.Group,
			Kind:  gvk.Kind,
		})
		if err != nil {
			log.Info("failed to get mapping", "find-name", ref.Name, "find-gvk", gvk.String(), "err", err.Error())
			continue
		}
		unstruct0, err := m.dyn.Resource(mapping.Resource).Namespace(unstruct.GetNamespace()).Get(context.TODO(), ref.Name, metav1.GetOptions{})
		if err != nil {
			log.Info("failed to get unstructured object", "find-name", ref.Name, "find-gvk", gvk.String(), "find-mapping", mapping.Resource.String(), "err", err.Error())
			continue
		}
		if u := m.findOwner(log, unstruct0); u != nil {
			return u
		}
	}

	return nil
}

func (m *M) checkUnstruct(log logr.Logger, unstruct *unstructured.Unstructured) bool {
	annotations := unstruct.GetAnnotations()
	lbls := unstruct.GetLabels()

	if managedBy, ok := lbls["app.kubernetes.io/managed-by"]; !ok {
		// No managed-by label; ignore this object
		return false
	} else if managedBy != "Helm" {
		// There is a managed-by label, and it's not helm.  Ignore this also.
		return false
	}

	helmReleaseName := annotations["meta.helm.sh/release-name"]
	helmReleaseNamespace := annotations["meta.helm.sh/release-namespace"]

	r0, err0 := labels.NewRequirement("tugboat.engineering/release-name", selection.Equals, []string{helmReleaseName})
	r1, err1 := labels.NewRequirement("tugboat.engineering/release-namespace", selection.Equals, []string{helmReleaseNamespace})
	if err0 != nil || err1 != nil {
		log.Info("failed to create requirement", "err0", err0.Error(), "err1", err1.Error())
		return false
	}
	rhs, err := m.lister.List(labels.NewSelector().Add(*r0, *r1))
	if err != nil {
		log.Info("failed to list", "err", err.Error())
		return false
	}

	log.Info("searched for releasehistories", "count", len(rhs))

	return len(rhs) != 0
}

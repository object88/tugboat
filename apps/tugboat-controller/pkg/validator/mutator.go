package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/internal/constants"
	"github.com/object88/tugboat/pkg/k8s/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/pkg/k8s/client/clientset/versioned"
	listerv1alpha1 "github.com/object88/tugboat/pkg/k8s/client/listers/engineering.tugboat/v1alpha1"
	"helm.sh/helm/v3/pkg/release"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/restmapper"
)

type M struct {
	Webhook

	dyn                dynamic.Interface
	mapper             *restmapper.DeferredDiscoveryRESTMapper
	lister             listerv1alpha1.ReleaseHistoryLister
	secretlister       listercorev1.SecretLister
	versionedclientset *versioned.Clientset
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func NewMutator(log logr.Logger, clientset *versioned.Clientset, lister listerv1alpha1.ReleaseHistoryLister, secretlister listercorev1.SecretLister, dynamicclient dynamic.Interface, mapper *restmapper.DeferredDiscoveryRESTMapper) *M {
	m := M{
		Webhook:            NewWebhook(log),
		versionedclientset: clientset,
		lister:             lister,
		secretlister:       secretlister,
		dyn:                dynamicclient,
		mapper:             mapper,
	}
	m.WebhookProcessor = &m
	return &m
}

// Process implements WebhookProcessor
func (m *M) Process(ctx context.Context, req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	ar := &admissionv1.AdmissionResponse{
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

	ownerunstruct := m.findOwner(ctx, log, unstruct)
	if ownerunstruct == nil {
		log.Info("Incoming object does not originate from a tracked helm release")
		return ar
	}

	// This one is interesting.
	annotations := ownerunstruct.GetAnnotations()
	helmReleaseName := annotations[constants.HelmLabelReleaseName]

	rel, err := m.versionedclientset.TugboatV1alpha1().ReleaseHistories(ownerunstruct.GetNamespace()).Get(ctx, helmReleaseName, metav1.GetOptions{})
	if err != nil {
		log.Info("failed to find release history", "name", helmReleaseName, "namespace", ownerunstruct.GetNamespace())
		return ar
	}

	deployingRevision := m.findDeployingRevision(ownerunstruct.GetNamespace(), helmReleaseName, rel.Status.Revisions)
	if deployingRevision == v1alpha1.Revision(0) {
		log.Info("failed to find an actively deploying revision")
	}

	if i := indexOfRevision(rel.Status.Revisions, deployingRevision); i != -1 {
		copyrel := rel.DeepCopy()
		// "GROUP/VERSION, Kind=KIND"
		// ex: "/v1, Kind=Pod", "apps/v1, Kind=StatefulSet"
		copyrel.Status.Revisions[i].GVKs[unstruct.GroupVersionKind().String()] = "true"
		log.Info("adding gvk", "gvk", unstruct.GroupVersionKind().String())
		_, err = m.versionedclientset.TugboatV1alpha1().ReleaseHistories(ownerunstruct.GetNamespace()).UpdateStatus(ctx, copyrel, metav1.UpdateOptions{})
		if err != nil {
			log.Info("error while updating release history", "name", helmReleaseName, "namespace", ownerunstruct.GetNamespace(), "err", err.Error())
		}
	}

	pos := []patchOperation{
		{
			Op:    "add",
			Path:  "/metadata/labels/tugboat.engineering~1releasehistory",
			Value: helmReleaseName,
		},
		{
			Op:    "add",
			Path:  "/metadata/labels/tugboat.engineering~1revision",
			Value: strconv.Itoa(int(deployingRevision)),
		},
	}
	buf, err := json.Marshal(pos)
	if err != nil {
		log.Info("Failed to marshal patch", "err", err)
		return ar
	}

	log.Info("patching object")

	pt := admissionv1.PatchTypeJSONPatch
	ar.Patch = buf
	ar.PatchType = &pt
	return ar
}

// findDeployingRevision determines the revision that an object was created by.
func (m *M) findDeployingRevision(namespace string, name string, revs []v1alpha1.ReleaseHistoryRevision) v1alpha1.Revision {
	r0, err0 := labels.NewRequirement("name", selection.Equals, []string{name})
	r1, err1 := labels.NewRequirement("owner", selection.Equals, []string{"helm"})
	r2, err2 := labels.NewRequirement("status", selection.In, []string{string(release.StatusPendingInstall), string(release.StatusPendingUpgrade), string(release.StatusPendingRollback)})
	if err0 != nil || err1 != nil || err2 != nil {
		m.Log.Info("failed to create requirement", "err0", err0.Error(), "err1", err1.Error(), "err2", err2.Error())
		return v1alpha1.Revision(0)
	}

	dumpedsecrets, _ := m.secretlister.Secrets(namespace).List(labels.NewSelector().Add(*r0, *r1))
	for _, x := range dumpedsecrets {
		m.Log.Info("Secret", "name", x.GetName(), "status", x.GetLabels()["status"])
	}

	secrets, err := m.secretlister.Secrets(namespace).List(labels.NewSelector().Add(*r0, *r1, *r2))
	if err != nil {
		return v1alpha1.Revision(0)
	}

	rev := 0
	for _, x := range secrets {
		rev0, _ := strconv.Atoi(x.GetLabels()["version"])
		if rev < rev0 {
			rev = rev0
		}
	}

	return v1alpha1.Revision(rev)
}

func indexOfRevision(revs []v1alpha1.ReleaseHistoryRevision, rev v1alpha1.Revision) int {
	for i, x := range revs {
		if x.Revision == rev {
			return i
		}
	}
	return -1
}

func (m *M) findOwner(ctx context.Context, log logr.Logger, unstruct *unstructured.Unstructured) *unstructured.Unstructured {
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
		unstruct0, err := m.dyn.Resource(mapping.Resource).Namespace(unstruct.GetNamespace()).Get(ctx, ref.Name, metav1.GetOptions{})
		if err != nil {
			log.Info("failed to get unstructured object", "find-name", ref.Name, "find-gvk", gvk.String(), "find-mapping", mapping.Resource.String(), "err", err.Error())
			continue
		}
		if u := m.findOwner(ctx, log, unstruct0); u != nil {
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

	helmReleaseName := annotations[constants.HelmLabelReleaseName]
	helmReleaseNamespace := annotations[constants.HelmLabelReleaseNamespace]

	r0, err0 := labels.NewRequirement(constants.LabelReleaseName, selection.Equals, []string{helmReleaseName})
	r1, err1 := labels.NewRequirement(constants.LabelReleaseNamespace, selection.Equals, []string{helmReleaseNamespace})
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

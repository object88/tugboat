package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/internal/constants"
	"github.com/object88/tugboat/pkg/k8s/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/pkg/k8s/client/clientset/versioned"
	listerv1alpha1 "github.com/object88/tugboat/pkg/k8s/client/listers/engineering.tugboat/v1alpha1"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
)

type V2 struct {
	Webhook
	scheme             *runtime.Scheme
	versionedclientset *versioned.Clientset
	lister             listerv1alpha1.ReleaseHistoryLister
}

func NewV2(log logr.Logger, scheme *runtime.Scheme, clientset *versioned.Clientset, lister listerv1alpha1.ReleaseHistoryLister) *V2 {
	v := V2{
		Webhook:            NewWebhook(log),
		scheme:             scheme,
		versionedclientset: clientset,
		lister:             lister,
	}
	v.WebhookProcessor = &v
	return &v
}

func (v *V2) Process(ctx context.Context, req *v1.AdmissionRequest) *v1.AdmissionResponse {
	// Validator should be very careful about what it does not allow through.

	var obj *corev1.Secret
	if err := json.Unmarshal(req.Object.Raw, &obj); err != nil {
		// If there is a problem unmarshaling the object, then we can reject it.
		v.Log.Error(err, "Could not unmarshal raw object", "name", req.Name, "namespace", req.Namespace)
		return &v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: err.Error(),
			},
		}
	} else if obj.TypeMeta.APIVersion != "v1" || obj.TypeMeta.Kind != "Secret" {
		// This is not a secret.  The tugboat chart should ensure that we never get
		// an object to this endpoint that's not a secret.  Log the error, but
		// allow the admission.
		err = fmt.Errorf("Unexpected type")
		v.Log.Error(err, "Unmarshalled object was not a secret", "name", req.Name, "namespace", req.Namespace, "actualapiversion", obj.TypeMeta.APIVersion, "actualkind", obj.TypeMeta.Kind)
		return &v1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: err.Error(),
			},
		}
	} else if obj.Type != constants.HelmSecretType {
		// This is not a helm secret, so we should ignore it.
		v.Log.Info("Secret is not a helm secret; ignoring", "name", req.Name, "namespace", req.Namespace, "actualapiversion", obj.TypeMeta.APIVersion, "actualkind", obj.TypeMeta.Kind)
		return &v1.AdmissionResponse{
			Allowed: true,
			UID:     req.UID,
		}
	}

	lbls := obj.Labels
	chartname := lbls[constants.HelmSecretLabelName]
	chartnamespace := obj.Namespace
	chartrevision, err := strconv.Atoi(lbls[constants.HelmSecretLabelRevision])
	if err != nil {
		// TODO: figure it out.
		v.Log.Info("version label is either missing or not an integer", "raw", lbls["version"])
		return &v1.AdmissionResponse{
			Allowed: true,
			UID:     req.UID,
		}
	}

	// Check to see if there is a release history.  If there isn't one, then we
	// want to create one and wait for it to be available.
	annotations := obj.Annotations
	helmReleaseName := annotations["meta.helm.sh/release-name"]
	helmReleaseNamespace := annotations["meta.helm.sh/release-namespace"]
	log := v.Log.WithValues("release-name", helmReleaseName, "release-namespace", helmReleaseNamespace)

	log.Info("found annotations")

	r0, err0 := labels.NewRequirement(constants.LabelReleaseName, selection.Equals, []string{helmReleaseName})
	r1, err1 := labels.NewRequirement(constants.LabelReleaseNamespace, selection.Equals, []string{helmReleaseNamespace})
	if err0 != nil || err1 != nil {
		// This is an internal error, but we do not want to interfere with the rest
		// of the system. Return a success.
		log.Error(fmt.Errorf("Internal error"), "failed to create requirement", "err0", err0.Error(), "err1", err1.Error())
		return &v1.AdmissionResponse{
			Allowed: true,
			UID:     req.UID,
		}
	}
	_, err = v.lister.List(labels.NewSelector().Add(*r0, *r1))
	if err != nil {
		// As above, this is (probably) an internal error, but we do not want to
		// interfere with the rest of the system. Return a success.
		log.Error(fmt.Errorf("Internal error: %w", err), "failed to get ")
		return &v1.AdmissionResponse{
			Allowed: true,
			UID:     req.UID,
		}
	}

	namespacedHistories := v.versionedclientset.TugboatV1alpha1().ReleaseHistories(obj.Namespace)

	rh, err := namespacedHistories.Get(context.Background(), chartname, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		// There was an error, and it was not that the release history could not be
		// found.
		v.Log.Error(err, "failed to check for existing v1alpha1.ReleaseHistory", "name", chartname, "namespace", obj.Namespace)
	} else if err == nil {
		// There was no error, indicating that the history does already exist.
		v.Log.Info("already have release history; ignoring", "name", chartname, "namespace", obj.Namespace)

		found := false
		for _, rev := range rh.Status.Revisions {
			if int(rev.Revision) == chartrevision {
				found = true
				break
			}
		}
		// TODO: consider checking the "status" of the secret so that we don't
		// create revision records for past objects.
		// if !found && obj.Labels["status"] == "deploying" {
		if !found {
			// This is a novel revision; add it to the pile
			rev := v1alpha1.ReleaseHistoryRevision{
				DeployedAt: obj.CreationTimestamp,
				GVKs:       map[string]string{},
				Revision:   v1alpha1.Revision(chartrevision),
			}

			newrh := rh.DeepCopy()
			newrh.Status.Revisions = append(newrh.Status.Revisions, rev)

			_, err = namespacedHistories.UpdateStatus(ctx, newrh, metav1.UpdateOptions{})
			if err != nil {
				v.Log.Info("failed to upate existing release history status with new revision", "err", err.Error())
			}
		}

	} else {
		v.Log.Info("release history does not already exist; creating", "name", chartname, "namespace", obj.Namespace)
		rh := &v1alpha1.ReleaseHistory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      chartname,
				Namespace: chartnamespace,
				Labels: map[string]string{
					constants.LabelReleaseName:      chartname,
					constants.LabelReleaseNamespace: chartnamespace,
					constants.LabelState:            constants.LabelStateActive,
				},
			},
			Spec: v1alpha1.ReleaseHistorySpec{
				ReleaseName: chartname,
			},
		}
		newrh, err := namespacedHistories.Create(context.Background(), rh, metav1.CreateOptions{})
		if err != nil {
			v.Log.Error(err, "failed to create v1alpha1.ReleaseHistory")
		}

		newrh.Status = v1alpha1.ReleaseHistoryStatus{
			DeployedAt: obj.CreationTimestamp,
			Revisions: []v1alpha1.ReleaseHistoryRevision{
				{
					DeployedAt: obj.CreationTimestamp,
					GVKs:       map[string]string{},
					Revision:   v1alpha1.Revision(uint(chartrevision)),
				},
			},
		}
		_, err = namespacedHistories.UpdateStatus(context.Background(), newrh, metav1.UpdateOptions{})
		if err != nil {
			v.Log.Error(err, "failed to update v1alpha1.ReleaseHistory with status after create")
		}

		v.Log.Info("added", "name", chartname, "namespace", chartnamespace, "uid", obj.UID)
	}

	// Regardless, we want this to succeed.
	return &v1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
	}
}

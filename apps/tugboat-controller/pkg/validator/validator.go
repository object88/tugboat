package validator

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos/cache"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type V struct {
	log      logr.Logger
	rc       *repos.Cache
	settings *cli.EnvSettings
}

func New(log logr.Logger, rc *repos.Cache, settings *cli.EnvSettings) *V {
	return &V{
		log:      log,
		rc:       rc,
		settings: settings,
	}
}

func (v *V) ProcessAdmission(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	}

	admissionCodecs := serializer.NewCodecFactory(runtime.NewScheme())

	var reviewResponse *v1.AdmissionResponse
	ar := v1.AdmissionReview{}
	if _, _, err := admissionCodecs.UniversalDeserializer().Decode(body, nil, &ar); err != nil {
		reviewResponse = &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		reviewResponse = v.mutate(&ar)
	}

	response := v1.AdmissionReview{
		TypeMeta: ar.TypeMeta,
		Response: reviewResponse,
	}

	// reset the Object and OldObject, they are not needed in a response.
	ar.Request.Object = runtime.RawExtension{}
	ar.Request.OldObject = runtime.RawExtension{}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		v.log.Error(err, "failed to write response")
	}
}

// main mutation process
func (v *V) mutate(ar *v1.AdmissionReview) *v1.AdmissionResponse {
	req := ar.Request
	var launch *v1alpha1.Launch
	if err := json.Unmarshal(req.Object.Raw, &launch); err != nil {
		v.log.Error(err, "Could not unmarshal raw object")
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	allowed := true
	resp := &v1.AdmissionResponse{
		UID: ar.Request.UID,
	}
	switch req.Operation {
	case v1.Create:
		allowed = v.creationAllowed(launch, resp)
	case v1.Update:
		// determine whether to perform mutation
		allowed = mutationAllowed(launch)
	default:
		// Let it pass
	}

	if !allowed {
		v.log.Info("Validation failed due to policy check", "namespace", launch.Namespace, "name", launch.Name, "operation", req.Operation)
		return resp
	}

	return resp
}

func (v *V) creationAllowed(launch *v1alpha1.Launch, resp *v1.AdmissionResponse) bool {
	resp.Allowed = false

	if err := chartutil.ValidateReleaseName(launch.Spec.Chart); err != nil {
		v.log.Error(err, "Launch chart name is invalid", "repository", launch.Spec.Repository, "chart", launch.Spec.Chart, "error", err)
		return false
	}

	if _, err := v.rc.GetChartRepository(launch.Spec.Repository); err != nil {
		if err == cache.ErrMissingRepository {
			return false
		}
		v.log.Error(err, err.Error(), "repository", launch.Spec.Repository)
		return false
	}

	_, err := v.rc.GetChartVersion(launch.Spec.Repository, launch.Spec.Chart, launch.Spec.Version)
	if err != nil {
		v.log.Error(err, err.Error(), "repository", launch.Spec.Repository, "chart", launch.Spec.Chart, "version", launch.Spec.Version.String())
		return false
	}

	resp.Allowed = true
	return true
}

func mutationAllowed(launch *v1alpha1.Launch) bool {
	return true
}

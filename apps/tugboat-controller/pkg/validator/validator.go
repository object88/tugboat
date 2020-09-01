package validator

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type V struct {
	log logr.Logger
}

func New(log logr.Logger) *V {
	return &V{
		log: log,
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

	admissionScheme := runtime.NewScheme()
	admissionCodecs := serializer.NewCodecFactory(admissionScheme)

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

	if resp, err := json.Marshal(response); err != nil {
		v.log.Error(err, "")
	} else if _, err := w.Write(resp); err != nil {
		v.log.Error(err, "")
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

	// determine whether to perform mutation
	if !mutationAllowed(launch) {
		v.log.Info("Skipping mutation due to policy check", "namespace", launch.Namespace, "name", launch.Name)
		return &v1.AdmissionResponse{
			Allowed: false,
		}
	}

	return &v1.AdmissionResponse{
		UID:     ar.Request.UID,
		Allowed: true,
	}
}

func mutationAllowed(launch *v1alpha1.Launch) bool {
	return true
}

package validator

import (
	"encoding/json"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

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
	// TODO: This should be temporary; unclear if we need to run Secrets through
	// the validator.
	var obj *corev1.Secret
	if err := json.Unmarshal(req.Object.Raw, &obj); err != nil {
		v.Log.Error(err, "Could not unmarshal raw object", "name", req.Name, "namespace", req.Namespace)
		return &v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: err.Error(),
			},
		}
	}

	return &v1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
	}
}

// func (v *V) ProcessAdmission(w http.ResponseWriter, r *http.Request) {
// 	var body []byte
// 	if r.Body != nil {
// 		if data, err := ioutil.ReadAll(r.Body); err == nil {
// 			body = data
// 		}
// 	}
// 	if len(body) == 0 {
// 		w.WriteHeader(http.StatusBadRequest)
// 	}

// 	admissionCodecs := serializer.NewCodecFactory(runtime.NewScheme())

// 	var reviewResponse *v1.AdmissionResponse
// 	ar := v1.AdmissionReview{}
// 	if _, _, err := admissionCodecs.UniversalDeserializer().Decode(body, nil, &ar); err != nil {
// 		reviewResponse = &v1.AdmissionResponse{
// 			Result: &metav1.Status{
// 				Message: err.Error(),
// 			},
// 		}
// 	} else {
// 		reviewResponse = v.mutate(&ar)
// 	}

// 	response := v1.AdmissionReview{
// 		TypeMeta: ar.TypeMeta,
// 		Response: reviewResponse,
// 	}

// 	// reset the Object and OldObject, they are not needed in a response.
// 	ar.Request.Object = runtime.RawExtension{}
// 	ar.Request.OldObject = runtime.RawExtension{}

// 	if err := json.NewEncoder(w).Encode(response); err != nil {
// 		v.log.Error(err, "failed to write response")
// 	}
// }

// // main mutation process
// func (v *V) mutate(ar *v1.AdmissionReview) *v1.AdmissionResponse {
// 	req := ar.Request

// 	allowed := true
// 	resp := &v1.AdmissionResponse{
// 		UID: ar.Request.UID,
// 	}
// 	switch req.Operation {
// 	case v1.Create:
// 		allowed = v.creationAllowed(resp)
// 	case v1.Update:
// 		// determine whether to perform mutation
// 		allowed = mutationAllowed()
// 	default:
// 		// Let it pass
// 	}

// 	if !allowed {
// 		v.log.Info("Validation failed due to policy check", "operation", req.Operation)
// 		return resp
// 	}

// 	return resp
// }

// func (v *V) creationAllowed(resp *v1.AdmissionResponse) bool {
// 	return true
// }

// func mutationAllowed() bool {
// 	return true
// }

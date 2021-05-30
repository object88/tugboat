package validator

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/pkg/errs"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// WebhookProcessor is the interface that the HTTP web hooks call into via the
// ProcessAdmission func on Webhook struct
type WebhookProcessor interface {
	Process(ctx context.Context, ar *v1.AdmissionRequest) *v1.AdmissionResponse
}

// Webhook manages the decoding and encoding of the Kubernetes admission
// structs during the processing of an HTTP request, and hands them over to
// a WebbookProcessor for handling
type Webhook struct {
	WebhookProcessor
	Log              logr.Logger
	admissionDecoder runtime.Decoder
}

// NewWebhook returns an instance of a Webhook with an unassigned
// WebhookProcessor interface.  This must be assigned before the
// ProcessAdmission func is invoked.
func NewWebhook(log logr.Logger) Webhook {
	ac := serializer.NewCodecFactory(runtime.NewScheme())
	d := ac.UniversalDeserializer()
	return Webhook{
		Log:              log,
		admissionDecoder: d,
	}
}

func (wh *Webhook) ProcessAdmission(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	}

	var reviewResponse *v1.AdmissionResponse
	ar := v1.AdmissionReview{}
	if _, _, err := wh.admissionDecoder.Decode(body, nil, &ar); err != nil {
		wh.Log.Error(err, "could not decode AdmissionReview", "error", err)
		reviewResponse = &v1.AdmissionResponse{
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: err.Error(),
			},
		}
	} else if ar.Request == nil {
		wh.Log.Error(errs.ErrNilPointerReceived, "AdmissionReview has nil reference for AdmissionRequest")
		reviewResponse = &v1.AdmissionResponse{
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Message: errs.ErrNilPointerReceived.Error(),
			},
		}
	} else {
		reviewResponse = wh.Process(r.Context(), ar.Request)
		reviewResponse.UID = ar.Request.UID
	}

	response := v1.AdmissionReview{
		TypeMeta: ar.TypeMeta,
		Response: reviewResponse,
	}

	// Reset the Object and OldObject, they are not needed in a response.
	ar.Request.Object = runtime.RawExtension{}
	ar.Request.OldObject = runtime.RawExtension{}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		wh.Log.Error(err, "failed to write response")
	}
}

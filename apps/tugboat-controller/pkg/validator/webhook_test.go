package validator

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/object88/tugboat/mocks"
	"github.com/object88/tugboat/pkg/logging/testlogger"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testhook struct {
	Webhook
}

func Test_Validator_Webbook_New(t *testing.T) {
	l := testlogger.TestLogger{T: t}

	wh := NewWebhook(l)
	if wh.admissionDecoder == nil {
		t.Errorf("no admission decoder")
	}
}

func Test_Validator_Webhook(t *testing.T) {
	l := testlogger.TestLogger{T: t}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockWebhookProcessor(ctrl)

	th := testhook{
		Webhook: NewWebhook(l),
	}
	th.WebhookProcessor = m

	ar := v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "v1",
		},
		Request: &v1.AdmissionRequest{
			UID: "123",
			Kind: metav1.GroupVersionKind{
				Version: "v1",
				Group:   "",
				Kind:    "AdmissionRequest",
			},
		},
	}

	rr := v1.AdmissionResponse{
		Result: &metav1.Status{
			Status: metav1.StatusSuccess,
		},
		UID: "123",
	}
	m.EXPECT().Process(gomock.Any(), gomock.AssignableToTypeOf(&v1.AdmissionRequest{})).Return(&rr)

	w, req := makeAdmissionRequest(t, &ar)
	th.ProcessAdmission(w, &req)

	rev := fromResponseWriter(t, w)
	if rev.APIVersion != "v1" {
		t.Errorf("Got wrong API Version for review: '%s'", rev.APIVersion)
	}
	if rev.Kind != "AdmissionReview" {
		t.Errorf("Got wrong kind for review: '%s'", rev.Kind)
	}
	if rev.Response.Result.Status != metav1.StatusSuccess {
		t.Errorf("Got unexpected review response result status: '%s'", rev.Response.Result.Status)
	}
	if rev.Response.UID != "123" {
		t.Errorf("Got unexpected review UID: '%s'", rev.Response.UID)
	}
}

func makeAdmissionRequest(t *testing.T, ar *v1.AdmissionReview) (http.ResponseWriter, http.Request) {
	buf, err := json.Marshal(&ar)
	if err != nil {
		t.Fatalf("failed to marshal: %s", err.Error())
	}

	w := httptest.NewRecorder()
	req := http.Request{
		Method: http.MethodPost,
		Body:   ioutil.NopCloser(bytes.NewReader(buf)),
	}

	return w, req
}

func fromResponseWriter(t *testing.T, w http.ResponseWriter) *v1.AdmissionReview {
	switch x := w.(type) {
	case *httptest.ResponseRecorder:
		if x.Code != http.StatusOK {
			t.Errorf("Did not get OK status: %d", x.Code)
		}
		rev := v1.AdmissionReview{}
		if err := json.NewDecoder(x.Body).Decode(&rev); err != nil {
			t.Fatalf("failed to decode the admission review from the response: %s", err.Error())
		}
		return &rev
	default:
		t.Fatalf("Unexpected writer type")
		return nil
	}
}

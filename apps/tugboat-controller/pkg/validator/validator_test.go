package validator

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/object88/tugboat/pkg/k8s/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/pkg/logging/testlogger"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_Validator_New(t *testing.T) {
	l := testlogger.TestLogger{T: t}

	v := New(l, runtime.NewScheme())
	if v == nil {
		t.Fatal("New returned nil validator")
	}
}

func Test_Validator_BadRequest(t *testing.T) {
	l := testlogger.TestLogger{T: t}

	rh := v1alpha1.ReleaseHistory{
		Spec: v1alpha1.ReleaseHistorySpec{
			ReleaseName: "foo",
		},
	}
	var rlbuf bytes.Buffer
	json.NewEncoder(&rlbuf).Encode(rh)
	relhistbytes := rlbuf.Bytes()

	v := New(l, runtime.NewScheme())

	tcs := []struct {
		name string
		req  v1.AdmissionRequest
	}{
		{
			name: "empty-object",
			req: v1.AdmissionRequest{
				Object: runtime.RawExtension{},
			},
		},
		{
			name: "garbage-object",
			req: v1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: []byte("garbage"),
				},
			},
		},
		{
			name: "partial-object",
			req: v1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: relhistbytes[0 : len(relhistbytes)/2],
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resp := v.Process(&tc.req)
			if resp == nil {
				t.Errorf("Did not get expected response")
				return
			}

			if resp.Allowed {
				t.Errorf("Bad request was allowed")
			}
		})
	}
}

func Test_Validator_IncorrectType(t *testing.T) {
	l := testlogger.TestLogger{T: t}

	pod := corev1.Pod{}
	var podbuf bytes.Buffer
	json.NewEncoder(&podbuf).Encode(pod)
	podbytes := podbuf.Bytes()

	v := New(l, runtime.NewScheme())

	tcs := []struct {
		name string
		req  v1.AdmissionRequest
	}{
		{
			name: "wrong-object-type",
			req: v1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: podbytes,
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resp := v.Process(&tc.req)
			if resp == nil {
				t.Errorf("Did not get expected response")
				return
			}

			if !resp.Allowed {
				t.Errorf("Bad request was allowed")
			}

			if resp.Result.Status != metav1.StatusFailure {
				t.Errorf("Result did not have failed status")
			}
		})
	}
}

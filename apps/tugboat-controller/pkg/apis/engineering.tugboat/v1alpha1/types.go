package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReleaseHistory describes the history of the kubernetes resources described
// in a particular release of a chart.
type ReleaseHistory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ReleaseHistorySpec `json:"spec"`

	Status ReleaseHistoryStatus `json:"status"`
}

// ReleaseHistorySpec is the spec for a ReleaseHistory
type ReleaseHistorySpec struct {
	ReleaseName string `json:"releasename"`
}

// ReleaseHistoryStatus is the status for a ReleaseHistory resource
type ReleaseHistoryStatus struct {
	DeployedAt metav1.Time              `json:"deployedat"`
	Revisions  []ReleaseHistoryRevision `json:"revisions"`
}

type Revision uint

type ReleaseHistoryRevision struct {
	Revision   Revision    `json:"revision"`
	DeployedAt metav1.Time `json:"deployedat"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReleaseHistoryList is a list of ReleaseHistory resources
type ReleaseHistoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ReleaseHistory `json:"items"`
}

package v1alpha1

import (
	"github.com/Masterminds/semver/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Launch describes a launch.
type Launch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec LaunchSpec `json:"spec"`
}

// LaunchSpec is the spec for a Launch resource
type LaunchSpec struct {
	Chart   string          `json:"chart"`
	Version *semver.Version `json:"version"`
	Values  string          `json:"values,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LaunchList is a list of Launch resources
type LaunchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Launch `json:"items"`
}

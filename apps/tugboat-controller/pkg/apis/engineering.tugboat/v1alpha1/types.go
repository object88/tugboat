package v1alpha1

import (
	"github.com/Masterminds/semver/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ChartReference struct {
	Repository string          `json:"repository"`
	Chart      string          `json:"chart"`
	Version    *semver.Version `json:"version"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Launch describes a launch.
type Launch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec LaunchSpec `json:"spec"`

	Status LaunchStatus `json:"status"`
}

// LaunchSpec is the spec for a Launch resource
type LaunchSpec struct {
	ChartReference `json:",inline"`
	Values         string `json:"values,omitempty"`
}

// LaunchStatus is the status for a Launch resource
type LaunchStatus struct {
	State string `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LaunchList is a list of Launch resources
type LaunchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Launch `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Repository describes a helm chart repository.
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RepositorySpec `json:"spec"`
}

// RepositorySpec is the spec for a Repository resource
type RepositorySpec struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RepositoryList is a list of Repository resources
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Repository `json:"items"`
}

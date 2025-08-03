package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LeviathanBuildSpec defines the desired state of LeviathanBuild
type LeviathanBuildSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of LeviathanBuild. Edit leviathanbuild_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`
}

// LeviathanBuildStatus defines the observed state of LeviathanBuild.
type LeviathanBuildStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// LeviathanBuild is the Schema for the leviathanbuilds API
type LeviathanBuild struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of LeviathanBuild
	// +required
	Spec LeviathanBuildSpec `json:"spec"`

	// status defines the observed state of LeviathanBuild
	// +optional
	Status LeviathanBuildStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// LeviathanBuildList contains a list of LeviathanBuild
type LeviathanBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LeviathanBuild `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LeviathanBuild{}, &LeviathanBuildList{})
}

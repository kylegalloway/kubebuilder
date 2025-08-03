package v1

import (
	// batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make generate" to regenerate code after modifying this file
// The following markers will use OpenAPI v3 schema to validate the value
// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

// LeviathanBuildSpec defines the desired state of LeviathanBuild
type LeviathanBuildSpec struct {

	// packageName is the name of the package being built/published
	// +required
	PackageName *string `json:"packageName"`

	// buildType is the type of build
	// - "Build" (default): runs a build of the given package;
	// - "BuildPublish": runs a build and publish of the given package;
	// - "Publish": runs a publish of the given package
	// +optional
	// +kubebuilder:default:=Build
	BuildType BuildType `json:"buildType,omitempty"`

	// sourceType indicates the type of source that should be pulled from
	// - "Local" (default): Use a local path for the source
	// - "Git": Pull the source from git
	// - "S3": Pull the source from an s3 bucket
	// +optional
	// +kubebuilder:default:=Build
	SourceType SourceType `json:"sourceType,omitempty"`

	// sourcePath indicates the path that the source should be pulled from
	// +optional
	SourcePath *string `json:"sourcePath,omitempty"`

	// sourceURL indicates the URL that the source should be pulled from
	// +optional
	SourceURL *string `json:"sourceURL,omitempty"`

	// // job defines the job that will be created when executing the given build.
	// // +required
	// Job batchv1.JobSpec `json:"job"`

	// successfulJobsHistoryLimit defines the number of successful finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	// +kubebuilder:validation:Minimum=0
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`

	// failedJobsHistoryLimit defines the number of failed finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	// +kubebuilder:validation:Minimum=0
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`
}

// BuildType describes how the job will be handled.
// Only one of the following build types may be specified.
// If none of the following types is specified, the default is build.
// +kubebuilder:validation:Enum=Allow;Forbid;Replace
type BuildType string

const (
	// Build runs a build of the given package
	Build BuildType = "Build"

	// BuildPublish runs a build of the given package, then publishes it
	BuildPublish BuildType = "BuildPublish"

	// Publish runs a publish of the given package
	Publish BuildType = "Publish"
)

// SourceType indicates the type of source that should be pulled from
// Only one of the following build types may be specified.
// If none of the following types is specified, the default is local.
// +kubebuilder:validation:Enum=Local;Git;S3
type SourceType string

const (
	// Use a local path for the source
	LocalSource SourceType = "Local"

	// Pull the source from git
	GitSource SourceType = "Git"

	// Pull the source from an s3 bucket
	S3Source SourceType = "S3"
)

// LeviathanBuildStatus defines the observed state of LeviathanBuild.
type LeviathanBuildStatus struct {

	// active defines a list of pointers to currently running jobs.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// lastJobTime defines when was the last time the job was successfully scheduled.
	// +optional
	LastJobTime *metav1.Time `json:"lastJobTime,omitempty"`

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the CronJob resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
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

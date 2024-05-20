package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KustomizeSpec defines the desired state of Kustomize
type KustomizeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// URL is the Kustomize installation url.
	// +kubebuilder:example="https://example.com/path/to/kustomize/manifests"
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Required
	URL string `json:"url"`
}

// KustomizeStatus defines the observed state of Kustomize
type KustomizeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Processed represents the succession of processing the KustomizeSpec.
	// +kubebuilder:validation:Type=boolean
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Processed bool `json:"processed"`

	// Error is any error(s) that have occurred during the processing of the KustomizeSpec.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Optional
	Error *string `json:"error,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Processed",type=boolean,JSONPath=`.status.processed`,description="Kustomize Processing Succession"
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`,description="The Kustomize Installation URL"

// Kustomize is the Schema for the kustomizes API
type Kustomize struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KustomizeSpec   `json:"spec,omitempty"`
	Status KustomizeStatus `json:"status,omitempty"`
}

func (k *Kustomize) SetURL(v string) {
	k.Spec.URL = v
}

// +kubebuilder:object:root=true

// KustomizeList contains a list of Kustomize
type KustomizeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kustomize `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kustomize{}, &KustomizeList{})
}

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HelmSpec defines the desired state of Helm
type HelmSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Helm. Edit helm_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// HelmStatus defines the observed state of Helm
type HelmStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Helm is the Schema for the helms API
type Helm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmSpec   `json:"spec,omitempty"`
	Status HelmStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HelmList contains a list of Helm
type HelmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Helm `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Helm{}, &HelmList{})
}

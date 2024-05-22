package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Type describes the type of Manifest the custom-resource is.
// +kubebuilder:validation:Enum=Flux;Standard
type Type string

const (
	// Standard is the default Type of Manifest.
	Standard Type = "Standard"
)

// ManifestSpec defines the desired state of Manifest
type ManifestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// URL is the Manifest installation url.
	// +kubebuilder:example="https://example.com/path/to/manifests/install.yaml"
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Required
	URL string `json:"url"`

	// Specifies how to evaluate the given Manifest.
	// Valid values are:
	// - "Standard" (default)
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Required
	// +kubebuilder:default="Standard"
	Type Type `json:"type"`
}

// ManifestStatus defines the observed state of Manifest
type ManifestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Processed represents the succession of processing the ManifestSpec.
	// +kubebuilder:validation:Type=boolean
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Processed bool `json:"processed"`

	// Total represents the amount of manifests evaluated for the given custom-resource.
	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=-1
	Total int `json:"total-manifests"`

	// Error is any error(s) that have occurred during the processing of the ManifestSpec.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=null
	Error *string `json:"error"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Processed",type=boolean,JSONPath=`.status.processed`,description="Manifest Processing Succession"
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`,description="The Manifest Installation URL"

// Manifest is the Schema for the Manifests API
type Manifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManifestSpec   `json:"spec,omitempty"`
	Status ManifestStatus `json:"status,omitempty"`
}

func (k *Manifest) SetURL(v string) {
	k.Spec.URL = v
}

func (k *Manifest) SetType(v Type) {
	k.Spec.Type = v
}

func (k *Manifest) SetError(e error) {
	if e == nil {
		k.Status.Error = nil
		return
	}

	v := e.Error()

	k.Status.Error = &v
}

func (k *Manifest) SetTotal(v int) {
	k.Status.Total = v
}

func (k *Manifest) SetProcessed(v bool) {
	k.Status.Processed = v
}

// +kubebuilder:object:root=true

// ManifestList contains a list of Manifest
type ManifestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Manifest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Manifest{}, &ManifestList{})
}

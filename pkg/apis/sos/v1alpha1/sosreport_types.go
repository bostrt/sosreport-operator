package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SosreportSpec defines the desired state of Sosreport
type SosreportSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
	CaseNumber int32 `json:"caseNumber"`
}

type Location struct {
	Node string `json:"node"`
	Path string `json:"path"`
	Failed bool `json:"failed"`
}

// SosreportStatus defines the observed state of Sosreport
type SosreportStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Locations []Location `json:"locations"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Sosreport is the Schema for the sosreports API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sosreports,scope=Namespaced
type Sosreport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SosreportSpec   `json:"spec,omitempty"`
	Status SosreportStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SosreportList contains a list of Sosreport
type SosreportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sosreport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sosreport{}, &SosreportList{})
}

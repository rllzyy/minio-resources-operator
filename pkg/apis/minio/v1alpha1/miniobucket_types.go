package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MinioBucketSpec defines the desired state of MinioBucket
type MinioBucketSpec struct {
	Server string `json:"server"`
	Name   string `json:"name"`
	Policy string `json:"policy,omitempty"`
}

// MinioBucketStatus defines the observed state of MinioBucket
type MinioBucketStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinioBucket is the Schema for the miniobuckets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=miniobuckets,scope=Namespaced
type MinioBucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinioBucketSpec   `json:"spec,omitempty"`
	Status MinioBucketStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinioBucketList contains a list of MinioBucket
type MinioBucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MinioBucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MinioBucket{}, &MinioBucketList{})
}

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MinioUserSpec defines the desired state of MinioUser
type MinioUserSpec struct {
	Server string `json:"server"`
	Policy string `json:"policy,omitempty"`
}

// MinioUserStatus defines the observed state of MinioUser
type MinioUserStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinioUser is the Schema for the miniousers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=miniousers,scope=Namespaced
type MinioUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinioUserSpec   `json:"spec,omitempty"`
	Status MinioUserStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinioUserList contains a list of MinioUser
type MinioUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MinioUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MinioUser{}, &MinioUserList{})
}

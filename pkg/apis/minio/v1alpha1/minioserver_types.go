package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MinioServerSpec defines the desired state of MinioServer
type MinioServerSpec struct {
	Hostname  string `json:"hostname"`
	Port      int    `json:"port"`
	SSL       bool   `json:"ssl,omitempty"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

// MinioServerStatus defines the observed state of MinioServer
type MinioServerStatus struct {
	Online bool `json:"online"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinioServer is the Schema for the minioservers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=minioservers,scope=Cluster
type MinioServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinioServerSpec   `json:"spec,omitempty"`
	Status MinioServerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinioServerList contains a list of MinioServer
type MinioServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MinioServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MinioServer{}, &MinioServerList{})
}

/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MinioServerSpec defines the desired state of MinioServer
type MinioServerSpec struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	SSL      bool   `json:"ssl,omitempty"`
}

// MinioServerStatus defines the observed state of MinioServer
type MinioServerStatus struct {
	Online bool `json:"online"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MinioServer is the Schema for the minioservers API
type MinioServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinioServerSpec   `json:"spec,omitempty"`
	Status MinioServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MinioServerList contains a list of MinioServer
type MinioServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MinioServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MinioServer{}, &MinioServerList{})
}

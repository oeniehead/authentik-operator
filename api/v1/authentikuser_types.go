/*
Copyright 2024.

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

// AuthentikUserSpec defines the desired state of AuthentikUser
type AuthentikUserSpec struct {
	// The name of the user
	Name string `json:"name,omitempty"`
	// The username of the user
	Username string `json:"username,omitempty"`
	// The email address of the user
	Email string `json:"email,omitempty"`
	// The groups this user belongs to
	Groups []string `json:"groups,omitempty"`
}

// AuthentikUserStatus defines the observed state of AuthentikUser
type AuthentikUserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AuthentikUser is the Schema for the authentikusers API
type AuthentikUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuthentikUserSpec   `json:"spec,omitempty"`
	Status AuthentikUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AuthentikUserList contains a list of AuthentikUser
type AuthentikUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuthentikUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuthentikUser{}, &AuthentikUserList{})
}

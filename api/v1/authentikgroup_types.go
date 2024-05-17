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

// AuthentikGroupSpec defines the desired state of AuthentikGroup
type AuthentikGroupSpec struct {
	// The name of the group
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// If this group is administrative
	// +kubebuilder:validation:Required
	IsAdmin bool `json:"isAdmin"`
	// The parent of this group
	// +optional
	Parent *string `json:"parent,omitempty"`
}

// AuthentikGroupStatus defines the observed state of AuthentikGroup
type AuthentikGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AuthentikGroup is the Schema for the authentikgroups API
type AuthentikGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuthentikGroupSpec   `json:"spec,omitempty"`
	Status AuthentikGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AuthentikGroupList contains a list of AuthentikGroup
type AuthentikGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuthentikGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuthentikGroup{}, &AuthentikGroupList{})
}

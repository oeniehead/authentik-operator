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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AuthentikProviderSpec defines the desired state of AuthentikProvider
type AuthentikProviderSpec struct {
	// Name of the provider
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Name string `json:"name,omitempty"`
	// Type of authentication, one of: oauth2
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Type string `json:"type,omitempty" binding:"oneof=oauth2"`
	// Authentication flow for this application
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	AuthenticationFlow string `json:"authenticationFlow,omitempty"`
	// Authorization flow for this application
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	AuthorizationFlow string `json:"authorizationFlow,omitempty"`
	// Type of client, one of: confidential
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ClientType string `json:"clientType,omitempty" binding:"oneof=confidential"`
	// Valid redirect URI
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	RedirectUri string `json:"redirectUri,omitempty"`
	// All requested scopes for the application
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ScopeMappings []string `json:"scopes,omitempty"`
}

// AuthentikProviderStatus defines the observed state of AuthentikProvider
type AuthentikProviderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AuthentikProvider is the Schema for the authentikproviders API
type AuthentikProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuthentikProviderSpec   `json:"spec,omitempty"`
	Status AuthentikProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AuthentikProviderList contains a list of AuthentikProvider
type AuthentikProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuthentikProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuthentikProvider{}, &AuthentikProviderList{})
}

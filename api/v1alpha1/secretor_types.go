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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretorSpec defines the desired state of Secretor
type SecretorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum=constant;generate
	// Type
	Type       string      `json:"type,omitempty"`
	Value      *string     `json:"value,omitempty"`
	Generating *Generating `json:"generating,omitempty"`
	InjectTo   []InjectTo  `json:"injectTo,omitempty"`
}

type InjectTo struct {
	SecretRef SecretRef `json:"secretRef,omitempty"`
}

type SecretRef struct {
	Name      string  `json:"name,omitempty"`
	Namespace *string `json:"namespace,omitempty"`
	Field     string  `json:"field,omitempty"`
}

type Generating struct {
	Length int `json:"length,omitempty"`
}

// SecretorStatus defines the observed state of Secretor
type SecretorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Secretor is the Schema for the secretors API
type Secretor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretorSpec   `json:"spec,omitempty"`
	Status SecretorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretorList contains a list of Secretor
type SecretorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Secretor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Secretor{}, &SecretorList{})
}

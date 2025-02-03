/*
Copyright 2025.

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

// MyChildResourceSpec defines the desired state of MyChildResource.
type MyChildResourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MyChildResource. Edit mychildresource_types.go to remove/update
	Foo string `json:"foo,omitempty"`
	// +kubebuilder:default={}
	FooMap  map[string]string `json:"fooMap,omitempty"`
	FooList []string          `json:"fooList,omitempty"`
	// +kubebuilder:default="ho-ho-ho"
	FooValueDefault string `json:"fooValueDefault,omitempty"`
}

// MyChildResourceStatus defines the observed state of MyChildResource.
type MyChildResourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MyChildResource is the Schema for the mychildresources API.
type MyChildResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyChildResourceSpec   `json:"spec,omitempty"`
	Status MyChildResourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MyChildResourceList contains a list of MyChildResource.
type MyChildResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyChildResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyChildResource{}, &MyChildResourceList{})
}

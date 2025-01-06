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

// ErrorKubeSpec defines the desired state of ErrorKube
type ErrorKubeSpec struct {
	// Replicas is the desired number of backend pod replicas
	Replicas int32 `json:"replicas"`

	// BackendPort is the port to bind with the backend application
	// This maps the user-provided port to the backend container's port 8080
	BackendPort int32 `json:"backendPort"`

	// MongoDBHostPath is the host path for MongoDB data storage
	MongoDBHostPath string `json:"mongoDBHostPath"`

	// BackendServiceType specifies the type of service for the backend (e.g., ClusterIP, NodePort, LoadBalancer)
	BackendServiceType string `json:"backendServiceType"`
}

// ErrorKubeStatus defines the observed state of ErrorKube
type ErrorKubeStatus struct {
	// MongoDBReady indicates whether the MongoDB deployment is ready
	MongoDBReady bool `json:"mongoDBReady"`

	// BackendReady indicates whether the backend deployment is ready
	BackendReady bool `json:"backendReady"`

	// LastUpdated is the last time the status was updated
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ErrorKube is the Schema for the errorkubes API
type ErrorKube struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ErrorKubeSpec   `json:"spec,omitempty"`
	Status ErrorKubeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ErrorKubeList contains a list of ErrorKube
type ErrorKubeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ErrorKube `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ErrorKube{}, &ErrorKubeList{})
}
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

type InstallStatus string

const (
	Success  InstallStatus = "success"
	NotFound InstallStatus = "not_found"
	Failure  InstallStatus = "failure"
)

type Installation struct {
	Resource string        `json:"resource"`
	Status   InstallStatus `json:"status"`
	Output   []string      `json:"logs,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RemoteInstallationSpec defines the desired state of RemoteInstallation
type RemoteInstallationSpec struct {
	Url  string `json:"url"`
	Type string `json:"type,omitempty"`
}

// RemoteInstallationStatus defines the observed state of RemoteInstallation
type RemoteInstallationStatus struct {
	Status []InstallStatus `json:"status,omitempty"`
	Logs   string          `json:"logs,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RemoteInstallation is the Schema for the remoteinstallations API
type RemoteInstallation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemoteInstallationSpec   `json:"spec,omitempty"`
	Status RemoteInstallationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RemoteInstallationList contains a list of RemoteInstallation
type RemoteInstallationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteInstallation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemoteInstallation{}, &RemoteInstallationList{})
}

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

	"github.com/kyma-project/module-manager/pkg/types"
)

// BackendType is the supported Eventing backend type.
// +kubebuilder:validation:Enum=nats;eventmesh
type BackendType string

const (
	BackendTypeNats      = BackendType("nats")
	BackendTypeEventMesh = BackendType("eventmesh")
)

// BackendSpec defines the desired state of the Eventing backend.
type BackendSpec struct {
	Type BackendType `json:"type"`
}

// EventingSpec defines the desired state of Eventing
type EventingSpec struct {
	BackendSpec BackendSpec `json:"backend"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Eventing is the Schema for the eventings API
type Eventing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventingSpec `json:"spec,omitempty"`
	Status types.Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EventingList contains a list of Eventing
type EventingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Eventing `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Eventing{}, &EventingList{})
}

var _ types.CustomObject = &Eventing{}

func (in *Eventing) GetStatus() types.Status {
	return in.Status
}

func (in *Eventing) SetStatus(status types.Status) {
	in.Status = status
}

func (in *Eventing) ComponentName() string {
	return "eventing"
}

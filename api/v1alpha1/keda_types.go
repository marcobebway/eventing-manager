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
	"github.com/kyma-project/module-manager/operator/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackendType is the supported Eventing backend type.
// +kubebuilder:validation:Enum=nats;jetstream
type BackendType string

const (
	BackendTypeNats      = BackendType("nats")
	BackendTypeJetStream = BackendType("jetstream")
)

// BackendSpec defines the desired state of the Eventing backend.
type BackendSpec struct {
	Type BackendType `json:"type"`
}

// KedaSpec defines the desired state of Keda
type KedaSpec struct {
	BackendSpec BackendSpec `json:"backend"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Keda is the Schema for the kedas API
type Keda struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KedaSpec     `json:"spec,omitempty"`
	Status types.Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KedaList contains a list of Keda
type KedaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Keda `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Keda{}, &KedaList{})
}

var _ types.CustomObject = &Keda{}

func (s *Keda) GetStatus() types.Status {
	return s.Status
}

func (s *Keda) SetStatus(status types.Status) {
	s.Status = status
}

func (s *Keda) ComponentName() string {
	return "keda"
}

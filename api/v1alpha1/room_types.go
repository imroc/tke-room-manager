/*
Copyright 2024 imroc.

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

// RoomSpec defines the desired state of Room
type RoomSpec struct {
	PodName         string `json:"podName"`
	Type            string `json:"type"`
	ExternalAddress string `json:"externalAddress"`
}

// RoomStatus defines the observed state of Room
type RoomStatus struct {
	// +optional
	Idle bool `json:"idle,omitempty"`
	// +optional
	Ready bool `json:"ready,omitempty"`
	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Room is the Schema for the rooms API
type Room struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoomSpec   `json:"spec,omitempty"`
	Status RoomStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RoomList contains a list of Room
type RoomList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Room `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Room{}, &RoomList{})
}

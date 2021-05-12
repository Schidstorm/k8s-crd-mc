/*
Copyright 2021.

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

//+kubebuilder:validation:Enum=survival;hardcore;survival
type MinecraftMode string

const (
	MinecraftModeSurvival = MinecraftMode("survival")
	MinecraftModeCreative = MinecraftMode("creative")
	MinecraftModeHardcore = MinecraftMode("hardcore")
)

type MinecraftDifficulty string

const (
	MinecraftDifficultyHard     = MinecraftDifficulty("hard")
	MinecraftDifficultyPeaceful = MinecraftDifficulty("peaceful")
	MinecraftDifficultyEasy     = MinecraftDifficulty("easy")
	MinecraftDifficultyNormal   = MinecraftDifficulty("normal")
)

type MinecraftPorts struct {
	Minecraft *uint16 `json:"minecraft"`
}

// MinecraftSpec defines the desired state of Minecraft
type MinecraftSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Mode *MinecraftMode `json:"mode"`
	// +optional
	Name *string `json:"name"`
	// +optional
	Motd *string `json:"motd"`
	// +optional
	Seed *string `json:"seed"`
	// +optional
	Difficulty *MinecraftDifficulty `json:"difficulty"`
	// +optional
	Ports *MinecraftPorts `json:"ports"`

	Template metav1.ObjectMeta `json:"template"`
}

// MinecraftStatus defines the observed state of Minecraft
type MinecraftStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	//+kubebuilder:validation:Optional
	Mode   *MinecraftMode `json:"mode,omitempty"`
	Status string         `json:"status"`
	//+kubebuilder:validation:Optional
	Pod string `json:"pod,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Minecraft is the Schema for the minecrafts API
type Minecraft struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinecraftSpec   `json:"spec,omitempty"`
	Status MinecraftStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MinecraftList contains a list of Minecraft
type MinecraftList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Minecraft `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Minecraft{}, &MinecraftList{})
}

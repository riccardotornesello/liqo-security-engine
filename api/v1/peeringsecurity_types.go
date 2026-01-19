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

// ResourceGroup represents a group of resources.
// +kubebuilder:validation:Enum=local-cluster;remote-cluster;offloaded;vc-local;vc-remote
type ResourceGroup string

const (
	ResourceGroupLocalCluster  ResourceGroup = "local-cluster"
	ResourceGroupRemoteCluster ResourceGroup = "remote-cluster"
	ResourceGroupOffloaded     ResourceGroup = "offloaded"
	ResourceGroupVcLocal       ResourceGroup = "vc-local"
	ResourceGroupVcRemote      ResourceGroup = "vc-remote"
)

// Action defines a peering connectivity rule action.
// +kubebuilder:validation:Enum=allow;deny
type Action string

const (
	ActionAllow Action = "allow"
	ActionDeny  Action = "deny"
)

// Party defines a connectivity party.
// +kubebuilder:validation:ExactlyOneOf=group
type Party struct {
	// group defines the resource group of the party.
	Group *ResourceGroup `json:"group,omitempty"`
}

// Rule defines a peering connectivity rule.
type Rule struct {
	// action defines whether to allow or deny the traffic.
	Action Action `json:"action,omitempty"`

	// source defines the source resource group of the allowed traffic.
	Source *Party `json:"source,omitempty"`

	// destination defines the destination resource group of the allowed traffic.
	Destination *Party `json:"destination,omitempty"`
}

// PeeringConnectivitySpec defines the desired state of PeeringConnectivity
type PeeringConnectivitySpec struct {
	// rules defines the list of allowed traffic rules
	Rules []Rule `json:"rules,omitempty"`
}

// PeeringConnectivityStatus defines the observed state of PeeringConnectivity.
type PeeringConnectivityStatus struct {
	// conditions represent the current state of the PeeringConnectivity resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// ObservedGeneration is the last observed generation of the PeeringConnectivity resource.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PeeringConnectivity is the Schema for the peeringconnectivities API
type PeeringConnectivity struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PeeringConnectivity
	// +required
	Spec PeeringConnectivitySpec `json:"spec"`

	// status defines the observed state of PeeringConnectivity
	// +optional
	Status PeeringConnectivityStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PeeringConnectivityList contains a list of PeeringConnectivity
type PeeringConnectivityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PeeringConnectivity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PeeringConnectivity{}, &PeeringConnectivityList{})
}

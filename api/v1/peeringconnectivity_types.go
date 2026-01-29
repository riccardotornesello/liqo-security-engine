// Copyright 2019-2026 The Liqo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceGroup represents a group of resources in a Liqo peering environment.
// It categorizes different types of pods and network entities to enable fine-grained
// network security policy management across cluster boundaries.
//
// +kubebuilder:validation:Enum=local-cluster;remote-cluster;offloaded;vc-local;vc-remote;private-subnets
type ResourceGroup string

const (
	// ResourceGroupLocalCluster represents pods running in the local cluster's own pod CIDR.
	ResourceGroupLocalCluster ResourceGroup = "local-cluster"

	// ResourceGroupRemoteCluster represents pods running in the remote cluster's pod CIDR.
	ResourceGroupRemoteCluster ResourceGroup = "remote-cluster"

	// ResourceGroupOffloaded represents pods that have been offloaded from a consumer cluster
	// and are running on a provider cluster.
	// For provider only!
	ResourceGroupOffloaded ResourceGroup = "offloaded"

	// ResourceGroupVcLocal (virtual cluster local) represents local pods in namespaces that are configured
	// for offloading but are still running in the local cluster.
	// For consumer only!
	ResourceGroupVcLocal ResourceGroup = "vc-local"

	// ResourceGroupVcRemote (virtual cluster remote) represents shadow pods on the consumer cluster
	// that represent pods offloaded to a provider cluster .
	// For consumer only!
	ResourceGroupVcRemote ResourceGroup = "vc-remote"

	// PrivateSubnets represents ALL private subnet ranges defined by RFC1918.
	// This group is used to match traffic destined to private IP ranges.
	ResourceGroupPrivateSubnets ResourceGroup = "private-subnets"
)

// Action defines the action to take when a firewall rule matches network traffic.
//
// +kubebuilder:validation:Enum=allow;deny
type Action string

const (
	// ActionAllow permits the matched network traffic to pass through.
	ActionAllow Action = "allow"

	// ActionDeny blocks the matched network traffic.
	ActionDeny Action = "deny"
)

// Party defines a participant in a network connectivity rule.
// A party can represent either the source or destination of network traffic.
//
// +kubebuilder:validation:XValidation:rule="(has(self.group) ? 1 : 0) + (has(self.__namespace__) ? 1 : 0) == 1",message="exactly one of group or namespace must be set"
type Party struct {
	// Group defines the resource group of this party.
	// It identifies which set of pods or resources this party represents.
	Group *ResourceGroup `json:"group,omitempty"`

	// Namespace specifies the Kubernetes namespace associated with this party.
	Namespace *string `json:"namespace,omitempty"`
}

// Rule defines a network connectivity rule for peering scenarios.
// Rules specify how the traffic should flow based on source
// and destination parties and the action to be taken.
type Rule struct {
	// Action defines whether to allow or deny the traffic matching this rule.
	Action Action `json:"action,omitempty"`

	// Source defines the source party for the traffic.
	// If omitted, the rule applies to traffic from any source.
	Source *Party `json:"source,omitempty"`

	// Destination defines the destination party for the traffic.
	// If omitted, the rule applies to traffic to any destination.
	Destination *Party `json:"destination,omitempty"`
}

// PeeringConnectivitySpec defines the desired state of PeeringConnectivity.
// It specifies the security rules that should be applied to network traffic
// in a Liqo peering environment.
type PeeringConnectivitySpec struct {
	// Rules defines the ordered list of network traffic rules.
	// Rules are evaluated in order, and the first matching rule determines
	// whether traffic is allowed or denied.
	Rules []Rule `json:"rules,omitempty"`
}

// PeeringConnectivityStatus defines the observed state of PeeringConnectivity.
// It reflects the current status of the security policy enforcement.
type PeeringConnectivityStatus struct {
	// Conditions represent the current state of the PeeringConnectivity resource.
	// Each condition has a unique type and reflects the status of a specific aspect
	// of the resource, such as whether firewall rules have been successfully synced.
	//
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// ObservedGeneration is the last observed generation of the PeeringConnectivity resource.
	// It is used to track whether the status reflects the latest spec changes.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PeeringConnectivity is the Schema for the peeringconnectivities API.
// It represents a security policy configuration for controlling network connectivity
// in a Liqo multi-cluster peering scenario. Each PeeringConnectivity resource is
// typically created in a tenant namespace (e.g., liqo-tenant-<cluster-id>) and
// defines firewall rules that control traffic between different resource groups.
type PeeringConnectivity struct {
	metav1.TypeMeta `json:",inline"`

	// Metadata is standard Kubernetes object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of PeeringConnectivity.
	// It contains the security rules to be enforced.
	// +required
	Spec PeeringConnectivitySpec `json:"spec"`

	// Status defines the observed state of PeeringConnectivity.
	// It reflects the current status of rule enforcement.
	// +optional
	Status PeeringConnectivityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PeeringConnectivityList contains a list of PeeringConnectivity resources.
// It is used by Kubernetes for list operations.
type PeeringConnectivityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PeeringConnectivity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PeeringConnectivity{}, &PeeringConnectivityList{})
}

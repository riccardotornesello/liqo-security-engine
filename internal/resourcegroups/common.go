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

package resourcegroups

import (
	"context"

	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
)

// groupFuncts defines the functions needed to implement a resource group.
// Each resource group needs to provide:
//   - MakeSets: creates firewall sets (collections of IP addresses) for the group.
//     May be nil if the resource group uses CIDR-based matching instead of sets.
//   - MakeMatchRule: creates firewall match rules for the group.
//     Required for all resource groups.
type groupFuncts struct {
	MakeSets      func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error)
	MakeMatchRule func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error)
}

// ResourceGroupFuncts maps each ResourceGroup to its implementation functions.
// This allows the controller to dynamically create firewall rules based on the
// resource groups specified in the PeeringConnectivity spec.
var ResourceGroupFuncts = map[securityv1.ResourceGroup]groupFuncts{
	securityv1.ResourceGroupLocalCluster:  ResourceGroupLocalCluster,
	securityv1.ResourceGroupRemoteCluster: ResourceGroupRemoteCluster,

	securityv1.ResourceGroupOffloaded: ResourceGroupOffloaded,

	securityv1.ResourceGroupVcLocal:  ResourceGroupVcLocal,
	securityv1.ResourceGroupVcRemote: ResourceGroupVcRemote,

	securityv1.ResourceGroupPrivateSubnets: ResourceGroupPrivateSubnets,
}

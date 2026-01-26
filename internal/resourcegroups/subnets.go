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
)

// private-subnets: Matches traffic destined to all private subnet ranges defined by RFC1918.
// This doesn't need a set because it uses simple CIDR matches.
var ResourceGroupPrivateSubnets = groupFuncts{
	MakeSets: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error) {
		return []networkingv1beta1firewall.Set{{
			Name:    "privatesubnets",
			KeyType: networkingv1beta1firewall.SetDataTypeIPAddr,
			Elements: []networkingv1beta1firewall.SetElement{
				{Key: "10.0.0.0/8"},
				{Key: "172.16.0.0/12"},
				{Key: "192.168.0.0/16"},
			},
		}}, nil
	},
	MakeMatchRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
		return []networkingv1beta1firewall.Match{{
			IP: &networkingv1beta1firewall.MatchIP{
				Value:    "@privatesubnets",
				Position: position,
			},
			Op: networkingv1beta1firewall.MatchOperationEq,
		}}, nil
	},
}

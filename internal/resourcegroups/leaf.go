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
	"github.com/riccardotornesello/liqo-connectivity-engine/internal/controller/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// leaf: matches the external CIDR
var ResourceGroupLeaf = groupFuncts{
	MakeMatchRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
		// Get the remote cluster's external CIDR and create a match rule for it.
		cidr, err := utils.GetRemoteClusterExternalCIDR(ctx, cl, clusterID)
		if err != nil {
			return nil, err
		}

		return []networkingv1beta1firewall.Match{{
			IP: &networkingv1beta1firewall.MatchIP{
				Value:    cidr,
				Position: position,
			},
			Op: networkingv1beta1firewall.MatchOperationEq,
		}}, nil
	},
}

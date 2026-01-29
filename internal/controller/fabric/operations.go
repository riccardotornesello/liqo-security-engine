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

package fabric

import (
	"context"

	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	securityv1 "github.com/riccardotornesello/liqo-connectivity-engine/api/v1"
	"github.com/riccardotornesello/liqo-connectivity-engine/internal/controller/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileFabricFirewallConfiguration ensures that the FirewallConfiguration
// resource for the fabric security rules exists and is up to date.
// It creates or updates the resource as needed based on the provided
// PeeringConnectivity configuration.
func ReconcileFabricFirewallConfiguration(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	cfg *securityv1.PeeringConnectivity,
	clusterID string,
) (controllerutil.OperationResult, error) {
	fabricFwcfg := networkingv1beta1.FirewallConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ForgeFabricResourceName(clusterID),
			Namespace: utils.GetClusterNamespace(clusterID),
		},
	}

	return controllerutil.CreateOrUpdate(ctx, c, &fabricFwcfg, func() error {
		// Set labels that identify this FirewallConfiguration as a fabric-level
		// security configuration targeting all nodes.
		fabricFwcfg.SetLabels(ForgeFabricLabels(clusterID))

		// Generate the FirewallConfiguration spec based on the PeeringConnectivity rules.
		spec, err := ForgeFabricSpec(ctx, c, cfg, clusterID)
		if err != nil {
			return err
		}
		fabricFwcfg.Spec = *spec

		// Set owner reference so the FirewallConfiguration is deleted when the
		// PeeringConnectivity is deleted.
		return controllerutil.SetOwnerReference(cfg, &fabricFwcfg, scheme)
	})
}

// EnsureFabricFirewallConfigurationDeleted deletes the fabric-level FirewallConfiguration
// resource associated with the given cluster ID, if it exists.
func EnsureFabricFirewallConfigurationDeleted(
	ctx context.Context,
	c client.Client,
	clusterID string,
) error {
	fabricFwcfg := networkingv1beta1.FirewallConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ForgeFabricResourceName(clusterID),
			Namespace: utils.GetClusterNamespace(clusterID),
		},
	}

	err := c.Delete(ctx, &fabricFwcfg)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	return nil
}

// Package utils provides utility functions for managing resource groups in the Liqo security engine.
// Resource groups categorize pods and network entities to enable fine-grained security policies.
package utils

import (
	"context"
	"fmt"

	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	// local-cluster: Matches pods in the local cluster's pod CIDR.
	// This doesn't need a set because it uses a simple CIDR match.
	securityv1.ResourceGroupLocalCluster: {
		MakeSets: nil,
		MakeMatchRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
			// Get the local cluster's pod CIDR and create a match rule for it.
			cidr, err := GetCurrentClusterPodCIDR(ctx, cl)
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
	},
	// remote-cluster: Matches pods in the remote cluster's pod CIDR.
	// This doesn't need a set because it uses a simple CIDR match.
	securityv1.ResourceGroupRemoteCluster: {
		MakeSets: nil,
		MakeMatchRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
			// Get the remote cluster's pod CIDR and create a match rule for it.
			cidr, err := GetRemoteClusterPodCIDR(ctx, cl, clusterID)
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
	},
	// offloaded: Matches pods that have been offloaded from the consumer cluster
	// and are running on this provider cluster.
	// Uses a set because pod IPs are dynamically allocated and may not be contiguous.
	securityv1.ResourceGroupOffloaded: {
		MakeSets: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error) {
			// Get all pods offloaded from the consumer cluster.
			pods, err := GetPodsFromConsumer(ctx, cl, clusterID)
			if err != nil {
				return nil, err
			}

			// Create a firewall set containing the IPs of these pods.
			setName := string(securityv1.ResourceGroupOffloaded)
			podIpsSet := ForgePodIpsSet(setName, pods)
			return []networkingv1beta1firewall.Set{podIpsSet}, nil
		},
		MakeMatchRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
			return []networkingv1beta1firewall.Match{{
				IP: &networkingv1beta1firewall.MatchIP{
					Value:    fmt.Sprintf("@%s", string(securityv1.ResourceGroupOffloaded)),
					Position: position,
				},
				Op: networkingv1beta1firewall.MatchOperationEq,
			}}, nil
		},
	},
	// vc-local: Matches local pods in namespaces that are configured for offloading.
	// These are the actual pods running locally that could be offloaded.
	// Uses a set because pod IPs are dynamically allocated.
	securityv1.ResourceGroupVcLocal: {
		MakeSets: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error) {
			// Get all pods in namespaces that are configured for offloading.
			pods, err := GetPodsInOffloadedNamespaces(ctx, cl)
			if err != nil {
				return nil, err
			}

			// Create a firewall set containing the IPs of these pods.
			setName := string(securityv1.ResourceGroupVcLocal)
			podIpsSet := ForgePodIpsSet(setName, pods)
			return []networkingv1beta1firewall.Set{podIpsSet}, nil
		},
		MakeMatchRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
			return []networkingv1beta1firewall.Match{{
				IP: &networkingv1beta1firewall.MatchIP{
					Value:    fmt.Sprintf("@%s", string(securityv1.ResourceGroupVcLocal)),
					Position: position,
				},
				Op: networkingv1beta1firewall.MatchOperationEq,
			}}, nil
		},
	},
	// vc-remote: Matches shadow pods on the consumer cluster that represent
	// pods offloaded to a provider cluster.
	// Uses a set because pod IPs are dynamically allocated.
	securityv1.ResourceGroupVcRemote: {
		MakeSets: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error) {
			// Get all shadow pods that represent offloaded pods on the specified provider cluster.
			pods, err := GetPodsOffloadedToProvider(ctx, cl, clusterID)
			if err != nil {
				return nil, err
			}

			// Create a firewall set containing the IPs of these shadow pods.
			setName := string(securityv1.ResourceGroupVcRemote)
			podIpsSet := ForgePodIpsSet(setName, pods)
			return []networkingv1beta1firewall.Set{podIpsSet}, nil
		},
		MakeMatchRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
			return []networkingv1beta1firewall.Match{{
				IP: &networkingv1beta1firewall.MatchIP{
					Value:    fmt.Sprintf("@%s", string(securityv1.ResourceGroupVcRemote)),
					Position: position,
				},
				Op: networkingv1beta1firewall.MatchOperationEq,
			}}, nil
		},
	},
}

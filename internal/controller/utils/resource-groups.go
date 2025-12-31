package utils

import (
	"context"
	"fmt"

	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type groupFuncts struct {
	MakeSets      func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error)
	MakeMatchRule func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error)
}

var ResourceGroupFuncts = map[securityv1.ResourceGroup]groupFuncts{
	securityv1.ResourceGroupLocalCluster: {
		MakeSets: nil,
		MakeMatchRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
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
	securityv1.ResourceGroupRemoteCluster: {
		MakeSets: nil,
		MakeMatchRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
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
	securityv1.ResourceGroupOffloaded: {
		MakeSets: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error) {
			pods, err := GetPodsFromConsumer(ctx, cl, clusterID)
			if err != nil {
				return nil, err
			}

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
	securityv1.ResourceGroupVcRemote: {
		MakeSets: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error) {
			pods, err := GetPodsOffloadedToProvider(ctx, cl, clusterID)
			if err != nil {
				return nil, err
			}

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

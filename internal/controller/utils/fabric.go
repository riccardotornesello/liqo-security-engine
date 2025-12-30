package utils

import (
	"context"
	"fmt"

	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	"github.com/liqotech/liqo/pkg/fabric"
	"github.com/liqotech/liqo/pkg/firewall"
	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// fabricResourceNameSuffix is the suffix appended to the cluster ID to form the fabric FirewallConfiguration name.
	fabricResourceNameSuffix = "security-fabric"

	// fabricTableName is the name of the firewall table used by the fabric FirewallConfiguration.
	fabricTableName = "cluster-security"

	// fabricChainName is the name of the firewall chain used by the fabric FirewallConfiguration.
	fabricChainName = "cluster-security-filter"

	// fabricChainPriority is the priority of the fabric firewall chain.
	fabricChainPriority = 200
)

// Generate the name of the Fabric FirewallConfiguration resource for the given cluster ID.
func ForgeFabricResourceName(clusterID string) string {
	return fmt.Sprintf("%s-%s", clusterID, fabricResourceNameSuffix)
}

// ForgeFabricLabels for the given cluster ID.
func ForgeFabricLabels(clusterID string) map[string]string {
	// TODO: liqo managed?
	// TODO: category security?

	return map[string]string{
		firewall.FirewallCategoryTargetKey: fabric.FirewallCategoryTargetValue,
		firewall.FirewallUniqueTargetKey:   fabric.FirewallSubCategoryTargetAllNodesValue,
	}
}

func ForgeFabricSpec(ctx context.Context, cl client.Client, cfg *securityv1.PeeringSecurity, clusterID string, clusterSubnet string) (*networkingv1beta1.FirewallConfigurationSpec, error) {
	// TODO: optimize by creatig only the required sets

	spec := networkingv1beta1.FirewallConfigurationSpec{
		Table: networkingv1beta1firewall.Table{
			Name:   ptr.To(fabricTableName),
			Family: ptr.To(networkingv1beta1firewall.TableFamilyIPv4),
			Chains: []networkingv1beta1firewall.Chain{{
				Name:     ptr.To(fabricChainName),
				Hook:     ptr.To(networkingv1beta1firewall.ChainHookPostrouting),
				Policy:   ptr.To(networkingv1beta1firewall.ChainPolicyAccept),
				Priority: ptr.To[networkingv1beta1firewall.ChainPriority](fabricChainPriority),
				Type:     ptr.To(networkingv1beta1firewall.ChainTypeFilter),
				Rules: networkingv1beta1firewall.RulesSet{
					FilterRules: []networkingv1beta1firewall.FilterRule{
						{
							Name:   ptr.To("only-offloaded"),
							Action: networkingv1beta1firewall.ActionAccept,
							Match: []networkingv1beta1firewall.Match{{
								IP: &networkingv1beta1firewall.MatchIP{
									Value:    clusterSubnet,
									Position: networkingv1beta1firewall.MatchPositionSrc,
								},
								Op: networkingv1beta1firewall.MatchOperationNeq,
							}},
						},
					},
				},
			}},
		},
	}

	// Update the default policy
	if cfg.Spec.BlockOffloadedPodsTraffic {
		spec.Table.Chains[0].Policy = ptr.To(networkingv1beta1firewall.ChainPolicyDrop)
	}

	return &spec, nil
}

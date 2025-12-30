package utils

import (
	"context"
	"fmt"

	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	"github.com/liqotech/liqo/pkg/firewall"
	"github.com/liqotech/liqo/pkg/gateway"
	tenantnamespace "github.com/liqotech/liqo/pkg/tenantNamespace"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
)

const (
	// gatewayResourceNameSuffix is the suffix appended to the cluster ID to form the gateway FirewallConfiguration name.
	gatewayResourceNameSuffix = "security-gateway"

	// gatewayTableName is the name of the firewall table used by the gateway FirewallConfiguration.
	gatewayTableName = "cluster-security"

	// gatewayChainName is the name of the firewall chain used by the gateway FirewallConfiguration.
	gatewayChainName = "cluster-security-filter"

	// gatewayChainPriority is the priority of the gateway firewall chain.
	gatewayChainPriority = 200
)

// Generate the name of the Gateway FirewallConfiguration resource for the given cluster ID.
func ForgeGatewayResourceName(clusterID string) string {
	return fmt.Sprintf("%s-%s", clusterID, gatewayResourceNameSuffix)
}

// ForgeGatewayLabels for the given cluster ID.
func ForgeGatewayLabels(clusterID string) map[string]string {
	// TODO: liqo managed?
	// TODO: category security?

	return map[string]string{
		firewall.FirewallCategoryTargetKey: gateway.FirewallCategoryGwTargetValue,
		firewall.FirewallUniqueTargetKey:   clusterID,
	}
}

// Forge allowed traffic rule
func ForgeAllowedTrafficRule(rule securityv1.AllowedTrafficRule, name *string) networkingv1beta1firewall.FilterRule {
	filterRule := networkingv1beta1firewall.FilterRule{
		Name:   name,
		Action: networkingv1beta1firewall.ActionAccept,
		Match: []networkingv1beta1firewall.Match{{
			IP: &networkingv1beta1firewall.MatchIP{
				Value:    fmt.Sprintf("@%s", rule.Source),
				Position: networkingv1beta1firewall.MatchPositionSrc,
			},
			Op: networkingv1beta1firewall.MatchOperationEq,
		}},
	}

	if rule.Destination != nil {
		filterRule.Match = append(filterRule.Match, networkingv1beta1firewall.Match{
			IP: &networkingv1beta1firewall.MatchIP{
				Value:    fmt.Sprintf("@%s", *rule.Destination),
				Position: networkingv1beta1firewall.MatchPositionDst,
			},
			Op: networkingv1beta1firewall.MatchOperationEq,
		})
	}

	return filterRule
}

func ForgeGatewaySpec(ctx context.Context, cl client.Client, cfg *securityv1.PeeringSecurity, clusterID string) (*networkingv1beta1.FirewallConfigurationSpec, error) {
	// TODO: optimize by creatig only the required sets

	spec := networkingv1beta1.FirewallConfigurationSpec{
		Table: networkingv1beta1firewall.Table{
			Name:   ptr.To(gatewayTableName),
			Family: ptr.To(networkingv1beta1firewall.TableFamilyIPv4),
			Sets:   make([]networkingv1beta1firewall.Set, 0),
			Chains: []networkingv1beta1firewall.Chain{{
				Name:     ptr.To(gatewayChainName),
				Hook:     ptr.To(networkingv1beta1firewall.ChainHookPostrouting),
				Policy:   ptr.To(networkingv1beta1firewall.ChainPolicyAccept),
				Priority: ptr.To[networkingv1beta1firewall.ChainPriority](gatewayChainPriority),
				Type:     ptr.To(networkingv1beta1firewall.ChainTypeFilter),
				Rules: networkingv1beta1firewall.RulesSet{
					FilterRules: []networkingv1beta1firewall.FilterRule{
						{
							Name:   ptr.To("allow-established-related"),
							Action: networkingv1beta1firewall.ActionAccept,
							Match: []networkingv1beta1firewall.Match{{
								CtState: &networkingv1beta1firewall.MatchCtState{
									Value: []networkingv1beta1firewall.CtStateValue{
										networkingv1beta1firewall.CtStateEstablished,
										networkingv1beta1firewall.CtStateRelated,
									},
								},
								Op: networkingv1beta1firewall.MatchOperationEq,
							}},
						},
						{
							Name:   ptr.To("only-from-tunnel"),
							Action: networkingv1beta1firewall.ActionAccept,
							Match: []networkingv1beta1firewall.Match{{
								Dev: &networkingv1beta1firewall.MatchDev{
									Position: networkingv1beta1firewall.MatchDevPositionIn,
									Value:    "liqo-tunnel",
								},
								Op: networkingv1beta1firewall.MatchOperationNeq,
							}},
						},
						{
							Name:   ptr.To("allow-eth"),
							Action: networkingv1beta1firewall.ActionAccept,
							Match: []networkingv1beta1firewall.Match{{
								Dev: &networkingv1beta1firewall.MatchDev{
									Position: networkingv1beta1firewall.MatchDevPositionOut,
									Value:    "eth0",
								},
								Op: networkingv1beta1firewall.MatchOperationEq,
							}},
						},
					},
				},
			}},
		},
	}

	// Update the default policy
	if cfg.Spec.BlockTunnelTraffic {
		spec.Table.Chains[0].Policy = ptr.To(networkingv1beta1firewall.ChainPolicyDrop)
	}

	// Forge the sets
	offloadedSet, err := GetPodsOffloadedToProvider(ctx, cl, clusterID)
	if err != nil {
		return nil, fmt.Errorf("unable to get the pods offloaded to the provider cluster %q: %w", clusterID, err)
	}
	spec.Table.Sets = append(spec.Table.Sets, ForgePodIpsSet(string(securityv1.ResourceGroupOffloaded), offloadedSet))

	fromConsumerSet, err := GetPodsFromConsumer(ctx, cl, clusterID)
	if err != nil {
		return nil, fmt.Errorf("unable to get the pods from the consumer cluster %q: %w", clusterID, err)
	}
	spec.Table.Sets = append(spec.Table.Sets, ForgePodIpsSet(string(securityv1.ResourceGroupFromConsumer), fromConsumerSet))

	// Add the allowed traffic rules
	for i, rule := range cfg.Spec.AllowedTraffic {
		ruleName := fmt.Sprintf("allowed-traffic-%d", i)
		filterRule := ForgeAllowedTrafficRule(rule, &ruleName)
		spec.Table.Chains[0].Rules.FilterRules = append(spec.Table.Chains[0].Rules.FilterRules, filterRule)
	}

	// Return the spec
	return &spec, nil
}

// ExtractClusterID extracts the cluster ID from the given namespace.
func ExtractClusterID(namespace string) (string, error) {
	// Remove the tenantnamespace.NamePrefix + "-" from the namespace to get the cluster ID
	const prefix = tenantnamespace.NamePrefix + "-"

	if len(namespace) <= len(prefix) || namespace[:len(prefix)] != prefix {
		return "", fmt.Errorf("namespace %q does not have the expected prefix %q", namespace, prefix)
	}
	return namespace[len(prefix):], nil
}

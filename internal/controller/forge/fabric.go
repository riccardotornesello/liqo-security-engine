// Package forge provides functions for creating Liqo FirewallConfiguration resources.
// It translates PeeringConnectivity security rules into low-level firewall configurations
// that can be applied to the Liqo fabric network.
package forge

import (
	"context"
	"fmt"

	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	"github.com/liqotech/liqo/pkg/fabric"
	"github.com/liqotech/liqo/pkg/firewall"
	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
	"github.com/riccardotornesello/liqo-security-manager/internal/controller/utils"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// fabricResourceNameSuffix is the suffix appended to the cluster ID to form the fabric FirewallConfiguration name.
	fabricResourceNameSuffix = "security-fabric"

	// fabricTableName is the name of the nftables table used by the fabric FirewallConfiguration.
	fabricTableName = "cluster-security"

	// fabricChainName is the name of the nftables chain used by the fabric FirewallConfiguration.
	fabricChainName = "cluster-security-filter"

	// fabricChainPriority is the priority of the fabric firewall chain.
	// Lower values have higher priority.
	fabricChainPriority = 200
)

// ForgeFabricResourceName generates the name of the Fabric FirewallConfiguration resource
// for the given cluster ID. The name follows the pattern: <cluster-id>-security-fabric
func ForgeFabricResourceName(clusterID string) string {
	return fmt.Sprintf("%s-%s", clusterID, fabricResourceNameSuffix)
}

// ForgeFabricLabels creates the labels for a Fabric FirewallConfiguration resource.
// These labels identify the configuration as a fabric-level security configuration
// that targets all nodes in the cluster.
func ForgeFabricLabels(clusterID string) map[string]string {
	// TODO: label liqo-managed?
	// TODO: create new category?

	// Labels identify this as a fabric-level firewall configuration targeting all nodes.
	return map[string]string{
		firewall.FirewallCategoryTargetKey:    fabric.FirewallCategoryTargetValue,
		firewall.FirewallSubCategoryTargetKey: fabric.FirewallSubCategoryTargetAllNodesValue,
	}
}

// ForgeFabricSpec creates the FirewallConfiguration spec from a PeeringConnectivity resource.
// It translates the high-level security rules into low-level nftables firewall rules,
// including:
// - Creating firewall sets for dynamic pod IP collections
// - Creating match rules for source and destination filtering
// - Setting up allow/deny actions based on the rule specifications
// - Adding a default rule to allow established/related connections
func ForgeFabricSpec(ctx context.Context, cl client.Client, cfg *securityv1.PeeringConnectivity, clusterID string) (*networkingv1beta1.FirewallConfigurationSpec, error) {
	// Initialize the FirewallConfiguration with basic structure.
	spec := networkingv1beta1.FirewallConfigurationSpec{
		Table: networkingv1beta1firewall.Table{
			Name:   ptr.To(fabricTableName),
			Family: ptr.To(networkingv1beta1firewall.TableFamilyIPv4),
			Sets:   make([]networkingv1beta1firewall.Set, 0),
			Chains: []networkingv1beta1firewall.Chain{{
				Name:     ptr.To(fabricChainName),
				Hook:     ptr.To(networkingv1beta1firewall.ChainHookPostrouting),
				Policy:   ptr.To(networkingv1beta1firewall.ChainPolicyAccept),
				Priority: ptr.To[networkingv1beta1firewall.ChainPriority](fabricChainPriority),
				Type:     networkingv1beta1firewall.ChainTypeFilter,
				Rules: networkingv1beta1firewall.RulesSet{
					FilterRules: []networkingv1beta1firewall.FilterRule{
						{
							// First rule: Always allow established and related connections.
							// This is essential to allow responses to outgoing connections.
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
					},
				},
			}},
		},
	}

	// Add the allowed traffic rules
	usedResourceGroups := make(map[securityv1.ResourceGroup]struct{})

	for i, rule := range cfg.Spec.Rules {
		ruleName := fmt.Sprintf("allowed-traffic-%d", i)

		filterRule := networkingv1beta1firewall.FilterRule{
			Name:   ptr.To(ruleName),
			Action: networkingv1beta1firewall.ActionAccept,
			Match:  []networkingv1beta1firewall.Match{},
		}

		// Set the action based on the rule specification.
		if rule.Action != securityv1.ActionAllow {
			filterRule.Action = networkingv1beta1firewall.ActionDrop
		}

		// Add match rules for the source (if specified).
		sourceRules, err := ForgeMatchRule(ctx, rule.Source, networkingv1beta1firewall.MatchPositionSrc, usedResourceGroups)
		if err != nil {
			return nil, err
		}
		filterRule.Match = append(filterRule.Match, sourceRules...)

		// Add match rules for the destination (if specified).
		destRules, err := ForgeMatchRule(ctx, rule.Destination, networkingv1beta1firewall.MatchPositionDst, usedResourceGroups)
		if err != nil {
			return nil, err
		}
		filterRule.Match = append(filterRule.Match, destRules...)

		// Add the filter rule to the chain.
		spec.Table.Chains[0].Rules.FilterRules = append(spec.Table.Chains[0].Rules.FilterRules, filterRule)
	}

	// Create firewall sets for all resource groups that require them.
	// Sets contain collections of IP addresses (e.g., pod IPs) that can be referenced in rules.
	for rg := range usedResourceGroups {
		if utils.ResourceGroupFuncts[rg].MakeSets != nil {
			sets, err := utils.ResourceGroupFuncts[rg].MakeSets(ctx, cl, clusterID)
			if err != nil {
				return nil, err
			}
			spec.Table.Sets = append(spec.Table.Sets, sets...)
		}
	}

	// Return the complete FirewallConfiguration spec.
	return &spec, nil
}

// ForgeMatchRule creates firewall match rules for a party (source or destination).
// It translates a high-level Party specification into low-level nftables match rules
// and tracks which resource groups are used so their sets can be created.
func ForgeMatchRule(ctx context.Context, party *securityv1.Party, position networkingv1beta1firewall.MatchPosition, usedResourceGroups map[securityv1.ResourceGroup]struct{}) (matchRules []networkingv1beta1firewall.Match, err error) {
	if party == nil {
		// No party specified, so no match rules needed (matches all).
		return nil, nil
	}

	if party.Group != nil {
		// Generate match rules for the specified resource group.
		matchRules, err = utils.ResourceGroupFuncts[*party.Group].MakeMatchRule(ctx, nil, "", position)
		if err != nil {
			return nil, err
		}
		// Mark this resource group as used so its set will be created.
		usedResourceGroups[*party.Group] = struct{}{}
	}

	return matchRules, nil
}

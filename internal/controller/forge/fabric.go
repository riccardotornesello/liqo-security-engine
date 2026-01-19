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
		firewall.FirewallCategoryTargetKey:    fabric.FirewallCategoryTargetValue,
		firewall.FirewallSubCategoryTargetKey: fabric.FirewallSubCategoryTargetAllNodesValue,
	}
}

func ForgeFabricSpec(ctx context.Context, cl client.Client, cfg *securityv1.PeeringConnectivity, clusterID string) (*networkingv1beta1.FirewallConfigurationSpec, error) {
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

		if rule.Action != securityv1.ActionAllow {
			filterRule.Action = networkingv1beta1firewall.ActionDrop
		}

		sourceRules, err := ForgeMatchRule(rule.Source, networkingv1beta1firewall.MatchPositionSrc, usedResourceGroups)
		if err != nil {
			return nil, err
		}
		filterRule.Match = append(filterRule.Match, sourceRules...)

		destRules, err := ForgeMatchRule(rule.Destination, networkingv1beta1firewall.MatchPositionDst, usedResourceGroups)
		if err != nil {
			return nil, err
		}
		filterRule.Match = append(filterRule.Match, destRules...)

		spec.Table.Chains[0].Rules.FilterRules = append(spec.Table.Chains[0].Rules.FilterRules, filterRule)
	}

	// Add the required sets
	for rg := range usedResourceGroups {
		if utils.ResourceGroupFuncts[rg].MakeSets != nil {
			sets, err := utils.ResourceGroupFuncts[rg].MakeSets(ctx, cl, clusterID)
			if err != nil {
				return nil, err
			}
			spec.Table.Sets = append(spec.Table.Sets, sets...)
		}
	}

	// Return the spec
	return &spec, nil
}

func ForgeMatchRule(party *securityv1.Party, position networkingv1beta1firewall.MatchPosition, usedResourceGroups map[securityv1.ResourceGroup]struct{}) (matchRules []networkingv1beta1firewall.Match, err error) {
	if party == nil {
		return nil, nil
	}

	if party.Group != nil {
		matchRules, err = utils.ResourceGroupFuncts[*party.Group].MakeMatchRule(context.TODO(), nil, "", position)
		if err != nil {
			return nil, err
		}
		usedResourceGroups[*party.Group] = struct{}{}
	}

	return matchRules, nil
}

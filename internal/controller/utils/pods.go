// Package utils provides utility functions for managing pod collections in Liqo environments.
// It includes functions to retrieve pods based on their role in the peering scenario
// (offloaded pods, shadow pods, virtual cluster pods, etc.).
package utils

import (
	"context"

	offloadingv1beta1 "github.com/liqotech/liqo/apis/offloading/v1beta1"
	"github.com/liqotech/liqo/pkg/consts"
	"github.com/liqotech/liqo/pkg/virtualKubelet/forge"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetPodsOffloadedToProvider returns the list of shadow pods on the consumer cluster
// that represent pods offloaded to a specific provider cluster.
// This function is used on the consumer side.
//
// Shadow pods are identified by the liqo.io/local-pod label and are scheduled
// on a virtual node representing the provider cluster.
//
// For consumer only!
func GetPodsOffloadedToProvider(ctx context.Context, cl client.Client, providerClusterID string) ([]corev1.Pod, error) {
	// TODO: optimize by adding labels in liqo when offloading pods (see TODOS.md)

	// Get all the pods offloaded to the provider cluster.
	podList := &corev1.PodList{}
	if err := cl.List(ctx, podList, client.MatchingLabels{
		consts.LocalPodLabelKey: consts.LocalPodLabelValue,
	}); err != nil {
		return nil, err
	}

	// Filter the pods that are scheduled on the specified provider cluster.
	// The node name corresponds to the cluster ID of the provider.
	pods := make([]corev1.Pod, 0)
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == providerClusterID {
			pods = append(pods, pod)
		}
	}

	return pods, nil
}

// GetPodsFromConsumer returns the list of actual pods on the provider cluster
// that were offloaded from a specific consumer cluster.
// This function is used on the provider side.
//
// These pods are identified by the liqo.io/origin-cluster-id label which contains
// the cluster ID of the consumer cluster that offloaded them.
//
// For provider only!
func GetPodsFromConsumer(ctx context.Context, cl client.Client, consumerClusterID string) ([]corev1.Pod, error) {
	podList := &corev1.PodList{}
	if err := cl.List(ctx, podList, client.MatchingLabels{
		forge.LiqoOriginClusterIDKey: consumerClusterID,
	}); err != nil {
		return nil, err
	}

	return podList.Items, nil
}

// GetPodsInOffloadedNamespaces returns the list of pods running in namespaces
// that are configured for offloading (have a NamespaceOffloading resource).
// This function excludes shadow pods (liqo.io/local-pod label).
//
// These are the actual local pods that could potentially be offloaded to remote clusters.
func GetPodsInOffloadedNamespaces(ctx context.Context, cl client.Client) ([]corev1.Pod, error) {
	namespaceList := &offloadingv1beta1.NamespaceOffloadingList{}
	if err := cl.List(ctx, namespaceList); err != nil {
		return nil, err
	}

	// Get all the pods in the offloaded namespaces.
	// Exclude local shadow pods (which represent offloaded pods).
	var pods []corev1.Pod

	// Create a label requirement to exclude shadow pods.
	// NotEquals includes also the case where the label is not present.
	labelsRequirement, err := labels.NewRequirement(consts.LocalPodLabelKey, selection.NotEquals, []string{consts.LocalPodLabelValue})
	if err != nil {
		return nil, err
	}

	for _, nso := range namespaceList.Items {
		podList := &corev1.PodList{}
		if err := cl.List(
			ctx,
			podList,
			client.InNamespace(nso.Namespace),
			client.MatchingLabelsSelector{Selector: labels.NewSelector().Add(*labelsRequirement)},
		); err != nil {
			return nil, err
		}
		pods = append(pods, podList.Items...)
	}

	return pods, nil
}

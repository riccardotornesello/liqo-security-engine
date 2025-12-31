package utils

import (
	"context"

	"github.com/liqotech/liqo/pkg/consts"
	"github.com/liqotech/liqo/pkg/virtualKubelet/forge"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Function for the Consumer. Returns the list of Pod IPs on the remote Provider cluster.
func GetPodsOffloadedToProvider(ctx context.Context, cl client.Client, providerClusterID string) ([]corev1.Pod, error) {
	// TODO: optimize by adding labels in liqo when offloading pods

	// Get all the pods offloaded to the provider cluster.
	podList := &corev1.PodList{}
	if err := cl.List(ctx, podList, client.MatchingLabels{
		consts.LocalPodLabelKey: consts.LocalPodLabelValue,
	}); err != nil {
		return nil, err
	}

	// Filter the pods owned by the provider cluster.
	pods := make([]corev1.Pod, 0)
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == providerClusterID {
			pods = append(pods, pod)
		}
	}

	return pods, nil
}

// Function for the Provider. Returns the list of Pods owned by the Consumer.
func GetPodsFromConsumer(ctx context.Context, cl client.Client, consumerClusterID string) ([]corev1.Pod, error) {
	// Get the pods coming from the remote cluster.
	podList := &corev1.PodList{}
	if err := cl.List(ctx, podList, client.MatchingLabels{
		forge.LiqoOriginClusterIDKey: consumerClusterID,
	}); err != nil {
		return nil, err
	}

	return podList.Items, nil
}

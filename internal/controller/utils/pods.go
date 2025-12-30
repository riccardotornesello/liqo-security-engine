package utils

import (
	"context"

	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
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

// Function for the Consumer. Forges a Set containing the IPs of the given Pods.
func ForgePodIpsSet(setName string, pods []corev1.Pod) networkingv1beta1firewall.Set {
	setElements := make([]networkingv1beta1firewall.SetElement, 0)
	for _, pod := range pods {
		podIp := pod.Status.PodIP
		if podIp == "" {
			continue
		}
		setElements = append(setElements, networkingv1beta1firewall.SetElement{
			Key: podIp,
		})
	}

	return networkingv1beta1firewall.Set{
		Name:     setName,
		KeyType:  networkingv1beta1firewall.SetDataTypeIPAddr,
		Elements: setElements,
	}
}

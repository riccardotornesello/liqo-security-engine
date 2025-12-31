package utils

import (
	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	corev1 "k8s.io/api/core/v1"
)

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

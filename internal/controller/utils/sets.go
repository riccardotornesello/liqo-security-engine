// Copyright 2019-2026 The Liqo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	corev1 "k8s.io/api/core/v1"
)

// ForgePodIpsSet creates a firewall Set containing the IP addresses of the given pods.
// This set can be referenced in firewall rules to match traffic to/from these pods.
// Pods without an IP address are excluded from the set.
//
// Parameters:
//   - setName: The name to assign to the firewall set (used for referencing in rules)
//   - pods: The list of pods whose IPs should be included in the set
//
// Returns a networkingv1beta1firewall.Set containing the pod IPs.
func ForgePodIpsSet(setName string, pods []corev1.Pod) networkingv1beta1firewall.Set {
	setElements := make([]networkingv1beta1firewall.SetElement, 0)
	for _, pod := range pods {
		podIp := pod.Status.PodIP
		if podIp == "" {
			// Skip pods that don't have an IP address yet.
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

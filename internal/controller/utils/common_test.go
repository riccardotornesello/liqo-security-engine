/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"testing"

	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Common Utilities", func() {
	Describe("GetClusterNamespace", func() {
		It("should return the correct namespace for a cluster ID", func() {
			clusterID := "test-cluster-123"
			expected := "liqo-tenant-test-cluster-123"
			result := GetClusterNamespace(clusterID)
			Expect(result).To(Equal(expected))
		})

		It("should handle short cluster IDs", func() {
			clusterID := "abc"
			expected := "liqo-tenant-abc"
			result := GetClusterNamespace(clusterID)
			Expect(result).To(Equal(expected))
		})

		It("should handle cluster IDs with hyphens", func() {
			clusterID := "test-cluster-with-hyphens"
			expected := "liqo-tenant-test-cluster-with-hyphens"
			result := GetClusterNamespace(clusterID)
			Expect(result).To(Equal(expected))
		})
	})

	Describe("ExtractClusterID", func() {
		Context("when namespace has correct format", func() {
			It("should extract cluster ID successfully", func() {
				namespace := "liqo-tenant-test-cluster-123"
				expected := "test-cluster-123"
				result, err := ExtractClusterID(namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expected))
			})

			It("should handle cluster IDs with hyphens", func() {
				namespace := "liqo-tenant-my-cluster-id"
				expected := "my-cluster-id"
				result, err := ExtractClusterID(namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expected))
			})

			It("should handle short cluster IDs", func() {
				namespace := "liqo-tenant-abc"
				expected := "abc"
				result, err := ExtractClusterID(namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expected))
			})
		})

		Context("when namespace has incorrect format", func() {
			It("should return error for namespace without prefix", func() {
				namespace := "default"
				_, err := ExtractClusterID(namespace)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("does not have the expected prefix"))
			})

			It("should return error for namespace with wrong prefix", func() {
				namespace := "kube-system-test"
				_, err := ExtractClusterID(namespace)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for partial prefix match", func() {
				namespace := "liqo-tenant"
				_, err := ExtractClusterID(namespace)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for namespace with only prefix and hyphen", func() {
				namespace := "liqo-tenant-"
				_, err := ExtractClusterID(namespace)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("does not have the expected prefix"))
			})

			It("should return error for empty namespace", func() {
				namespace := ""
				_, err := ExtractClusterID(namespace)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var _ = Describe("Sets Utilities", func() {
	Describe("ForgePodIpsSet", func() {
		Context("when creating a set with multiple pods", func() {
			It("should create a set with all pod IPs", func() {
				pods := []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.1"},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.2"},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod3"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.3"},
					},
				}

				setName := "test-set"
				result := ForgePodIpsSet(setName, pods)

				Expect(result.Name).To(Equal(setName))
				Expect(result.KeyType).To(Equal(networkingv1beta1firewall.SetDataTypeIPAddr))
				Expect(result.Elements).To(HaveLen(3))
				Expect(result.Elements[0].Key).To(Equal("10.0.0.1"))
				Expect(result.Elements[1].Key).To(Equal("10.0.0.2"))
				Expect(result.Elements[2].Key).To(Equal("10.0.0.3"))
			})
		})

		Context("when creating a set with empty pod list", func() {
			It("should create an empty set", func() {
				pods := []corev1.Pod{}
				setName := "empty-set"
				result := ForgePodIpsSet(setName, pods)

				Expect(result.Name).To(Equal(setName))
				Expect(result.KeyType).To(Equal(networkingv1beta1firewall.SetDataTypeIPAddr))
				Expect(result.Elements).To(HaveLen(0))
			})
		})

		Context("when pods have no IP addresses", func() {
			It("should exclude pods without IPs", func() {
				pods := []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
						Status:     corev1.PodStatus{PodIP: ""},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.2"},
					},
				}

				setName := "partial-set"
				result := ForgePodIpsSet(setName, pods)

				Expect(result.Name).To(Equal(setName))
				Expect(result.Elements).To(HaveLen(1))
				Expect(result.Elements[0].Key).To(Equal("10.0.0.2"))
			})
		})

		Context("when all pods have no IP addresses", func() {
			It("should create an empty set", func() {
				pods := []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
						Status:     corev1.PodStatus{PodIP: ""},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
						Status:     corev1.PodStatus{},
					},
				}

				setName := "no-ips-set"
				result := ForgePodIpsSet(setName, pods)

				Expect(result.Name).To(Equal(setName))
				Expect(result.Elements).To(HaveLen(0))
			})
		})

		Context("when pods have IPv6 addresses", func() {
			It("should include IPv6 addresses in the set", func() {
				pods := []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
						Status:     corev1.PodStatus{PodIP: "2001:db8::1"},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.1"},
					},
				}

				setName := "mixed-ip-set"
				result := ForgePodIpsSet(setName, pods)

				Expect(result.Name).To(Equal(setName))
				Expect(result.Elements).To(HaveLen(2))
				Expect(result.Elements[0].Key).To(Equal("2001:db8::1"))
				Expect(result.Elements[1].Key).To(Equal("10.0.0.1"))
			})
		})

		Context("when set name contains special characters", func() {
			It("should preserve the set name as-is", func() {
				pods := []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.1"},
					},
				}

				setName := "test-set_123"
				result := ForgePodIpsSet(setName, pods)

				Expect(result.Name).To(Equal(setName))
			})
		})
	})
})

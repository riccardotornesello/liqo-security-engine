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
	"context"

	offloadingv1beta1 "github.com/liqotech/liqo/apis/offloading/v1beta1"
	"github.com/liqotech/liqo/pkg/consts"
	"github.com/liqotech/liqo/pkg/virtualKubelet/forge"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Pods Utilities", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		RegisterScheme(scheme)
	})

	Describe("GetPodsOffloadedToProvider", func() {
		It("should return pods offloaded to a specific provider cluster", func() {
			providerClusterID := "provider-cluster-123"

			pod1 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
					Labels: map[string]string{
						consts.LocalPodLabelKey: consts.LocalPodLabelValue,
					},
				},
				Spec: corev1.PodSpec{
					NodeName: providerClusterID,
				},
			}

			pod2 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "default",
					Labels: map[string]string{
						consts.LocalPodLabelKey: consts.LocalPodLabelValue,
					},
				},
				Spec: corev1.PodSpec{
					NodeName: "other-cluster",
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod1, pod2).Build()

			pods, err := GetPodsOffloadedToProvider(ctx, cl, providerClusterID)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(HaveLen(1))
			Expect(pods[0].Name).To(Equal("pod1"))
		})

		It("should return empty list when no pods are offloaded to the provider", func() {
			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			pods, err := GetPodsOffloadedToProvider(ctx, cl, "nonexistent-cluster")
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(BeEmpty())
		})

		It("should filter pods by node name", func() {
			providerClusterID := "provider-cluster-abc"

			pod1 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "ns1",
					Labels: map[string]string{
						consts.LocalPodLabelKey: consts.LocalPodLabelValue,
					},
				},
				Spec: corev1.PodSpec{
					NodeName: providerClusterID,
				},
			}

			pod2 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "ns2",
					Labels: map[string]string{
						consts.LocalPodLabelKey: consts.LocalPodLabelValue,
					},
				},
				Spec: corev1.PodSpec{
					NodeName: providerClusterID,
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod1, pod2).Build()

			pods, err := GetPodsOffloadedToProvider(ctx, cl, providerClusterID)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(HaveLen(2))
		})
	})

	Describe("GetPodsFromConsumer", func() {
		It("should return pods from a specific consumer cluster", func() {
			consumerClusterID := "consumer-cluster-123"

			pod1 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
					Labels: map[string]string{
						forge.LiqoOriginClusterIDKey: consumerClusterID,
					},
				},
			}

			pod2 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "default",
					Labels: map[string]string{
						forge.LiqoOriginClusterIDKey: "other-consumer",
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod1, pod2).Build()

			pods, err := GetPodsFromConsumer(ctx, cl, consumerClusterID)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(HaveLen(1))
			Expect(pods[0].Name).To(Equal("pod1"))
		})

		It("should return empty list when no pods are from the consumer", func() {
			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			pods, err := GetPodsFromConsumer(ctx, cl, "nonexistent-consumer")
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(BeEmpty())
		})
	})

	Describe("GetPodsInOffloadedNamespaces", func() {
		It("should return pods in namespaces with NamespaceOffloading", func() {
			nso := &offloadingv1beta1.NamespaceOffloading{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "offloading",
					Namespace: "test-ns",
				},
			}

			pod1 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "test-ns",
				},
			}

			pod2 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "other-ns",
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(nso, pod1, pod2).Build()

			pods, err := GetPodsInOffloadedNamespaces(ctx, cl)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(HaveLen(1))
			Expect(pods[0].Name).To(Equal("pod1"))
		})

		It("should exclude shadow pods", func() {
			nso := &offloadingv1beta1.NamespaceOffloading{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "offloading",
					Namespace: "test-ns",
				},
			}

			pod1 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "real-pod",
					Namespace: "test-ns",
				},
			}

			pod2 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "shadow-pod",
					Namespace: "test-ns",
					Labels: map[string]string{
						consts.LocalPodLabelKey: consts.LocalPodLabelValue,
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(nso, pod1, pod2).Build()

			pods, err := GetPodsInOffloadedNamespaces(ctx, cl)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(HaveLen(1))
			Expect(pods[0].Name).To(Equal("real-pod"))
		})

		It("should return empty list when no NamespaceOffloading resources exist", func() {
			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			pods, err := GetPodsInOffloadedNamespaces(ctx, cl)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(BeEmpty())
		})
	})

	Describe("GetPodsInNamespace", func() {
		It("should return all pods in a specific namespace", func() {
			pod1 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "test-ns",
				},
			}

			pod2 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "test-ns",
				},
			}

			pod3 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod3",
					Namespace: "other-ns",
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod1, pod2, pod3).Build()

			pods, err := GetPodsInNamespace(ctx, cl, "test-ns")
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(HaveLen(2))
		})

		It("should return empty list for non-existent namespace", func() {
			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			pods, err := GetPodsInNamespace(ctx, cl, "nonexistent-ns")
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(BeEmpty())
		})

		It("should return empty list for empty namespace", func() {
			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			pods, err := GetPodsInNamespace(ctx, cl, "empty-ns")
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(BeEmpty())
		})
	})
})

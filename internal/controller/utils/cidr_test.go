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

	ipamv1alpha1 "github.com/liqotech/liqo/apis/ipam/v1alpha1"
	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CIDR Utilities", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		RegisterScheme(scheme)
	})

	Describe("GetCurrentClusterPodCIDR", func() {
		It("should retrieve the local pod CIDR from the Network resource", func() {
			network := &ipamv1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-cidr",
					Namespace: "liqo",
				},
				Spec: ipamv1alpha1.NetworkSpec{
					CIDR: networkingv1beta1.CIDR("10.0.0.0/16"),
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(network).Build()

			cidr, err := GetCurrentClusterPodCIDR(ctx, cl)
			Expect(err).NotTo(HaveOccurred())
			Expect(cidr).To(Equal("10.0.0.0/16"))
		})

		It("should return error when Network resource does not exist", func() {
			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			_, err := GetCurrentClusterPodCIDR(ctx, cl)
			Expect(err).To(HaveOccurred())
		})

		It("should handle IPv6 CIDR", func() {
			network := &ipamv1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-cidr",
					Namespace: "liqo",
				},
				Spec: ipamv1alpha1.NetworkSpec{
					CIDR: networkingv1beta1.CIDR("2001:db8::/32"),
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(network).Build()

			cidr, err := GetCurrentClusterPodCIDR(ctx, cl)
			Expect(err).NotTo(HaveOccurred())
			Expect(cidr).To(Equal("2001:db8::/32"))
		})
	})

	Describe("GetRemoteClusterPodCIDR", func() {
		It("should retrieve the remote pod CIDR from the Network resource", func() {
			clusterID := "remote-cluster-123"
			network := &ipamv1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "remote-cluster-123-pod",
					Namespace: "liqo-tenant-remote-cluster-123",
				},
				Status: ipamv1alpha1.NetworkStatus{
					CIDR: networkingv1beta1.CIDR("10.1.0.0/16"),
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(network).Build()

			cidr, err := GetRemoteClusterPodCIDR(ctx, cl, clusterID)
			Expect(err).NotTo(HaveOccurred())
			Expect(cidr).To(Equal("10.1.0.0/16"))
		})

		It("should return error when Network resource does not exist", func() {
			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			_, err := GetRemoteClusterPodCIDR(ctx, cl, "nonexistent-cluster")
			Expect(err).To(HaveOccurred())
		})

		It("should handle different cluster IDs", func() {
			clusterID := "cluster-abc"
			network := &ipamv1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cluster-abc-pod",
					Namespace: "liqo-tenant-cluster-abc",
				},
				Status: ipamv1alpha1.NetworkStatus{
					CIDR: networkingv1beta1.CIDR("10.2.0.0/16"),
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(network).Build()

			cidr, err := GetRemoteClusterPodCIDR(ctx, cl, clusterID)
			Expect(err).NotTo(HaveOccurred())
			Expect(cidr).To(Equal("10.2.0.0/16"))
		})
	})
})

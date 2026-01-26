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

package controller

import (
	"context"

	ipamv1alpha1 "github.com/liqotech/liqo/apis/ipam/v1alpha1"
	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
	"github.com/riccardotornesello/liqo-security-manager/internal/controller/utils"
)

var _ = Describe("PeeringConnectivity Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName       = "test-resource"
			clusterID          = "test-cluster-123"
			alwaysPresentRules = 1 // established connection tracking rule
		)

		var (
			ctx            context.Context
			namespace      string
			namespacedName types.NamespacedName
			reconciler     *PeeringConnectivityReconciler
		)

		BeforeEach(func() {
			ctx = context.Background()
			// Use proper liqo-tenant namespace format
			namespace = "liqo-tenant-" + clusterID
			namespacedName = types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}
			reconciler = &PeeringConnectivityReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: k8sMgr.GetEventRecorderFor("peeringconnectivity-controller"),
			}

			// Create the namespaces if they do not exist
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			liqoNs := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "liqo",
				},
			}
			err = k8sClient.Create(ctx, liqoNs)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			// Create a Network resource for the remote cluster's pod CIDR
			network := &ipamv1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterID + "-pod",
					Namespace: namespace,
				},
				Spec: ipamv1alpha1.NetworkSpec{
					CIDR: networkingv1beta1.CIDR("10.0.0.0/16"),
				},
				Status: ipamv1alpha1.NetworkStatus{
					CIDR: networkingv1beta1.CIDR("10.0.0.0/16"),
				},
			}
			err = k8sClient.Create(ctx, network)
			Expect(err).NotTo(HaveOccurred())

			// Create a network resource for the local cluster's pod CIDR
			localNetwork := &ipamv1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-cidr",
					Namespace: "liqo",
				},
				Spec: ipamv1alpha1.NetworkSpec{
					CIDR: networkingv1beta1.CIDR("10.1.0.0/16"),
				},
				Status: ipamv1alpha1.NetworkStatus{
					CIDR: networkingv1beta1.CIDR("10.1.0.0/16"),
				},
			}
			err = k8sClient.Create(ctx, localNetwork)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			// Cleanup resources
			resource := &securityv1.PeeringConnectivity{}
			err := k8sClient.Get(ctx, namespacedName, resource)
			if err == nil {
				By("Cleanup the specific resource instance PeeringConnectivity")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

				// Run the reconcile to process deletion finalizers
				_, err = reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
			}

			network := &ipamv1alpha1.Network{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      clusterID + "-pod",
				Namespace: namespace,
			}, network)
			if err == nil {
				By("Cleanup the Network resource")
				Expect(k8sClient.Delete(ctx, network)).To(Succeed())
			}

			localNetwork := &ipamv1alpha1.Network{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      "pod-cidr",
				Namespace: "liqo",
			}, localNetwork)
			if err == nil {
				By("Cleanup the local Network resource")
				Expect(k8sClient.Delete(ctx, localNetwork)).To(Succeed())
			}

			// Wait for the resource to be deleted
			Eventually(func() bool {
				err := k8sClient.Get(ctx, namespacedName, resource)
				return errors.IsNotFound(err)
			}).Should(BeTrue())
		})

		It("should successfully create PeeringConnectivity resource with basic rules", func() {
			By("creating the custom resource for the Kind PeeringConnectivity")
			resource := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{
					Rules: []securityv1.Rule{
						{
							Action:      securityv1.ActionAllow,
							Source:      &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
							Destination: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupLocalCluster)},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			By("Reconciling the created resource")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the FirewallConfiguration was created")
			fwcfg := &networkingv1beta1.FirewallConfiguration{}
			fwcfgName := types.NamespacedName{
				Name:      clusterID + "-security-fabric",
				Namespace: namespace,
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, fwcfgName, fwcfg)
			}).Should(Succeed())

			By("Verifying the status was updated")
			Eventually(func() metav1.ConditionStatus {
				err := k8sClient.Get(ctx, namespacedName, resource)
				if err != nil {
					return metav1.ConditionUnknown
				}
				for _, cond := range resource.Status.Conditions {
					if cond.Type == utils.ConditionTypeReady {
						return cond.Status
					}
				}
				return metav1.ConditionUnknown
			}).Should(Equal(metav1.ConditionTrue))
		})

		It("should handle PeeringConnectivity with deny rules", func() {
			By("creating a PeeringConnectivity with deny rules")
			resource := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{
					Rules: []securityv1.Rule{
						{
							Action: securityv1.ActionDeny,
							Source: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			By("Reconciling the created resource")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the FirewallConfiguration was created")
			fwcfg := &networkingv1beta1.FirewallConfiguration{}
			fwcfgName := types.NamespacedName{
				Name:      clusterID + "-security-fabric",
				Namespace: namespace,
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, fwcfgName, fwcfg)
			}).Should(Succeed())
		})

		It("should handle PeeringConnectivity with empty rules", func() {
			By("creating a PeeringConnectivity with no rules")
			resource := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{
					Rules: []securityv1.Rule{},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			By("Reconciling the created resource")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle invalid namespace format", func() {
			By("creating a PeeringConnectivity in wrong namespace")
			invalidNamespace := "default"
			invalidName := types.NamespacedName{
				Name:      "invalid-resource",
				Namespace: invalidNamespace,
			}

			// Ensure default namespace exists
			ns := &corev1.Namespace{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: invalidNamespace}, ns)
			if errors.IsNotFound(err) {
				ns = &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: invalidNamespace,
					},
				}
				Expect(k8sClient.Create(ctx, ns)).To(Succeed())
			}

			resource := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-resource",
					Namespace: invalidNamespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			By("Reconciling should fail with cluster ID extraction error")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: invalidName,
			})
			Expect(err).To(HaveOccurred())

			By("Verifying the status reflects the error")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, invalidName, resource)
				if err != nil {
					return false
				}
				for _, cond := range resource.Status.Conditions {
					if cond.Type == utils.ConditionTypeReady && cond.Status == metav1.ConditionFalse {
						return cond.Reason == ConditionReasonClusterIDError
					}
				}
				return false
			}).Should(BeTrue())

			// Cleanup
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should update FirewallConfiguration when PeeringConnectivity is updated", func() {
			By("creating initial PeeringConnectivity")
			resource := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{
					Rules: []securityv1.Rule{
						{
							Action: securityv1.ActionAllow,
							Source: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			By("Reconciling the created resource")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the FirewallConfiguration was created")
			fwcfg := &networkingv1beta1.FirewallConfiguration{}
			fwcfgName := types.NamespacedName{
				Name:      clusterID + "-security-fabric",
				Namespace: namespace,
			}
			Eventually(func() int {
				err := k8sClient.Get(ctx, fwcfgName, fwcfg)
				if err != nil {
					return 0
				}
				// The table should have chains with rules
				if len(fwcfg.Spec.Table.Chains) > 0 {
					return len(fwcfg.Spec.Table.Chains[0].Rules.FilterRules)
				}
				return 0
			}).Should(BeNumerically("==", 1+alwaysPresentRules))

			By("Updating the PeeringConnectivity rules")
			Eventually(func() error {
				err := k8sClient.Get(ctx, namespacedName, resource)
				if err != nil {
					return err
				}
				resource.Spec.Rules = append(resource.Spec.Rules, securityv1.Rule{
					Action: securityv1.ActionDeny,
					Source: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupLocalCluster)},
				})
				return k8sClient.Update(ctx, resource)
			}).Should(Succeed())

			By("Reconciling again")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the FirewallConfiguration was updated")
			Eventually(func() int {
				err := k8sClient.Get(ctx, fwcfgName, fwcfg)
				if err != nil {
					return 0
				}
				// The table should have chains with rules
				if len(fwcfg.Spec.Table.Chains) > 0 {
					return len(fwcfg.Spec.Table.Chains[0].Rules.FilterRules)
				}
				return 0
			}).Should(BeNumerically("==", 2+alwaysPresentRules))
		})
	})
})

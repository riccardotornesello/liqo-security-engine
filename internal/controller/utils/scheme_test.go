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
	ipamv1alpha1 "github.com/liqotech/liqo/apis/ipam/v1alpha1"
	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	offloadingv1beta1 "github.com/liqotech/liqo/apis/offloading/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
)

var _ = Describe("Scheme Utilities", func() {
	Describe("RegisterScheme", func() {
		It("should register all required schemes", func() {
			scheme := runtime.NewScheme()
			RegisterScheme(scheme)

			// Verify that core Kubernetes types are registered
			Expect(scheme.IsGroupRegistered(corev1.SchemeGroupVersion.Group)).To(BeTrue())

			// Verify that Liqo networking types are registered
			Expect(scheme.IsGroupRegistered(networkingv1beta1.GroupVersion.Group)).To(BeTrue())

			// Verify that Liqo IPAM types are registered
			Expect(scheme.IsGroupRegistered(ipamv1alpha1.SchemeGroupVersion.Group)).To(BeTrue())

			// Verify that Liqo offloading types are registered
			Expect(scheme.IsGroupRegistered(offloadingv1beta1.SchemeGroupVersion.Group)).To(BeTrue())

			// Verify that security types are registered
			Expect(scheme.IsGroupRegistered(securityv1.GroupVersion.Group)).To(BeTrue())
		})

		It("should allow creation of core Kubernetes objects", func() {
			scheme := runtime.NewScheme()
			RegisterScheme(scheme)

			// Verify the scheme recognizes core Kubernetes types by checking if Pod type is known
			gvk := corev1.SchemeGroupVersion.WithKind("Pod")
			Expect(scheme.Recognizes(gvk)).To(BeTrue())
		})

		It("should not panic when called multiple times", func() {
			scheme := runtime.NewScheme()
			Expect(func() {
				RegisterScheme(scheme)
				RegisterScheme(scheme)
			}).NotTo(Panic())
		})
	})
})

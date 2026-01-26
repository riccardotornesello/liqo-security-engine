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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespaces Utilities", func() {
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

	Describe("ExtractClusterIDFromNamespace", func() {
		Context("when namespace has correct format", func() {
			It("should extract cluster ID successfully", func() {
				namespace := "liqo-tenant-test-cluster-123"
				expected := "test-cluster-123"
				result, err := ExtractClusterIDFromNamespace(namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expected))
			})

			It("should handle cluster IDs with hyphens", func() {
				namespace := "liqo-tenant-my-cluster-id"
				expected := "my-cluster-id"
				result, err := ExtractClusterIDFromNamespace(namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expected))
			})

			It("should handle short cluster IDs", func() {
				namespace := "liqo-tenant-abc"
				expected := "abc"
				result, err := ExtractClusterIDFromNamespace(namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expected))
			})
		})

		Context("when namespace has incorrect format", func() {
			It("should return error for namespace without prefix", func() {
				namespace := "default"
				_, err := ExtractClusterIDFromNamespace(namespace)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("does not have the expected prefix"))
			})

			It("should return error for namespace with wrong prefix", func() {
				namespace := "kube-system-test"
				_, err := ExtractClusterIDFromNamespace(namespace)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for partial prefix match", func() {
				namespace := "liqo-tenant"
				_, err := ExtractClusterIDFromNamespace(namespace)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for namespace with only prefix and hyphen", func() {
				namespace := "liqo-tenant-"
				_, err := ExtractClusterIDFromNamespace(namespace)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("does not have the expected prefix"))
			})

			It("should return error for empty namespace", func() {
				namespace := ""
				_, err := ExtractClusterIDFromNamespace(namespace)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

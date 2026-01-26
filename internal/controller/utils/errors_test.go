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
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
)

var _ = Describe("Errors Utilities", func() {
	var (
		ctx      context.Context
		scheme   *runtime.Scheme
		logger   logr.Logger
		recorder *record.FakeRecorder
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		RegisterScheme(scheme)
		logger = logr.Discard()
		recorder = record.NewFakeRecorder(10)
	})

	Describe("HandleReconcileError", func() {
		It("should update status condition to False with error message", func() {
			cfg := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-peering",
					Namespace:  "default",
					Generation: 1,
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg).WithStatusSubresource(cfg).Build()

			testErr := errors.New("test error")
			err := HandleReconcileError(ctx, cl, logger, recorder, cfg, testErr, "Test failed", "TestEventReason", "TestConditionReason")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Test failed"))
			Expect(err.Error()).To(ContainSubstring("test error"))

			// Verify the status was updated
			var updatedCfg securityv1.PeeringConnectivity
			Expect(cl.Get(ctx, client.ObjectKeyFromObject(cfg), &updatedCfg)).To(Succeed())
			Expect(updatedCfg.Status.ObservedGeneration).To(Equal(int64(1)))

			// Verify condition was set
			conditions := updatedCfg.Status.Conditions
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0].Type).To(Equal(ConditionTypeReady))
			Expect(conditions[0].Status).To(Equal(metav1.ConditionFalse))
			Expect(conditions[0].Reason).To(Equal("TestConditionReason"))
			Expect(conditions[0].Message).To(ContainSubstring("Test failed"))
		})

		It("should record an event", func() {
			cfg := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-peering",
					Namespace:  "default",
					Generation: 1,
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg).WithStatusSubresource(cfg).Build()

			testErr := errors.New("test error")
			_ = HandleReconcileError(ctx, cl, logger, recorder, cfg, testErr, "Test failed", "TestEventReason", "TestConditionReason")

			// Verify event was recorded
			Eventually(recorder.Events).Should(Receive(ContainSubstring("Test failed")))
		})

		It("should wrap the original error", func() {
			cfg := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-peering",
					Namespace:  "default",
					Generation: 1,
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg).WithStatusSubresource(cfg).Build()

			originalErr := errors.New("original error")
			err := HandleReconcileError(ctx, cl, logger, recorder, cfg, originalErr, "Operation failed", "EventReason", "ConditionReason")

			Expect(err).To(HaveOccurred())
			Expect(errors.Unwrap(err).Error()).To(Equal("original error"))
		})

		It("should set the correct observed generation", func() {
			cfg := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-peering",
					Namespace:  "default",
					Generation: 5,
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg).WithStatusSubresource(cfg).Build()

			testErr := errors.New("test error")
			_ = HandleReconcileError(ctx, cl, logger, recorder, cfg, testErr, "Test failed", "EventReason", "ConditionReason")

			var updatedCfg securityv1.PeeringConnectivity
			Expect(cl.Get(ctx, client.ObjectKeyFromObject(cfg), &updatedCfg)).To(Succeed())
			Expect(updatedCfg.Status.ObservedGeneration).To(Equal(int64(5)))
		})
	})
})

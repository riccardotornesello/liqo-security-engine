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
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
)

const (
	// ConditionTypeReady indicates whether the PeeringConnectivity resource is ready.
	// A resource is considered ready when its FirewallConfiguration has been successfully
	// created and synced.
	ConditionTypeReady = "Ready"
)

// HandleReconcileError handles reconciliation errors by logging, recording events,
// and updating the status condition of the PeeringConnectivity resource.
// It sets the Ready condition to False with the provided reason and message,
// and returns a formatted error.
func HandleReconcileError(
	ctx context.Context,
	cl client.Client,
	logger logr.Logger,
	recorder record.EventRecorder,
	cfg *securityv1.PeeringConnectivity,
	err error,
	message string,
	eventReason string,
	conditionReason string,
) error {
	// Log
	logger.Error(err, message)

	// Event
	recorder.Eventf(cfg, corev1.EventTypeWarning, eventReason, "%s: %v", message, err)

	// Status condition
	cfg.Status.ObservedGeneration = cfg.Generation

	meta.SetStatusCondition(&cfg.Status.Conditions, metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  conditionReason,
		Message: fmt.Sprintf("%s: %v", message, err),
	})

	if updateErr := cl.Status().Update(ctx, cfg); updateErr != nil {
		logger.Error(updateErr, "failed to update status")
		return updateErr
	}

	// Return the original error
	return fmt.Errorf("%s: %w", message, err)
}

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

package controller

import (
	"context"
	"fmt"

	ipamv1alpha1 "github.com/liqotech/liqo/apis/ipam/v1alpha1"
	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	offloadingv1beta1 "github.com/liqotech/liqo/apis/offloading/v1beta1"
	"github.com/liqotech/liqo/pkg/consts"
	vkforge "github.com/liqotech/liqo/pkg/virtualKubelet/forge"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
	"github.com/riccardotornesello/liqo-security-manager/internal/controller/forge"
	"github.com/riccardotornesello/liqo-security-manager/internal/controller/utils"
)

// PeeringSecurityReconciler reconciles a PeeringConnectivity object
type PeeringSecurityReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	// Condition Types
	ConditionTypeReady = "Ready"

	// Reasons
	ReasonClusterIDError   = "ClusterIDExtractionFailed"
	ReasonFabricSyncFailed = "FabricSyncFailed"
	ReasonFabricSynced     = "FabricSynced"

	// Event Types (Normal vs Warning is managed by k8s, here we define the reasons for events)
	EventReasonReconcileError = "ReconcileError"
	EventReasonSynced         = "Synced"
)

// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringsecurities,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringsecurities/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringsecurities/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PeeringSecurityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO: make sure the cluster exists
	// TODO: handle the case of multiple PeeringConnectivity in the same cluster

	logger := log.FromContext(ctx)

	// Retrieve the PeeringConnectivity resource
	cfg := &securityv1.PeeringConnectivity{}
	if err := r.Client.Get(ctx, req.NamespacedName, cfg); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("missing PeeringConnectivity resource, skipping reconciliation")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("unable to get the PeeringConnectivity %q: %w", req.NamespacedName, err)
	}

	logger.Info("reconciling PeeringConnectivity")

	// Extract Cluster ID from Namespace
	clusterID, err := utils.ExtractClusterID(req.Namespace)
	if err != nil {
		r.Recorder.Eventf(cfg, corev1.EventTypeWarning, EventReasonReconcileError, "Failed to extract cluster ID: %v", err)

		meta.SetStatusCondition(&cfg.Status.Conditions, metav1.Condition{
			Type:    ConditionTypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonClusterIDError,
			Message: fmt.Sprintf("Unable to extract cluster ID: %v", err),
		})
		if updateErr := r.Status().Update(ctx, cfg); updateErr != nil {
			logger.Error(updateErr, "failed to update status")
		}

		return ctrl.Result{}, fmt.Errorf("unable to extract the cluster ID from the namespace %q: %w", req.Namespace, err)
	}

	// Fabric Firewall Configuration Management
	fabricFwcfg := networkingv1beta1.FirewallConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.ForgeFabricResourceName(clusterID),
			Namespace: req.Namespace,
		},
	}

	fabricOp, err := controllerutil.CreateOrUpdate(ctx, r.Client, &fabricFwcfg, func() error {
		fabricFwcfg.SetLabels(forge.ForgeFabricLabels(clusterID))

		spec, err := forge.ForgeFabricSpec(ctx, r.Client, cfg, clusterID)
		if err != nil {
			return err
		}
		fabricFwcfg.Spec = *spec

		return controllerutil.SetOwnerReference(cfg, &fabricFwcfg, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "unable to reconcile the fabric firewall configuration")

		r.Recorder.Eventf(cfg, corev1.EventTypeWarning, EventReasonReconcileError, "Failed to reconcile fabric: %v", err)

		meta.SetStatusCondition(&cfg.Status.Conditions, metav1.Condition{
			Type:    ConditionTypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonFabricSyncFailed,
			Message: fmt.Sprintf("Failed to sync FirewallConfiguration: %v", err),
		})
		if updateErr := r.Status().Update(ctx, cfg); updateErr != nil {
			logger.Error(updateErr, "failed to update status during error handling")
		}

		return ctrl.Result{}, fmt.Errorf("unable to reconcile the fabric firewall configuration: %w", err)
	}

	logger.Info("reconciliation completed", "fabricOp", fabricOp)

	// Success and Final Status Update
	cfg.Status.ObservedGeneration = cfg.Generation

	meta.SetStatusCondition(&cfg.Status.Conditions, metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonFabricSynced,
		Message: "FirewallConfiguration successfully synced",
	})

	if err := r.Status().Update(ctx, cfg); err != nil {
		logger.Error(err, "failed to update PeeringConnectivity status")
		return ctrl.Result{}, err
	}

	if fabricOp != controllerutil.OperationResultNone {
		r.Recorder.Eventf(cfg, corev1.EventTypeNormal, EventReasonSynced, "FirewallConfiguration %s successfully", fabricOp)
	}

	return ctrl.Result{}, nil
}

// podEnqueuer enqueues the PeeringConnectivity reconciliation requests for Pods.
func (r *PeeringSecurityReconciler) podEnqueuer(ctx context.Context, obj client.Object) []ctrl.Request {
	logger := log.FromContext(ctx)

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		logger.Error(nil, "Expected a Pod object but got a different type", "type", fmt.Sprintf("%T", obj))
		return nil
	}

	labels := pod.GetLabels()

	localPodLabel, exists := labels[consts.LocalPodLabelKey]
	if exists && localPodLabel == consts.LocalPodLabelValue {
		// The Pod is a shadow Pod on the consumer cluster
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			return nil
		}

		logger.Info("Enqueuing Configuration for Offloaded Pod", "pod", pod.Name, "node", nodeName)
		return []ctrl.Request{{NamespacedName: types.NamespacedName{Name: nodeName, Namespace: utils.GetClusterNamespace(nodeName)}}}
	}

	originClusterLabel, exists := labels[vkforge.LiqoOriginClusterIDKey]
	if exists {
		// The Pod is offloaded to the provider cluster
		logger.Info("Enqueuing Configuration for Pod from Consumer", "pod", pod.Name, "originCluster", originClusterLabel)
		return []ctrl.Request{{NamespacedName: types.NamespacedName{Name: originClusterLabel, Namespace: utils.GetClusterNamespace(originClusterLabel)}}}
	}

	return nil
}

func (r *PeeringSecurityReconciler) networkEnqueuer(ctx context.Context, obj client.Object) []ctrl.Request {
	logger := log.FromContext(ctx)

	_, ok := obj.(*ipamv1alpha1.Network)
	if !ok {
		logger.Error(nil, "Expected a Network object but got a different type", "type", fmt.Sprintf("%T", obj))
		return nil
	}

	namespace := obj.GetNamespace()
	if namespace == "liqo" {
		// Ignore Liqo system namespace
		return nil
	}

	clusterId, err := utils.ExtractClusterID(namespace)
	if err != nil {
		logger.Error(err, "unable to extract cluster ID from Network namespace", "namespace", namespace)
		return nil
	}

	return []ctrl.Request{{NamespacedName: types.NamespacedName{Name: clusterId, Namespace: utils.GetClusterNamespace(clusterId)}}}
}

// Enqueuer that triggers reconciliation to all PeeringConnectivity resources
func (r *PeeringSecurityReconciler) allPeeringSecurityEnqueuer(ctx context.Context, _ client.Object) []ctrl.Request {
	logger := log.FromContext(ctx)

	peeringSecurityList := &securityv1.PeeringSecurityList{}
	if err := r.Client.List(ctx, peeringSecurityList); err != nil {
		logger.Error(err, "unable to list PeeringConnectivity resources for enqueuing all")
		return nil
	}

	var requests []ctrl.Request
	for _, ps := range peeringSecurityList.Items {
		requests = append(requests, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      ps.Name,
				Namespace: ps.Namespace,
			},
		})
	}

	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *PeeringSecurityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1.PeeringConnectivity{}).
		Owns(&networkingv1beta1.FirewallConfiguration{}).
		Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(r.podEnqueuer)).
		Watches(&ipamv1alpha1.Network{}, handler.EnqueueRequestsFromMapFunc(r.networkEnqueuer)).
		Watches(&offloadingv1beta1.NamespaceOffloading{}, handler.EnqueueRequestsFromMapFunc(r.allPeeringSecurityEnqueuer)).
		Named("peeringsecurity").
		Complete(r)
}

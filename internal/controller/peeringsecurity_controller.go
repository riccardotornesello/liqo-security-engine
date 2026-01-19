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

// Package controller implements the Kubernetes controller for PeeringConnectivity resources.
//
// This controller reconciles PeeringConnectivity custom resources and creates corresponding
// FirewallConfiguration resources in Liqo. It watches various Kubernetes resources (Pods,
// Networks, NamespaceOffloadings) to dynamically update firewall rules based on the current
// state of the cluster and peering configuration.
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

// PeeringConnectivityReconciler reconciles a PeeringConnectivity object.
// It manages the lifecycle of FirewallConfiguration resources that implement
// the security policies defined in PeeringConnectivity specs.
type PeeringConnectivityReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	// ConditionTypeReady indicates whether the PeeringConnectivity resource is ready.
	// A resource is considered ready when its FirewallConfiguration has been successfully
	// created and synced.
	ConditionTypeReady = "Ready"

	// ReasonClusterIDError indicates that the cluster ID could not be extracted from the namespace.
	ReasonClusterIDError = "ClusterIDExtractionFailed"
	// ReasonFabricSyncFailed indicates that the FirewallConfiguration failed to sync.
	ReasonFabricSyncFailed = "FabricSyncFailed"
	// ReasonFabricSynced indicates that the FirewallConfiguration was successfully synced.
	ReasonFabricSynced = "FabricSynced"

	// EventReasonReconcileError is emitted when a reconciliation error occurs.
	EventReasonReconcileError = "ReconcileError"
	// EventReasonSynced is emitted when the FirewallConfiguration is successfully synced.
	EventReasonSynced = "Synced"
)

// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringconnectivities,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringconnectivities/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringconnectivities/finalizers,verbs=update

// NewPeeringConnectivityReconciler creates a new PeeringConnectivityReconciler.
// It initializes the reconciler with the necessary client, scheme, and event recorder
// from the provided controller manager.
func NewPeeringConnectivityReconciler(mgr ctrl.Manager) *PeeringConnectivityReconciler {
	return &PeeringConnectivityReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("peeringconnectivity-controller"),
	}
}

// Reconcile is part of the main kubernetes reconciliation loop.
// It moves the current state of the cluster closer to the desired state by:
// 1. Reading the PeeringConnectivity resource
// 2. Extracting the cluster ID from the namespace
// 3. Creating or updating the corresponding FirewallConfiguration
// 4. Updating the status to reflect the current state
//
// Return values:
//   - (ctrl.Result{}, nil): Reconciliation succeeded, no requeue needed
//   - (ctrl.Result{}, err): Reconciliation failed, will be requeued automatically
func (r *PeeringConnectivityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	// Extract Cluster ID from Namespace.
	// The namespace should follow the pattern: liqo-tenant-<cluster-id>
	clusterID, err := utils.ExtractClusterID(req.Namespace)
	if err != nil {
		r.Recorder.Eventf(cfg, corev1.EventTypeWarning, EventReasonReconcileError, "Failed to extract cluster ID: %w", err)

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

	// Create or update the Fabric FirewallConfiguration.
	// The FirewallConfiguration is the Liqo resource that implements the actual
	// firewall rules at the network level.
	fabricFwcfg := networkingv1beta1.FirewallConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.ForgeFabricResourceName(clusterID),
			Namespace: req.Namespace,
		},
	}

	fabricOp, err := controllerutil.CreateOrUpdate(ctx, r.Client, &fabricFwcfg, func() error {
		// Set labels that identify this FirewallConfiguration as a fabric-level
		// security configuration targeting all nodes.
		fabricFwcfg.SetLabels(forge.ForgeFabricLabels(clusterID))

		// Generate the FirewallConfiguration spec based on the PeeringConnectivity rules.
		spec, err := forge.ForgeFabricSpec(ctx, r.Client, cfg, clusterID)
		if err != nil {
			return err
		}
		fabricFwcfg.Spec = *spec

		// Set owner reference so the FirewallConfiguration is deleted when the
		// PeeringConnectivity is deleted.
		return controllerutil.SetOwnerReference(cfg, &fabricFwcfg, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "unable to reconcile the fabric firewall configuration")

		r.Recorder.Eventf(cfg, corev1.EventTypeWarning, EventReasonReconcileError, "Failed to reconcile fabric: %w", err)

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

	// Update status to reflect successful reconciliation.
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

	// Emit an event if the FirewallConfiguration was created or updated.
	if fabricOp != controllerutil.OperationResultNone {
		r.Recorder.Eventf(cfg, corev1.EventTypeNormal, EventReasonSynced, "FirewallConfiguration %s successfully", fabricOp)
	}

	return ctrl.Result{}, nil
}

// podEnqueuer enqueues PeeringConnectivity reconciliation requests based on Pod changes.
// This function is called when a Pod is created, updated, or deleted. It determines
// which PeeringConnectivity resource(s) should be reconciled based on the Pod's labels
// and characteristics.
//
// It handles two scenarios:
// 1. Shadow pods on the consumer cluster (identified by liqo.io/local-pod label)
// 2. Offloaded pods on the provider cluster (identified by liqo.io/origin-cluster-id label)
func (r *PeeringConnectivityReconciler) podEnqueuer(ctx context.Context, obj client.Object) []ctrl.Request {
	logger := log.FromContext(ctx)

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		logger.Error(nil, "Expected a Pod object but got a different type", "type", fmt.Sprintf("%T", obj))
		return nil
	}

	labels := pod.GetLabels()

	localPodLabel, exists := labels[consts.LocalPodLabelKey]
	if exists && localPodLabel == consts.LocalPodLabelValue {
		// The Pod is a shadow Pod on the consumer cluster.
		// Enqueue the PeeringConnectivity for the provider cluster (node name).
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			return nil
		}

		logger.Info("Enqueuing Configuration for Offloaded Pod", "pod", pod.Name, "node", nodeName)
		return []ctrl.Request{{NamespacedName: types.NamespacedName{Name: nodeName, Namespace: utils.GetClusterNamespace(nodeName)}}}
	}

	originClusterLabel, exists := labels[vkforge.LiqoOriginClusterIDKey]
	if exists {
		// The Pod is offloaded to the provider cluster.
		// Enqueue the PeeringConnectivity for the consumer cluster (origin cluster).
		logger.Info("Enqueuing Configuration for Pod from Consumer", "pod", pod.Name, "originCluster", originClusterLabel)
		return []ctrl.Request{{NamespacedName: types.NamespacedName{Name: originClusterLabel, Namespace: utils.GetClusterNamespace(originClusterLabel)}}}
	}

	return nil
}

// networkEnqueuer enqueues PeeringConnectivity reconciliation requests based on Network changes.
// This function is called when a Liqo Network resource is created, updated, or deleted.
// Network resources contain CIDR information that is used in firewall rules.
func (r *PeeringConnectivityReconciler) networkEnqueuer(ctx context.Context, obj client.Object) []ctrl.Request {
	logger := log.FromContext(ctx)

	_, ok := obj.(*ipamv1alpha1.Network)
	if !ok {
		logger.Error(nil, "Expected a Network object but got a different type", "type", fmt.Sprintf("%T", obj))
		return nil
	}

	namespace := obj.GetNamespace()
	if namespace == "liqo" {
		// Ignore Network resources in the Liqo system namespace.
		return nil
	}

	// Extract the cluster ID from the namespace and enqueue the corresponding PeeringConnectivity.
	clusterId, err := utils.ExtractClusterID(namespace)
	if err != nil {
		logger.Error(err, "unable to extract cluster ID from Network namespace", "namespace", namespace)
		return nil
	}

	return []ctrl.Request{{NamespacedName: types.NamespacedName{Name: clusterId, Namespace: utils.GetClusterNamespace(clusterId)}}}
}

// allPeeringConnectivityEnqueuer enqueues reconciliation for all PeeringConnectivity resources.
// This function is called when a NamespaceOffloading resource changes, which may affect
// multiple PeeringConnectivity resources. It lists all PeeringConnectivity resources and
// enqueues them for reconciliation.
func (r *PeeringConnectivityReconciler) allPeeringConnectivityEnqueuer(ctx context.Context, _ client.Object) []ctrl.Request {
	logger := log.FromContext(ctx)

	peeringConnectivityList := &securityv1.PeeringConnectivityList{}
	if err := r.Client.List(ctx, peeringConnectivityList); err != nil {
		logger.Error(err, "unable to list PeeringConnectivity resources for enqueuing all")
		return nil
	}

	var requests []ctrl.Request
	for _, ps := range peeringConnectivityList.Items {
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
// It configures the controller to:
// - Reconcile PeeringConnectivity resources
// - Own FirewallConfiguration resources (so they're deleted when the PC is deleted)
// - Watch Pods, Networks, and NamespaceOffloadings to trigger reconciliation when they change
func (r *PeeringConnectivityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1.PeeringConnectivity{}).
		Owns(&networkingv1beta1.FirewallConfiguration{}).
		Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(r.podEnqueuer)).
		Watches(&ipamv1alpha1.Network{}, handler.EnqueueRequestsFromMapFunc(r.networkEnqueuer)).
		Watches(&offloadingv1beta1.NamespaceOffloading{}, handler.EnqueueRequestsFromMapFunc(r.allPeeringConnectivityEnqueuer)).
		Named("peeringconnectivity").
		Complete(r)
}

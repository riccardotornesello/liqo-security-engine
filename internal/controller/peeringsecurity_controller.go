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

	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
	"github.com/riccardotornesello/liqo-security-manager/internal/controller/utils"
)

// PeeringSecurityReconciler reconciles a PeeringSecurity object
type PeeringSecurityReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringsecurities,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringsecurities/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringsecurities/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PeeringSecurity object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *PeeringSecurityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO: make sure the cluster exists
	// TODO: handle duplicates

	cfg := &securityv1.PeeringSecurity{}
	if err := r.Client.Get(ctx, req.NamespacedName, cfg); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("missing configuration")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("unable to get the configuration %q: %w", req.NamespacedName, err)
	}
	logger.Info("reconciling configuration")

	clusterID, err := utils.ExtractClusterID(req.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to extract the cluster ID from the namespace %q: %w", req.Namespace, err)
	}

	gatewayFwcfg := networkingv1beta1.FirewallConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.ForgeGatewayResourceName(clusterID),
			Namespace: req.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, &gatewayFwcfg, func() error {
		gatewayFwcfg.SetLabels(utils.ForgeGatewayLabels(clusterID))

		spec, err := utils.ForgeGatewaySpec(ctx, r.Client, cfg, clusterID)
		if err != nil {
			return err
		}
		gatewayFwcfg.Spec = *spec

		return controllerutil.SetOwnerReference(cfg, &gatewayFwcfg, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to reconcile the gateway firewall configuration: %w", err)
	}

	logger.Info("gateway firewall configuration reconciled", "operation", op)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PeeringSecurityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1.PeeringSecurity{}).
		Named("peeringsecurity").
		Complete(r)
}

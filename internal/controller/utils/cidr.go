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

	ipamv1alpha1 "github.com/liqotech/liqo/apis/ipam/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// localPodCIDRNetworkName is the name of the Network resource that contains
	// the local cluster's pod CIDR information.
	localPodCIDRNetworkName = "pod-cidr"
	// localPodCIDRNetworkNamespace is the namespace where the local pod CIDR Network resource is stored.
	localPodCIDRNetworkNamespace = "liqo"
)

// GetCurrentClusterPodCIDR retrieves the pod CIDR for the current (local) cluster.
// It reads the Network resource in the liqo namespace to obtain the CIDR.
func GetCurrentClusterPodCIDR(ctx context.Context, cl client.Client) (string, error) {
	var network ipamv1alpha1.Network

	if err := cl.Get(ctx, client.ObjectKey{
		Namespace: localPodCIDRNetworkNamespace,
		Name:      localPodCIDRNetworkName,
	}, &network); err != nil {
		return "", err
	}

	return string(network.Spec.CIDR), nil
}

// GetRemoteClusterPodCIDR retrieves the pod CIDR for a remote peered cluster.
// It reads the Network resource in the tenant namespace for the specified cluster ID.
func GetRemoteClusterPodCIDR(ctx context.Context, cl client.Client, clusterID string) (string, error) {
	var network ipamv1alpha1.Network

	if err := cl.Get(ctx, client.ObjectKey{
		Namespace: GetClusterNamespace(clusterID),
		Name:      fmt.Sprintf("%s-pod", clusterID),
	}, &network); err != nil {
		return "", err
	}

	return string(network.Status.CIDR), nil
}

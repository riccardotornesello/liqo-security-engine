package utils

import (
	"context"
	"fmt"

	ipamv1alpha1 "github.com/liqotech/liqo/apis/ipam/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	localPodCIDRNetworkName      = "pod-cidr"
	localPodCIDRNetworkNamespace = "liqo"
)

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

func GetRemoteClusterPodCIDR(ctx context.Context, cl client.Client, clusterID string) (string, error) {
	var network ipamv1alpha1.Network

	if err := cl.Get(ctx, client.ObjectKey{
		Namespace: GetClusterNamespace(clusterID),
		Name:      fmt.Sprintf("%s-pod", clusterID),
	}, &network); err != nil {
		return "", err
	}

	return string(network.Spec.CIDR), nil
}

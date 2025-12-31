package utils

import (
	"fmt"

	tenantnamespace "github.com/liqotech/liqo/pkg/tenantNamespace"
)

func GetClusterNamespace(clusterID string) string {
	return fmt.Sprintf("%s-%s", tenantnamespace.NamePrefix, clusterID)
}

// ExtractClusterID extracts the cluster ID from the given namespace.
func ExtractClusterID(namespace string) (string, error) {
	// Remove the tenantnamespace.NamePrefix + "-" from the namespace to get the cluster ID
	const prefix = tenantnamespace.NamePrefix + "-"

	if len(namespace) <= len(prefix) || namespace[:len(prefix)] != prefix {
		return "", fmt.Errorf("namespace %q does not have the expected prefix %q", namespace, prefix)
	}
	return namespace[len(prefix):], nil
}

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
	"fmt"

	tenantnamespace "github.com/liqotech/liqo/pkg/tenantNamespace"
)

// GetClusterNamespace returns the Liqo tenant namespace for a given cluster ID.
// Liqo uses namespaces with the format "liqo-tenant-<cluster-id>" to isolate
// resources for each peered cluster.
func GetClusterNamespace(clusterID string) string {
	return fmt.Sprintf("%s-%s", tenantnamespace.NamePrefix, clusterID)
}

// ExtractClusterID extracts the cluster ID from a Liqo tenant namespace name.
// It removes the "liqo-tenant-" prefix to obtain the cluster ID.
// Returns an error if the namespace doesn't follow the expected format.
func ExtractClusterID(namespace string) (string, error) {
	const prefix = tenantnamespace.NamePrefix + "-"

	if len(namespace) <= len(prefix) || namespace[:len(prefix)] != prefix {
		return "", fmt.Errorf("namespace %q does not have the expected prefix %q", namespace, prefix)
	}
	return namespace[len(prefix):], nil
}

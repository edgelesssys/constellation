/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import "fmt"

// AutoscalingNodeGroup converts an azure scale set into a node group used by the k8s cluster-autoscaler.
func AutoscalingNodeGroup(scaleSet string, min int, max int) string {
	return fmt.Sprintf("%d:%d:%s", min, max, scaleSet)
}

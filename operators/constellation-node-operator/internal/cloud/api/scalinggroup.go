/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package api

import updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"

// ScalingGroup is a cloud provider scaling group.
type ScalingGroup struct {
	// Name is the csp specific name of the scaling group.
	Name string
	// NodeGroupName is the human friendly name of the node group
	// as defined in the Constellation configuration.
	NodeGroupName string
	// GroupID is the CSP specific, canonical identifier of a scaling group.
	GroupID string
	// AutoscalingGroupName is name that is expected by the autoscaler.
	AutoscalingGroupName string
	// Role is the role of the nodes in the scaling group.
	Role updatev1alpha1.NodeRole
}

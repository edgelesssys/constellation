/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
)

const (
	provisioningStateCreating  = "ProvisioningState/creating"
	provisioningStateUpdating  = "ProvisioningState/updating"
	provisioningStateMigrating = "ProvisioningState/migrating"
	provisioningStateFailed    = "ProvisioningState/failed"
	provisioningStateDeleting  = "ProvisioningState/deleting"
	powerStateStarting         = "PowerState/starting"
	powerStateStopping         = "PowerState/stopping"
	powerStateStopped          = "PowerState/stopped"
	powerStateDeallocating     = "PowerState/deallocating"
	powerStateDeallocated      = "PowerState/deallocated"
	powerStateRunning          = "PowerState/running"
)

// nodeStateFromStatuses returns the node state from instance view statuses.
// reference:
// - https://docs.microsoft.com/en-us/azure/virtual-machines/states-billing#provisioning-states
// - https://docs.microsoft.com/en-us/azure/virtual-machines/states-billing#power-states-and-billing
func nodeStateFromStatuses(statuses []*armcompute.InstanceViewStatus) updatev1alpha1.CSPNodeState {
	for _, status := range statuses {
		if status == nil || status.Code == nil {
			continue
		}
		switch *status.Code {
		case provisioningStateCreating:
			return updatev1alpha1.NodeStateCreating
		case provisioningStateUpdating:
			return updatev1alpha1.NodeStateStopped
		case provisioningStateMigrating:
			return updatev1alpha1.NodeStateStopped
		case provisioningStateFailed:
			return updatev1alpha1.NodeStateFailed
		case provisioningStateDeleting:
			return updatev1alpha1.NodeStateTerminating
		case powerStateStarting:
			return updatev1alpha1.NodeStateCreating
		case powerStateStopping:
			return updatev1alpha1.NodeStateStopped
		case powerStateStopped:
			return updatev1alpha1.NodeStateStopped
		case powerStateDeallocating:
			return updatev1alpha1.NodeStateStopped
		case powerStateDeallocated:
			return updatev1alpha1.NodeStateStopped
		case powerStateRunning:
			return updatev1alpha1.NodeStateReady
		}
	}
	return updatev1alpha1.NodeStateUnknown
}

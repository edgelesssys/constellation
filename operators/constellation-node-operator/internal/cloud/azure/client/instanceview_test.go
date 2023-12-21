/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/stretchr/testify/assert"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
)

// this state is included in most VMs but not needed
// to determine the node state as every provisioned VM also has a power state.
const provisioningStateSucceeded = "ProvisioningState/succeeded"

func TestNodeStateFromStatuses(t *testing.T) {
	testCases := map[string]struct {
		statuses  []*armcompute.InstanceViewStatus
		wantState updatev1alpha1.CSPNodeState
	}{
		"no statuses": {
			wantState: updatev1alpha1.NodeStateUnknown,
		},
		"invalid status": {
			statuses:  []*armcompute.InstanceViewStatus{nil, {Code: nil}},
			wantState: updatev1alpha1.NodeStateUnknown,
		},
		"provisioning creating": {
			statuses:  []*armcompute.InstanceViewStatus{{Code: to.Ptr(provisioningStateCreating)}},
			wantState: updatev1alpha1.NodeStateCreating,
		},
		"provisioning updating": {
			statuses:  []*armcompute.InstanceViewStatus{{Code: to.Ptr(provisioningStateUpdating)}},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"provisioning migrating": {
			statuses:  []*armcompute.InstanceViewStatus{{Code: to.Ptr(provisioningStateMigrating)}},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"provisioning failed": {
			statuses:  []*armcompute.InstanceViewStatus{{Code: to.Ptr(provisioningStateFailed)}},
			wantState: updatev1alpha1.NodeStateFailed,
		},
		"provisioning deleting": {
			statuses:  []*armcompute.InstanceViewStatus{{Code: to.Ptr(provisioningStateDeleting)}},
			wantState: updatev1alpha1.NodeStateTerminating,
		},
		"provisioning succeeded (but no power state)": {
			statuses:  []*armcompute.InstanceViewStatus{{Code: to.Ptr(provisioningStateSucceeded)}},
			wantState: updatev1alpha1.NodeStateUnknown,
		},
		"power starting": {
			statuses: []*armcompute.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateStarting)},
			},
			wantState: updatev1alpha1.NodeStateCreating,
		},
		"power stopping": {
			statuses: []*armcompute.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateStopping)},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"power stopped": {
			statuses: []*armcompute.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateStopped)},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"power deallocating": {
			statuses: []*armcompute.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateDeallocating)},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"power deallocated": {
			statuses: []*armcompute.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateDeallocated)},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"power running": {
			statuses: []*armcompute.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateRunning)},
			},
			wantState: updatev1alpha1.NodeStateReady,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			gotState := nodeStateFromStatuses(tc.statuses)
			assert.Equal(tc.wantState, gotState)
		})
	}
}

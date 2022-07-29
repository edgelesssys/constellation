package client

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/stretchr/testify/assert"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
)

// this state is included in most VMs but not needed
// to determine the node state as every provisioned VM also has a power state.
const provisioningStateSucceeded = "ProvisioningState/succeeded"

func TestNodeStateFromStatuses(t *testing.T) {
	testCases := map[string]struct {
		statuses  []*armcomputev2.InstanceViewStatus
		wantState updatev1alpha1.CSPNodeState
	}{
		"no statuses": {
			wantState: updatev1alpha1.NodeStateUnknown,
		},
		"invalid status": {
			statuses:  []*armcomputev2.InstanceViewStatus{nil, {Code: nil}},
			wantState: updatev1alpha1.NodeStateUnknown,
		},
		"provisioning creating": {
			statuses:  []*armcomputev2.InstanceViewStatus{{Code: to.Ptr(provisioningStateCreating)}},
			wantState: updatev1alpha1.NodeStateCreating,
		},
		"provisioning updating": {
			statuses:  []*armcomputev2.InstanceViewStatus{{Code: to.Ptr(provisioningStateUpdating)}},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"provisioning migrating": {
			statuses:  []*armcomputev2.InstanceViewStatus{{Code: to.Ptr(provisioningStateMigrating)}},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"provisioning failed": {
			statuses:  []*armcomputev2.InstanceViewStatus{{Code: to.Ptr(provisioningStateFailed)}},
			wantState: updatev1alpha1.NodeStateFailed,
		},
		"provisioning deleting": {
			statuses:  []*armcomputev2.InstanceViewStatus{{Code: to.Ptr(provisioningStateDeleting)}},
			wantState: updatev1alpha1.NodeStateTerminating,
		},
		"provisioning succeeded (but no power state)": {
			statuses:  []*armcomputev2.InstanceViewStatus{{Code: to.Ptr(provisioningStateSucceeded)}},
			wantState: updatev1alpha1.NodeStateUnknown,
		},
		"power starting": {
			statuses: []*armcomputev2.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateStarting)},
			},
			wantState: updatev1alpha1.NodeStateCreating,
		},
		"power stopping": {
			statuses: []*armcomputev2.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateStopping)},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"power stopped": {
			statuses: []*armcomputev2.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateStopped)},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"power deallocating": {
			statuses: []*armcomputev2.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateDeallocating)},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"power deallocated": {
			statuses: []*armcomputev2.InstanceViewStatus{
				{Code: to.Ptr(provisioningStateSucceeded)},
				{Code: to.Ptr(powerStateDeallocated)},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"power running": {
			statuses: []*armcomputev2.InstanceViewStatus{
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

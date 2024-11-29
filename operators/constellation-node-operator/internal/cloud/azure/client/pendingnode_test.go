/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodeState(t *testing.T) {
	testCases := map[string]struct {
		providerID         string
		instanceView       armcompute.VirtualMachineScaleSetVMInstanceView
		getInstanceViewErr error
		wantState          updatev1alpha1.CSPNodeState
		wantErr            bool
	}{
		"getting node state works": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			instanceView: armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{Code: to.Ptr("ProvisioningState/succeeded")},
					{Code: to.Ptr("PowerState/running")},
				},
			},
			wantState: updatev1alpha1.NodeStateReady,
		},
		"splitting providerID fails": {
			providerID: "invalid",
			wantErr:    true,
		},
		"get instance view fails": {
			providerID:         "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			getInstanceViewErr: errors.New("get instance view error"),
			wantErr:            true,
		},
		"get instance view returns 404": {
			providerID:         "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			getInstanceViewErr: &azcore.ResponseError{StatusCode: http.StatusNotFound},
			wantState:          updatev1alpha1.NodeStateTerminated,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				virtualMachineScaleSetVMsAPI: &stubvirtualMachineScaleSetVMsAPI{
					instanceView: armcompute.VirtualMachineScaleSetVMsClientGetInstanceViewResponse{
						VirtualMachineScaleSetVMInstanceView: tc.instanceView,
					},
					instanceViewErr: tc.getInstanceViewErr,
				},
			}
			gotState, err := client.GetNodeState(context.Background(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantState, gotState)
		})
	}
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestTerminator(t *testing.T) {
	someAzureState := func() state.ConstellationState {
		return state.ConstellationState{
			CloudProvider: cloudprovider.Azure.String(),
			AzureWorkerInstances: cloudtypes.Instances{
				"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
				"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			},
			AzureControlPlaneInstances: cloudtypes.Instances{
				"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			},
			AzureADAppObjectID: "00000000-0000-0000-0000-000000000001",
		}
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		tfClient          terraformClient
		newTfClientErr    error
		azureclient       azureclient
		newAzureClientErr error
		state             state.ConstellationState
		wantErr           bool
	}{
		"gcp": {
			tfClient: &stubTerraformClient{},
			state:    state.ConstellationState{CloudProvider: cloudprovider.GCP.String()},
		},
		"gcp newTfClientErr": {
			newTfClientErr: someErr,
			state:          state.ConstellationState{CloudProvider: cloudprovider.GCP.String()},
			wantErr:        true,
		},
		"gcp destroy cluster error": {
			tfClient: &stubTerraformClient{destroyClusterErr: someErr},
			state:    state.ConstellationState{CloudProvider: cloudprovider.GCP.String()},
			wantErr:  true,
		},
		"gcp clean up workspace error": {
			tfClient: &stubTerraformClient{cleanUpWorkspaceErr: someErr},
			state:    state.ConstellationState{CloudProvider: cloudprovider.GCP.String()},
			wantErr:  true,
		},
		"qemu": {
			tfClient: &stubTerraformClient{},
			state:    state.ConstellationState{CloudProvider: cloudprovider.QEMU.String()},
		},
		"qemu destroy cluster error": {
			tfClient: &stubTerraformClient{destroyClusterErr: someErr},
			state:    state.ConstellationState{CloudProvider: cloudprovider.QEMU.String()},
			wantErr:  true,
		},
		"qemu clean up workspace error": {
			tfClient: &stubTerraformClient{cleanUpWorkspaceErr: someErr},
			state:    state.ConstellationState{CloudProvider: cloudprovider.QEMU.String()},
			wantErr:  true,
		},
		"azure": {
			azureclient: &stubAzureClient{},
			state:       someAzureState(),
		},
		"azure newAzureClient error": {
			newAzureClientErr: someErr,
			state:             someAzureState(),
			wantErr:           true,
		},
		"azure terminateResourceGroupResources error": {
			azureclient: &stubAzureClient{terminateResourceGroupResourcesErr: someErr},
			state:       someAzureState(),
			wantErr:     true,
		},
		"unknown cloud provider": {
			state:   state.ConstellationState{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			terminator := &Terminator{
				newTerraformClient: func(ctx context.Context) (terraformClient, error) {
					return tc.tfClient, tc.newTfClientErr
				},
				newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
					return tc.azureclient, tc.newAzureClientErr
				},
			}

			err := terminator.Terminate(context.Background(), tc.state)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				switch cloudprovider.FromString(tc.state.CloudProvider) {
				case cloudprovider.GCP:
					fallthrough
				case cloudprovider.QEMU:
					cl := tc.tfClient.(*stubTerraformClient)
					assert.True(cl.destroyClusterCalled)
					assert.True(cl.removeInstallerCalled)
				case cloudprovider.Azure:
					cl := tc.azureclient.(*stubAzureClient)
					assert.True(cl.terminateResourceGroupResourcesCalled)
				}
			}
		})
	}
}

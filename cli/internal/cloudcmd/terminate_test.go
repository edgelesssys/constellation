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
		libvirt           *stubLibvirtRunner
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
			libvirt:  &stubLibvirtRunner{},
			state:    state.ConstellationState{CloudProvider: cloudprovider.QEMU.String()},
		},
		"qemu destroy cluster error": {
			tfClient: &stubTerraformClient{destroyClusterErr: someErr},
			libvirt:  &stubLibvirtRunner{},
			state:    state.ConstellationState{CloudProvider: cloudprovider.QEMU.String()},
			wantErr:  true,
		},
		"qemu clean up workspace error": {
			tfClient: &stubTerraformClient{cleanUpWorkspaceErr: someErr},
			libvirt:  &stubLibvirtRunner{},
			state:    state.ConstellationState{CloudProvider: cloudprovider.QEMU.String()},
			wantErr:  true,
		},
		"qemu stop libvirt error": {
			tfClient: &stubTerraformClient{},
			libvirt:  &stubLibvirtRunner{stopErr: someErr},
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
				newTerraformClient: func(ctx context.Context, provider cloudprovider.Provider) (terraformClient, error) {
					return tc.tfClient, tc.newTfClientErr
				},
				newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
					return tc.azureclient, tc.newAzureClientErr
				},
				newLibvirtRunner: func() libvirtRunner {
					return tc.libvirt
				},
			}

			err := terminator.Terminate(context.Background(), tc.state)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				switch cloudprovider.FromString(tc.state.CloudProvider) {
				case cloudprovider.QEMU:
					assert.True(tc.libvirt.stopCalled)
					fallthrough
				case cloudprovider.GCP:
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

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
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestTerminator(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		tfClient       terraformClient
		newTfClientErr error
		libvirt        *stubLibvirtRunner
		state          state.ConstellationState
		wantErr        bool
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
				newLibvirtRunner: func() libvirtRunner {
					return tc.libvirt
				},
			}

			err := terminator.Terminate(context.Background(), tc.state)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				cl := tc.tfClient.(*stubTerraformClient)
				assert.True(cl.destroyClusterCalled)
				assert.True(cl.removeInstallerCalled)
				if cloudprovider.FromString(tc.state.CloudProvider) == cloudprovider.QEMU {
					assert.True(tc.libvirt.stopCalled)
				}
			}
		})
	}
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/stretchr/testify/assert"
)

func TestRollbackTerraform(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		tfClient       *stubTerraformClient
		wantCleanupErr bool
		wantDestroyErr bool
	}{
		"success": {
			tfClient: &stubTerraformClient{},
		},
		"destroy cluster error": {
			tfClient:       &stubTerraformClient{destroyErr: someErr},
			wantDestroyErr: true,
		},
		"clean up workspace error": {
			tfClient:       &stubTerraformClient{cleanUpWorkspaceErr: someErr},
			wantCleanupErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			rollbacker := &rollbackerTerraform{
				client: tc.tfClient,
			}

			destroyClusterErrOutput := &bytes.Buffer{}
			err := rollbacker.rollback(context.Background(), destroyClusterErrOutput, terraform.LogLevelNone)
			if tc.wantCleanupErr {
				assert.Error(err)
				if tc.tfClient.cleanUpWorkspaceErr == nil {
					assert.False(tc.tfClient.cleanUpWorkspaceCalled)
				}
				return
			}
			if tc.wantDestroyErr {
				assert.Error(err)
				assert.Equal("Could not destroy the resources. Please delete the \"constellation-terraform\" directory manually if no resources were created\n", destroyClusterErrOutput.String())
				return
			}
			assert.NoError(err)
			assert.True(tc.tfClient.destroyCalled)
			assert.True(tc.tfClient.cleanUpWorkspaceCalled)
		})
	}
}

func TestRollbackQEMU(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		libvirt        *stubLibvirtRunner
		tfClient       *stubTerraformClient
		wantDestroyErr bool
		wantErr        bool
	}{
		"success": {
			libvirt:  &stubLibvirtRunner{},
			tfClient: &stubTerraformClient{},
		},
		"stop libvirt error": {
			libvirt:  &stubLibvirtRunner{stopErr: someErr},
			tfClient: &stubTerraformClient{},
			wantErr:  true,
		},
		"destroy cluster error": {
			libvirt:        &stubLibvirtRunner{stopErr: someErr},
			tfClient:       &stubTerraformClient{destroyErr: someErr},
			wantDestroyErr: true,
		},
		"clean up workspace error": {
			libvirt:  &stubLibvirtRunner{},
			tfClient: &stubTerraformClient{cleanUpWorkspaceErr: someErr},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			rollbacker := &rollbackerQEMU{
				libvirt: tc.libvirt,
				client:  tc.tfClient,
			}

			destroyClusterErrOutput := &bytes.Buffer{}

			err := rollbacker.rollback(context.Background(), destroyClusterErrOutput, terraform.LogLevelNone)
			if tc.wantErr {
				assert.Error(err)
				if tc.tfClient.cleanUpWorkspaceErr == nil {
					assert.False(tc.tfClient.cleanUpWorkspaceCalled)
				}
				return
			}
			if tc.wantDestroyErr {
				assert.Error(err)
				assert.Equal("Could not destroy the resources. Please delete the \"constellation-terraform\" directory manually if no resources were created\n", destroyClusterErrOutput.String())
				return
			}
			assert.NoError(err)
			assert.True(tc.libvirt.stopCalled)
			assert.True(tc.tfClient.destroyCalled)
			assert.True(tc.tfClient.cleanUpWorkspaceCalled)
		})
	}
}

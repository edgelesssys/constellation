/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/stretchr/testify/assert"
)

func TestRollbackTerraform(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		tfClient *stubTerraformClient
		wantErr  bool
	}{
		"success": {
			tfClient: &stubTerraformClient{},
		},
		"destroy cluster error": {
			tfClient: &stubTerraformClient{destroyErr: someErr},
			wantErr:  true,
		},
		"clean up workspace error": {
			tfClient: &stubTerraformClient{cleanUpWorkspaceErr: someErr},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			rollbacker := &rollbackerTerraform{
				client: tc.tfClient,
			}

			err := rollbacker.rollback(context.Background(), terraform.LogLevelNone)
			if tc.wantErr {
				assert.Error(err)
				if tc.tfClient.cleanUpWorkspaceErr == nil {
					assert.False(tc.tfClient.cleanUpWorkspaceCalled)
				}
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
		libvirt          *stubLibvirtRunner
		tfClient         *stubTerraformClient
		createdWorkspace bool
		wantErr          bool
	}{
		"success": {
			libvirt:          &stubLibvirtRunner{},
			tfClient:         &stubTerraformClient{},
			createdWorkspace: true,
		},
		"stop libvirt error": {
			libvirt:  &stubLibvirtRunner{stopErr: someErr},
			tfClient: &stubTerraformClient{},
			wantErr:  true,
		},
		"destroy cluster error": {
			libvirt:  &stubLibvirtRunner{stopErr: someErr},
			tfClient: &stubTerraformClient{destroyErr: someErr},
			wantErr:  true,
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
				libvirt:          tc.libvirt,
				client:           tc.tfClient,
				createdWorkspace: tc.createdWorkspace,
			}

			err := rollbacker.rollback(context.Background(), terraform.LogLevelNone)
			if tc.wantErr {
				assert.Error(err)
				if tc.tfClient.cleanUpWorkspaceErr == nil {
					assert.False(tc.tfClient.cleanUpWorkspaceCalled)
				}
				return
			}
			assert.NoError(err)
			assert.True(tc.libvirt.stopCalled)
			if tc.createdWorkspace {
				assert.True(tc.tfClient.destroyCalled)
			} else {
				assert.False(tc.tfClient.destroyCalled)
			}
			assert.True(tc.tfClient.cleanUpWorkspaceCalled)
		})
	}
}

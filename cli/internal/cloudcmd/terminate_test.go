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

func TestTerminator(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		tfClient       tfDestroyer
		newTfClientErr error
		libvirt        *stubLibvirtRunner
		wantErr        bool
	}{
		"gcp": {
			libvirt:  &stubLibvirtRunner{},
			tfClient: &stubTerraformClient{},
		},
		"newTfClientErr": {
			libvirt:        &stubLibvirtRunner{},
			newTfClientErr: someErr,
			wantErr:        true,
		},
		"destroy cluster error": {
			tfClient: &stubTerraformClient{destroyErr: someErr},
			libvirt:  &stubLibvirtRunner{},
			wantErr:  true,
		},
		"clean up workspace error": {
			tfClient: &stubTerraformClient{cleanUpWorkspaceErr: someErr},
			libvirt:  &stubLibvirtRunner{},
			wantErr:  true,
		},
		"qemu stop libvirt error": {
			tfClient: &stubTerraformClient{},
			libvirt:  &stubLibvirtRunner{stopErr: someErr},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			terminator := &Terminator{
				newTerraformClient: func(_ context.Context, _ string) (tfDestroyer, error) {
					return tc.tfClient, tc.newTfClientErr
				},
				newLibvirtRunner: func() libvirtRunner {
					return tc.libvirt
				},
			}

			err := terminator.Terminate(context.Background(), "", terraform.LogLevelNone)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			cl := tc.tfClient.(*stubTerraformClient)
			assert.True(cl.destroyCalled)
			assert.True(cl.removeInstallerCalled)
			assert.True(tc.libvirt.stopCalled)
		})
	}
}

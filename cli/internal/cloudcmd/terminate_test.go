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
	"github.com/stretchr/testify/assert"
)

func TestTerminator(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		tfClient       terraformClient
		newTfClientErr error
		libvirt        *stubLibvirtRunner
		provider       cloudprovider.Provider
		wantErr        bool
	}{
		"gcp": {
			tfClient: &stubTerraformClient{},
			provider: cloudprovider.GCP,
		},
		"gcp newTfClientErr": {
			newTfClientErr: someErr,
			provider:       cloudprovider.GCP,
			wantErr:        true,
		},
		"gcp destroy cluster error": {
			tfClient: &stubTerraformClient{destroyClusterErr: someErr},
			provider: cloudprovider.GCP,
			wantErr:  true,
		},
		"gcp clean up workspace error": {
			tfClient: &stubTerraformClient{cleanUpWorkspaceErr: someErr},
			provider: cloudprovider.GCP,
			wantErr:  true,
		},
		"qemu": {
			tfClient: &stubTerraformClient{},
			libvirt:  &stubLibvirtRunner{},
			provider: cloudprovider.QEMU,
		},
		"qemu destroy cluster error": {
			tfClient: &stubTerraformClient{destroyClusterErr: someErr},
			libvirt:  &stubLibvirtRunner{},
			provider: cloudprovider.QEMU,
			wantErr:  true,
		},
		"qemu clean up workspace error": {
			tfClient: &stubTerraformClient{cleanUpWorkspaceErr: someErr},
			libvirt:  &stubLibvirtRunner{},
			provider: cloudprovider.QEMU,
			wantErr:  true,
		},
		"qemu stop libvirt error": {
			tfClient: &stubTerraformClient{},
			libvirt:  &stubLibvirtRunner{stopErr: someErr},
			provider: cloudprovider.QEMU,
			wantErr:  true,
		},
		"unknown cloud provider": {
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

			err := terminator.Terminate(context.Background(), tc.provider)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				cl := tc.tfClient.(*stubTerraformClient)
				assert.True(cl.destroyClusterCalled)
				assert.True(cl.removeInstallerCalled)
				if tc.provider == cloudprovider.QEMU {
					assert.True(tc.libvirt.stopCalled)
				}
			}
		})
	}
}

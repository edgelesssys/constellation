/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"bytes"
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCreator(t *testing.T) {
	failOnNonAMD64 := (runtime.GOARCH != "amd64") || (runtime.GOOS != "linux")
	ip := "192.0.2.1"
	someErr := errors.New("failed")

	testCases := map[string]struct {
		tfClient       terraformClient
		newTfClientErr error
		libvirt        *stubLibvirtRunner
		provider       cloudprovider.Provider
		config         *config.Config
		wantErr        bool
		wantRollback   bool // Use only together with stubClients.
	}{
		"gcp": {
			tfClient: &stubTerraformClient{ip: ip},
			provider: cloudprovider.GCP,
			config:   config.Default(),
		},
		"gcp newTerraformClient error": {
			newTfClientErr: someErr,
			provider:       cloudprovider.GCP,
			config:         config.Default(),
			wantErr:        true,
		},
		"gcp create cluster error": {
			tfClient:     &stubTerraformClient{createClusterErr: someErr},
			provider:     cloudprovider.GCP,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"qemu": {
			tfClient: &stubTerraformClient{ip: ip},
			libvirt:  &stubLibvirtRunner{},
			provider: cloudprovider.QEMU,
			config:   config.Default(),
			wantErr:  failOnNonAMD64,
		},
		"qemu newTerraformClient error": {
			newTfClientErr: someErr,
			libvirt:        &stubLibvirtRunner{},
			provider:       cloudprovider.QEMU,
			config:         config.Default(),
			wantErr:        true,
		},
		"qemu create cluster error": {
			tfClient:     &stubTerraformClient{createClusterErr: someErr},
			libvirt:      &stubLibvirtRunner{},
			provider:     cloudprovider.QEMU,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: !failOnNonAMD64, // if we run on non-AMD64/linux, we don't get to a point where rollback is needed
		},
		"qemu start libvirt error": {
			tfClient:     &stubTerraformClient{ip: ip},
			libvirt:      &stubLibvirtRunner{startErr: someErr},
			provider:     cloudprovider.QEMU,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: !failOnNonAMD64,
		},
		"unknown provider": {
			provider: cloudprovider.Unknown,
			config:   config.Default(),
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			creator := &Creator{
				out: &bytes.Buffer{},
				newTerraformClient: func(ctx context.Context) (terraformClient, error) {
					return tc.tfClient, tc.newTfClientErr
				},
				newLibvirtRunner: func() libvirtRunner {
					return tc.libvirt
				},
			}

			idFile, err := creator.Create(context.Background(), tc.provider, tc.config, "name", "type", 2, 3)

			if tc.wantErr {
				assert.Error(err)
				if tc.wantRollback {
					cl := tc.tfClient.(*stubTerraformClient)
					assert.True(cl.destroyClusterCalled)
					assert.True(cl.cleanUpWorkspaceCalled)
					if tc.provider == cloudprovider.QEMU {
						assert.True(tc.libvirt.stopCalled)
					}
				}
			} else {
				assert.NoError(err)
				assert.Equal(tc.provider, idFile.CloudProvider)
				assert.Equal(ip, idFile.IP)
			}
		})
	}
}

func TestNormalizeAzureURIs(t *testing.T) {
	testCases := map[string]struct {
		in   terraform.AzureVariables
		want terraform.AzureVariables
	}{
		"empty": {
			in:   terraform.AzureVariables{},
			want: terraform.AzureVariables{},
		},
		"no change": {
			in: terraform.AzureVariables{
				ImageID: "/communityGalleries/foo/images/constellation/versions/2.1.0",
			},
			want: terraform.AzureVariables{
				ImageID: "/communityGalleries/foo/images/constellation/versions/2.1.0",
			},
		},
		"fix image id": {
			in: terraform.AzureVariables{
				ImageID: "/CommunityGalleries/foo/Images/constellation/Versions/2.1.0",
			},
			want: terraform.AzureVariables{
				ImageID: "/communityGalleries/foo/images/constellation/versions/2.1.0",
			},
		},
		"fix resource group": {
			in: terraform.AzureVariables{
				UserAssignedIdentity: "/subscriptions/foo/resourcegroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai",
			},
			want: terraform.AzureVariables{
				UserAssignedIdentity: "/subscriptions/foo/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			out := normalizeAzureURIs(tc.in)
			assert.Equal(tc.want, out)
		})
	}
}

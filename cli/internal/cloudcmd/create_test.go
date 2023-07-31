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

	"github.com/stretchr/testify/assert"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
)

func TestCreator(t *testing.T) {
	// TODO(malt3): remove once OpenStack is fully supported.
	t.Setenv("CONSTELLATION_OPENSTACK_DEV", "1")
	failOnNonAMD64 := (runtime.GOARCH != "amd64") || (runtime.GOOS != "linux")
	ip := "192.0.2.1"
	someErr := errors.New("failed")

	testCases := map[string]struct {
		tfClient              tfResourceClient
		newTfClientErr        error
		libvirt               *stubLibvirtRunner
		provider              cloudprovider.Provider
		config                *config.Config
		policyPatcher         *stubPolicyPatcher
		wantErr               bool
		wantRollback          bool // Use only together with stubClients.
		wantTerraformRollback bool // When libvirt fails, don't call into Terraform.
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
			tfClient:              &stubTerraformClient{createClusterErr: someErr},
			provider:              cloudprovider.GCP,
			config:                config.Default(),
			wantErr:               true,
			wantRollback:          true,
			wantTerraformRollback: true,
		},
		"azure": {
			tfClient: &stubTerraformClient{ip: ip},
			provider: cloudprovider.Azure,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				return cfg
			}(),
			policyPatcher: &stubPolicyPatcher{},
		},
		"azure trusted launch": {
			tfClient: &stubTerraformClient{ip: ip},
			provider: cloudprovider.Azure,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.Attestation = config.AttestationConfig{
					AzureTrustedLaunch: &config.AzureTrustedLaunch{},
				}
				return cfg
			}(),
			policyPatcher: &stubPolicyPatcher{},
		},
		"azure new policy patch error": {
			tfClient: &stubTerraformClient{ip: ip},
			provider: cloudprovider.Azure,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				return cfg
			}(),
			policyPatcher: &stubPolicyPatcher{someErr},
			wantErr:       true,
		},
		"azure newTerraformClient error": {
			newTfClientErr: someErr,
			provider:       cloudprovider.Azure,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				return cfg
			}(),
			policyPatcher: &stubPolicyPatcher{},
			wantErr:       true,
		},
		"azure create cluster error": {
			tfClient: &stubTerraformClient{createClusterErr: someErr},
			provider: cloudprovider.Azure,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				return cfg
			}(),
			policyPatcher:         &stubPolicyPatcher{},
			wantErr:               true,
			wantRollback:          true,
			wantTerraformRollback: true,
		},
		"openstack": {
			tfClient: &stubTerraformClient{ip: ip},
			libvirt:  &stubLibvirtRunner{},
			provider: cloudprovider.OpenStack,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.Provider.OpenStack.Cloud = "testcloud"
				return cfg
			}(),
		},
		"openstack without clouds.yaml": {
			tfClient: &stubTerraformClient{ip: ip},
			libvirt:  &stubLibvirtRunner{},
			provider: cloudprovider.OpenStack,
			config:   config.Default(),
			wantErr:  true,
		},
		"openstack newTerraformClient error": {
			newTfClientErr: someErr,
			libvirt:        &stubLibvirtRunner{},
			provider:       cloudprovider.OpenStack,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.Provider.OpenStack.Cloud = "testcloud"
				return cfg
			}(),
			wantErr: true,
		},
		"openstack create cluster error": {
			tfClient: &stubTerraformClient{createClusterErr: someErr},
			libvirt:  &stubLibvirtRunner{},
			provider: cloudprovider.OpenStack,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.Provider.OpenStack.Cloud = "testcloud"
				return cfg
			}(),
			wantErr:               true,
			wantRollback:          true,
			wantTerraformRollback: true,
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
			tfClient:              &stubTerraformClient{createClusterErr: someErr},
			libvirt:               &stubLibvirtRunner{},
			provider:              cloudprovider.QEMU,
			config:                config.Default(),
			wantErr:               true,
			wantRollback:          !failOnNonAMD64, // if we run on non-AMD64/linux, we don't get to a point where rollback is needed
			wantTerraformRollback: true,
		},
		"qemu start libvirt error": {
			tfClient:              &stubTerraformClient{ip: ip},
			libvirt:               &stubLibvirtRunner{startErr: someErr},
			provider:              cloudprovider.QEMU,
			config:                config.Default(),
			wantRollback:          !failOnNonAMD64,
			wantTerraformRollback: false,
			wantErr:               true,
		},
		"unknown provider": {
			tfClient: &stubTerraformClient{},
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
				image: &stubImageFetcher{
					reference: "some-image",
				},
				newTerraformClient: func(ctx context.Context) (tfResourceClient, error) {
					return tc.tfClient, tc.newTfClientErr
				},
				newLibvirtRunner: func() libvirtRunner {
					return tc.libvirt
				},
				newRawDownloader: func() rawDownloader {
					return &stubRawDownloader{
						destination: "some-destination",
					}
				},
				policyPatcher: tc.policyPatcher,
			}

			opts := CreateOptions{
				Provider:          tc.provider,
				Config:            tc.config,
				InsType:           "type",
				ControlPlaneCount: 2,
				WorkerCount:       3,
				TFLogLevel:        terraform.LogLevelNone,
			}
			idFile, err := creator.Create(context.Background(), opts)

			if tc.wantErr {
				assert.Error(err)
				if tc.wantRollback {
					cl := tc.tfClient.(*stubTerraformClient)
					if tc.wantTerraformRollback {
						assert.True(cl.destroyCalled)
					}
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

type stubPolicyPatcher struct {
	patchErr error
}

func (s stubPolicyPatcher) Patch(_ context.Context, _ string) error {
	return s.patchErr
}

func TestNormalizeAzureURIs(t *testing.T) {
	testCases := map[string]struct {
		in   *terraform.AzureClusterVariables
		want *terraform.AzureClusterVariables
	}{
		"empty": {
			in:   &terraform.AzureClusterVariables{},
			want: &terraform.AzureClusterVariables{},
		},
		"no change": {
			in: &terraform.AzureClusterVariables{
				ImageID: "/communityGalleries/foo/images/constellation/versions/2.1.0",
			},
			want: &terraform.AzureClusterVariables{
				ImageID: "/communityGalleries/foo/images/constellation/versions/2.1.0",
			},
		},
		"fix image id": {
			in: &terraform.AzureClusterVariables{
				ImageID: "/CommunityGalleries/foo/Images/constellation/Versions/2.1.0",
			},
			want: &terraform.AzureClusterVariables{
				ImageID: "/communityGalleries/foo/images/constellation/versions/2.1.0",
			},
		},
		"fix resource group": {
			in: &terraform.AzureClusterVariables{
				UserAssignedIdentity: "/subscriptions/foo/resourcegroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai",
			},
			want: &terraform.AzureClusterVariables{
				UserAssignedIdentity: "/subscriptions/foo/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai",
			},
		},
		"fix arbitrary casing": {
			in: &terraform.AzureClusterVariables{
				ImageID:              "/CoMMUnitygaLLeries/foo/iMAges/constellation/vERsions/2.1.0",
				UserAssignedIdentity: "/subsCRiptions/foo/resoURCegroups/test/proViDers/MICROsoft.mANAgedIdentity/USerASsignediDENtities/uai",
			},
			want: &terraform.AzureClusterVariables{
				ImageID:              "/communityGalleries/foo/images/constellation/versions/2.1.0",
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

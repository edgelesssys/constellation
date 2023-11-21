/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/terraform"
)

func TestApplier(t *testing.T) {
	t.Setenv("CONSTELLATION_OPENSTACK_DEV", "1")
	failOnNonAMD64 := (runtime.GOARCH != "amd64") || (runtime.GOOS != "linux")
	ip := "192.0.2.1"
	configWithProvider := func(provider cloudprovider.Provider) *config.Config {
		cfg := config.Default()
		cfg.RemoveProviderAndAttestationExcept(provider)
		return cfg
	}

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
			config:   configWithProvider(cloudprovider.GCP),
		},
		"gcp create cluster error": {
			tfClient:              &stubTerraformClient{applyClusterErr: assert.AnError},
			provider:              cloudprovider.GCP,
			config:                configWithProvider(cloudprovider.GCP),
			wantErr:               true,
			wantRollback:          true,
			wantTerraformRollback: true,
		},
		"azure": {
			tfClient:      &stubTerraformClient{ip: ip},
			provider:      cloudprovider.Azure,
			config:        configWithProvider(cloudprovider.Azure),
			policyPatcher: &stubPolicyPatcher{},
		},
		"azure trusted launch": {
			tfClient: &stubTerraformClient{ip: ip},
			provider: cloudprovider.Azure,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				cfg.Attestation = config.AttestationConfig{
					AzureTrustedLaunch: &config.AzureTrustedLaunch{},
				}
				return cfg
			}(),
			policyPatcher: &stubPolicyPatcher{},
		},
		"azure new policy patch error": {
			tfClient:      &stubTerraformClient{ip: ip},
			provider:      cloudprovider.Azure,
			config:        configWithProvider(cloudprovider.Azure),
			policyPatcher: &stubPolicyPatcher{assert.AnError},
			wantErr:       true,
		},
		"azure create cluster error": {
			tfClient:              &stubTerraformClient{applyClusterErr: assert.AnError},
			provider:              cloudprovider.Azure,
			config:                configWithProvider(cloudprovider.Azure),
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
				cfg.RemoveProviderAndAttestationExcept(cloudprovider.OpenStack)
				cfg.Provider.OpenStack.Cloud = "testcloud"
				return cfg
			}(),
		},
		"openstack without clouds.yaml": {
			tfClient: &stubTerraformClient{ip: ip},
			libvirt:  &stubLibvirtRunner{},
			provider: cloudprovider.OpenStack,
			config:   configWithProvider(cloudprovider.OpenStack),
			wantErr:  true,
		},
		"openstack create cluster error": {
			tfClient: &stubTerraformClient{applyClusterErr: assert.AnError},
			libvirt:  &stubLibvirtRunner{},
			provider: cloudprovider.OpenStack,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.RemoveProviderAndAttestationExcept(cloudprovider.OpenStack)
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
			config:   configWithProvider(cloudprovider.QEMU),
			wantErr:  failOnNonAMD64,
		},
		"qemu create cluster error": {
			tfClient:              &stubTerraformClient{applyClusterErr: assert.AnError},
			libvirt:               &stubLibvirtRunner{},
			provider:              cloudprovider.QEMU,
			config:                configWithProvider(cloudprovider.QEMU),
			wantErr:               true,
			wantRollback:          !failOnNonAMD64, // if we run on non-AMD64/linux, we don't get to a point where rollback is needed
			wantTerraformRollback: true,
		},
		"qemu start libvirt error": {
			tfClient:              &stubTerraformClient{ip: ip},
			libvirt:               &stubLibvirtRunner{startErr: assert.AnError},
			provider:              cloudprovider.QEMU,
			config:                configWithProvider(cloudprovider.QEMU),
			wantRollback:          !failOnNonAMD64,
			wantTerraformRollback: false,
			wantErr:               true,
		},
		"unknown provider": {
			tfClient: &stubTerraformClient{},
			provider: cloudprovider.Unknown,
			config: func() *config.Config {
				cfg := config.Default()
				cfg.RemoveProviderAndAttestationExcept(cloudprovider.AWS)
				cfg.Provider.AWS = nil
				return cfg
			}(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			applier := &Applier{
				fileHandler: file.NewHandler(afero.NewMemMapFs()),
				imageFetcher: &stubImageFetcher{
					reference: "some-image",
				},
				terraformClient: tc.tfClient,
				libvirtRunner:   tc.libvirt,
				rawDownloader: &stubRawDownloader{
					destination: "some-destination",
				},
				policyPatcher: tc.policyPatcher,
				logLevel:      terraform.LogLevelNone,
				workingDir:    "test",
				backupDir:     "test-backup",
				out:           &bytes.Buffer{},
			}

			diff, err := applier.Plan(context.Background(), tc.config)
			if err != nil {
				assert.True(tc.wantErr, "unexpected error: %s", err)
				return
			}
			assert.False(diff)

			idFile, err := applier.Apply(context.Background(), tc.provider, true)

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
				assert.Equal(ip, idFile.ClusterEndpoint)
			}
		})
	}
}

func TestPlan(t *testing.T) {
	setUpFilesystem := func(existingFiles []string) file.Handler {
		fs := file.NewHandler(afero.NewMemMapFs())
		require.NoError(t, fs.Write("test/terraform.tfstate", []byte{}, file.OptMkdirAll))
		for _, f := range existingFiles {
			require.NoError(t, fs.Write(f, []byte{}))
		}
		return fs
	}

	testCases := map[string]struct {
		upgradeID string
		tf        *stubTerraformClient
		fs        file.Handler
		want      bool
		wantErr   bool
	}{
		"success no diff": {
			upgradeID: "1234",
			tf:        &stubTerraformClient{},
			fs:        setUpFilesystem([]string{}),
		},
		"success diff": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				planDiff: true,
			},
			fs:   setUpFilesystem([]string{}),
			want: true,
		},
		"prepare workspace error": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				prepareWorkspaceErr: assert.AnError,
			},
			fs:      setUpFilesystem([]string{}),
			wantErr: true,
		},
		"plan error": {
			tf: &stubTerraformClient{
				planErr: assert.AnError,
			},
			fs:      setUpFilesystem([]string{}),
			wantErr: true,
		},
		"show plan error no diff": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				showPlanErr: assert.AnError,
			},
			fs: setUpFilesystem([]string{}),
		},
		"show plan error diff": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				showPlanErr: assert.AnError,
				planDiff:    true,
			},
			fs:      setUpFilesystem([]string{}),
			wantErr: true,
		},
		"workspace not clean": {
			upgradeID: "1234",
			tf:        &stubTerraformClient{},
			fs:        setUpFilesystem([]string{filepath.Join(constants.UpgradeDir, "1234", constants.TerraformUpgradeBackupDir)}),
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			u := &Applier{
				terraformClient: tc.tf,
				policyPatcher:   stubPolicyPatcher{},
				fileHandler:     tc.fs,
				imageFetcher:    &stubImageFetcher{reference: "some-image"},
				rawDownloader:   &stubRawDownloader{destination: "some-destination"},
				libvirtRunner:   &stubLibvirtRunner{},
				logLevel:        terraform.LogLevelDebug,
				backupDir:       filepath.Join(constants.UpgradeDir, tc.upgradeID),
				workingDir:      "test",
				out:             io.Discard,
			}

			cfg := config.Default()
			cfg.RemoveProviderAndAttestationExcept(cloudprovider.QEMU)

			diff, err := u.Plan(context.Background(), cfg)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Equal(tc.want, diff)
			}
		})
	}
}

func TestApply(t *testing.T) {
	testCases := map[string]struct {
		upgradeID     string
		tf            *stubTerraformClient
		policyPatcher stubPolicyPatcher
		fs            file.Handler
		wantErr       bool
	}{
		"success": {
			upgradeID:     "1234",
			tf:            &stubTerraformClient{},
			policyPatcher: stubPolicyPatcher{},
		},
		"apply error": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				applyClusterErr: assert.AnError,
			},
			policyPatcher: stubPolicyPatcher{},
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := require.New(t)

			u := &Applier{
				terraformClient: tc.tf,
				logLevel:        terraform.LogLevelDebug,
				libvirtRunner:   &stubLibvirtRunner{},
				policyPatcher:   stubPolicyPatcher{},
				fileHandler:     tc.fs,
				backupDir:       filepath.Join(constants.UpgradeDir, tc.upgradeID),
				workingDir:      "test",
				out:             io.Discard,
			}

			_, err := u.Apply(context.Background(), cloudprovider.QEMU, WithoutRollbackOnError)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
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

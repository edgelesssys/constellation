/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// defaultStateFile returns a valid default state for testing.
func defaultStateFile() *state.State {
	return &state.State{
		Version: "v1",
		Infrastructure: state.Infrastructure{
			UID:               "123",
			Name:              "test-cluster",
			ClusterEndpoint:   "192.0.2.1",
			InClusterEndpoint: "192.0.2.1",
			InitSecret:        []byte{0x41},
			APIServerCertSANs: []string{
				"127.0.0.1",
				"www.example.com",
			},
			IPCidrNode: "0.0.0.0/24",
			Azure: &state.Azure{
				ResourceGroup:            "test-rg",
				SubscriptionID:           "test-sub",
				NetworkSecurityGroupName: "test-nsg",
				LoadBalancerName:         "test-lb",
				UserAssignedIdentity:     "test-uami",
				AttestationURL:           "test-maaUrl",
			},
			GCP: &state.GCP{
				ProjectID: "test-project",
				IPCidrPod: "0.0.0.0/24",
			},
		},
		ClusterValues: state.ClusterValues{
			ClusterID:       "deadbeef",
			OwnerID:         "deadbeef",
			MeasurementSalt: []byte{0x41},
		},
	}
}

func defaultAzureStateFile() *state.State {
	s := defaultStateFile()
	s.Infrastructure.GCP = nil
	return s
}

func defaultGCPStateFile() *state.State {
	s := defaultStateFile()
	s.Infrastructure.Azure = nil
	return s
}

func TestParseApplyFlags(t *testing.T) {
	require := require.New(t)
	defaultFlags := func() *pflag.FlagSet {
		flags := NewApplyCmd().Flags()
		// Register persistent flags
		flags.String("workspace", "", "")
		flags.String("tf-log", "NONE", "")
		flags.Bool("force", false, "")
		flags.Bool("debug", false, "")
		return flags
	}

	testCases := map[string]struct {
		flags     *pflag.FlagSet
		wantFlags applyFlags
		wantErr   bool
	}{
		"default flags": {
			flags: defaultFlags(),
			wantFlags: applyFlags{
				helmWaitMode:   helm.WaitModeAtomic,
				upgradeTimeout: 5 * time.Minute,
			},
		},
		"skip phases": {
			flags: func() *pflag.FlagSet {
				flags := defaultFlags()
				require.NoError(flags.Set("skip-phases", fmt.Sprintf("%s,%s", skipHelmPhase, skipK8sPhase)))
				return flags
			}(),
			wantFlags: applyFlags{
				skipPhases:     skipPhases{skipHelmPhase: struct{}{}, skipK8sPhase: struct{}{}},
				helmWaitMode:   helm.WaitModeAtomic,
				upgradeTimeout: 5 * time.Minute,
			},
		},
		"skip helm wait": {
			flags: func() *pflag.FlagSet {
				flags := defaultFlags()
				require.NoError(flags.Set("skip-helm-wait", "true"))
				return flags
			}(),
			wantFlags: applyFlags{
				helmWaitMode:   helm.WaitModeNone,
				upgradeTimeout: 5 * time.Minute,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			var flags applyFlags

			err := flags.parse(tc.flags)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantFlags, flags)
		})
	}
}

func TestBackupHelmCharts(t *testing.T) {
	testCases := map[string]struct {
		helmApplier      helm.Applier
		backupClient     *stubKubernetesUpgrader
		includesUpgrades bool
		wantErr          bool
	}{
		"success, no upgrades": {
			helmApplier:  &stubRunner{},
			backupClient: &stubKubernetesUpgrader{},
		},
		"success with upgrades": {
			helmApplier:      &stubRunner{},
			backupClient:     &stubKubernetesUpgrader{},
			includesUpgrades: true,
		},
		"saving charts fails": {
			helmApplier: &stubRunner{
				saveChartsErr: assert.AnError,
			},
			backupClient: &stubKubernetesUpgrader{},
			wantErr:      true,
		},
		"backup CRDs fails": {
			helmApplier: &stubRunner{},
			backupClient: &stubKubernetesUpgrader{
				backupCRDsErr: assert.AnError,
			},
			includesUpgrades: true,
			wantErr:          true,
		},
		"backup CRs fails": {
			helmApplier: &stubRunner{},
			backupClient: &stubKubernetesUpgrader{
				backupCRsErr: assert.AnError,
			},
			includesUpgrades: true,
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			a := applyCmd{
				fileHandler: file.NewHandler(afero.NewMemMapFs()),
				log:         logger.NewTest(t),
			}

			err := a.backupHelmCharts(context.Background(), tc.backupClient, tc.helmApplier, tc.includesUpgrades, "")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			if tc.includesUpgrades {
				assert.True(tc.backupClient.backupCRDsCalled)
				assert.True(tc.backupClient.backupCRsCalled)
			}
		})
	}
}

func TestSkipPhases(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	cmd := NewApplyCmd()
	// register persistent flags manually
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().Bool("force", true, "")
	cmd.Flags().String("tf-log", "NONE", "")
	cmd.Flags().Bool("debug", false, "")

	require.NoError(cmd.Flags().Set("skip-phases", strings.Join(allPhases(), ",")))
	wantPhases := skipPhases{}
	wantPhases.add(skipInfrastructurePhase, skipInitPhase, skipAttestationConfigPhase, skipCertSANsPhase, skipHelmPhase, skipK8sPhase, skipImagePhase)

	var flags applyFlags
	err := flags.parse(cmd.Flags())
	require.NoError(err)
	assert.Equal(wantPhases, flags.skipPhases)

	phases := skipPhases{}
	phases.add(skipAttestationConfigPhase, skipCertSANsPhase)
	assert.True(phases.contains(skipAttestationConfigPhase, skipCertSANsPhase))
	assert.False(phases.contains(skipAttestationConfigPhase, skipInitPhase))
	assert.False(phases.contains(skipInitPhase, skipInfrastructurePhase))
}

func TestValidateInputs(t *testing.T) {
	defaultConfig := func(csp cloudprovider.Provider) func(require *require.Assertions, fh file.Handler) {
		return func(require *require.Assertions, fh file.Handler) {
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), csp)

			if csp == cloudprovider.GCP {
				require.NoError(fh.WriteJSON("saKey.json", &gcpshared.ServiceAccountKey{
					Type:                    "service_account",
					ProjectID:               "project_id",
					PrivateKeyID:            "key_id",
					PrivateKey:              "key",
					ClientEmail:             "client_email",
					ClientID:                "client_id",
					AuthURI:                 "auth_uri",
					TokenURI:                "token_uri",
					AuthProviderX509CertURL: "cert",
					ClientX509CertURL:       "client_cert",
				}))
				cfg.Provider.GCP.ServiceAccountKeyPath = "saKey.json"
			}

			require.NoError(fh.WriteYAML(constants.ConfigFilename, cfg))
		}
	}
	defaultState := func(require *require.Assertions, fh file.Handler) {
		require.NoError(fh.WriteYAML(constants.StateFilename, &state.State{}))
	}
	defaultMasterSecret := func(require *require.Assertions, fh file.Handler) {
		require.NoError(fh.WriteJSON(constants.MasterSecretFilename, &uri.MasterSecret{}))
	}
	defaultAdminConfig := func(require *require.Assertions, fh file.Handler) {
		require.NoError(fh.Write(constants.AdminConfFilename, []byte("admin config")))
	}
	defaultTfState := func(require *require.Assertions, fh file.Handler) {
		require.NoError(fh.Write(filepath.Join(constants.TerraformWorkingDir, "tfvars"), []byte("tf state")))
	}

	testCases := map[string]struct {
		createConfig       func(require *require.Assertions, fh file.Handler)
		createState        func(require *require.Assertions, fh file.Handler)
		createMasterSecret func(require *require.Assertions, fh file.Handler)
		createAdminConfig  func(require *require.Assertions, fh file.Handler)
		createTfState      func(require *require.Assertions, fh file.Handler)
		stdin              string
		flags              applyFlags
		wantPhases         skipPhases
		wantErr            bool
	}{
		"gcp: all files exist": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        defaultState,
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases: skipPhases{
				skipInitPhase: struct{}{},
			},
		},
		"aws: all files exist": {
			createConfig:       defaultConfig(cloudprovider.AWS),
			createState:        defaultState,
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases: skipPhases{
				skipInitPhase: struct{}{},
			},
		},
		"azure: all files exist": {
			createConfig:       defaultConfig(cloudprovider.Azure),
			createState:        defaultState,
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases: skipPhases{
				skipInitPhase: struct{}{},
			},
		},
		"qemu: all files exist": {
			createConfig:       defaultConfig(cloudprovider.QEMU),
			createState:        defaultState,
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases: skipPhases{
				skipInitPhase:  struct{}{},
				skipImagePhase: struct{}{}, // No image upgrades on QEMU
			},
		},
		"no config file": {
			createConfig:       func(require *require.Assertions, fh file.Handler) {},
			createState:        defaultState,
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantErr:            true,
		},
		"no admin config file, but mastersecret file exists": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        defaultState,
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  func(require *require.Assertions, fh file.Handler) {},
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantErr:            true,
		},
		"no admin config file, no master secret": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        defaultState,
			createMasterSecret: func(require *require.Assertions, fh file.Handler) {},
			createAdminConfig:  func(require *require.Assertions, fh file.Handler) {},
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases: skipPhases{
				skipImagePhase: struct{}{},
				skipK8sPhase:   struct{}{},
			},
		},
		"no tf state, but admin config exists": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        defaultState,
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      func(require *require.Assertions, fh file.Handler) {},
			flags:              applyFlags{},
			wantErr:            true,
		},
		"only config file": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        func(require *require.Assertions, fh file.Handler) {},
			createMasterSecret: func(require *require.Assertions, fh file.Handler) {},
			createAdminConfig:  func(require *require.Assertions, fh file.Handler) {},
			createTfState:      func(require *require.Assertions, fh file.Handler) {},
			flags:              applyFlags{},
			wantPhases: skipPhases{
				skipImagePhase: struct{}{},
				skipK8sPhase:   struct{}{},
			},
		},
		"skip terraform": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        defaultState,
			createMasterSecret: func(require *require.Assertions, fh file.Handler) {},
			createAdminConfig:  func(require *require.Assertions, fh file.Handler) {},
			createTfState:      func(require *require.Assertions, fh file.Handler) {},
			flags: applyFlags{
				skipPhases: skipPhases{
					skipInfrastructurePhase: struct{}{},
				},
			},
			wantPhases: skipPhases{
				skipInfrastructurePhase: struct{}{},
				skipImagePhase:          struct{}{},
				skipK8sPhase:            struct{}{},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			tc.createConfig(require, fileHandler)
			tc.createState(require, fileHandler)
			tc.createMasterSecret(require, fileHandler)
			tc.createAdminConfig(require, fileHandler)
			tc.createTfState(require, fileHandler)

			cmd := NewApplyCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)
			cmd.SetIn(bytes.NewBufferString(tc.stdin))

			a := applyCmd{
				log:          logger.NewTest(t),
				fileHandler:  fileHandler,
				flags:        tc.flags,
				quotaChecker: &stubLicenseClient{},
			}

			_, _, err := a.validateInputs(cmd, &stubAttestationFetcher{})
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			var cfgErr *config.ValidationError
			if errors.As(err, &cfgErr) {
				t.Log(cfgErr.LongMessage())
			}
			assert.Equal(tc.wantPhases, a.flags.skipPhases)
		})
	}
}

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
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation"
	"github.com/edgelesssys/constellation/v2/internal/constellation/helm"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// defaultStateFile returns a valid default state for testing.
func defaultStateFile(csp cloudprovider.Provider) *state.State {
	stateFile := &state.State{
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
	switch csp {
	case cloudprovider.GCP:
		stateFile.Infrastructure.Azure = nil
	case cloudprovider.Azure:
		stateFile.Infrastructure.GCP = nil
	default:
		stateFile.Infrastructure.Azure = nil
		stateFile.Infrastructure.GCP = nil
	}
	return stateFile
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
				helmWaitMode: helm.WaitModeAtomic,
				helmTimeout:  10 * time.Minute,
			},
		},
		"skip phases": {
			flags: func() *pflag.FlagSet {
				flags := defaultFlags()
				require.NoError(flags.Set("skip-phases", fmt.Sprintf("%s,%s", skipHelmPhase, skipK8sPhase)))
				return flags
			}(),
			wantFlags: applyFlags{
				skipPhases:   newPhases(skipHelmPhase, skipK8sPhase),
				helmWaitMode: helm.WaitModeAtomic,
				helmTimeout:  10 * time.Minute,
			},
		},
		"skip helm wait": {
			flags: func() *pflag.FlagSet {
				flags := defaultFlags()
				require.NoError(flags.Set("skip-helm-wait", "true"))
				return flags
			}(),
			wantFlags: applyFlags{
				helmWaitMode: helm.WaitModeNone,
				helmTimeout:  10 * time.Minute,
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
				applier: &stubConstellApplier{
					stubKubernetesUpgrader: tc.backupClient,
				},
				log: logger.NewTest(t),
			}

			err := a.backupHelmCharts(context.Background(), tc.helmApplier, tc.includesUpgrades, "")
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
	wantPhases := newPhases(skipInfrastructurePhase, skipInitPhase, skipAttestationConfigPhase, skipCertSANsPhase, skipHelmPhase, skipK8sPhase, skipImagePhase)

	var flags applyFlags
	err := flags.parse(cmd.Flags())
	require.NoError(err)
	assert.Equal(wantPhases, flags.skipPhases)

	phases := newPhases(skipAttestationConfigPhase, skipCertSANsPhase)
	assert.True(phases.contains(skipAttestationConfigPhase, skipCertSANsPhase))
	assert.False(phases.contains(skipAttestationConfigPhase, skipInitPhase))
	assert.False(phases.contains(skipInitPhase, skipInfrastructurePhase))
}

func TestValidateInputs(t *testing.T) {
	defaultConfig := func(csp cloudprovider.Provider) func(require *require.Assertions, _ file.Handler) {
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
	preInitState := func(csp cloudprovider.Provider) func(_ *require.Assertions, _ file.Handler) {
		return func(require *require.Assertions, fh file.Handler) {
			stateFile := defaultStateFile(csp)
			stateFile.ClusterValues = state.ClusterValues{}
			require.NoError(fh.WriteYAML(constants.StateFilename, stateFile))
		}
	}
	postInitState := func(csp cloudprovider.Provider) func(require *require.Assertions, fh file.Handler) {
		return func(require *require.Assertions, fh file.Handler) {
			require.NoError(fh.WriteYAML(constants.StateFilename, defaultStateFile(csp)))
		}
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
		assert             func(require *require.Assertions, assert *assert.Assertions, conf *config.Config, stateFile *state.State)
		wantErr            bool
	}{
		"[upgrade] gcp: all files exist": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        postInitState(cloudprovider.GCP),
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases:         newPhases(skipInitPhase),
		},
		"[upgrade] aws: all files exist": {
			createConfig:       defaultConfig(cloudprovider.AWS),
			createState:        postInitState(cloudprovider.AWS),
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases:         newPhases(skipInitPhase),
		},
		"[upgrade] azure: all files exist": {
			createConfig:       defaultConfig(cloudprovider.Azure),
			createState:        postInitState(cloudprovider.Azure),
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases:         newPhases(skipInitPhase),
		},
		"[upgrade] qemu: all files exist": {
			createConfig:       defaultConfig(cloudprovider.QEMU),
			createState:        postInitState(cloudprovider.QEMU),
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases:         newPhases(skipInitPhase, skipImagePhase), // No image upgrades on QEMU
		},
		"no config file errors": {
			createConfig:       func(_ *require.Assertions, _ file.Handler) {},
			createState:        postInitState(cloudprovider.GCP),
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantErr:            true,
		},
		"[init] no admin config file, but mastersecret file exists errors": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        preInitState(cloudprovider.GCP),
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  func(_ *require.Assertions, _ file.Handler) {},
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantErr:            true,
		},
		"[init] no admin config file, no master secret": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        preInitState(cloudprovider.GCP),
			createMasterSecret: func(_ *require.Assertions, _ file.Handler) {},
			createAdminConfig:  func(_ *require.Assertions, _ file.Handler) {},
			createTfState:      defaultTfState,
			flags:              applyFlags{},
			wantPhases:         newPhases(skipImagePhase, skipK8sPhase),
		},
		"[create] no tf state, but admin config exists errors": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        preInitState(cloudprovider.GCP),
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      func(_ *require.Assertions, _ file.Handler) {},
			flags:              applyFlags{},
			wantErr:            true,
		},
		"[create] only config, skip everything but infrastructure": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        func(_ *require.Assertions, _ file.Handler) {},
			createMasterSecret: func(_ *require.Assertions, _ file.Handler) {},
			createAdminConfig:  func(_ *require.Assertions, _ file.Handler) {},
			createTfState:      func(_ *require.Assertions, _ file.Handler) {},
			flags: applyFlags{
				skipPhases: newPhases(skipInitPhase, skipAttestationConfigPhase, skipCertSANsPhase, skipHelmPhase, skipK8sPhase, skipImagePhase),
			},
			wantPhases: newPhases(skipInitPhase, skipAttestationConfigPhase, skipCertSANsPhase, skipHelmPhase, skipK8sPhase, skipImagePhase),
		},
		"[create + init] only config file": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        func(_ *require.Assertions, _ file.Handler) {},
			createMasterSecret: func(_ *require.Assertions, _ file.Handler) {},
			createAdminConfig:  func(_ *require.Assertions, _ file.Handler) {},
			createTfState:      func(_ *require.Assertions, _ file.Handler) {},
			flags:              applyFlags{},
			wantPhases:         newPhases(skipImagePhase, skipK8sPhase),
		},
		"[init] self-managed: config and state file exist, skip-phases=infrastructure": {
			createConfig:       defaultConfig(cloudprovider.GCP),
			createState:        preInitState(cloudprovider.GCP),
			createMasterSecret: func(_ *require.Assertions, _ file.Handler) {},
			createAdminConfig:  func(_ *require.Assertions, _ file.Handler) {},
			createTfState:      func(_ *require.Assertions, _ file.Handler) {},
			flags: applyFlags{
				skipPhases: newPhases(skipInfrastructurePhase),
			},
			wantPhases: newPhases(skipInfrastructurePhase, skipImagePhase, skipK8sPhase),
		},
		"[upgrade] k8s patch version no longer supported, user confirms to skip k8s and continue upgrade. Valid K8s patch version is used in config afterwards": {
			createConfig: func(require *require.Assertions, fh file.Handler) {
				cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.GCP)

				// use first version in list (oldest) as it should never have a patch version
				versionParts := strings.Split(versions.SupportedK8sVersions()[0], ".")
				versionParts[len(versionParts)-1] = "0"
				cfg.KubernetesVersion = versions.ValidK8sVersion(strings.Join(versionParts, "."))
				require.NoError(fh.WriteYAML(constants.ConfigFilename, cfg))
			},
			createState:        postInitState(cloudprovider.GCP),
			createMasterSecret: defaultMasterSecret,
			createAdminConfig:  defaultAdminConfig,
			createTfState:      defaultTfState,
			stdin:              "y\n",
			wantPhases:         newPhases(skipInitPhase, skipK8sPhase),
			assert: func(_ *require.Assertions, assert *assert.Assertions, conf *config.Config, _ *state.State) {
				assert.NotEmpty(conf.KubernetesVersion)
				_, err := versions.NewValidK8sVersion(string(conf.KubernetesVersion), true)
				assert.NoError(err)
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
				log:         logger.NewTest(t),
				fileHandler: fileHandler,
				flags:       tc.flags,
			}

			conf, state, err := a.validateInputs(cmd, &stubAttestationFetcher{})
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

			if tc.assert != nil {
				tc.assert(require, assert, conf, state)
			}
		})
	}
}

func TestSkipPhasesCompletion(t *testing.T) {
	testCases := map[string]struct {
		toComplete      string
		wantSuggestions []string
	}{
		"empty": {
			toComplete:      "",
			wantSuggestions: allPhases(),
		},
		"partial": {
			toComplete:      "hel",
			wantSuggestions: []string{string(skipHelmPhase)},
		},
		"one full word": {
			toComplete: string(skipHelmPhase),
		},
		"one full word with comma": {
			toComplete: string(skipHelmPhase) + ",",
			wantSuggestions: func() []string {
				allPhases := allPhases()
				var suggestions []string
				for _, phase := range allPhases {
					if phase == string(skipHelmPhase) {
						continue
					}
					suggestions = append(suggestions, fmt.Sprintf("%s,%s", skipHelmPhase, phase))
				}
				return suggestions
			}(),
		},
		"one full word, one partial": {
			toComplete:      string(skipHelmPhase) + ",ima",
			wantSuggestions: []string{fmt.Sprintf("%s,%s", skipHelmPhase, skipImagePhase)},
		},
		"all phases": {
			toComplete:      strings.Join(allPhases(), ","),
			wantSuggestions: []string{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			suggestions, _ := skipPhasesCompletion(nil, nil, tc.toComplete)
			assert.ElementsMatch(tc.wantSuggestions, suggestions, "got: %v, want: %v", suggestions, tc.wantSuggestions)
		})
	}
}

func newPhases(phases ...skipPhase) skipPhases {
	skipPhases := skipPhases{}
	skipPhases.add(phases...)
	return skipPhases
}

type stubConstellApplier struct {
	checkLicenseErr            error
	masterSecret               uri.MasterSecret
	measurementSalt            []byte
	generateMasterSecretErr    error
	generateMeasurementSaltErr error
	initErr                    error
	initOutput                 constellation.InitOutput
	*stubKubernetesUpgrader
	helmApplier
}

func (s *stubConstellApplier) SetKubeConfig([]byte) error { return nil }

func (s *stubConstellApplier) CheckLicense(context.Context, cloudprovider.Provider, bool, string) (int, error) {
	return 0, s.checkLicenseErr
}

func (s *stubConstellApplier) GenerateMasterSecret() (uri.MasterSecret, error) {
	return s.masterSecret, s.generateMasterSecretErr
}

func (s *stubConstellApplier) GenerateMeasurementSalt() ([]byte, error) {
	return s.measurementSalt, s.generateMeasurementSaltErr
}

func (s *stubConstellApplier) Init(context.Context, atls.Validator, *state.State, io.Writer, constellation.InitPayload) (constellation.InitOutput, error) {
	return s.initOutput, s.initErr
}

type helmApplier interface {
	AnnotateCoreDNSResources(context.Context) error
	CleanupCoreDNSResources(ctx context.Context) error
	PrepareHelmCharts(
		flags helm.Options, stateFile *state.State, serviceAccURI string, masterSecret uri.MasterSecret,
	) (
		helm.Applier, bool, error)
}

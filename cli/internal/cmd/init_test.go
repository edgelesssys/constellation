/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
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
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	k8sclientapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestInitArgumentValidation(t *testing.T) {
	assert := assert.New(t)

	cmd := NewInitCmd()
	assert.NoError(cmd.ValidateArgs(nil))
	assert.Error(cmd.ValidateArgs([]string{"something"}))
	assert.Error(cmd.ValidateArgs([]string{"sth", "sth"}))
}

// preInitStateFile returns a state file satisfying the pre-init state file
// constraints.
func preInitStateFile(csp cloudprovider.Provider) *state.State {
	s := defaultStateFile(csp)
	s.ClusterValues = state.ClusterValues{}
	return s
}

func TestInitialize(t *testing.T) {
	respKubeconfig := k8sclientapi.Config{
		Clusters: map[string]*k8sclientapi.Cluster{
			"cluster": {
				Server: "https://192.0.2.1:6443",
			},
		},
	}
	respKubeconfigBytes, err := clientcmd.Write(respKubeconfig)
	require.NoError(t, err)

	gcpServiceAccKey := &gcpshared.ServiceAccountKey{
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
	}
	testInitOutput := constellation.InitOutput{
		Kubeconfig: respKubeconfigBytes,
		OwnerID:    "ownerID",
		ClusterID:  "clusterID",
	}
	serviceAccPath := "/test/service-account.json"

	testCases := map[string]struct {
		provider                cloudprovider.Provider
		stateFile               *state.State
		configMutator           func(*config.Config)
		serviceAccKey           *gcpshared.ServiceAccountKey
		initOutput              constellation.InitOutput
		initErr                 error
		retriable               bool
		masterSecretShouldExist bool
		wantErr                 bool
	}{
		"initialize some gcp instances": {
			provider:      cloudprovider.GCP,
			stateFile:     preInitStateFile(cloudprovider.GCP),
			configMutator: func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			serviceAccKey: gcpServiceAccKey,
			initOutput:    testInitOutput,
		},
		"initialize some azure instances": {
			provider:   cloudprovider.Azure,
			stateFile:  preInitStateFile(cloudprovider.Azure),
			initOutput: testInitOutput,
		},
		"initialize some qemu instances": {
			provider:   cloudprovider.QEMU,
			stateFile:  preInitStateFile(cloudprovider.QEMU),
			initOutput: testInitOutput,
		},
		"non retriable error": {
			provider:                cloudprovider.QEMU,
			stateFile:               preInitStateFile(cloudprovider.QEMU),
			initErr:                 &constellation.NonRetriableInitError{Err: assert.AnError},
			retriable:               false,
			masterSecretShouldExist: true,
			wantErr:                 true,
		},
		"non retriable error with failed log collection": {
			provider:                cloudprovider.QEMU,
			stateFile:               preInitStateFile(cloudprovider.QEMU),
			initErr:                 &constellation.NonRetriableInitError{Err: assert.AnError, LogCollectionErr: assert.AnError},
			retriable:               false,
			masterSecretShouldExist: true,
			wantErr:                 true,
		},
		"invalid state file": {
			provider:      cloudprovider.GCP,
			stateFile:     &state.State{Version: "invalid"},
			configMutator: func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			serviceAccKey: gcpServiceAccKey,
			initOutput:    testInitOutput,
			retriable:     true,
			wantErr:       true,
		},
		"empty state file": {
			provider:      cloudprovider.GCP,
			stateFile:     &state.State{},
			configMutator: func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			serviceAccKey: gcpServiceAccKey,
			initOutput:    testInitOutput,
			retriable:     true,
			wantErr:       true,
		},
		"no state file": {
			provider:      cloudprovider.GCP,
			configMutator: func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			serviceAccKey: gcpServiceAccKey,
			initOutput:    testInitOutput,
			retriable:     true,
			wantErr:       true,
		},
		"init call fails": {
			provider:                cloudprovider.GCP,
			configMutator:           func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			stateFile:               preInitStateFile(cloudprovider.GCP),
			serviceAccKey:           gcpServiceAccKey,
			initErr:                 &constellation.NonRetriableInitError{Err: assert.AnError},
			retriable:               false,
			masterSecretShouldExist: true,
			wantErr:                 true,
		},
		"k8s version without v works": {
			provider:   cloudprovider.Azure,
			stateFile:  preInitStateFile(cloudprovider.Azure),
			initOutput: testInitOutput,
			configMutator: func(c *config.Config) {
				res, err := versions.NewValidK8sVersion(strings.TrimPrefix(string(versions.Default), "v"), true)
				require.NoError(t, err)
				c.KubernetesVersion = res
			},
		},
		"outdated k8s patch version doesn't work": {
			provider:   cloudprovider.Azure,
			stateFile:  preInitStateFile(cloudprovider.Azure),
			initOutput: testInitOutput,
			configMutator: func(c *config.Config) {
				v, err := semver.New(versions.SupportedK8sVersions()[0])
				require.NoError(t, err)
				outdatedPatchVer := semver.NewFromInt(v.Major(), v.Minor(), v.Patch()-1, "").String()
				c.KubernetesVersion = versions.ValidK8sVersion(outdatedPatchVer)
			},
			wantErr:   true,
			retriable: true, // doesn't need to show retriable error message
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Command
			cmd := NewInitCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)

			// File system preparation
			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			config := defaultConfigWithExpectedMeasurements(t, config.Default(), tc.provider)
			if tc.configMutator != nil {
				tc.configMutator(config)
			}
			require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, config, file.OptNone))
			if tc.stateFile != nil {
				require.NoError(tc.stateFile.WriteToFile(fileHandler, constants.StateFilename))
			}
			if tc.serviceAccKey != nil {
				require.NoError(fileHandler.WriteJSON(serviceAccPath, tc.serviceAccKey, file.OptNone))
			}

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
			defer cancel()
			cmd.SetContext(ctx)

			i := &applyCmd{
				fileHandler: fileHandler,
				flags: applyFlags{
					rootFlags:  rootFlags{force: true},
					skipPhases: newPhases(skipInfrastructurePhase),
				},
				log:     logger.NewTest(t),
				spinner: &nopSpinner{},
				merger:  &stubMerger{},
				applier: &stubConstellApplier{
					masterSecret: uri.MasterSecret{
						Key:  bytes.Repeat([]byte{0x01}, 32),
						Salt: bytes.Repeat([]byte{0x02}, 32),
					},
					measurementSalt: bytes.Repeat([]byte{0x03}, 32),
					initErr:         tc.initErr,
					initOutput:      tc.initOutput,
					stubKubernetesUpgrader: &stubKubernetesUpgrader{
						// On init, no attestation config exists yet
						getClusterAttestationConfigErr: k8serrors.NewNotFound(schema.GroupResource{}, ""),
					},
					helmApplier: &stubHelmApplier{},
				},
			}

			err := i.apply(cmd, stubAttestationFetcher{}, "test")

			if tc.wantErr {
				assert.Error(err)
				fmt.Println(err)
				if !tc.retriable {
					assert.Contains(errOut.String(), "This error is not recoverable")
				} else {
					assert.Empty(errOut.String())
				}
				if !tc.masterSecretShouldExist {
					_, err = fileHandler.Stat(constants.MasterSecretFilename)
					assert.Error(err)
				}
				return
			}
			require.NoError(err)
			// assert.Contains(out.String(), base64.StdEncoding.EncodeToString([]byte("ownerID")))
			assert.Contains(out.String(), "clusterID")
			var secret uri.MasterSecret
			assert.NoError(fileHandler.ReadJSON(constants.MasterSecretFilename, &secret))
			assert.NotEmpty(secret.Key)
			assert.NotEmpty(secret.Salt)
		})
	}
}

type stubHelmApplier struct {
	err error
}

func (s stubHelmApplier) AnnotateCoreDNSResources(_ context.Context) error {
	return nil
}

func (s stubHelmApplier) CleanupCoreDNSResources(_ context.Context) error {
	return nil
}

func (s stubHelmApplier) PrepareHelmCharts(
	_ helm.Options, _ *state.State, _ string, _ uri.MasterSecret,
) (helm.Applier, bool, error) {
	return stubRunner{}, false, s.err
}

type stubRunner struct {
	applyErr      error
	saveChartsErr error
}

func (s stubRunner) Apply(_ context.Context) error {
	return s.applyErr
}

func (s stubRunner) SaveCharts(_ string, _ file.Handler) error {
	return s.saveChartsErr
}

func TestWriteOutput(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	clusterEndpoint := "cluster-endpoint"

	expectedKubeconfig := k8sclientapi.Config{
		Clusters: map[string]*k8sclientapi.Cluster{
			"cluster": {
				Server: fmt.Sprintf("https://%s:6443", clusterEndpoint),
			},
		},
	}
	expectedKubeconfigBytes, err := clientcmd.Write(expectedKubeconfig)
	require.NoError(err)

	respKubeconfig := k8sclientapi.Config{
		Clusters: map[string]*k8sclientapi.Cluster{
			"cluster": {
				Server: "https://cluster-endpoint:6443",
			},
		},
	}
	respKubeconfigBytes, err := clientcmd.Write(respKubeconfig)
	require.NoError(err)

	resp := &initproto.InitResponse{
		Kind: &initproto.InitResponse_InitSuccess{
			InitSuccess: &initproto.InitSuccessResponse{
				OwnerId:    []byte("ownerID"),
				ClusterId:  []byte("clusterID"),
				Kubeconfig: respKubeconfigBytes,
			},
		},
	}
	ownerID := string(resp.GetInitSuccess().GetOwnerId())
	clusterID := string(resp.GetInitSuccess().GetClusterId())
	initOutput := constellation.InitOutput{
		OwnerID:    "ownerID",
		ClusterID:  "clusterID",
		Kubeconfig: respKubeconfigBytes,
	}

	measurementSalt := []byte{0x41}

	expectedStateFile := &state.State{
		Version: state.Version1,
		ClusterValues: state.ClusterValues{
			ClusterID:       clusterID,
			OwnerID:         ownerID,
			MeasurementSalt: measurementSalt,
		},
		Infrastructure: state.Infrastructure{
			APIServerCertSANs: []string{},
			InitSecret:        []byte{},
			ClusterEndpoint:   clusterEndpoint,
		},
	}

	var out bytes.Buffer
	testFs := afero.NewMemMapFs()
	fileHandler := file.NewHandler(testFs)

	stateFile := state.New().SetInfrastructure(state.Infrastructure{
		ClusterEndpoint: clusterEndpoint,
	})

	i := &applyCmd{
		fileHandler: fileHandler,
		spinner:     &nopSpinner{},
		merger:      &stubMerger{},
		log:         logger.NewTest(t),
		applier:     constellation.NewApplier(logger.NewTest(t), &nopSpinner{}, constellation.ApplyContextCLI, nil),
	}
	err = i.writeInitOutput(stateFile, initOutput, false, &out, measurementSalt)
	require.NoError(err)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), constants.AdminConfFilename)

	afs := afero.Afero{Fs: testFs}
	adminConf, err := afs.ReadFile(constants.AdminConfFilename)
	assert.NoError(err)
	assert.Contains(string(adminConf), clusterEndpoint)
	assert.Equal(string(expectedKubeconfigBytes), string(adminConf))

	fh := file.NewHandler(afs)
	readStateFile, err := state.ReadFromFile(fh, constants.StateFilename)
	assert.NoError(err)
	assert.Equal(expectedStateFile, readStateFile)
	out.Reset()
	require.NoError(afs.Remove(constants.AdminConfFilename))

	// test custom workspace
	i.flags.pathPrefixer = pathprefix.New("/some/path")
	err = i.writeInitOutput(stateFile, initOutput, true, &out, measurementSalt)
	require.NoError(err)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), i.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	out.Reset()
	// File is written to current working dir, we simply pass the workspace for generating readable user output
	require.NoError(afs.Remove(constants.AdminConfFilename))
	i.flags.pathPrefixer = pathprefix.PathPrefixer{}

	// test config merging
	err = i.writeInitOutput(stateFile, initOutput, true, &out, measurementSalt)
	require.NoError(err)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), constants.AdminConfFilename)
	assert.Contains(out.String(), "Constellation kubeconfig merged with default config")
	assert.Contains(out.String(), "You can now connect to your cluster")
	out.Reset()
	require.NoError(afs.Remove(constants.AdminConfFilename))

	// test config merging with env vars set
	i.merger = &stubMerger{envVar: "/some/path/to/kubeconfig"}
	err = i.writeInitOutput(stateFile, initOutput, true, &out, measurementSalt)
	require.NoError(err)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), constants.AdminConfFilename)
	assert.Contains(out.String(), "Constellation kubeconfig merged with default config")
	assert.Contains(out.String(), "Warning: KUBECONFIG environment variable is set")
}

func TestGenerateMasterSecret(t *testing.T) {
	testCases := map[string]struct {
		createFileFunc func(handler file.Handler) error
		fs             func() afero.Fs
		wantErr        bool
	}{
		"file already exists": {
			fs: afero.NewMemMapFs,
			createFileFunc: func(handler file.Handler) error {
				return handler.WriteJSON(
					constants.MasterSecretFilename,
					uri.MasterSecret{Key: []byte("constellation-master-secret"), Salt: []byte("constellation-32Byte-length-salt")},
					file.OptNone,
				)
			},
			wantErr: true,
		},
		"file does not exist": {
			createFileFunc: func(_ file.Handler) error { return nil },
			fs:             afero.NewMemMapFs,
			wantErr:        false,
		},
		"file not writeable": {
			createFileFunc: func(_ file.Handler) error { return nil },
			fs:             func() afero.Fs { return afero.NewReadOnlyFs(afero.NewMemMapFs()) },
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(tc.fs())
			require.NoError(tc.createFileFunc(fileHandler))

			var out bytes.Buffer
			i := &applyCmd{
				fileHandler: fileHandler,
				log:         logger.NewTest(t),
				applier:     constellation.NewApplier(logger.NewTest(t), &nopSpinner{}, constellation.ApplyContextCLI, nil),
			}
			secret, err := i.generateAndPersistMasterSecret(&out)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)

				require.Contains(out.String(), constants.MasterSecretFilename)

				var masterSecret uri.MasterSecret
				require.NoError(fileHandler.ReadJSON(constants.MasterSecretFilename, &masterSecret))
				assert.Equal(masterSecret.Key, secret.Key)
				assert.Equal(masterSecret.Salt, secret.Salt)
			}
		})
	}
}

type stubMerger struct {
	envVar   string
	mergeErr error
}

func (m *stubMerger) mergeConfigs(string, file.Handler) error {
	return m.mergeErr
}

func (m *stubMerger) kubeconfigEnvVar() string {
	return m.envVar
}

func defaultConfigWithExpectedMeasurements(t *testing.T, conf *config.Config, csp cloudprovider.Provider) *config.Config {
	t.Helper()

	conf.RemoveProviderAndAttestationExcept(csp)

	conf.Image = constants.BinaryVersion().String()
	conf.Name = "kubernetes"

	var zone, instanceType, diskType string
	switch csp {
	case cloudprovider.AWS:
		conf.Provider.AWS.Region = "test-region-2"
		conf.Provider.AWS.Zone = "test-zone-2c"
		conf.Provider.AWS.IAMProfileControlPlane = "test-iam-profile"
		conf.Provider.AWS.IAMProfileWorkerNodes = "test-iam-profile"
		conf.Attestation.AWSSEVSNP.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.AWSSEVSNP.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.AWSSEVSNP.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
		zone = "test-zone-2c"
		instanceType = "c6a.xlarge"
		diskType = "gp3"
	case cloudprovider.Azure:
		conf.Provider.Azure.SubscriptionID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.TenantID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.Location = "test-location"
		conf.Provider.Azure.UserAssignedIdentity = "test-identity"
		conf.Provider.Azure.ResourceGroup = "test-resource-group"
		conf.Attestation.AzureSEVSNP.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.AzureSEVSNP.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.AzureSEVSNP.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
		instanceType = "Standard_DC4as_v5"
		diskType = "StandardSSD_LRS"
	case cloudprovider.GCP:
		conf.Provider.GCP.Region = "test-region"
		conf.Provider.GCP.Project = "test-project"
		conf.Provider.GCP.Zone = "test-zone"
		conf.Provider.GCP.ServiceAccountKeyPath = "test-key-path"
		conf.Attestation.GCPSEVSNP.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.GCPSEVSNP.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.GCPSEVSNP.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
		zone = "europe-west3-b"
		instanceType = "n2d-standard-4"
		diskType = "pd-ssd"
	case cloudprovider.QEMU:
		conf.Attestation.QEMUVTPM.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.QEMUVTPM.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.QEMUVTPM.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
	}

	for groupName, group := range conf.NodeGroups {
		group.Zone = zone
		group.InstanceType = instanceType
		group.StateDiskType = diskType
		conf.NodeGroups[groupName] = group
	}

	return conf
}

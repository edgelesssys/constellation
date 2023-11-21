/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/cli/internal/pathprefix"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/helm"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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
	testInitResp := &initproto.InitSuccessResponse{
		Kubeconfig: respKubeconfigBytes,
		OwnerId:    []byte("ownerID"),
		ClusterId:  []byte("clusterID"),
	}
	serviceAccPath := "/test/service-account.json"

	testCases := map[string]struct {
		provider                cloudprovider.Provider
		stateFile               *state.State
		configMutator           func(*config.Config)
		serviceAccKey           *gcpshared.ServiceAccountKey
		initServerAPI           *stubInitServer
		retriable               bool
		masterSecretShouldExist bool
		wantErr                 bool
	}{
		"initialize some gcp instances": {
			provider:      cloudprovider.GCP,
			stateFile:     preInitStateFile(cloudprovider.GCP),
			configMutator: func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			serviceAccKey: gcpServiceAccKey,
			initServerAPI: &stubInitServer{res: []*initproto.InitResponse{{Kind: &initproto.InitResponse_InitSuccess{InitSuccess: testInitResp}}}},
		},
		"initialize some azure instances": {
			provider:      cloudprovider.Azure,
			stateFile:     preInitStateFile(cloudprovider.Azure),
			initServerAPI: &stubInitServer{res: []*initproto.InitResponse{{Kind: &initproto.InitResponse_InitSuccess{InitSuccess: testInitResp}}}},
		},
		"initialize some qemu instances": {
			provider:      cloudprovider.QEMU,
			stateFile:     preInitStateFile(cloudprovider.QEMU),
			initServerAPI: &stubInitServer{res: []*initproto.InitResponse{{Kind: &initproto.InitResponse_InitSuccess{InitSuccess: testInitResp}}}},
		},
		"non retriable error": {
			provider:                cloudprovider.QEMU,
			stateFile:               preInitStateFile(cloudprovider.QEMU),
			initServerAPI:           &stubInitServer{initErr: &nonRetriableError{err: assert.AnError}},
			retriable:               false,
			masterSecretShouldExist: true,
			wantErr:                 true,
		},
		"non retriable error with failed log collection": {
			provider:  cloudprovider.QEMU,
			stateFile: preInitStateFile(cloudprovider.QEMU),
			initServerAPI: &stubInitServer{
				res: []*initproto.InitResponse{
					{
						Kind: &initproto.InitResponse_InitFailure{
							InitFailure: &initproto.InitFailureResponse{
								Error: "error",
							},
						},
					},
					{
						Kind: &initproto.InitResponse_InitFailure{
							InitFailure: &initproto.InitFailureResponse{
								Error: "error",
							},
						},
					},
				},
			},
			retriable:               false,
			masterSecretShouldExist: true,
			wantErr:                 true,
		},
		"invalid state file": {
			provider:      cloudprovider.GCP,
			stateFile:     &state.State{Version: "invalid"},
			configMutator: func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			serviceAccKey: gcpServiceAccKey,
			initServerAPI: &stubInitServer{},
			retriable:     true,
			wantErr:       true,
		},
		"empty state file": {
			provider:      cloudprovider.GCP,
			stateFile:     &state.State{},
			configMutator: func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			serviceAccKey: gcpServiceAccKey,
			initServerAPI: &stubInitServer{},
			retriable:     true,
			wantErr:       true,
		},
		"no state file": {
			provider:      cloudprovider.GCP,
			configMutator: func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			serviceAccKey: gcpServiceAccKey,
			retriable:     true,
			wantErr:       true,
		},
		"init call fails": {
			provider:                cloudprovider.GCP,
			configMutator:           func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			stateFile:               preInitStateFile(cloudprovider.GCP),
			serviceAccKey:           gcpServiceAccKey,
			initServerAPI:           &stubInitServer{initErr: assert.AnError},
			retriable:               false,
			masterSecretShouldExist: true,
			wantErr:                 true,
		},
		"k8s version without v works": {
			provider:      cloudprovider.Azure,
			stateFile:     preInitStateFile(cloudprovider.Azure),
			initServerAPI: &stubInitServer{res: []*initproto.InitResponse{{Kind: &initproto.InitResponse_InitSuccess{InitSuccess: testInitResp}}}},
			configMutator: func(c *config.Config) {
				res, err := versions.NewValidK8sVersion(strings.TrimPrefix(string(versions.Default), "v"), true)
				require.NoError(t, err)
				c.KubernetesVersion = res
			},
		},
		"outdated k8s patch version doesn't work": {
			provider:      cloudprovider.Azure,
			stateFile:     preInitStateFile(cloudprovider.Azure),
			initServerAPI: &stubInitServer{res: []*initproto.InitResponse{{Kind: &initproto.InitResponse_InitSuccess{InitSuccess: testInitResp}}}},
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
			// Networking
			netDialer := testdialer.NewBufconnDialer()
			newDialer := func(atls.Validator) *dialer.Dialer {
				return dialer.New(nil, nil, netDialer)
			}
			serverCreds := atlscredentials.New(nil, nil)
			initServer := grpc.NewServer(grpc.Creds(serverCreds))
			initproto.RegisterAPIServer(initServer, tc.initServerAPI)
			port := strconv.Itoa(constants.BootstrapperPort)
			listener := netDialer.GetListener(net.JoinHostPort("192.0.2.1", port))
			go initServer.Serve(listener)
			defer initServer.GracefulStop()

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
				newHelmClient: func(string, debugLog) (helmApplier, error) {
					return &stubApplier{}, nil
				},
				newDialer: newDialer,
				newKubeUpgrader: func(io.Writer, string, debugLog) (kubernetesUpgrader, error) {
					return &stubKubernetesUpgrader{
						// On init, no attestation config exists yet
						getClusterAttestationConfigErr: k8serrors.NewNotFound(schema.GroupResource{}, ""),
					}, nil
				},
			}

			err := i.apply(cmd, stubAttestationFetcher{}, &stubLicenseClient{}, "test")

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
			assert.Contains(out.String(), hex.EncodeToString([]byte("clusterID")))
			var secret uri.MasterSecret
			assert.NoError(fileHandler.ReadJSON(constants.MasterSecretFilename, &secret))
			assert.NotEmpty(secret.Key)
			assert.NotEmpty(secret.Salt)
		})
	}
}

type stubApplier struct {
	err error
}

func (s stubApplier) PrepareApply(_ *config.Config, _ *state.State, _ helm.Options, _ string, _ uri.MasterSecret) (helm.Applier, bool, error) {
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

func TestGetLogs(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		resp         initproto.API_InitClient
		fh           file.Handler
		wantedOutput []byte
		wantErr      bool
	}{
		"success": {
			resp:         stubInitClient{res: bytes.NewReader([]byte("asdf"))},
			fh:           file.NewHandler(afero.NewMemMapFs()),
			wantedOutput: []byte("asdf"),
		},
		"receive error": {
			resp:    stubInitClient{err: someErr},
			fh:      file.NewHandler(afero.NewMemMapFs()),
			wantErr: true,
		},
		"nil log": {
			resp:    stubInitClient{res: bytes.NewReader([]byte{1}), setResNil: true},
			fh:      file.NewHandler(afero.NewMemMapFs()),
			wantErr: true,
		},
		"failed write": {
			resp:    stubInitClient{res: bytes.NewReader([]byte("asdf"))},
			fh:      file.NewHandler(afero.NewReadOnlyFs(afero.NewMemMapFs())),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			doer := initDoer{
				fh:  tc.fh,
				log: logger.NewTest(t),
			}

			err := doer.getLogs(tc.resp)

			if tc.wantErr {
				assert.Error(err)
			}

			text, err := tc.fh.Read(constants.ErrorLog)

			if tc.wantedOutput == nil {
				assert.Error(err)
			}

			assert.Equal(tc.wantedOutput, text)
		})
	}
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
				Server: "https://192.0.2.1:6443",
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

	ownerID := hex.EncodeToString(resp.GetInitSuccess().GetOwnerId())
	clusterID := hex.EncodeToString(resp.GetInitSuccess().GetClusterId())
	measurementSalt := []byte{0x41}

	expectedStateFile := &state.State{
		Version: state.Version1,
		ClusterValues: state.ClusterValues{
			ClusterID:       clusterID,
			OwnerID:         ownerID,
			MeasurementSalt: []byte{0x41},
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
	}
	err = i.writeInitOutput(stateFile, resp.GetInitSuccess(), false, &out, measurementSalt)
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
	err = i.writeInitOutput(stateFile, resp.GetInitSuccess(), true, &out, measurementSalt)
	require.NoError(err)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), i.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	out.Reset()
	// File is written to current working dir, we simply pass the workspace for generating readable user output
	require.NoError(afs.Remove(constants.AdminConfFilename))
	i.flags.pathPrefixer = pathprefix.PathPrefixer{}

	// test config merging
	err = i.writeInitOutput(stateFile, resp.GetInitSuccess(), true, &out, measurementSalt)
	require.NoError(err)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), constants.AdminConfFilename)
	assert.Contains(out.String(), "Constellation kubeconfig merged with default config")
	assert.Contains(out.String(), "You can now connect to your cluster")
	out.Reset()
	require.NoError(afs.Remove(constants.AdminConfFilename))

	// test config merging with env vars set
	i.merger = &stubMerger{envVar: "/some/path/to/kubeconfig"}
	err = i.writeInitOutput(stateFile, resp.GetInitSuccess(), true, &out, measurementSalt)
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
			createFileFunc: func(handler file.Handler) error { return nil },
			fs:             afero.NewMemMapFs,
			wantErr:        false,
		},
		"file not writeable": {
			createFileFunc: func(handler file.Handler) error { return nil },
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

func TestAttestation(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	initServerAPI := &stubInitServer{res: []*initproto.InitResponse{
		{
			Kind: &initproto.InitResponse_InitSuccess{
				InitSuccess: &initproto.InitSuccessResponse{
					Kubeconfig: []byte("kubeconfig"),
					OwnerId:    []byte("ownerID"),
					ClusterId:  []byte("clusterID"),
				},
			},
		},
	}}

	existingStateFile := &state.State{Version: state.Version1, Infrastructure: state.Infrastructure{ClusterEndpoint: "192.0.2.4"}}

	netDialer := testdialer.NewBufconnDialer()

	issuer := &testIssuer{
		Getter: variant.QEMUVTPM{},
		pcrs: map[uint32][]byte{
			0: bytes.Repeat([]byte{0xFF}, 32),
			1: bytes.Repeat([]byte{0xFF}, 32),
			2: bytes.Repeat([]byte{0xFF}, 32),
			3: bytes.Repeat([]byte{0xFF}, 32),
		},
	}
	serverCreds := atlscredentials.New(issuer, nil)
	initServer := grpc.NewServer(grpc.Creds(serverCreds))
	initproto.RegisterAPIServer(initServer, initServerAPI)
	port := strconv.Itoa(constants.BootstrapperPort)
	listener := netDialer.GetListener(net.JoinHostPort("192.0.2.4", port))
	go initServer.Serve(listener)
	defer initServer.GracefulStop()

	cmd := NewInitCmd()
	cmd.Flags().String("workspace", "", "") // register persistent flag manually
	cmd.Flags().Bool("force", true, "")     // register persistent flag manually
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	fs := afero.NewMemMapFs()
	fileHandler := file.NewHandler(fs)
	require.NoError(existingStateFile.WriteToFile(fileHandler, constants.StateFilename))

	cfg := config.Default()
	cfg.Image = constants.BinaryVersion().String()
	cfg.RemoveProviderAndAttestationExcept(cloudprovider.QEMU)
	cfg.Attestation.QEMUVTPM.Measurements[0] = measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength)
	cfg.Attestation.QEMUVTPM.Measurements[1] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
	cfg.Attestation.QEMUVTPM.Measurements[2] = measurements.WithAllBytes(0x22, measurements.Enforce, measurements.PCRMeasurementLength)
	cfg.Attestation.QEMUVTPM.Measurements[3] = measurements.WithAllBytes(0x33, measurements.Enforce, measurements.PCRMeasurementLength)
	cfg.Attestation.QEMUVTPM.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
	cfg.Attestation.QEMUVTPM.Measurements[9] = measurements.WithAllBytes(0x99, measurements.Enforce, measurements.PCRMeasurementLength)
	cfg.Attestation.QEMUVTPM.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
	require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, cfg, file.OptNone))

	newDialer := func(v atls.Validator) *dialer.Dialer {
		validator := &testValidator{
			Getter: variant.QEMUVTPM{},
			pcrs:   cfg.GetAttestationConfig().GetMeasurements(),
		}
		return dialer.New(nil, validator, netDialer)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	cmd.SetContext(ctx)

	i := &applyCmd{
		fileHandler: fileHandler,
		spinner:     &nopSpinner{},
		merger:      &stubMerger{},
		log:         logger.NewTest(t),
		newKubeUpgrader: func(io.Writer, string, debugLog) (kubernetesUpgrader, error) {
			return &stubKubernetesUpgrader{}, nil
		},
		newDialer: newDialer,
	}
	_, err := i.runInit(cmd, cfg, existingStateFile)
	assert.Error(err)
	// make sure the error is actually a TLS handshake error
	assert.Contains(err.Error(), "transport: authentication handshake failed")
	if validationErr, ok := err.(*config.ValidationError); ok {
		t.Log(validationErr.LongMessage())
	}
}

type testValidator struct {
	variant.Getter
	pcrs measurements.M
}

func (v *testValidator) Validate(_ context.Context, attDoc []byte, _ []byte) ([]byte, error) {
	var attestation struct {
		UserData []byte
		PCRs     map[uint32][]byte
	}
	if err := json.Unmarshal(attDoc, &attestation); err != nil {
		return nil, err
	}

	for k, pcr := range v.pcrs {
		if !bytes.Equal(attestation.PCRs[k], pcr.Expected[:]) {
			return nil, errors.New("invalid PCR value")
		}
	}
	return attestation.UserData, nil
}

type testIssuer struct {
	variant.Getter
	pcrs map[uint32][]byte
}

func (i *testIssuer) Issue(_ context.Context, userData []byte, _ []byte) ([]byte, error) {
	return json.Marshal(
		struct {
			UserData []byte
			PCRs     map[uint32][]byte
		}{
			UserData: userData,
			PCRs:     i.pcrs,
		},
	)
}

type stubInitServer struct {
	res     []*initproto.InitResponse
	initErr error

	initproto.UnimplementedAPIServer
}

func (s *stubInitServer) Init(_ *initproto.InitRequest, stream initproto.API_InitServer) error {
	for _, r := range s.res {
		_ = stream.Send(r)
	}
	return s.initErr
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
		conf.Attestation.GCPSEVES.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.GCPSEVES.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.GCPSEVES.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
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

type stubLicenseClient struct{}

func (c *stubLicenseClient) QuotaCheck(_ context.Context, _ license.QuotaCheckRequest) (license.QuotaCheckResponse, error) {
	return license.QuotaCheckResponse{
		Quota: 25,
	}, nil
}

type stubInitClient struct {
	res       io.Reader
	err       error
	setResNil bool
	grpc.ClientStream
}

func (c stubInitClient) Recv() (*initproto.InitResponse, error) {
	if c.err != nil {
		return &initproto.InitResponse{}, c.err
	}

	text := make([]byte, 1024)
	n, err := c.res.Read(text)
	text = text[:n]

	res := &initproto.InitResponse{
		Kind: &initproto.InitResponse_Log{
			Log: &initproto.LogResponseType{
				Log: text,
			},
		},
	}
	if c.setResNil {
		res = &initproto.InitResponse{
			Kind: &initproto.InitResponse_Log{
				Log: &initproto.LogResponseType{
					Log: nil,
				},
			},
		}
	}

	return res, err
}

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
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestInitArgumentValidation(t *testing.T) {
	assert := assert.New(t)

	cmd := NewInitCmd()
	assert.NoError(cmd.ValidateArgs(nil))
	assert.Error(cmd.ValidateArgs([]string{"something"}))
	assert.Error(cmd.ValidateArgs([]string{"sth", "sth"}))
}

func TestInitialize(t *testing.T) {
	gcpServiceAccKey := &gcpshared.ServiceAccountKey{
		Type: "service_account",
	}
	testInitResp := &initproto.InitResponse{
		Kubeconfig: []byte("kubeconfig"),
		OwnerId:    []byte("ownerID"),
		ClusterId:  []byte("clusterID"),
	}
	serviceAccPath := "/test/service-account.json"
	someErr := errors.New("failed")

	testCases := map[string]struct {
		provider                cloudprovider.Provider
		idFile                  *clusterid.File
		configMutator           func(*config.Config)
		serviceAccKey           *gcpshared.ServiceAccountKey
		initServerAPI           *stubInitServer
		retriable               bool
		masterSecretShouldExist bool
		wantErr                 bool
	}{
		"initialize some gcp instances": {
			provider:      cloudprovider.GCP,
			idFile:        &clusterid.File{IP: "192.0.2.1"},
			configMutator: func(c *config.Config) { c.Provider.GCP.ServiceAccountKeyPath = serviceAccPath },
			serviceAccKey: gcpServiceAccKey,
			initServerAPI: &stubInitServer{initResp: testInitResp},
		},
		"initialize some azure instances": {
			provider:      cloudprovider.Azure,
			idFile:        &clusterid.File{IP: "192.0.2.1"},
			initServerAPI: &stubInitServer{initResp: testInitResp},
		},
		"initialize some qemu instances": {
			provider:      cloudprovider.QEMU,
			idFile:        &clusterid.File{IP: "192.0.2.1"},
			initServerAPI: &stubInitServer{initResp: testInitResp},
		},
		"non retriable error": {
			provider:                cloudprovider.QEMU,
			idFile:                  &clusterid.File{IP: "192.0.2.1"},
			initServerAPI:           &stubInitServer{initErr: &nonRetriableError{someErr}},
			retriable:               false,
			masterSecretShouldExist: true,
			wantErr:                 true,
		},
		"empty id file": {
			provider:      cloudprovider.GCP,
			idFile:        &clusterid.File{},
			initServerAPI: &stubInitServer{},
			retriable:     true,
			wantErr:       true,
		},
		"no id file": {
			provider:  cloudprovider.GCP,
			retriable: true,
			wantErr:   true,
		},
		"init call fails": {
			provider:      cloudprovider.GCP,
			idFile:        &clusterid.File{IP: "192.0.2.1"},
			initServerAPI: &stubInitServer{initErr: someErr},
			retriable:     true,
			wantErr:       true,
		},
		"k8s version without v works": {
			provider:      cloudprovider.Azure,
			idFile:        &clusterid.File{IP: "192.0.2.1"},
			initServerAPI: &stubInitServer{initResp: testInitResp},
			configMutator: func(c *config.Config) { c.KubernetesVersion = strings.TrimPrefix(string(versions.Default), "v") },
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

			// Flags
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")                        // register persistent flag manually

			// File system preparation
			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			config := defaultConfigWithExpectedMeasurements(t, config.Default(), tc.provider)
			if tc.configMutator != nil {
				tc.configMutator(config)
			}
			require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, config, file.OptNone))
			if tc.idFile != nil {
				tc.idFile.CloudProvider = tc.provider
				require.NoError(fileHandler.WriteJSON(constants.ClusterIDsFileName, tc.idFile, file.OptNone))
			}
			if tc.serviceAccKey != nil {
				require.NoError(fileHandler.WriteJSON(serviceAccPath, tc.serviceAccKey, file.OptNone))
			}

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
			defer cancel()
			cmd.SetContext(ctx)
			i := &initCmd{log: logger.NewTest(t), spinner: &nopSpinner{}}
			err := i.initialize(cmd, newDialer, fileHandler, &stubLicenseClient{})

			if tc.wantErr {
				assert.Error(err)
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

func TestWriteOutput(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := &initproto.InitResponse{
		OwnerId:    []byte("ownerID"),
		ClusterId:  []byte("clusterID"),
		Kubeconfig: []byte("kubeconfig"),
	}

	ownerID := hex.EncodeToString(resp.OwnerId)
	clusterID := hex.EncodeToString(resp.ClusterId)

	expectedIDFile := clusterid.File{
		ClusterID: clusterID,
		OwnerID:   ownerID,
		IP:        "cluster-ip",
		UID:       "test-uid",
	}

	var out bytes.Buffer
	testFs := afero.NewMemMapFs()
	fileHandler := file.NewHandler(testFs)

	idFile := clusterid.File{
		UID: "test-uid",
		IP:  "cluster-ip",
	}
	i := &initCmd{
		log:    logger.NewTest(t),
		merger: &stubMerger{},
	}
	err := i.writeOutput(idFile, resp, false, &out, fileHandler)
	require.NoError(err)
	// assert.Contains(out.String(), ownerID)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), constants.AdminConfFilename)

	afs := afero.Afero{Fs: testFs}
	adminConf, err := afs.ReadFile(constants.AdminConfFilename)
	assert.NoError(err)
	assert.Equal(string(resp.Kubeconfig), string(adminConf))

	idsFile, err := afs.ReadFile(constants.ClusterIDsFileName)
	assert.NoError(err)
	var testIDFile clusterid.File
	err = json.Unmarshal(idsFile, &testIDFile)
	assert.NoError(err)
	assert.Equal(expectedIDFile, testIDFile)

	// test config merging
	out.Reset()
	require.NoError(afs.Remove(constants.AdminConfFilename))
	err = i.writeOutput(idFile, resp, true, &out, fileHandler)
	require.NoError(err)
	// assert.Contains(out.String(), ownerID)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), constants.AdminConfFilename)
	assert.Contains(out.String(), "Constellation kubeconfig merged with default config")
	assert.Contains(out.String(), "You can now connect to your cluster")

	// test config merging with env vars set
	i.merger = &stubMerger{envVar: "/some/path/to/kubeconfig"}
	out.Reset()
	require.NoError(afs.Remove(constants.AdminConfFilename))
	err = i.writeOutput(idFile, resp, true, &out, fileHandler)
	require.NoError(err)
	// assert.Contains(out.String(), ownerID)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), constants.AdminConfFilename)
	assert.Contains(out.String(), "Constellation kubeconfig merged with default config")
	assert.Contains(out.String(), "Warning: KUBECONFIG environment variable is set")
}

func TestReadOrGenerateMasterSecret(t *testing.T) {
	testCases := map[string]struct {
		filename       string
		createFileFunc func(handler file.Handler) error
		fs             func() afero.Fs
		wantErr        bool
	}{
		"file with secret exists": {
			filename: "someSecret",
			fs:       afero.NewMemMapFs,
			createFileFunc: func(handler file.Handler) error {
				return handler.WriteJSON(
					"someSecret",
					uri.MasterSecret{Key: []byte("constellation-master-secret"), Salt: []byte("constellation-32Byte-length-salt")},
					file.OptNone,
				)
			},
			wantErr: false,
		},
		"no file given": {
			filename:       "",
			createFileFunc: func(handler file.Handler) error { return nil },
			fs:             afero.NewMemMapFs,
			wantErr:        false,
		},
		"file does not exist": {
			filename:       "nonExistingSecret",
			createFileFunc: func(handler file.Handler) error { return nil },
			fs:             afero.NewMemMapFs,
			wantErr:        true,
		},
		"file is empty": {
			filename: "emptySecret",
			createFileFunc: func(handler file.Handler) error {
				return handler.Write("emptySecret", []byte{}, file.OptNone)
			},
			fs:      afero.NewMemMapFs,
			wantErr: true,
		},
		"salt too short": {
			filename: "shortSecret",
			createFileFunc: func(handler file.Handler) error {
				return handler.WriteJSON(
					"shortSecret",
					uri.MasterSecret{Key: []byte("constellation-master-secret"), Salt: []byte("short")},
					file.OptNone,
				)
			},
			fs:      afero.NewMemMapFs,
			wantErr: true,
		},
		"key too short": {
			filename: "shortSecret",
			createFileFunc: func(handler file.Handler) error {
				return handler.WriteJSON(
					"shortSecret",
					uri.MasterSecret{Key: []byte("short"), Salt: []byte("constellation-32Byte-length-salt")},
					file.OptNone,
				)
			},
			fs:      afero.NewMemMapFs,
			wantErr: true,
		},
		"invalid file content": {
			filename: "unencodedSecret",
			createFileFunc: func(handler file.Handler) error {
				return handler.Write("unencodedSecret", []byte("invalid-constellation-master-secret"), file.OptNone)
			},
			fs:      afero.NewMemMapFs,
			wantErr: true,
		},
		"file not writeable": {
			filename:       "",
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
			i := &initCmd{log: logger.NewTest(t)}
			secret, err := i.readOrGenerateMasterSecret(&out, fileHandler, tc.filename)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)

				if tc.filename == "" {
					require.Contains(out.String(), constants.MasterSecretFilename)
					filename := strings.Split(out.String(), "./")
					tc.filename = strings.Trim(filename[1], "\n")
				}

				var masterSecret uri.MasterSecret
				require.NoError(fileHandler.ReadJSON(tc.filename, &masterSecret))
				assert.Equal(masterSecret.Key, secret.Key)
				assert.Equal(masterSecret.Salt, secret.Salt)
			}
		})
	}
}

func TestAttestation(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	initServerAPI := &stubInitServer{initResp: &initproto.InitResponse{
		Kubeconfig: []byte("kubeconfig"),
		OwnerId:    []byte("ownerID"),
		ClusterId:  []byte("clusterID"),
	}}
	existingIDFile := &clusterid.File{IP: "192.0.2.4", CloudProvider: cloudprovider.QEMU}

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
	cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
	cmd.Flags().Bool("force", true, "")                        // register persistent flag manually
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	fs := afero.NewMemMapFs()
	fileHandler := file.NewHandler(fs)
	require.NoError(fileHandler.WriteJSON(constants.ClusterIDsFileName, existingIDFile, file.OptNone))

	cfg := config.Default()
	cfg.Image = "image"
	cfg.RemoveProviderExcept(cloudprovider.QEMU)
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

	i := &initCmd{log: logger.NewTest(t), spinner: &nopSpinner{}}
	err := i.initialize(cmd, newDialer, fileHandler, &stubLicenseClient{})
	assert.Error(err)
	// make sure the error is actually a TLS handshake error
	assert.Contains(err.Error(), "transport: authentication handshake failed")
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
	initResp *initproto.InitResponse
	initErr  error

	initproto.UnimplementedAPIServer
}

func (s *stubInitServer) Init(_ context.Context, _ *initproto.InitRequest) (*initproto.InitResponse, error) {
	return s.initResp, s.initErr
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

	conf.Image = "v" + constants.VersionInfo()
	conf.Name = "kubernetes"

	switch csp {
	case cloudprovider.Azure:
		conf.Provider.Azure.SubscriptionID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.TenantID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.Location = "test-location"
		conf.Provider.Azure.UserAssignedIdentity = "test-identity"
		conf.Provider.Azure.ResourceGroup = "test-resource-group"
		conf.Provider.Azure.AppClientID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.ClientSecretValue = "test-client-secret"
		conf.Attestation.AzureSEVSNP.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.AzureSEVSNP.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.AzureSEVSNP.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
	case cloudprovider.GCP:
		conf.Provider.GCP.Region = "test-region"
		conf.Provider.GCP.Project = "test-project"
		conf.Provider.GCP.Zone = "test-zone"
		conf.Provider.GCP.ServiceAccountKeyPath = "test-key-path"
		conf.Attestation.GCPSEVES.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.GCPSEVES.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.GCPSEVES.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
	case cloudprovider.QEMU:
		conf.Attestation.QEMUVTPM.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.QEMUVTPM.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.QEMUVTPM.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
	}

	conf.RemoveProviderExcept(csp)
	return conf
}

type stubLicenseClient struct{}

func (c *stubLicenseClient) QuotaCheck(_ context.Context, _ license.QuotaCheckRequest) (license.QuotaCheckResponse, error) {
	return license.QuotaCheckResponse{
		Quota: 25,
	}, nil
}

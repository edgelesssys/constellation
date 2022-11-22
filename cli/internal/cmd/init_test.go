/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/edgelesssys/constellation/v2/internal/oid"
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
		"empty id file": {
			provider:      cloudprovider.GCP,
			idFile:        &clusterid.File{},
			initServerAPI: &stubInitServer{},
			wantErr:       true,
		},
		"no id file": {
			provider: cloudprovider.GCP,
			wantErr:  true,
		},
		"init call fails": {
			provider:      cloudprovider.GCP,
			idFile:        &clusterid.File{IP: "192.0.2.1"},
			initServerAPI: &stubInitServer{initErr: someErr},
			wantErr:       true,
		},
		"fail missing enforced PCR": {
			provider: cloudprovider.GCP,
			idFile:   &clusterid.File{IP: "192.0.2.1"},
			configMutator: func(c *config.Config) {
				c.Provider.GCP.EnforcedMeasurements = append(c.Provider.GCP.EnforcedMeasurements, 10)
			},
			serviceAccKey: gcpServiceAccKey,
			initServerAPI: &stubInitServer{initResp: testInitResp},
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Networking
			netDialer := testdialer.NewBufconnDialer()
			newDialer := func(*cloudcmd.Validator) *dialer.Dialer {
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

			err := initialize(cmd, newDialer, fileHandler, &stubLicenseClient{}, nopSpinner{})

			if tc.wantErr {
				assert.Error(err)
				if !tc.masterSecretShouldExist {
					_, err = fileHandler.Stat(constants.MasterSecretFilename)
					assert.Error(err)
				}
				return
			}
			require.NoError(err)
			// assert.Contains(out.String(), base64.StdEncoding.EncodeToString([]byte("ownerID")))
			assert.Contains(out.String(), base64.StdEncoding.EncodeToString([]byte("clusterID")))
			var secret masterSecret
			assert.NoError(fileHandler.ReadJSON(constants.MasterSecretFilename, &secret))
			assert.NotEmpty(secret.Key)
			assert.NotEmpty(secret.Salt)
		})
	}
}

func TestWriteOutput(t *testing.T) {
	assert := assert.New(t)

	resp := &initproto.InitResponse{
		OwnerId:    []byte("ownerID"),
		ClusterId:  []byte("clusterID"),
		Kubeconfig: []byte("kubeconfig"),
	}

	ownerID := base64.StdEncoding.EncodeToString(resp.OwnerId)
	clusterID := base64.StdEncoding.EncodeToString(resp.ClusterId)

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
	err := writeOutput(idFile, resp, &out, fileHandler)
	assert.NoError(err)
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
					masterSecret{Key: []byte("constellation-master-secret"), Salt: []byte("constellation-32Byte-length-salt")},
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
					masterSecret{Key: []byte("constellation-master-secret"), Salt: []byte("short")},
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
					masterSecret{Key: []byte("short"), Salt: []byte("constellation-32Byte-length-salt")},
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
			secret, err := readOrGenerateMasterSecret(&out, fileHandler, tc.filename)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)

				if tc.filename == "" {
					require.Contains(out.String(), constants.MasterSecretFilename)
					filename := strings.Split(out.String(), "./")
					tc.filename = strings.Trim(filename[1], "\n")
				}

				var masterSecret masterSecret
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
	newDialer := func(v *cloudcmd.Validator) *dialer.Dialer {
		validator := &testValidator{
			Getter: oid.QEMU{},
			pcrs:   v.PCRS(),
		}
		return dialer.New(nil, validator, netDialer)
	}

	issuer := &testIssuer{
		Getter: oid.QEMU{},
		pcrs: measurements.M{
			0: measurements.PCRWithAllBytes(0xFF),
			1: measurements.PCRWithAllBytes(0xFF),
			2: measurements.PCRWithAllBytes(0xFF),
			3: measurements.PCRWithAllBytes(0xFF),
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
	cfg.Provider.QEMU.Measurements[0] = measurements.PCRWithAllBytes(0x00)
	cfg.Provider.QEMU.Measurements[1] = measurements.PCRWithAllBytes(0x11)
	cfg.Provider.QEMU.Measurements[2] = measurements.PCRWithAllBytes(0x22)
	cfg.Provider.QEMU.Measurements[3] = measurements.PCRWithAllBytes(0x33)
	cfg.Provider.QEMU.Measurements[4] = measurements.PCRWithAllBytes(0x44)
	cfg.Provider.QEMU.Measurements[9] = measurements.PCRWithAllBytes(0x99)
	cfg.Provider.QEMU.Measurements[12] = measurements.PCRWithAllBytes(0xcc)
	require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, cfg, file.OptNone))

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	cmd.SetContext(ctx)

	err := initialize(cmd, newDialer, fileHandler, &stubLicenseClient{}, nopSpinner{})
	assert.Error(err)
	// make sure the error is actually a TLS handshake error
	assert.Contains(err.Error(), "transport: authentication handshake failed")
}

type testValidator struct {
	oid.Getter
	pcrs measurements.M
}

func (v *testValidator) Validate(attDoc []byte, nonce []byte) ([]byte, error) {
	var attestation struct {
		UserData []byte
		PCRs     measurements.M
	}
	if err := json.Unmarshal(attDoc, &attestation); err != nil {
		return nil, err
	}

	for k, pcr := range v.pcrs {
		if !bytes.Equal(attestation.PCRs[k], pcr) {
			return nil, errors.New("invalid PCR value")
		}
	}
	return attestation.UserData, nil
}

type testIssuer struct {
	oid.Getter
	pcrs measurements.M
}

func (i *testIssuer) Issue(userData []byte, nonce []byte) ([]byte, error) {
	return json.Marshal(
		struct {
			UserData []byte
			PCRs     measurements.M
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

func (s *stubInitServer) Init(ctx context.Context, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	return s.initResp, s.initErr
}

func defaultConfigWithExpectedMeasurements(t *testing.T, conf *config.Config, csp cloudprovider.Provider) *config.Config {
	t.Helper()

	conf.Image = "image"

	switch csp {
	case cloudprovider.Azure:
		conf.Provider.Azure.SubscriptionID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.TenantID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.Location = "test-location"
		conf.Provider.Azure.UserAssignedIdentity = "test-identity"
		conf.Provider.Azure.ResourceGroup = "test-resource-group"
		conf.Provider.Azure.AppClientID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.ClientSecretValue = "test-client-secret"
		conf.Provider.Azure.Measurements[4] = measurements.PCRWithAllBytes(0x44)
		conf.Provider.Azure.Measurements[9] = measurements.PCRWithAllBytes(0x11)
		conf.Provider.Azure.Measurements[12] = measurements.PCRWithAllBytes(0xcc)
	case cloudprovider.GCP:
		conf.Provider.GCP.Region = "test-region"
		conf.Provider.GCP.Project = "test-project"
		conf.Provider.GCP.Zone = "test-zone"
		conf.Provider.GCP.ServiceAccountKeyPath = "test-key-path"
		conf.Provider.GCP.Measurements[4] = measurements.PCRWithAllBytes(0x44)
		conf.Provider.GCP.Measurements[9] = measurements.PCRWithAllBytes(0x11)
		conf.Provider.GCP.Measurements[12] = measurements.PCRWithAllBytes(0xcc)
	case cloudprovider.QEMU:
		conf.Provider.QEMU.Measurements[4] = measurements.PCRWithAllBytes(0x44)
		conf.Provider.QEMU.Measurements[9] = measurements.PCRWithAllBytes(0x11)
		conf.Provider.QEMU.Measurements[12] = measurements.PCRWithAllBytes(0xcc)
	}

	conf.RemoveProviderExcept(csp)
	return conf
}

type stubLicenseClient struct{}

func (c *stubLicenseClient) QuotaCheck(ctx context.Context, checkRequest license.QuotaCheckRequest) (license.QuotaCheckResponse, error) {
	return license.QuotaCheckResponse{
		Quota: 25,
	}, nil
}

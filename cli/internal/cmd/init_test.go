package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/state"
	wgquick "github.com/nmiculinic/wg-quick-go"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitArgumentValidation(t *testing.T) {
	assert := assert.New(t)

	cmd := NewInitCmd()
	assert.NoError(cmd.ValidateArgs(nil))
	assert.Error(cmd.ValidateArgs([]string{"something"}))
	assert.Error(cmd.ValidateArgs([]string{"sth", "sth"}))
}

func TestInitialize(t *testing.T) {
	testKey := base64.StdEncoding.EncodeToString([]byte("32bytesWireGuardKeyForTheTesting"))
	testGcpState := state.ConstellationState{
		CloudProvider: "GCP",
		GCPNodes: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		GCPCoordinators: cloudtypes.Instances{
			"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
	}
	testAzureState := state.ConstellationState{
		CloudProvider: "Azure",
		AzureNodes: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureCoordinators: cloudtypes.Instances{
			"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureResourceGroup: "test",
	}
	testQemuState := state.ConstellationState{
		CloudProvider: "QEMU",
		QEMUNodes: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		QEMUCoordinators: cloudtypes.Instances{
			"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
	}
	testActivationResps := []fakeActivationRespMessage{
		{log: "testlog1"},
		{log: "testlog2"},
		{
			kubeconfig:        "kubeconfig",
			clientVpnIp:       "192.0.2.2",
			coordinatorVpnKey: testKey,
			ownerID:           "ownerID",
			clusterID:         "clusterID",
		},
		{log: "testlog3"},
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		existingState         state.ConstellationState
		client                protoClient
		serviceAccountCreator stubServiceAccountCreator
		waiter                statusWaiter
		privKey               string
		vpnHandler            vpnHandler
		initVPN               bool
		wantErr               bool
	}{
		"initialize some gcp instances": {
			existingState: testGcpState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     &stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{},
			privKey:    testKey,
		},
		"initialize some azure instances": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     &stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{},
			privKey:    testKey,
		},
		"initialize some qemu instances": {
			existingState: testQemuState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     &stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{},
			privKey:    testKey,
		},
		"initialize vpn": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     &stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{},
			initVPN:    true,
			privKey:    testKey,
		},
		"invalid initialize vpn": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     &stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{applyErr: someErr},
			initVPN:    true,
			privKey:    testKey,
			wantErr:    true,
		},
		"invalid create vpn config": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     &stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{createErr: someErr},
			initVPN:    true,
			privKey:    testKey,
			wantErr:    true,
		},
		"invalid write vpn config": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     &stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{marshalErr: someErr},
			initVPN:    true,
			privKey:    testKey,
			wantErr:    true,
		},
		"no state exists": {
			existingState: state.ConstellationState{},
			client:        &stubProtoClient{},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"no instances to pick one": {
			existingState: state.ConstellationState{GCPNodes: cloudtypes.Instances{}},
			client:        &stubProtoClient{},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"public key to short": {
			existingState: testGcpState,
			client:        &stubProtoClient{},
			waiter:        &stubStatusWaiter{},
			privKey:       base64.StdEncoding.EncodeToString([]byte("tooShortKey")),
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"public key to long": {
			existingState: testGcpState,
			client:        &stubProtoClient{},
			waiter:        &stubStatusWaiter{},
			privKey:       base64.StdEncoding.EncodeToString([]byte("thisWireguardKeyIsToLongAndHasTooManyBytes")),
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"public key not base64": {
			existingState: testGcpState,
			client:        &stubProtoClient{},
			waiter:        &stubStatusWaiter{},
			privKey:       "this is not base64 encoded",
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail Connect": {
			existingState: testGcpState,
			client:        &stubProtoClient{connectErr: someErr},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail Activate": {
			existingState: testGcpState,
			client:        &stubProtoClient{activateErr: someErr},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail respClient WriteLogStream": {
			existingState: testGcpState,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{writeLogStreamErr: someErr}},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail respClient getKubeconfig": {
			existingState: testGcpState,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getKubeconfigErr: someErr}},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail respClient getCoordinatorVpnKey": {
			existingState: testGcpState,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getCoordinatorVpnKeyErr: someErr}},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail respClient getClientVpnIp": {
			existingState: testGcpState,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getClientVpnIpErr: someErr}},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail respClient getOwnerID": {
			existingState: testGcpState,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getOwnerIDErr: someErr}},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail respClient getClusterID": {
			existingState: testGcpState,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getClusterIDErr: someErr}},
			waiter:        &stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail to wait for required status": {
			existingState: testGcpState,
			client:        &stubProtoClient{},
			waiter:        &stubStatusWaiter{waitForAllErr: someErr},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			wantErr:       true,
		},
		"fail to create service account": {
			existingState:         testGcpState,
			client:                &stubProtoClient{},
			serviceAccountCreator: stubServiceAccountCreator{createErr: someErr},
			waiter:                &stubStatusWaiter{},
			privKey:               testKey,
			vpnHandler:            &stubVPNHandler{},
			wantErr:               true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewInitCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)
			cmd.Flags().String("config", "", "") // register persisten flag manually
			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			require.NoError(fileHandler.WriteJSON(constants.StateFilename, tc.existingState, file.OptNone))

			// Write key file to filesystem and set path in flag.
			require.NoError(afero.Afero{Fs: fs}.WriteFile("privK", []byte(tc.privKey), 0o600))
			require.NoError(cmd.Flags().Set("privatekey", "privK"))
			if tc.initVPN {
				require.NoError(cmd.Flags().Set("wg-autoconfig", "true"))
			}

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
			defer cancel()

			err := initialize(ctx, cmd, tc.client, &tc.serviceAccountCreator, fileHandler, tc.waiter, tc.vpnHandler)

			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.Equal(tc.initVPN, tc.vpnHandler.(*stubVPNHandler).configured)
				assert.Contains(out.String(), "192.0.2.2")
				assert.Contains(out.String(), "ownerID")
				assert.Contains(out.String(), "clusterID")
			}
		})
	}
}

func TestWriteOutput(t *testing.T) {
	assert := assert.New(t)

	result := activationResult{
		clientVpnIP:       "foo-qq",
		coordinatorPubKey: "bar-qq",
		coordinatorPubIP:  "baz-qq",
		kubeconfig:        "foo-bar-baz-qq",
	}
	var out bytes.Buffer
	testFs := afero.NewMemMapFs()
	fileHandler := file.NewHandler(testFs)

	err := result.writeOutput(&out, fileHandler)
	assert.NoError(err)
	assert.Contains(out.String(), result.clientVpnIP)
	assert.Contains(out.String(), result.coordinatorPubIP)
	assert.Contains(out.String(), result.coordinatorPubKey)

	afs := afero.Afero{Fs: testFs}
	adminConf, err := afs.ReadFile(constants.AdminConfFilename)
	assert.NoError(err)
	assert.Equal(result.kubeconfig, string(adminConf))
}

func TestIpsToEndpoints(t *testing.T) {
	assert := assert.New(t)

	ips := []string{"192.0.2.1", "192.0.2.2", "", "192.0.2.3"}
	port := "8080"
	endpoints := ipsToEndpoints(ips, port)
	assert.Equal([]string{"192.0.2.1:8080", "192.0.2.2:8080", "192.0.2.3:8080"}, endpoints)
}

func TestInitCompletion(t *testing.T) {
	testCases := map[string]struct {
		args        []string
		toComplete  string
		wantResult  []string
		wantShellCD cobra.ShellCompDirective
	}{
		"first arg": {
			args:        []string{},
			toComplete:  "hello",
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveDefault,
		},
		"secnod arg": {
			args:        []string{"23"},
			toComplete:  "/test/h",
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveError,
		},
		"third arg": {
			args:        []string{"./file", "sth"},
			toComplete:  "./file",
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := &cobra.Command{}
			result, shellCD := initCompletion(cmd, tc.args, tc.toComplete)
			assert.Equal(tc.wantResult, result)
			assert.Equal(tc.wantShellCD, shellCD)
		})
	}
}

func TestReadOrGenerateVPNKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	testKey := []byte(base64.StdEncoding.EncodeToString([]byte("32bytesWireGuardKeyForTheTesting")))
	fileHandler := file.NewHandler(afero.NewMemMapFs())
	require.NoError(fileHandler.Write("testKey", testKey, file.OptNone))

	privK, pubK, err := readOrGenerateVPNKey(fileHandler, "testKey")
	assert.NoError(err)
	assert.Equal(testKey, privK)
	assert.NotEmpty(pubK)

	// no path provided
	privK, pubK, err = readOrGenerateVPNKey(fileHandler, "")
	assert.NoError(err)
	assert.NotEmpty(privK)
	assert.NotEmpty(pubK)
}

func TestReadOrGenerateMasterSecret(t *testing.T) {
	testCases := map[string]struct {
		filename    string
		filecontent string
		createFile  bool
		fs          func() afero.Fs
		wantErr     bool
	}{
		"file with secret exists": {
			filename:    "someSecret",
			filecontent: base64.StdEncoding.EncodeToString([]byte("ConstellationSecret")),
			createFile:  true,
			fs:          afero.NewMemMapFs,
			wantErr:     false,
		},
		"no file given": {
			filename:    "",
			filecontent: "",
			fs:          afero.NewMemMapFs,
			wantErr:     false,
		},
		"file does not exist": {
			filename:    "nonExistingSecret",
			filecontent: "",
			createFile:  false,
			fs:          afero.NewMemMapFs,
			wantErr:     true,
		},
		"file is empty": {
			filename:    "emptySecret",
			filecontent: "",
			createFile:  true,
			fs:          afero.NewMemMapFs,
			wantErr:     true,
		},
		"secret too short": {
			filename:    "shortSecret",
			filecontent: base64.StdEncoding.EncodeToString([]byte("short")),
			createFile:  true,
			fs:          afero.NewMemMapFs,
			wantErr:     true,
		},
		"secret not encoded": {
			filename:    "unencodedSecret",
			filecontent: "Constellation",
			createFile:  true,
			fs:          afero.NewMemMapFs,
			wantErr:     true,
		},
		"file not writeable": {
			filename:    "",
			filecontent: "",
			createFile:  false,
			fs:          func() afero.Fs { return afero.NewReadOnlyFs(afero.NewMemMapFs()) },
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(tc.fs())

			if tc.createFile {
				require.NoError(fileHandler.Write(tc.filename, []byte(tc.filecontent), file.OptNone))
			}

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

				content, err := fileHandler.Read(tc.filename)
				require.NoError(err)
				assert.Equal(content, []byte(base64.StdEncoding.EncodeToString(secret)))
			}
		})
	}
}

func TestAutoscaleFlag(t *testing.T) {
	testKey := base64.StdEncoding.EncodeToString([]byte("32bytesWireGuardKeyForTheTesting"))
	testGcpState := state.ConstellationState{
		CloudProvider: "gcp",
		GCPNodes: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		GCPCoordinators: cloudtypes.Instances{
			"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
	}
	testAzureState := state.ConstellationState{
		CloudProvider: "azure",
		AzureNodes: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureCoordinators: cloudtypes.Instances{
			"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureResourceGroup: "test",
	}
	testActivationResps := []fakeActivationRespMessage{
		{log: "testlog1"},
		{log: "testlog2"},
		{
			kubeconfig:        "kubeconfig",
			clientVpnIp:       "192.0.2.2",
			coordinatorVpnKey: testKey,
			ownerID:           "ownerID",
			clusterID:         "clusterID",
		},
		{log: "testlog3"},
	}

	testCases := map[string]struct {
		autoscaleFlag         bool
		existingState         state.ConstellationState
		client                *stubProtoClient
		serviceAccountCreator stubServiceAccountCreator
		waiter                statusWaiter
		privKey               string
	}{
		"initialize some gcp instances without autoscale flag": {
			autoscaleFlag: false,
			existingState: testGcpState,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  &stubStatusWaiter{},
			privKey: testKey,
		},
		"initialize some azure instances without autoscale flag": {
			autoscaleFlag: false,
			existingState: testAzureState,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  &stubStatusWaiter{},
			privKey: testKey,
		},
		"initialize some gcp instances with autoscale flag": {
			autoscaleFlag: true,
			existingState: testGcpState,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  &stubStatusWaiter{},
			privKey: testKey,
		},
		"initialize some azure instances with autoscale flag": {
			autoscaleFlag: true,
			existingState: testAzureState,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  &stubStatusWaiter{},
			privKey: testKey,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewInitCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)
			cmd.Flags().String("config", "", "") // register persisten flag manually
			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			vpnHandler := stubVPNHandler{}
			require.NoError(fileHandler.WriteJSON(constants.StateFilename, tc.existingState, file.OptNone))

			// Write key file to filesystem and set path in flag.
			require.NoError(afero.Afero{Fs: fs}.WriteFile("privK", []byte(tc.privKey), 0o600))
			require.NoError(cmd.Flags().Set("privatekey", "privK"))

			require.NoError(cmd.Flags().Set("autoscale", strconv.FormatBool(tc.autoscaleFlag)))
			ctx := context.Background()

			require.NoError(initialize(ctx, cmd, tc.client, &tc.serviceAccountCreator, fileHandler, tc.waiter, &vpnHandler))
			if tc.autoscaleFlag {
				assert.Len(tc.client.activateAutoscalingNodeGroups, 1)
			} else {
				assert.Len(tc.client.activateAutoscalingNodeGroups, 0)
			}
		})
	}
}

func TestWriteWGQuickFile(t *testing.T) {
	testCases := map[string]struct {
		fileHandler file.Handler
		vpnHandler  *stubVPNHandler
		vpnConfig   *wgquick.Config
		wantErr     bool
	}{
		"write wg quick file": {
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
			vpnHandler:  &stubVPNHandler{marshalRes: "config"},
		},
		"marshal failed": {
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
			vpnHandler:  &stubVPNHandler{marshalErr: errors.New("some err")},
			wantErr:     true,
		},
		"write fails": {
			fileHandler: file.NewHandler(afero.NewReadOnlyFs(afero.NewMemMapFs())),
			vpnHandler:  &stubVPNHandler{marshalRes: "config"},
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := writeWGQuickFile(tc.fileHandler, tc.vpnHandler, tc.vpnConfig)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				file, err := tc.fileHandler.Read(constants.WGQuickConfigFilename)
				assert.NoError(err)
				assert.Contains(string(file), tc.vpnHandler.marshalRes)
			}
		})
	}
}

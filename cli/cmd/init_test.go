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

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
	wgquick "github.com/nmiculinic/wg-quick-go"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitArgumentValidation(t *testing.T) {
	assert := assert.New(t)

	cmd := newInitCmd()
	assert.NoError(cmd.ValidateArgs(nil))
	assert.Error(cmd.ValidateArgs([]string{"something"}))
	assert.Error(cmd.ValidateArgs([]string{"sth", "sth"}))
}

func TestInitialize(t *testing.T) {
	testKey := base64.StdEncoding.EncodeToString([]byte("32bytesWireGuardKeyForTheTesting"))
	config := config.Default()
	testEc2State := state.ConstellationState{
		CloudProvider: "AWS",
		EC2Instances: ec2.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.2",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.2",
			},
			"id-2": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.2",
			},
		},
		EC2SecurityGroup: "sg-test",
	}
	testGcpState := state.ConstellationState{
		GCPNodes: gcp.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		GCPCoordinators: gcp.Instances{
			"id-c": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
	}
	testAzureState := state.ConstellationState{
		CloudProvider: "Azure",
		AzureNodes: azure.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		AzureCoordinators: azure.Instances{
			"id-c": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
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
	someErr := errors.New("failed")

	testCases := map[string]struct {
		existingState         state.ConstellationState
		client                protoClient
		serviceAccountCreator stubServiceAccountCreator
		waiter                statusWaiter
		privKey               string
		vpnHandler            vpnHandler
		initVPN               bool
		errExpected           bool
	}{
		"initialize some ec2 instances": {
			existingState: testEc2State,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{},
			privKey:    testKey,
		},
		"initialize some gcp instances": {
			existingState: testGcpState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{},
			privKey:    testKey,
		},
		"initialize some azure instances": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{},
			privKey:    testKey,
		},
		"initialize vpn": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:     stubStatusWaiter{},
			vpnHandler: &stubVPNHandler{},
			initVPN:    true,
			privKey:    testKey,
		},
		"invalid initialize vpn": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:      stubStatusWaiter{},
			vpnHandler:  &stubVPNHandler{applyErr: someErr},
			initVPN:     true,
			privKey:     testKey,
			errExpected: true,
		},
		"invalid create vpn config": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:      stubStatusWaiter{},
			vpnHandler:  &stubVPNHandler{createErr: someErr},
			initVPN:     true,
			privKey:     testKey,
			errExpected: true,
		},
		"invalid write vpn config": {
			existingState: testAzureState,
			client: &fakeProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:      stubStatusWaiter{},
			vpnHandler:  &stubVPNHandler{marshalErr: someErr},
			initVPN:     true,
			privKey:     testKey,
			errExpected: true,
		},
		"no state exists": {
			existingState: state.ConstellationState{},
			client:        &stubProtoClient{},
			waiter:        stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"no instances to pick one": {
			existingState: state.ConstellationState{
				EC2Instances:     ec2.Instances{},
				EC2SecurityGroup: "sg-test",
			},
			client:      &stubProtoClient{},
			waiter:      stubStatusWaiter{},
			privKey:     testKey,
			vpnHandler:  &stubVPNHandler{},
			errExpected: true,
		},
		"only one instance": {
			existingState: state.ConstellationState{
				EC2Instances:     ec2.Instances{"id-1": {}},
				EC2SecurityGroup: "sg-test",
			},
			client:      &stubProtoClient{},
			waiter:      stubStatusWaiter{},
			privKey:     testKey,
			vpnHandler:  &stubVPNHandler{},
			errExpected: true,
		},
		"public key to short": {
			existingState: testEc2State,
			client:        &stubProtoClient{},
			waiter:        stubStatusWaiter{},
			privKey:       base64.StdEncoding.EncodeToString([]byte("tooShortKey")),
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"public key to long": {
			existingState: testEc2State,
			client:        &stubProtoClient{},
			waiter:        stubStatusWaiter{},
			privKey:       base64.StdEncoding.EncodeToString([]byte("thisWireguardKeyIsToLongAndHasTooManyBytes")),
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"public key not base64": {
			existingState: testEc2State,
			client:        &stubProtoClient{},
			waiter:        stubStatusWaiter{},
			privKey:       "this is not base64 encoded",
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail Connect": {
			existingState: testEc2State,
			client:        &stubProtoClient{connectErr: someErr},
			waiter:        stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail Activate": {
			existingState: testEc2State,
			client:        &stubProtoClient{activateErr: someErr},
			waiter:        stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail respClient WriteLogStream": {
			existingState: testEc2State,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{writeLogStreamErr: someErr}},
			waiter:        stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail respClient getKubeconfig": {
			existingState: testEc2State,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getKubeconfigErr: someErr}},
			waiter:        stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail respClient getCoordinatorVpnKey": {
			existingState: testEc2State,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getCoordinatorVpnKeyErr: someErr}},
			waiter:        stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail respClient getClientVpnIp": {
			existingState: testEc2State,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getClientVpnIpErr: someErr}},
			waiter:        stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail respClient getOwnerID": {
			existingState: testEc2State,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getOwnerIDErr: someErr}},
			waiter:        stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail respClient getClusterID": {
			existingState: testEc2State,
			client:        &stubProtoClient{respClient: &stubActivationRespClient{getClusterIDErr: someErr}},
			waiter:        stubStatusWaiter{},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail to wait for required status": {
			existingState: testGcpState,
			client:        &stubProtoClient{},
			waiter:        stubStatusWaiter{waitForAllErr: someErr},
			privKey:       testKey,
			vpnHandler:    &stubVPNHandler{},
			errExpected:   true,
		},
		"fail to create service account": {
			existingState: testGcpState,
			client:        &stubProtoClient{},
			serviceAccountCreator: stubServiceAccountCreator{
				createErr: someErr,
			},
			waiter:      stubStatusWaiter{},
			privKey:     testKey,
			vpnHandler:  &stubVPNHandler{},
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newInitCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)
			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			require.NoError(fileHandler.WriteJSON(*config.StatePath, tc.existingState, file.OptNone))

			// Write key file to filesystem and set path in flag.
			require.NoError(afero.Afero{Fs: fs}.WriteFile("privK", []byte(tc.privKey), 0o600))
			require.NoError(cmd.Flags().Set("privatekey", "privK"))
			if tc.initVPN {
				require.NoError(cmd.Flags().Set("wg-autoconfig", "true"))
			}

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
			defer cancel()

			err := initialize(ctx, cmd, tc.client, &tc.serviceAccountCreator, fileHandler, config, tc.waiter, tc.vpnHandler)

			if tc.errExpected {
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
	config := config.Default()

	err := result.writeOutput(&out, fileHandler, config)
	assert.NoError(err)
	assert.Contains(out.String(), result.clientVpnIP)
	assert.Contains(out.String(), result.coordinatorPubIP)
	assert.Contains(out.String(), result.coordinatorPubKey)

	afs := afero.Afero{Fs: testFs}
	adminConf, err := afs.ReadFile(*config.AdminConfPath)
	assert.NoError(err)
	assert.Equal(result.kubeconfig, string(adminConf))
}

func TestIpsToEndpoints(t *testing.T) {
	assert := assert.New(t)

	ips := []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"}
	port := "8080"
	endpoints := ipsToEndpoints(ips, port)
	assert.Equal([]string{"192.0.2.1:8080", "192.0.2.2:8080", "192.0.2.3:8080"}, endpoints)
}

func TestInitCompletion(t *testing.T) {
	testCases := map[string]struct {
		args            []string
		toComplete      string
		resultExpected  []string
		shellCDExpected cobra.ShellCompDirective
	}{
		"first arg": {
			args:            []string{},
			toComplete:      "hello",
			resultExpected:  []string{},
			shellCDExpected: cobra.ShellCompDirectiveDefault,
		},
		"secnod arg": {
			args:            []string{"23"},
			toComplete:      "/test/h",
			resultExpected:  []string{},
			shellCDExpected: cobra.ShellCompDirectiveError,
		},
		"third arg": {
			args:            []string{"./file", "sth"},
			toComplete:      "./file",
			resultExpected:  []string{},
			shellCDExpected: cobra.ShellCompDirectiveError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := &cobra.Command{}
			result, shellCD := initCompletion(cmd, tc.args, tc.toComplete)
			assert.Equal(tc.resultExpected, result)
			assert.Equal(tc.shellCDExpected, shellCD)
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

func TestReadOrGeneratedMasterSecret(t *testing.T) {
	testCases := map[string]struct {
		filename    string
		filecontent string
		createFile  bool
		fs          func() afero.Fs
		errExpected bool
	}{
		"file with secret exists": {
			filename:    "someSecret",
			filecontent: base64.StdEncoding.EncodeToString([]byte("ConstellationSecret")),
			createFile:  true,
			fs:          afero.NewMemMapFs,
			errExpected: false,
		},
		"no file given": {
			filename:    "",
			filecontent: "",
			fs:          afero.NewMemMapFs,
			errExpected: false,
		},
		"file does not exist": {
			filename:    "nonExistingSecret",
			filecontent: "",
			createFile:  false,
			fs:          afero.NewMemMapFs,
			errExpected: true,
		},
		"file is empty": {
			filename:    "emptySecret",
			filecontent: "",
			createFile:  true,
			fs:          afero.NewMemMapFs,
			errExpected: true,
		},
		"secret too short": {
			filename:    "shortSecret",
			filecontent: base64.StdEncoding.EncodeToString([]byte("short")),
			createFile:  true,
			fs:          afero.NewMemMapFs,
			errExpected: true,
		},
		"secret not encoded": {
			filename:    "unencodedSecret",
			filecontent: "Constellation",
			createFile:  true,
			fs:          afero.NewMemMapFs,
			errExpected: true,
		},
		"file not writeable": {
			filename:    "",
			filecontent: "",
			createFile:  false,
			fs:          func() afero.Fs { return afero.NewReadOnlyFs(afero.NewMemMapFs()) },
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(tc.fs())
			config := config.Default()

			if tc.createFile {
				require.NoError(fileHandler.Write(tc.filename, []byte(tc.filecontent), file.OptNone))
			}

			var out bytes.Buffer
			secret, err := readOrGeneratedMasterSecret(&out, fileHandler, tc.filename, config)

			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)

				if tc.filename == "" {
					require.Contains(out.String(), *config.MasterSecretPath)
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
	config := config.Default()
	testEc2State := state.ConstellationState{
		EC2Instances: ec2.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.2",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.2",
			},
			"id-2": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.2",
			},
		},
		EC2SecurityGroup: "sg-test",
	}
	testGcpState := state.ConstellationState{
		GCPNodes: gcp.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		GCPCoordinators: gcp.Instances{
			"id-c": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
	}
	testAzureState := state.ConstellationState{
		AzureNodes: azure.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		AzureCoordinators: azure.Instances{
			"id-c": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
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
		"initialize some ec2 instances without autoscale flag": {
			autoscaleFlag: false,
			existingState: testEc2State,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  stubStatusWaiter{},
			privKey: testKey,
		},
		"initialize some gcp instances without autoscale flag": {
			autoscaleFlag: false,
			existingState: testGcpState,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  stubStatusWaiter{},
			privKey: testKey,
		},
		"initialize some azure instances without autoscale flag": {
			autoscaleFlag: false,
			existingState: testAzureState,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  stubStatusWaiter{},
			privKey: testKey,
		},
		"initialize some ec2 instances with autoscale flag": {
			autoscaleFlag: true,
			existingState: testEc2State,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  stubStatusWaiter{},
			privKey: testKey,
		},
		"initialize some gcp instances with autoscale flag": {
			autoscaleFlag: true,
			existingState: testGcpState,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  stubStatusWaiter{},
			privKey: testKey,
		},
		"initialize some azure instances with autoscale flag": {
			autoscaleFlag: true,
			existingState: testAzureState,
			client: &stubProtoClient{
				respClient: &fakeActivationRespClient{responses: testActivationResps},
			},
			waiter:  stubStatusWaiter{},
			privKey: testKey,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newInitCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)
			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			vpnHandler := stubVPNHandler{}
			require.NoError(fileHandler.WriteJSON(*config.StatePath, tc.existingState, file.OptNone))

			// Write key file to filesystem and set path in flag.
			require.NoError(afero.Afero{Fs: fs}.WriteFile("privK", []byte(tc.privKey), 0o600))
			require.NoError(cmd.Flags().Set("privatekey", "privK"))

			require.NoError(cmd.Flags().Set("autoscale", strconv.FormatBool(tc.autoscaleFlag)))
			ctx := context.Background()

			require.NoError(initialize(ctx, cmd, tc.client, &tc.serviceAccountCreator, fileHandler, config, tc.waiter, &vpnHandler))
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
		config      *config.Config
		vpnHandler  *stubVPNHandler
		vpnConfig   *wgquick.Config
		wantErr     bool
	}{
		"write wg quick file": {
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
			config:      &config.Config{WGQuickConfigPath: func(s string) *string { return &s }("a.conf")},
			vpnHandler:  &stubVPNHandler{marshalRes: "config"},
		},
		"marshal failed": {
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
			config:      &config.Config{WGQuickConfigPath: func(s string) *string { return &s }("a.conf")},
			vpnHandler:  &stubVPNHandler{marshalErr: errors.New("some err")},
			wantErr:     true,
		},
		"write fails": {
			fileHandler: file.NewHandler(afero.NewReadOnlyFs(afero.NewMemMapFs())),
			config:      &config.Config{WGQuickConfigPath: func(s string) *string { return &s }("a.conf")},
			vpnHandler:  &stubVPNHandler{marshalRes: "config"},
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := writeWGQuickFile(tc.fileHandler, tc.config, tc.vpnHandler, tc.vpnConfig)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				file, err := tc.fileHandler.Read(*tc.config.WGQuickConfigPath)
				assert.NoError(err)
				assert.Contains(string(file), tc.vpnHandler.marshalRes)
			}
		})
	}
}

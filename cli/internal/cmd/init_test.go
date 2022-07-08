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

	"github.com/edgelesssys/constellation/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
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
	testGcpState := state.ConstellationState{
		CloudProvider:    "GCP",
		BootstrapperHost: "192.0.2.1",
		GCPWorkers: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		GCPControlPlanes: cloudtypes.Instances{
			"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
	}
	testAzureState := state.ConstellationState{
		CloudProvider:    "Azure",
		BootstrapperHost: "192.0.2.1",
		AzureWorkers: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureControlPlane: cloudtypes.Instances{
			"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureResourceGroup: "test",
	}
	testQemuState := state.ConstellationState{
		CloudProvider:    "QEMU",
		BootstrapperHost: "192.0.2.1",
		QEMUWorkers: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		QEMUControlPlane: cloudtypes.Instances{
			"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
	}
	testInitResp := &initproto.InitResponse{
		Kubeconfig: []byte("kubeconfig"),
		OwnerId:    []byte("ownerID"),
		ClusterId:  []byte("clusterID"),
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		existingState         state.ConstellationState
		serviceAccountCreator stubServiceAccountCreator
		initServerAPI         *stubInitServer
		setAutoscaleFlag      bool
		wantErr               bool
	}{
		"initialize some gcp instances": {
			existingState: testGcpState,
			initServerAPI: &stubInitServer{initResp: testInitResp},
		},
		"initialize some azure instances": {
			existingState: testAzureState,
			initServerAPI: &stubInitServer{initResp: testInitResp},
		},
		"initialize some qemu instances": {
			existingState: testQemuState,
			initServerAPI: &stubInitServer{initResp: testInitResp},
		},
		"initialize gcp with autoscaling": {
			existingState:    testGcpState,
			initServerAPI:    &stubInitServer{initResp: testInitResp},
			setAutoscaleFlag: true,
		},
		"initialize azure with autoscaling": {
			existingState:    testAzureState,
			initServerAPI:    &stubInitServer{initResp: testInitResp},
			setAutoscaleFlag: true,
		},
		"empty state": {
			existingState: state.ConstellationState{},
			initServerAPI: &stubInitServer{},
			wantErr:       true,
		},
		"init call fails": {
			existingState: testGcpState,
			initServerAPI: &stubInitServer{initErr: someErr},
			wantErr:       true,
		},
		"fail to create service account": {
			existingState:         testGcpState,
			initServerAPI:         &stubInitServer{},
			serviceAccountCreator: stubServiceAccountCreator{createErr: someErr},
			wantErr:               true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, nil, netDialer)
			serverCreds := atlscredentials.New(nil, nil)
			initServer := grpc.NewServer(grpc.Creds(serverCreds))
			initproto.RegisterAPIServer(initServer, tc.initServerAPI)
			port := strconv.Itoa(constants.BootstrapperPort)
			listener := netDialer.GetListener(net.JoinHostPort("192.0.2.1", port))
			go initServer.Serve(listener)
			defer initServer.GracefulStop()

			cmd := NewInitCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)
			cmd.Flags().String("config", "", "") // register persisten flag manually
			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			require.NoError(fileHandler.WriteJSON(constants.StateFilename, tc.existingState, file.OptNone))
			require.NoError(cmd.Flags().Set("autoscale", strconv.FormatBool(tc.setAutoscaleFlag)))

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
			defer cancel()
			cmd.SetContext(ctx)

			err := initialize(cmd, dialer, &tc.serviceAccountCreator, fileHandler)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Contains(out.String(), base64.StdEncoding.EncodeToString([]byte("ownerID")))
			assert.Contains(out.String(), base64.StdEncoding.EncodeToString([]byte("clusterID")))
			if tc.setAutoscaleFlag {
				assert.Len(tc.initServerAPI.activateAutoscalingNodeGroups, 1)
			} else {
				assert.Len(tc.initServerAPI.activateAutoscalingNodeGroups, 0)
			}
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

	expectedIDFile := clusterIDsFile{
		ClusterID: clusterID,
		OwnerID:   ownerID,
		Endpoint:  net.JoinHostPort("ip", strconv.Itoa(constants.VerifyServiceNodePortGRPC)),
	}

	var out bytes.Buffer
	testFs := afero.NewMemMapFs()
	fileHandler := file.NewHandler(testFs)

	err := writeOutput(resp, "ip", &out, fileHandler)
	assert.NoError(err)
	assert.Contains(out.String(), ownerID)
	assert.Contains(out.String(), clusterID)
	assert.Contains(out.String(), constants.AdminConfFilename)

	afs := afero.Afero{Fs: testFs}
	adminConf, err := afs.ReadFile(constants.AdminConfFilename)
	assert.NoError(err)
	assert.Equal(string(resp.Kubeconfig), string(adminConf))

	idsFile, err := afs.ReadFile(constants.ClusterIDsFileName)
	assert.NoError(err)
	var testIDFile clusterIDsFile
	err = json.Unmarshal(idsFile, &testIDFile)
	assert.NoError(err)
	assert.Equal(expectedIDFile, testIDFile)
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

func TestReadOrGeneratedMasterSecret(t *testing.T) {
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

type stubInitServer struct {
	initResp *initproto.InitResponse
	initErr  error

	activateAutoscalingNodeGroups []string

	initproto.UnimplementedAPIServer
}

func (s *stubInitServer) Init(ctx context.Context, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	s.activateAutoscalingNodeGroups = req.AutoscalingNodeGroups
	return s.initResp, s.initErr
}

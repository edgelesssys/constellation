package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"testing"

	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/edgelesssys/constellation/verify/verifyproto"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	rpcStatus "google.golang.org/grpc/status"
)

func TestVerifyCmdArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"no args":          {[]string{}, true},
		"valid azure":      {[]string{"azure"}, false},
		"valid gcp":        {[]string{"gcp"}, false},
		"invalid provider": {[]string{"invalid", "192.0.2.1", "1234"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := NewVerifyCmd()
			err := cmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	zeroBase64 := base64.StdEncoding.EncodeToString([]byte("00000000000000000000000000000000"))
	someErr := errors.New("failed")

	testCases := map[string]struct {
		setupFs          func(*require.Assertions) afero.Fs
		provider         cloudprovider.Provider
		protoClient      *stubVerifyClient
		nodeEndpointFlag string
		configFlag       string
		ownerIDFlag      string
		clusterIDFlag    string
		idFile           *clusterIDsFile
		wantEndpoint     string
		wantErr          bool
	}{
		"gcp": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantEndpoint:     "192.0.2.1:1234",
		},
		"azure": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantEndpoint:     "192.0.2.1:1234",
		},
		"default port": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantEndpoint:     "192.0.2.1:30081",
		},
		"endpoint not set": {
			setupFs:     func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:    cloudprovider.GCP,
			ownerIDFlag: zeroBase64,
			protoClient: &stubVerifyClient{},
			wantErr:     true,
		},
		"endpoint from id file": {
			setupFs:      func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:     cloudprovider.GCP,
			ownerIDFlag:  zeroBase64,
			protoClient:  &stubVerifyClient{},
			idFile:       &clusterIDsFile{Endpoint: "192.0.2.1:1234"},
			wantEndpoint: "192.0.2.1:1234",
		},
		"override endpoint from details file": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.2:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubVerifyClient{},
			idFile:           &clusterIDsFile{Endpoint: "192.0.2.1:1234"},
			wantEndpoint:     "192.0.2.2:1234",
		},
		"invalid endpoint": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: ":::::",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantErr:          true,
		},
		"neither owner id nor cluster id set": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			wantErr:          true,
		},
		"use owner id from id file": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			protoClient:      &stubVerifyClient{},
			idFile:           &clusterIDsFile{OwnerID: zeroBase64},
			wantEndpoint:     "192.0.2.1:1234",
		},
		"config file not existing": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			ownerIDFlag:      zeroBase64,
			nodeEndpointFlag: "192.0.2.1:1234",
			configFlag:       "./file",
			wantErr:          true,
		},
		"error protoClient GetState": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubVerifyClient{verifyErr: rpcStatus.Error(codes.Internal, "failed")},
			wantErr:          true,
		},
		"error protoClient GetState not rpc": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubVerifyClient{verifyErr: someErr},
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewVerifyCmd()
			cmd.Flags().String("config", "", "") // register persistent flag manually
			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(&bytes.Buffer{})
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}
			if tc.ownerIDFlag != "" {
				require.NoError(cmd.Flags().Set("owner-id", tc.ownerIDFlag))
			}
			if tc.clusterIDFlag != "" {
				require.NoError(cmd.Flags().Set("cluster-id", tc.clusterIDFlag))
			}
			if tc.nodeEndpointFlag != "" {
				require.NoError(cmd.Flags().Set("node-endpoint", tc.nodeEndpointFlag))
			}
			fileHandler := file.NewHandler(tc.setupFs(require))

			if tc.idFile != nil {
				require.NoError(fileHandler.WriteJSON(constants.ClusterIDsFileName, tc.idFile, file.OptNone))
			}

			err := verify(cmd, tc.provider, fileHandler, tc.protoClient)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Contains(out.String(), "OK")
				assert.Equal(tc.wantEndpoint, tc.protoClient.endpoint)
			}
		})
	}
}

func TestVerifyCompletion(t *testing.T) {
	testCases := map[string]struct {
		args        []string
		toComplete  string
		wantResult  []string
		wantShellCD cobra.ShellCompDirective
	}{
		"first arg": {
			args:        []string{},
			toComplete:  "az",
			wantResult:  []string{"gcp", "azure"},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"additional arg": {
			args:        []string{"gcp", "foo"},
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := &cobra.Command{}
			result, shellCD := verifyCompletion(cmd, tc.args, tc.toComplete)
			assert.Equal(tc.wantResult, result)
			assert.Equal(tc.wantShellCD, shellCD)
		})
	}
}

func TestVerifyClient(t *testing.T) {
	testCases := map[string]struct {
		attestationDoc atls.FakeAttestationDoc
		userData       []byte
		nonce          []byte
		attestationErr error
		wantErr        bool
	}{
		"success": {
			attestationDoc: atls.FakeAttestationDoc{
				UserData: []byte("user data"),
				Nonce:    []byte("nonce"),
			},
			userData: []byte("user data"),
			nonce:    []byte("nonce"),
		},
		"attestation error": {
			attestationDoc: atls.FakeAttestationDoc{
				UserData: []byte("user data"),
				Nonce:    []byte("nonce"),
			},
			userData:       []byte("user data"),
			nonce:          []byte("nonce"),
			attestationErr: errors.New("error"),
			wantErr:        true,
		},
		"user data does not match": {
			attestationDoc: atls.FakeAttestationDoc{
				UserData: []byte("wrong user data"),
				Nonce:    []byte("nonce"),
			},
			userData: []byte("user data"),
			nonce:    []byte("nonce"),
			wantErr:  true,
		},
		"nonce does not match": {
			attestationDoc: atls.FakeAttestationDoc{
				UserData: []byte("user data"),
				Nonce:    []byte("wrong nonce"),
			},
			userData: []byte("user data"),
			nonce:    []byte("nonce"),
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			attestation, err := json.Marshal(tc.attestationDoc)
			require.NoError(err)
			verifyAPI := &stubVerifyAPI{
				attestation:    &verifyproto.GetAttestationResponse{Attestation: attestation},
				attestationErr: tc.attestationErr,
			}

			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, nil, netDialer)
			verifyServer := grpc.NewServer()
			verifyproto.RegisterAPIServer(verifyServer, verifyAPI)

			addr := net.JoinHostPort("192.0.2.1", strconv.Itoa(constants.VerifyServiceNodePortGRPC))
			listener := netDialer.GetListener(addr)
			go verifyServer.Serve(listener)
			defer verifyServer.GracefulStop()

			verifier := &constellationVerifier{dialer: dialer}
			request := &verifyproto.GetAttestationRequest{
				UserData: tc.userData,
				Nonce:    tc.nonce,
			}

			err = verifier.Verify(context.Background(), addr, request, atls.NewFakeValidator(oid.Dummy{}))

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubVerifyClient struct {
	verifyErr error
	endpoint  string
}

func (c *stubVerifyClient) Verify(ctx context.Context, endpoint string, req *verifyproto.GetAttestationRequest, validator atls.Validator) error {
	c.endpoint = endpoint
	return c.verifyErr
}

type stubVerifyAPI struct {
	attestation    *verifyproto.GetAttestationResponse
	attestationErr error
	verifyproto.UnimplementedAPIServer
}

func (a stubVerifyAPI) GetAttestation(context.Context, *verifyproto.GetAttestationRequest) (*verifyproto.GetAttestationResponse, error) {
	return a.attestation, a.attestationErr
}

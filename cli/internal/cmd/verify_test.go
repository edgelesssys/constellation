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
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/edgelesssys/constellation/v2/verify/verifyproto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	rpcStatus "google.golang.org/grpc/status"
)

func TestVerify(t *testing.T) {
	zeroBase64 := base64.StdEncoding.EncodeToString([]byte("00000000000000000000000000000000"))
	someErr := errors.New("failed")

	testCases := map[string]struct {
		provider         cloudprovider.Provider
		protoClient      *stubVerifyClient
		nodeEndpointFlag string
		configFlag       string
		ownerIDFlag      string
		clusterIDFlag    string
		idFile           *clusterid.File
		wantEndpoint     string
		wantErr          bool
	}{
		"gcp": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantEndpoint:     "192.0.2.1:1234",
		},
		"azure": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantEndpoint:     "192.0.2.1:1234",
		},
		"default port": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantEndpoint:     "192.0.2.1:" + strconv.Itoa(constants.VerifyServiceNodePortGRPC),
		},
		"endpoint not set": {
			provider:      cloudprovider.GCP,
			clusterIDFlag: zeroBase64,
			protoClient:   &stubVerifyClient{},
			wantErr:       true,
		},
		"endpoint from id file": {
			provider:      cloudprovider.GCP,
			clusterIDFlag: zeroBase64,
			protoClient:   &stubVerifyClient{},
			idFile:        &clusterid.File{IP: "192.0.2.1"},
			wantEndpoint:  "192.0.2.1:" + strconv.Itoa(constants.VerifyServiceNodePortGRPC),
		},
		"override endpoint from details file": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.2:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			idFile:           &clusterid.File{IP: "192.0.2.1"},
			wantEndpoint:     "192.0.2.2:1234",
		},
		"invalid endpoint": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: ":::::",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantErr:          true,
		},
		"neither owner id nor cluster id set": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			wantErr:          true,
		},
		"use owner id from id file": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			protoClient:      &stubVerifyClient{},
			idFile:           &clusterid.File{OwnerID: zeroBase64},
			wantEndpoint:     "192.0.2.1:1234",
		},
		"config file not existing": {
			provider:         cloudprovider.GCP,
			clusterIDFlag:    zeroBase64,
			nodeEndpointFlag: "192.0.2.1:1234",
			configFlag:       "./file",
			wantErr:          true,
		},
		"error protoClient GetState": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{verifyErr: rpcStatus.Error(codes.Internal, "failed")},
			wantErr:          true,
		},
		"error protoClient GetState not rpc": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{verifyErr: someErr},
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewVerifyCmd()
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")                        // register persistent flag manually
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
			fileHandler := file.NewHandler(afero.NewMemMapFs())

			config := defaultConfigWithExpectedMeasurements(t, config.Default(), tc.provider)
			require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, config))
			if tc.idFile != nil {
				require.NoError(fileHandler.WriteJSON(constants.ClusterIDsFileName, tc.idFile, file.OptNone))
			}

			v := &verifyCmd{log: logger.NewTest(t)}
			err := v.verify(cmd, fileHandler, tc.protoClient)

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

func TestVerifyClient(t *testing.T) {
	testCases := map[string]struct {
		attestationDoc atls.FakeAttestationDoc
		nonce          []byte
		attestationErr error
		wantErr        bool
	}{
		"success": {
			attestationDoc: atls.FakeAttestationDoc{
				UserData: []byte(constants.ConstellationVerifyServiceUserData),
				Nonce:    []byte("nonce"),
			},
			nonce: []byte("nonce"),
		},
		"attestation error": {
			attestationDoc: atls.FakeAttestationDoc{
				UserData: []byte(constants.ConstellationVerifyServiceUserData),
				Nonce:    []byte("nonce"),
			},
			nonce:          []byte("nonce"),
			attestationErr: errors.New("error"),
			wantErr:        true,
		},
		"user data does not match": {
			attestationDoc: atls.FakeAttestationDoc{
				UserData: []byte("wrong user data"),
				Nonce:    []byte("nonce"),
			},
			nonce:   []byte("nonce"),
			wantErr: true,
		},
		"nonce does not match": {
			attestationDoc: atls.FakeAttestationDoc{
				UserData: []byte(constants.ConstellationVerifyServiceUserData),
				Nonce:    []byte("wrong nonce"),
			},
			nonce:   []byte("nonce"),
			wantErr: true,
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

			verifier := &constellationVerifier{dialer: dialer, log: logger.NewTest(t)}
			request := &verifyproto.GetAttestationRequest{
				Nonce: tc.nonce,
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

func (c *stubVerifyClient) Verify(_ context.Context, endpoint string, _ *verifyproto.GetAttestationRequest, _ atls.Validator) error {
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

func TestAddPortIfMissing(t *testing.T) {
	testCases := map[string]struct {
		endpoint    string
		defaultPort int
		wantResult  string
		wantErr     bool
	}{
		"ip and port": {
			endpoint:    "192.0.2.1:2",
			defaultPort: 3,
			wantResult:  "192.0.2.1:2",
		},
		"hostname and port": {
			endpoint:    "foo:2",
			defaultPort: 3,
			wantResult:  "foo:2",
		},
		"ip": {
			endpoint:    "192.0.2.1",
			defaultPort: 3,
			wantResult:  "192.0.2.1:3",
		},
		"hostname": {
			endpoint:    "foo",
			defaultPort: 3,
			wantResult:  "foo:3",
		},
		"empty endpoint": {
			endpoint:    "",
			defaultPort: 3,
			wantErr:     true,
		},
		"invalid endpoint": {
			endpoint:    "foo:2:2",
			defaultPort: 3,
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			res, err := addPortIfMissing(tc.endpoint, tc.defaultPort)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.wantResult, res)
		})
	}
}

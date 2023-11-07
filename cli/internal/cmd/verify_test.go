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

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/edgelesssys/constellation/v2/verify/verifyproto"
	tpmProto "github.com/google/go-tpm-tools/proto/tpm"
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
		provider           cloudprovider.Provider
		protoClient        *stubVerifyClient
		formatter          *stubAttDocFormatter
		nodeEndpointFlag   string
		clusterIDFlag      string
		stateFile          *state.State
		wantEndpoint       string
		skipConfigCreation bool
		wantErr            bool
	}{
		"gcp": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			stateFile:        defaultStateFile(cloudprovider.GCP),
			wantEndpoint:     "192.0.2.1:1234",
			formatter:        &stubAttDocFormatter{},
		},
		"azure": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			stateFile:        defaultStateFile(cloudprovider.Azure),
			wantEndpoint:     "192.0.2.1:1234",
			formatter:        &stubAttDocFormatter{},
		},
		"default port": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			stateFile:        defaultStateFile(cloudprovider.GCP),
			wantEndpoint:     "192.0.2.1:" + strconv.Itoa(constants.VerifyServiceNodePortGRPC),
			formatter:        &stubAttDocFormatter{},
		},
		"endpoint not set": {
			provider:      cloudprovider.GCP,
			clusterIDFlag: zeroBase64,
			protoClient:   &stubVerifyClient{},
			stateFile: func() *state.State {
				s := defaultStateFile(cloudprovider.GCP)
				s.Infrastructure.ClusterEndpoint = ""
				return s
			}(),
			formatter: &stubAttDocFormatter{},
			wantErr:   true,
		},
		"endpoint from state file": {
			provider:      cloudprovider.GCP,
			clusterIDFlag: zeroBase64,
			protoClient:   &stubVerifyClient{},
			stateFile: func() *state.State {
				s := defaultStateFile(cloudprovider.GCP)
				s.Infrastructure.ClusterEndpoint = "192.0.2.1"
				return s
			}(),
			wantEndpoint: "192.0.2.1:" + strconv.Itoa(constants.VerifyServiceNodePortGRPC),
			formatter:    &stubAttDocFormatter{},
		},
		"override endpoint from details file": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.2:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			stateFile: func() *state.State {
				s := defaultStateFile(cloudprovider.GCP)
				s.Infrastructure.ClusterEndpoint = "192.0.2.1"
				return s
			}(),
			wantEndpoint: "192.0.2.2:1234",
			formatter:    &stubAttDocFormatter{},
		},
		"invalid endpoint": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: ":::::",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			stateFile:        defaultStateFile(cloudprovider.GCP),
			formatter:        &stubAttDocFormatter{},
			wantErr:          true,
		},
		"neither owner id nor cluster id set": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			stateFile: func() *state.State {
				s := defaultStateFile(cloudprovider.GCP)
				s.ClusterValues.OwnerID = ""
				s.ClusterValues.ClusterID = ""
				return s
			}(),
			formatter:   &stubAttDocFormatter{},
			protoClient: &stubVerifyClient{},
			wantErr:     true,
		},
		"use owner id from state file": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			protoClient:      &stubVerifyClient{},
			stateFile: func() *state.State {
				s := defaultStateFile(cloudprovider.GCP)
				s.ClusterValues.OwnerID = zeroBase64
				return s
			}(),
			wantEndpoint: "192.0.2.1:1234",
			formatter:    &stubAttDocFormatter{},
		},
		"config file not existing": {
			provider:           cloudprovider.GCP,
			clusterIDFlag:      zeroBase64,
			nodeEndpointFlag:   "192.0.2.1:1234",
			stateFile:          defaultStateFile(cloudprovider.GCP),
			formatter:          &stubAttDocFormatter{},
			skipConfigCreation: true,
			wantErr:            true,
		},
		"error protoClient GetState": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{verifyErr: rpcStatus.Error(codes.Internal, "failed")},
			stateFile:        defaultStateFile(cloudprovider.Azure),
			formatter:        &stubAttDocFormatter{},
			wantErr:          true,
		},
		"error protoClient GetState not rpc": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{verifyErr: someErr},
			stateFile:        defaultStateFile(cloudprovider.Azure),
			formatter:        &stubAttDocFormatter{},
			wantErr:          true,
		},
		"format error": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			stateFile:        defaultStateFile(cloudprovider.Azure),
			wantEndpoint:     "192.0.2.1:1234",
			formatter:        &stubAttDocFormatter{formatErr: someErr},
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewVerifyCmd()
			out := &bytes.Buffer{}
			cmd.SetErr(out)
			fileHandler := file.NewHandler(afero.NewMemMapFs())

			if !tc.skipConfigCreation {
				cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), tc.provider)
				require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, cfg))
			}
			require.NoError(tc.stateFile.WriteToFile(fileHandler, constants.StateFilename))

			v := &verifyCmd{
				fileHandler: fileHandler,
				log:         logger.NewTest(t),
				flags: verifyFlags{
					clusterID: tc.clusterIDFlag,
					endpoint:  tc.nodeEndpointFlag,
				},
			}
			formatterFac := func(_ string, _ cloudprovider.Provider, _ debugLog) (attestationDocFormatter, error) {
				return tc.formatter, nil
			}
			err := v.verify(cmd, tc.protoClient, formatterFac, stubAttestationFetcher{})
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

type stubAttDocFormatter struct {
	formatErr error
}

func (f *stubAttDocFormatter) format(_ context.Context, _ string, _ bool, _ config.AttestationCfg) (string, error) {
	return "", f.formatErr
}

func TestFormat(t *testing.T) {
	formatter := func() *defaultAttestationDocFormatter {
		return &defaultAttestationDocFormatter{
			log: logger.NewTest(t),
		}
	}

	testCases := map[string]struct {
		formatter *defaultAttestationDocFormatter
		doc       string
		wantErr   bool
	}{
		"invalid doc": {
			formatter: formatter(),
			doc:       "invalid",
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.formatter.format(context.Background(), tc.doc, false, nil)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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

			_, err = verifier.Verify(context.Background(), addr, request, atls.NewFakeValidator(variant.Dummy{}))

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

func (c *stubVerifyClient) Verify(_ context.Context, endpoint string, _ *verifyproto.GetAttestationRequest, _ atls.Validator) (string, error) {
	c.endpoint = endpoint
	return "", c.verifyErr
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

func TestParseQuotes(t *testing.T) {
	testCases := map[string]struct {
		quotes       []*tpmProto.Quote
		expectedPCRs measurements.M
		wantOutput   string
		wantErr      bool
	}{
		"parse quotes in order": {
			quotes: []*tpmProto.Quote{
				{
					Pcrs: &tpmProto.PCRs{
						Hash: tpmProto.HashAlgo_SHA256,
						Pcrs: map[uint32][]byte{
							0: {0x00},
							1: {0x01},
						},
					},
				},
			},
			expectedPCRs: measurements.M{
				0: measurements.WithAllBytes(0x00, measurements.Enforce, 1),
				1: measurements.WithAllBytes(0x01, measurements.WarnOnly, 1),
			},
			wantOutput: "\tQuote:\n\t\tPCR 0 (Strict: true):\n\t\t\tExpected:\t00\n\t\t\tActual:\t\t00\n\t\tPCR 1 (Strict: false):\n\t\t\tExpected:\t01\n\t\t\tActual:\t\t01\n",
		},
		"additional quotes are skipped": {
			quotes: []*tpmProto.Quote{
				{
					Pcrs: &tpmProto.PCRs{
						Hash: tpmProto.HashAlgo_SHA256,
						Pcrs: map[uint32][]byte{
							0: {0x00},
							1: {0x01},
							2: {0x02},
							3: {0x03},
						},
					},
				},
			},
			expectedPCRs: measurements.M{
				0: measurements.WithAllBytes(0x00, measurements.Enforce, 1),
				1: measurements.WithAllBytes(0x01, measurements.WarnOnly, 1),
			},
			wantOutput: "\tQuote:\n\t\tPCR 0 (Strict: true):\n\t\t\tExpected:\t00\n\t\t\tActual:\t\t00\n\t\tPCR 1 (Strict: false):\n\t\t\tExpected:\t01\n\t\t\tActual:\t\t01\n",
		},
		"missing quotes error": {
			quotes: []*tpmProto.Quote{
				{
					Pcrs: &tpmProto.PCRs{
						Hash: tpmProto.HashAlgo_SHA256,
						Pcrs: map[uint32][]byte{
							0: {0x00},
						},
					},
				},
			},
			expectedPCRs: measurements.M{
				0: measurements.WithAllBytes(0x00, measurements.Enforce, 1),
				1: measurements.WithAllBytes(0x01, measurements.WarnOnly, 1),
			},
			wantErr: true,
		},
		"no quotes error": {
			quotes: []*tpmProto.Quote{},
			expectedPCRs: measurements.M{
				0: measurements.WithAllBytes(0x00, measurements.Enforce, 1),
				1: measurements.WithAllBytes(0x01, measurements.WarnOnly, 1),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			b := &strings.Builder{}
			parser := &defaultAttestationDocFormatter{}

			err := parser.parseQuotes(b, tc.quotes, tc.expectedPCRs)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantOutput, b.String())
			}
		})
	}
}

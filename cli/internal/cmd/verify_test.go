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

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
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
		idFile             *clusterid.File
		wantEndpoint       string
		skipConfigCreation bool
		wantErr            bool
	}{
		"gcp": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantEndpoint:     "192.0.2.1:1234",
			formatter:        &stubAttDocFormatter{},
		},
		"azure": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantEndpoint:     "192.0.2.1:1234",
			formatter:        &stubAttDocFormatter{},
		},
		"default port": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			wantEndpoint:     "192.0.2.1:" + strconv.Itoa(constants.VerifyServiceNodePortGRPC),
			formatter:        &stubAttDocFormatter{},
		},
		"endpoint not set": {
			provider:      cloudprovider.GCP,
			clusterIDFlag: zeroBase64,
			protoClient:   &stubVerifyClient{},
			formatter:     &stubAttDocFormatter{},
			wantErr:       true,
		},
		"endpoint from id file": {
			provider:      cloudprovider.GCP,
			clusterIDFlag: zeroBase64,
			protoClient:   &stubVerifyClient{},
			idFile:        &clusterid.File{IP: "192.0.2.1"},
			wantEndpoint:  "192.0.2.1:" + strconv.Itoa(constants.VerifyServiceNodePortGRPC),
			formatter:     &stubAttDocFormatter{},
		},
		"override endpoint from details file": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.2:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			idFile:           &clusterid.File{IP: "192.0.2.1"},
			wantEndpoint:     "192.0.2.2:1234",
			formatter:        &stubAttDocFormatter{},
		},
		"invalid endpoint": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: ":::::",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
			formatter:        &stubAttDocFormatter{},
			wantErr:          true,
		},
		"neither owner id nor cluster id set": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			formatter:        &stubAttDocFormatter{},
			wantErr:          true,
		},
		"use owner id from id file": {
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			protoClient:      &stubVerifyClient{},
			idFile:           &clusterid.File{OwnerID: zeroBase64},
			wantEndpoint:     "192.0.2.1:1234",
			formatter:        &stubAttDocFormatter{},
		},
		"config file not existing": {
			provider:           cloudprovider.GCP,
			clusterIDFlag:      zeroBase64,
			nodeEndpointFlag:   "192.0.2.1:1234",
			formatter:          &stubAttDocFormatter{},
			skipConfigCreation: true,
			wantErr:            true,
		},
		"error protoClient GetState": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{verifyErr: rpcStatus.Error(codes.Internal, "failed")},
			formatter:        &stubAttDocFormatter{},
			wantErr:          true,
		},
		"error protoClient GetState not rpc": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{verifyErr: someErr},
			formatter:        &stubAttDocFormatter{},
			wantErr:          true,
		},
		"format error": {
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			clusterIDFlag:    zeroBase64,
			protoClient:      &stubVerifyClient{},
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
			cmd.Flags().String("workspace", "", "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")     // register persistent flag manually
			out := &bytes.Buffer{}
			cmd.SetErr(out)
			if tc.clusterIDFlag != "" {
				require.NoError(cmd.Flags().Set("cluster-id", tc.clusterIDFlag))
			}
			if tc.nodeEndpointFlag != "" {
				require.NoError(cmd.Flags().Set("node-endpoint", tc.nodeEndpointFlag))
			}
			fileHandler := file.NewHandler(afero.NewMemMapFs())

			if !tc.skipConfigCreation {
				cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), tc.provider)
				require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, cfg))
			}
			if tc.idFile != nil {
				require.NoError(fileHandler.WriteJSON(constants.ClusterIDsFilename, tc.idFile, file.OptNone))
			}

			v := &verifyCmd{log: logger.NewTest(t)}
			formatterFac := func(_ bool) attestationDocFormatter {
				return tc.formatter
			}
			err := v.verify(cmd, fileHandler, tc.protoClient, formatterFac, stubAttestationFetcher{})
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

func (f *stubAttDocFormatter) format(_ context.Context, _ string, _ bool, _ bool, _ measurements.M, _ string) (string, error) {
	return "", f.formatErr
}

func TestFormat(t *testing.T) {
	formatter := func() *attestationDocFormatterImpl {
		return &attestationDocFormatterImpl{
			log: logger.NewTest(t),
		}
	}

	testCases := map[string]struct {
		formatter *attestationDocFormatterImpl
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
			_, err := tc.formatter.format(context.Background(), tc.doc, false, false, nil, "")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseCerts(t *testing.T) {
	validCert := `-----BEGIN CERTIFICATE-----
MIIFTDCCAvugAwIBAgIBADBGBgkqhkiG9w0BAQowOaAPMA0GCWCGSAFlAwQCAgUA
oRwwGgYJKoZIhvcNAQEIMA0GCWCGSAFlAwQCAgUAogMCATCjAwIBATB7MRQwEgYD
VQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDASBgNVBAcMC1NhbnRhIENs
YXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5jZWQgTWljcm8gRGV2aWNl
czESMBAGA1UEAwwJU0VWLU1pbGFuMB4XDTIyMTEyMzIyMzM0N1oXDTI5MTEyMzIy
MzM0N1owejEUMBIGA1UECwwLRW5naW5lZXJpbmcxCzAJBgNVBAYTAlVTMRQwEgYD
VQQHDAtTYW50YSBDbGFyYTELMAkGA1UECAwCQ0ExHzAdBgNVBAoMFkFkdmFuY2Vk
IE1pY3JvIERldmljZXMxETAPBgNVBAMMCFNFVi1WQ0VLMHYwEAYHKoZIzj0CAQYF
K4EEACIDYgAEVGm4GomfpkiziqEYP61nfKaz5OjDLr8Y0POrv4iAnFVHAmBT81Ms
gfSLKL5r3V3mNzl1Zh7jwSBft14uhGdwpARoK0YNQc4OvptqVIiv2RprV53DMzge
rtwiumIargiCo4IBFjCCARIwEAYJKwYBBAGceAEBBAMCAQAwFwYJKwYBBAGceAEC
BAoWCE1pbGFuLUIwMBEGCisGAQQBnHgBAwEEAwIBAzARBgorBgEEAZx4AQMCBAMC
AQAwEQYKKwYBBAGceAEDBAQDAgEAMBEGCisGAQQBnHgBAwUEAwIBADARBgorBgEE
AZx4AQMGBAMCAQAwEQYKKwYBBAGceAEDBwQDAgEAMBEGCisGAQQBnHgBAwMEAwIB
CDARBgorBgEEAZx4AQMIBAMCAXMwTQYJKwYBBAGceAEEBEB80kCZ1oAyCjWC6w3m
xOz+i4t6dFjk/Bqhm7+Jscf8D62CXtlwcKc4aM9CdO4LuKlwpdTU80VNQc6ZEuMF
VzbRMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0B
AQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBA4ICAQCN1qBYOywoZWGnQvk6u0Oh
5zkEKykXU6sK8hA6L65rQcqWUjEHDa9AZUpx3UuCmpPc24dx6DTHc58M7TxcyKry
8s4CvruBKFbQ6B8MHnH6k07MzsmiBnsiIhAscZ0ipGm6h8e/VM/6ULrAcVSxZ+Mh
D/IogZAuCQARsGQ4QYXBT8Qc5mLnTkx30m1rZVlp1VcN4ngOo/1tz1jj1mfpG2zv
wNcQa9LwAzRLnnmLpxXA2OMbl7AaTWQenpL9rzBON2sg4OBl6lVhaSU0uBbFyCmR
RvBqKC0iDD6TvyIikkMq05v5YwIKFYw++ICndz+fKcLEULZbziAsZ52qjM8iPVHC
pN0yhVOr2g22F9zxlGH3WxTl9ymUytuv3vJL/aJiQM+n/Ri90Sc05EK4oIJ3+BS8
yu5cVy9o2cQcOcQ8rhQh+Kv1sR9xrs25EXZF8KEETfhoJnN6KY1RwG7HsOfAQ3dV
LWInQRaC/8JPyVS2zbd0+NRBJOnq4/quv/P3C4SBP98/ZuGrqN59uifyqC3Kodkl
WkG/2UdhiLlCmOtsU+BYDZrSiYK1R9FNnlQCOGrkuVxpDwa2TbbvEEzQP7RXxotA
KlxejvrY4VuK8agNqvffVofbdIIperK65K4+0mYIb+A6fU8QQHlCbti4ERSZ6UYD
F/SjRih31+SAtWb42jueAA==
-----END CERTIFICATE-----
`
	validCertExpected := "\tRaw Some Cert:\n\t\t-----BEGIN CERTIFICATE-----\n\t\tMIIFTDCCAvugAwIBAgIBADBGBgkqhkiG9w0BAQowOaAPMA0GCWCGSAFlAwQCAgUA\n\t\toRwwGgYJKoZIhvcNAQEIMA0GCWCGSAFlAwQCAgUAogMCATCjAwIBATB7MRQwEgYD\n\t\tVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDASBgNVBAcMC1NhbnRhIENs\n\t\tYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5jZWQgTWljcm8gRGV2aWNl\n\t\tczESMBAGA1UEAwwJU0VWLU1pbGFuMB4XDTIyMTEyMzIyMzM0N1oXDTI5MTEyMzIy\n\t\tMzM0N1owejEUMBIGA1UECwwLRW5naW5lZXJpbmcxCzAJBgNVBAYTAlVTMRQwEgYD\n\t\tVQQHDAtTYW50YSBDbGFyYTELMAkGA1UECAwCQ0ExHzAdBgNVBAoMFkFkdmFuY2Vk\n\t\tIE1pY3JvIERldmljZXMxETAPBgNVBAMMCFNFVi1WQ0VLMHYwEAYHKoZIzj0CAQYF\n\t\tK4EEACIDYgAEVGm4GomfpkiziqEYP61nfKaz5OjDLr8Y0POrv4iAnFVHAmBT81Ms\n\t\tgfSLKL5r3V3mNzl1Zh7jwSBft14uhGdwpARoK0YNQc4OvptqVIiv2RprV53DMzge\n\t\trtwiumIargiCo4IBFjCCARIwEAYJKwYBBAGceAEBBAMCAQAwFwYJKwYBBAGceAEC\n\t\tBAoWCE1pbGFuLUIwMBEGCisGAQQBnHgBAwEEAwIBAzARBgorBgEEAZx4AQMCBAMC\n\t\tAQAwEQYKKwYBBAGceAEDBAQDAgEAMBEGCisGAQQBnHgBAwUEAwIBADARBgorBgEE\n\t\tAZx4AQMGBAMCAQAwEQYKKwYBBAGceAEDBwQDAgEAMBEGCisGAQQBnHgBAwMEAwIB\n\t\tCDARBgorBgEEAZx4AQMIBAMCAXMwTQYJKwYBBAGceAEEBEB80kCZ1oAyCjWC6w3m\n\t\txOz+i4t6dFjk/Bqhm7+Jscf8D62CXtlwcKc4aM9CdO4LuKlwpdTU80VNQc6ZEuMF\n\t\tVzbRMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0B\n\t\tAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBA4ICAQCN1qBYOywoZWGnQvk6u0Oh\n\t\t5zkEKykXU6sK8hA6L65rQcqWUjEHDa9AZUpx3UuCmpPc24dx6DTHc58M7TxcyKry\n\t\t8s4CvruBKFbQ6B8MHnH6k07MzsmiBnsiIhAscZ0ipGm6h8e/VM/6ULrAcVSxZ+Mh\n\t\tD/IogZAuCQARsGQ4QYXBT8Qc5mLnTkx30m1rZVlp1VcN4ngOo/1tz1jj1mfpG2zv\n\t\twNcQa9LwAzRLnnmLpxXA2OMbl7AaTWQenpL9rzBON2sg4OBl6lVhaSU0uBbFyCmR\n\t\tRvBqKC0iDD6TvyIikkMq05v5YwIKFYw++ICndz+fKcLEULZbziAsZ52qjM8iPVHC\n\t\tpN0yhVOr2g22F9zxlGH3WxTl9ymUytuv3vJL/aJiQM+n/Ri90Sc05EK4oIJ3+BS8\n\t\tyu5cVy9o2cQcOcQ8rhQh+Kv1sR9xrs25EXZF8KEETfhoJnN6KY1RwG7HsOfAQ3dV\n\t\tLWInQRaC/8JPyVS2zbd0+NRBJOnq4/quv/P3C4SBP98/ZuGrqN59uifyqC3Kodkl\n\t\tWkG/2UdhiLlCmOtsU+BYDZrSiYK1R9FNnlQCOGrkuVxpDwa2TbbvEEzQP7RXxotA\n\t\tKlxejvrY4VuK8agNqvffVofbdIIperK65K4+0mYIb+A6fU8QQHlCbti4ERSZ6UYD\n\t\tF/SjRih31+SAtWb42jueAA==\n\t\t-----END CERTIFICATE-----\n\tSome Cert (1):\n\t\tSerial Number: 0\n\t\tSubject: CN=SEV-VCEK,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US\n\t\tIssuer: CN=SEV-Milan,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US\n\t\tNot Before: 2022-11-23 22:33:47 +0000 UTC\n\t\tNot After: 2029-11-23 22:33:47 +0000 UTC\n\t\tSignature Algorithm: SHA384-RSAPSS\n\t\tPublic Key Algorithm: ECDSA\n"

	testCases := map[string]struct {
		cert     []byte
		expected string
		wantErr  bool
	}{
		"one cert": {
			cert:     []byte(validCert),
			expected: validCertExpected,
		},
		"one cert with extra newlines": {
			cert:     []byte("\n\n" + validCert + "\n\n"),
			expected: validCertExpected,
		},
		"invalid cert": {
			cert:    []byte("invalid"),
			wantErr: true,
		},
		"no cert": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			b := &strings.Builder{}
			formatter := &attestationDocFormatterImpl{
				log: logger.NewTest(t),
			}
			err := formatter.parseCerts(b, "Some Cert", tc.cert)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.expected, b.String())
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
			parser := &attestationDocFormatterImpl{}

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

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/aws/snp/testdata"
	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/proto/sevsnp"
	spb "github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/verify"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTrustedKey(t *testing.T) {
	validator := func() *Validator { return &Validator{reportValidator: stubawsValidator{}} }
	testCases := map[string]struct {
		akPub              []byte
		info               []byte
		wantErr            bool
		assertCorrectError func(error)
	}{
		"null byte docs": {
			akPub:   []byte{0x00, 0x00, 0x00, 0x00},
			info:    []byte{0x00, 0x00, 0x00, 0x00},
			wantErr: true,
			assertCorrectError: func(err error) {
				target := &decodeError{}
				assert.ErrorAs(t, err, &target)
			},
		},
		"nil": {
			akPub:   nil,
			info:    nil,
			wantErr: true,
			assertCorrectError: func(err error) {
				target := &decodeError{}
				assert.ErrorAs(t, err, &target)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			out, err := validator().getTrustedKey(
				context.Background(),
				vtpm.AttestationDocument{
					Attestation: &attest.Attestation{
						AkPub: tc.akPub,
					},
					InstanceInfo: tc.info,
				},
				nil,
			)

			if tc.wantErr {
				assert.Error(err)
				tc.assertCorrectError(err)
			} else {
				assert.NoError(err)
			}

			assert.Nil(out)
		})
	}
}

// TestValidateSNPReport has to setup the following to run ValidateSNPReport:
// - parse ARK certificate from constants.go.
// - parse cached ASK certificate.
// - parse cached SNP report.
// - parse cached AK hash. Hash and SNP report have to match.
// - parse cache VLEK cert.
func TestValidateSNPReport(t *testing.T) {
	require := require.New(t)
	certs, err := loadCerts(testdata.CertChain)
	require.NoError(err)
	ark := certs[1]
	ask := certs[0]

	// reportTransformer unpacks the base64 encoded report, applies the given transformations and re-encodes it.
	reportTransformer := func(reportHex string, transformations func(*spb.Report)) string {
		rawReport, err := base64.StdEncoding.DecodeString(reportHex)
		require.NoError(err)
		report, err := abi.ReportToProto(rawReport)
		require.NoError(err)
		transformations(report)
		reportBytes, err := abi.ReportToAbiBytes(report)
		require.NoError(err)
		return base64.StdEncoding.EncodeToString(reportBytes)
	}

	testCases := map[string]struct {
		ak                string
		report            string
		reportTransformer func(string, func(*spb.Report)) string
		verifier          reportVerifier
		validator         reportValidator
		wantErr           bool
	}{
		"success": {
			ak:        testdata.AKDigest,
			report:    testdata.SNPReport,
			verifier:  &reportVerifierImpl{},
			validator: &reportValidatorImpl{},
		},
		"invalid report data": {
			ak: testdata.AKDigest,
			report: reportTransformer(testdata.SNPReport, func(r *spb.Report) {
				r.ReportData = make([]byte, 64)
			}),
			verifier:  &stubReportVerifier{},
			validator: &reportValidatorImpl{},
			wantErr:   true,
		},
		"invalid report signature": {
			ak:        testdata.AKDigest,
			report:    reportTransformer(testdata.SNPReport, func(r *spb.Report) { r.Signature[0]++ }),
			verifier:  &reportVerifierImpl{},
			validator: &reportValidatorImpl{},
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			hash, err := hex.DecodeString(tc.ak)
			require.NoError(err)

			report, err := base64.StdEncoding.DecodeString(tc.report)
			require.NoError(err)

			info := snp.InstanceInfo{AttestationReport: report, ReportSigner: testdata.VLEK}
			infoMarshalled, err := json.Marshal(info)
			require.NoError(err)

			v := awsValidator{httpsGetter: newStubHTTPSGetter(&urlResponseMatcher{}, nil), verifier: tc.verifier, validator: tc.validator}
			err = v.validate(vtpm.AttestationDocument{InstanceInfo: infoMarshalled}, ask, ark, [64]byte(hash), config.DefaultForAWSSEVSNP(), logger.NewTest(t))
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubHTTPSGetter struct {
	urlResponseMatcher *urlResponseMatcher // maps responses to requested URLs
	err                error
}

func newStubHTTPSGetter(urlResponseMatcher *urlResponseMatcher, err error) *stubHTTPSGetter {
	return &stubHTTPSGetter{
		urlResponseMatcher: urlResponseMatcher,
		err:                err,
	}
}

func (s *stubHTTPSGetter) Get(url string) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.urlResponseMatcher.match(url)
}

type urlResponseMatcher struct {
	certChainResponse    []byte
	wantCertChainRequest bool
	vcekResponse         []byte
	wantVcekRequest      bool
}

func (m *urlResponseMatcher) match(url string) ([]byte, error) {
	switch {
	case url == "https://kdsintf.amd.com/vcek/v1/Milan/cert_chain":
		if !m.wantCertChainRequest {
			return nil, fmt.Errorf("unexpected cert_chain request")
		}
		return m.certChainResponse, nil
	case regexp.MustCompile(`https:\/\/kdsintf.amd.com\/vcek\/v1\/Milan\/.*`).MatchString(url):
		if !m.wantVcekRequest {
			return nil, fmt.Errorf("unexpected VCEK request")
		}
		return m.vcekResponse, nil
	default:
		return nil, fmt.Errorf("unexpected URL: %s", url)
	}
}

func TestSha512sum(t *testing.T) {
	testCases := map[string]struct {
		key   string
		hash  string
		match bool
	}{
		"success": {
			// Generated using: rsa.GenerateKey(rand.Reader, 1024).
			key:   "30819f300d06092a864886f70d010101050003818d0030818902818100d4b2f072a32fa98456eb7f5938e2ff361fb64d698ea91e003d34bfc5374b814c16ba9ae3ec392ef6d48cf79b63067e338aa941219a7bcdf18aa43cd38bbe5567504838a3b1dca482035458853c5a171709dfae9df551815010bdfbc6df733cde84c4f7a5b0591d9cda9db087fb411ee3e2a4f19ad50c8331712ecdc5dd7ce34b0203010001",
			hash:  "2d6fe5ec59d7240b8a4c27c2ff27ba1071105fa50d45543768fcbabf9ee3cb8f8fa0afa51e08e053af30f6d11066ebfd47e75bda5ccc085c115d7e1896f3c62f",
			match: true,
		},
		"mismatching hash": {
			key:   "30819f300d06092a864886f70d010101050003818d0030818902818100d4b2f072a32fa98456eb7f5938e2ff361fb64d698ea91e003d34bfc5374b814c16ba9ae3ec392ef6d48cf79b63067e338aa941219a7bcdf18aa43cd38bbe5567504838a3b1dca482035458853c5a171709dfae9df551815010bdfbc6df733cde84c4f7a5b0591d9cda9db087fb411ee3e2a4f19ad50c8331712ecdc5dd7ce34b0203010001",
			hash:  "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			match: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			newKey, err := loadKeyFromHex(tc.key)
			require.NoError(err)

			// Function under test:
			hash, err := sha512sum(newKey)
			assert.NoError(err)

			expected, err := hex.DecodeString(tc.hash)
			require.NoError(err)

			if tc.match {
				assert.True(bytes.Equal(expected, hash[:]), fmt.Sprintf("expected hash %x, got %x", expected, hash))
			} else {
				assert.False(bytes.Equal(expected, hash[:]), fmt.Sprintf("expected mismatching hashes, got %x", hash))
			}
		})
	}
}

func loadKeyFromHex(key string) (crypto.PublicKey, error) {
	decoded, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	return x509.ParsePKIXPublicKey(decoded)
}

// loadCachedCertChain loads a valid ARK and ASK from the testdata folder.
func loadCerts(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, errors.New("no valid certificates found")
	}

	return certs, nil
}

type stubawsValidator struct{}

func (stubawsValidator) validate(_ vtpm.AttestationDocument, _ *x509.Certificate, _ *x509.Certificate, _ [64]byte, _ *config.AWSSEVSNP, _ attestation.Logger) error {
	return nil
}

type stubReportVerifier struct{}

func (stubReportVerifier) SnpAttestation(_ *sevsnp.Attestation, _ *verify.Options) error {
	return nil
}

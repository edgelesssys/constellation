/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/snp/testdata"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/google/go-sev-guest/kds"
	"github.com/google/go-sev-guest/verify/trust"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseCertChain tests the parsing of the certificate chain.
func TestParseCertChain(t *testing.T) {
	defaultCertChain := testdata.CertChain
	askOnly := strings.Split(string(defaultCertChain), "-----END CERTIFICATE-----")[0] + "-----END CERTIFICATE-----"
	arkOnly := strings.Split(string(defaultCertChain), "-----END CERTIFICATE-----")[1] + "-----END CERTIFICATE-----"

	testCases := map[string]struct {
		certChain []byte
		wantAsk   bool
		wantArk   bool
		wantErr   bool
	}{
		"success": {
			certChain: defaultCertChain,
			wantAsk:   true,
			wantArk:   true,
		},
		"empty cert chain": {
			certChain: []byte{},
			wantErr:   true,
		},
		"more than two certificates": {
			certChain: append(defaultCertChain, defaultCertChain...),
			wantErr:   true,
		},
		"invalid certificate": {
			certChain: []byte("invalid"),
			wantErr:   true,
		},
		"ark missing": {
			certChain: []byte(askOnly),
			wantAsk:   true,
		},
		"ask missing": {
			certChain: []byte(arkOnly),
			wantArk:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			instanceInfo := &InstanceInfo{
				CertChain: tc.certChain,
			}

			ask, ark, err := instanceInfo.ParseCertChain()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantAsk, ask != nil)
				assert.Equal(tc.wantArk, ark != nil)
			}
		})
	}
}

// TestParseVCEK tests the parsing of the VCEK certificate.
func TestParseVCEK(t *testing.T) {
	testCases := map[string]struct {
		VCEK     []byte
		wantVCEK bool
		wantErr  bool
	}{
		"success": {
			VCEK:     testdata.AzureThimVCEK,
			wantVCEK: true,
		},
		"empty": {
			VCEK: []byte{},
		},
		"malformed": {
			VCEK:    testdata.AzureThimVCEK[:len(testdata.AzureThimVCEK)-100],
			wantErr: true,
		},
		"invalid": {
			VCEK:    []byte("invalid"),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			instanceInfo := &InstanceInfo{
				ReportSigner: tc.VCEK,
			}

			vcek, err := instanceInfo.ParseReportSigner()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantVCEK, vcek != nil)
			}
		})
	}
}

// TestAttestationWithCerts tests the basic unmarshalling of the attestation report and the ASK / ARK precedence.
func TestAttestationWithCerts(t *testing.T) {
	defaultReport := testdata.AttestationReport
	vlekReport, err := hex.DecodeString(testdata.AttestationReportVLEK)
	require.NoError(t, err)
	testdataArk, testdataAsk := mustCertChainToPem(t, testdata.CertChain)
	testdataArvk, testdataAsvk := mustCertChainToPem(t, testdata.VlekCertChain)
	exampleCert := &x509.Certificate{
		Raw: []byte{1, 2, 3},
	}
	cfg := config.DefaultForAzureSEVSNP()

	testCases := map[string]struct {
		report        []byte
		idkeydigest   string
		reportSigner  []byte
		certChain     []byte
		fallbackCerts CertificateChain
		getter        *stubHTTPSGetter
		expectedArk   *x509.Certificate
		expectedAsk   *x509.Certificate
		wantErr       bool
	}{
		"success": {
			report:       defaultReport,
			idkeydigest:  "57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1",
			reportSigner: testdata.AzureThimVCEK,
			certChain:    testdata.CertChain,
			expectedArk:  testdataArk,
			expectedAsk:  testdataAsk,
		},
		"vlek success": {
			report:       vlekReport,
			idkeydigest:  "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			reportSigner: testdata.Vlek,
			expectedArk:  testdataArvk,
			expectedAsk:  testdataAsvk,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					certChainResponse:    testdata.VlekCertChain,
					vcekResponse:         testdata.Vlek,
					wantCertChainRequest: true,
					wantVcekRequest:      true,
				},
				nil,
			),
		},
		"retrieve vcek": {
			report:      defaultReport,
			idkeydigest: "57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1",
			certChain:   testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					vcekResponse:    testdata.AmdKdsVCEK,
					wantVcekRequest: true,
				},
				nil,
			),
			expectedArk: testdataArk,
			expectedAsk: testdataAsk,
		},
		"retrieve certchain": {
			report:       defaultReport,
			idkeydigest:  "57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1",
			reportSigner: testdata.AzureThimVCEK,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					certChainResponse:    testdata.CertChain,
					wantCertChainRequest: true,
				},
				nil,
			),
			expectedArk: testdataArk,
			expectedAsk: testdataAsk,
		},
		"use fallback certs": {
			report:        defaultReport,
			idkeydigest:   "57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1",
			reportSigner:  testdata.AzureThimVCEK,
			fallbackCerts: NewCertificateChain(exampleCert, exampleCert),
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			expectedArk: exampleCert,
			expectedAsk: exampleCert,
		},
		"use certchain with fallback certs": {
			report:        defaultReport,
			idkeydigest:   "57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1",
			certChain:     testdata.CertChain,
			reportSigner:  testdata.AzureThimVCEK,
			fallbackCerts: NewCertificateChain(&x509.Certificate{}, &x509.Certificate{}),
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			expectedArk: testdataArk,
			expectedAsk: testdataAsk,
		},
		"retrieve vcek and certchain": {
			report:      defaultReport,
			idkeydigest: "57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1",
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					certChainResponse:    testdata.CertChain,
					vcekResponse:         testdata.AmdKdsVCEK,
					wantCertChainRequest: true,
					wantVcekRequest:      true,
				},
				nil,
			),
			expectedArk: testdataArk,
			expectedAsk: testdataAsk,
		},
		"report too short": {
			report:  defaultReport[:len(defaultReport)-100],
			wantErr: true,
		},
		"corrupted report": {
			report:  defaultReport[10 : len(defaultReport)-10],
			wantErr: true,
		},
		"certificate fetch error": {
			report:  defaultReport,
			getter:  newStubHTTPSGetter(nil, assert.AnError),
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			instanceInfo := InstanceInfo{
				AttestationReport: tc.report,
				CertChain:         tc.certChain,
				ReportSigner:      tc.reportSigner,
			}

			defer trust.ClearProductCertCache()
			att, err := instanceInfo.AttestationWithCerts(tc.getter, tc.fallbackCerts, slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)))
			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.NotNil(att)
				assert.NotNil(att.CertificateChain)
				assert.NotNil(att.Report)

				assert.Equal(tc.idkeydigest, hex.EncodeToString(att.Report.IdKeyDigest[:]))

				// This is a canary for us: If this fails in the future we possibly downgraded a SVN.
				// See https://github.com/google/go-sev-guest/blob/14ac50e9ffcc05cd1d12247b710c65093beedb58/validate/validate.go#L336 for decomposition of the values.
				tcbValues := kds.DecomposeTCBVersion(kds.TCBVersion(att.Report.GetLaunchTcb()))
				assert.True(tcbValues.BlSpl >= cfg.BootloaderVersion.Value)
				assert.True(tcbValues.TeeSpl >= cfg.TEEVersion.Value)
				assert.True(tcbValues.SnpSpl >= cfg.SNPVersion.Value)
				assert.True(tcbValues.UcodeSpl >= cfg.MicrocodeVersion.Value)
				assert.Equal(tc.expectedArk.Raw, att.CertificateChain.ArkCert)
				assert.Equal(tc.expectedAsk.Raw, att.CertificateChain.AskCert)
			}
		})
	}
}

func mustCertChainToPem(t *testing.T, certchain []byte) (ark, ask *x509.Certificate) {
	t.Helper()
	a := InstanceInfo{CertChain: certchain}
	ask, ark, err := a.ParseCertChain()
	require.NoError(t, err)
	return ark, ask
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
	case regexp.MustCompile(`https:\/\/kdsintf.amd.com\/(vcek|vlek)\/v1\/Milan\/cert_chain`).MatchString(url):
		if !m.wantCertChainRequest {
			return nil, fmt.Errorf("unexpected cert_chain request")
		}
		return m.certChainResponse, nil
	case regexp.MustCompile(`https:\/\/kdsintf.amd.com\/(vcek|vlek)\/v1\/Milan\/.*`).MatchString(url):
		if !m.wantVcekRequest {
			return nil, fmt.Errorf("unexpected VCEK request")
		}
		return m.vcekResponse, nil
	default:
		return nil, fmt.Errorf("unexpected URL: %s", url)
	}
}

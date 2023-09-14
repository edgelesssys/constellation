/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp/testdata"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	spb "github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/validate"
	"github.com/google/go-sev-guest/verify"
	"github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewValidator tests the creation of a new validator.
func TestNewValidator(t *testing.T) {
	require := require.New(t)

	testCases := map[string]struct {
		cfg    *config.AzureSEVSNP
		logger attestation.Logger
	}{
		"success": {
			cfg:    config.DefaultForAzureSEVSNP(),
			logger: logger.NewTest(t),
		},
		"nil logger": {
			cfg:    config.DefaultForAzureSEVSNP(),
			logger: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			validator := NewValidator(tc.cfg, tc.logger)
			require.NotNil(validator)
			require.NotNil(validator.log)
		})
	}
}

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

			instanceInfo := &azureInstanceInfo{
				CertChain: tc.certChain,
			}

			ask, ark, err := instanceInfo.parseCertChain()
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

			instanceInfo := &azureInstanceInfo{
				VCEK: tc.VCEK,
			}

			vcek, err := instanceInfo.parseVCEK()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantVCEK, vcek != nil)
			}
		})
	}
}

// TestInstanceInfoAttestation tests the basic unmarshalling of the attestation report and the ASK / ARK precedence.
func TestInstanceInfoAttestation(t *testing.T) {
	defaultReport := testdata.AttestationReport
	cfg := config.DefaultForAzureSEVSNP()

	testCases := map[string]struct {
		report        []byte
		vcek          []byte
		certChain     []byte
		fallbackCerts sevSnpCerts
		getter        *stubHTTPSGetter
		wantErr       bool
	}{
		"success": {
			report:    defaultReport,
			vcek:      testdata.AzureThimVCEK,
			certChain: testdata.CertChain,
		},
		"retrieve vcek": {
			report:    defaultReport,
			certChain: testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					vcekResponse:    testdata.AmdKdsVCEK,
					wantVcekRequest: true,
				},
				nil,
			),
		},
		"retrieve certchain": {
			report: defaultReport,
			vcek:   testdata.AzureThimVCEK,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					certChainResponse:    testdata.CertChain,
					wantCertChainRequest: true,
				},
				nil,
			),
		},
		"use fallback certs": {
			report: defaultReport,
			vcek:   testdata.AzureThimVCEK,
			fallbackCerts: sevSnpCerts{
				ask: &x509.Certificate{},
				ark: &x509.Certificate{},
			},
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
		},
		"retrieve vcek and certchain": {
			report: defaultReport,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					certChainResponse:    testdata.CertChain,
					vcekResponse:         testdata.AmdKdsVCEK,
					wantCertChainRequest: true,
					wantVcekRequest:      true,
				},
				nil,
			),
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

			instanceInfo := azureInstanceInfo{
				AttestationReport: tc.report,
				CertChain:         tc.certChain,
				VCEK:              tc.vcek,
			}

			att, err := instanceInfo.attestationWithCerts(logger.NewTest(t), tc.getter, tc.fallbackCerts)
			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.NotNil(att)
				assert.NotNil(att.CertificateChain)
				assert.NotNil(att.Report)

				assert.Equal(hex.EncodeToString(att.Report.IdKeyDigest[:]), "57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1")

				// This is a canary for us: If this fails in the future we possibly downgraded a SVN.
				// See https://github.com/google/go-sev-guest/blob/14ac50e9ffcc05cd1d12247b710c65093beedb58/validate/validate.go#L336 for decomposition of the values.
				tcbValues := kds.DecomposeTCBVersion(kds.TCBVersion(att.Report.GetLaunchTcb()))
				assert.True(tcbValues.BlSpl >= cfg.BootloaderVersion.Value)
				assert.True(tcbValues.TeeSpl >= cfg.TEEVersion.Value)
				assert.True(tcbValues.SnpSpl >= cfg.SNPVersion.Value)
				assert.True(tcbValues.UcodeSpl >= cfg.MicrocodeVersion.Value)
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

// TestCheckIDKeyDigest tests validation of an IDKeyDigest under different enforcement policies.
func TestCheckIDKeyDigest(t *testing.T) {
	cfgWithAcceptedIDKeyDigests := func(enforcementPolicy idkeydigest.Enforcement, digestStrings []string) *config.AzureSEVSNP {
		digests := idkeydigest.List{}
		for _, digest := range digestStrings {
			digests = append(digests, []byte(digest))
		}
		cfg := config.DefaultForAzureSEVSNP()
		cfg.FirmwareSignerConfig.AcceptedKeyDigests = digests
		cfg.FirmwareSignerConfig.EnforcementPolicy = enforcementPolicy
		return cfg
	}
	reportWithIDKeyDigest := func(idKeyDigest string) *spb.Attestation {
		report := &spb.Attestation{}
		report.Report = &spb.Report{}
		report.Report.IdKeyDigest = []byte(idKeyDigest)
		return report
	}
	newTestValidator := func(cfg *config.AzureSEVSNP, validateTokenErr error) *Validator {
		validator := NewValidator(cfg, logger.NewTest(t))
		validator.maa = &stubMaaValidator{
			validateTokenErr: validateTokenErr,
		}
		return validator
	}

	testCases := map[string]struct {
		idKeyDigest          string
		acceptedIDKeyDigests []string
		enforcementPolicy    idkeydigest.Enforcement
		validateMaaTokenErr  error
		wantErr              bool
	}{
		"matching digest": {
			idKeyDigest:          "test",
			acceptedIDKeyDigests: []string{"test"},
		},
		"no accepted digests": {
			idKeyDigest:          "test",
			acceptedIDKeyDigests: []string{},
			wantErr:              true,
		},
		"mismatching digest, enforce": {
			idKeyDigest:          "test",
			acceptedIDKeyDigests: []string{"other"},
			wantErr:              true,
		},
		"mismatching digest, maaFallback": {
			idKeyDigest:          "test",
			acceptedIDKeyDigests: []string{"other"},
			enforcementPolicy:    idkeydigest.MAAFallback,
		},
		"mismatching digest, maaFallback errors": {
			idKeyDigest:          "test",
			acceptedIDKeyDigests: []string{"other"},
			enforcementPolicy:    idkeydigest.MAAFallback,
			validateMaaTokenErr:  errors.New("maa fallback failed"),
			wantErr:              true,
		},
		"mismatching digest, warnOnly": {
			idKeyDigest:          "test",
			acceptedIDKeyDigests: []string{"other"},
			enforcementPolicy:    idkeydigest.WarnOnly,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			cfg := cfgWithAcceptedIDKeyDigests(tc.enforcementPolicy, tc.acceptedIDKeyDigests)
			report := reportWithIDKeyDigest(tc.idKeyDigest)
			validator := newTestValidator(cfg, tc.validateMaaTokenErr)

			err := validator.checkIDKeyDigest(context.Background(), report, "", nil)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

type stubMaaValidator struct {
	validateTokenErr error
}

func (v *stubMaaValidator) validateToken(_ context.Context, _ string, _ string, _ []byte) error {
	return v.validateTokenErr
}

// TestValidateAk tests the attestation key validation with a simulated TPM device.
func TestValidateAk(t *testing.T) {
	cgo := os.Getenv("CGO_ENABLED")
	if cgo == "0" {
		t.Skip("skipping test because CGO is disabled and tpm simulator requires it")
	}

	int32ToBytes := func(val uint32) []byte {
		r := make([]byte, 4)
		binary.PutUvarint(r, uint64(val))
		return r
	}

	require := require.New(t)

	tpm, err := simulator.OpenSimulatedTPM()
	require.NoError(err)
	defer tpm.Close()
	key, err := client.AttestationKeyRSA(tpm)
	require.NoError(err)
	defer key.Close()

	e := base64.RawURLEncoding.EncodeToString(int32ToBytes(key.PublicArea().RSAParameters.ExponentRaw))
	n := base64.RawURLEncoding.EncodeToString(key.PublicArea().RSAParameters.ModulusRaw)

	ak := akPub{E: e, N: n}
	runtimeData := runtimeData{Keys: []akPub{ak}}

	defaultRuntimeDataRaw, err := json.Marshal(runtimeData)
	require.NoError(err)
	defaultInstanceInfo := azureInstanceInfo{RuntimeData: defaultRuntimeDataRaw}

	sig := sha256.Sum256(defaultRuntimeDataRaw)
	defaultReportData := sig[:]
	defaultRsaParams := key.PublicArea().RSAParameters

	testCases := map[string]struct {
		instanceInfo   azureInstanceInfo
		runtimeDataRaw []byte
		reportData     []byte
		rsaParameters  *tpm2.RSAParams
		wantErr        bool
	}{
		"success": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: defaultRuntimeDataRaw,
			reportData:     defaultReportData,
			rsaParameters:  defaultRsaParams,
		},
		"invalid json": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: []byte(""),
			reportData:     defaultReportData,
			rsaParameters:  defaultRsaParams,
			wantErr:        true,
		},
		"invalid hash": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: defaultRuntimeDataRaw,
			reportData:     bytes.Repeat([]byte{0}, 64),
			rsaParameters:  defaultRsaParams,
			wantErr:        true,
		},
		"invalid E": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: defaultRuntimeDataRaw,
			reportData:     defaultReportData,
			rsaParameters: func() *tpm2.RSAParams {
				tmp := *defaultRsaParams
				tmp.ExponentRaw = 1
				return &tmp
			}(),
			wantErr: true,
		},
		"invalid N": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: defaultRuntimeDataRaw,
			reportData:     defaultReportData,
			rsaParameters: func() *tpm2.RSAParams {
				tmp := *defaultRsaParams
				tmp.ModulusRaw = []byte{0, 1, 2, 3}
				return &tmp
			}(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err = tc.instanceInfo.validateAk(tc.runtimeDataRaw, tc.reportData, tc.rsaParameters)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

// TestGetTrustedKey tests the verification and validation of attestation report.
func TestTrustedKeyFromSNP(t *testing.T) {
	cgo := os.Getenv("CGO_ENABLED")
	if cgo == "0" {
		t.Skip("skipping test because CGO is disabled and tpm simulator requires it")
	}
	require := require.New(t)

	tpm, err := simulator.OpenSimulatedTPM()
	require.NoError(err)
	defer tpm.Close()
	key, err := client.AttestationKeyRSA(tpm)
	require.NoError(err)
	defer key.Close()
	akPub, err := key.PublicArea().Encode()
	require.NoError(err)

	defaultCfg := config.DefaultForAzureSEVSNP()
	defaultReport := hex.EncodeToString(testdata.AttestationReport)
	defaultRuntimeData := hex.EncodeToString(testdata.RuntimeData)
	defaultIDKeyDigestOld, err := hex.DecodeString("57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1")
	require.NoError(err)
	defaultIDKeyDigest := idkeydigest.NewList([][]byte{defaultIDKeyDigestOld})
	defaultVerifier := &stubAttestationVerifier{}
	skipVerifier := &stubAttestationVerifier{skipCheck: true}
	defaultValidator := &stubAttestationValidator{}

	// reportTransformer unpacks the hex-encoded report, applies the given transformations and re-encodes it.
	reportTransformer := func(reportHex string, transformations func(*spb.Report)) string {
		rawReport, err := hex.DecodeString(reportHex)
		require.NoError(err)
		report, err := abi.ReportToProto(rawReport)
		require.NoError(err)
		transformations(report)
		reportBytes, err := abi.ReportToAbiBytes(report)
		require.NoError(err)
		return hex.EncodeToString(reportBytes)
	}

	testCases := map[string]struct {
		report               string
		runtimeData          string
		vcek                 []byte
		certChain            []byte
		acceptedIDKeyDigests idkeydigest.List
		enforcementPolicy    idkeydigest.Enforcement
		getter               *stubHTTPSGetter
		verifier             *stubAttestationVerifier
		validator            *stubAttestationValidator
		wantErr              bool
		assertion            func(*assert.Assertions, error)
	}{
		"success": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
		},
		"certificate fetch error": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			getter: newStubHTTPSGetter(
				nil,
				assert.AnError,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "retrieving VCEK certificate from AMD KDS")
			},
		},
		"fetch vcek and certchain": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					vcekResponse:         testdata.AmdKdsVCEK,
					wantVcekRequest:      true,
					certChainResponse:    testdata.CertChain,
					wantCertChainRequest: true,
				},
				nil,
			),
		},
		"fetch vcek": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					vcekResponse:    testdata.AmdKdsVCEK,
					wantVcekRequest: true,
				},
				nil,
			),
		},
		"fetch certchain": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					certChainResponse:    testdata.CertChain,
					wantCertChainRequest: true,
				},
				nil,
			),
		},
		"invalid report signature": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				r.Signature = make([]byte, 512)
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "report signature verification error")
			},
		},
		"invalid vcek": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 []byte("invalid"),
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					vcekResponse:    []byte("invalid"),
					wantVcekRequest: true,
				},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "could not interpret VCEK DER bytes: x509: malformed certificate")
			},
		},
		"invalid certchain fall back to embedded": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            []byte("invalid"),
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{
					certChainResponse:    []byte("invalid"),
					wantCertChainRequest: true,
				},
				nil,
			),
		},
		"invalid runtime data": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData[:len(defaultRuntimeData)-10],
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "validating HCLAkPub: unmarshalling json: unexpected end of JSON input")
			},
		},
		"inacceptable idkeydigest (wrong size), enforce": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: idkeydigest.List{[]byte{0x00}},
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "bad hash size in TrustedIDKeyHashes")
			},
		},
		"inacceptable idkeydigest (wrong value), enforce": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: idkeydigest.List{make([]byte, 48)},
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "report ID key not trusted")
			},
		},
		"inacceptable idkeydigest, warn only": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: idkeydigest.List{[]byte{0x00}},
			enforcementPolicy:    idkeydigest.WarnOnly,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
		},
		"launch tcb < minimum launch tcb": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				launchTcb := kds.DecomposeTCBVersion(kds.TCBVersion(r.LaunchTcb))
				defaultCfg.MicrocodeVersion.Value = 10
				launchTcb.UcodeSpl = 9
				newLaunchTcb, err := kds.ComposeTCBParts(launchTcb)
				require.NoError(err)
				r.LaunchTcb = uint64(newLaunchTcb)
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.WarnOnly,
			verifier:             skipVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "is lower than the policy minimum launch TCB")
			},
		},
		"reported tcb < minimum tcb": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				reportedTcb := kds.DecomposeTCBVersion(kds.TCBVersion(r.ReportedTcb))
				reportedTcb.UcodeSpl = defaultCfg.MicrocodeVersion.Value - 1
				newReportedTcb, err := kds.ComposeTCBParts(reportedTcb)
				require.NoError(err)
				r.ReportedTcb = uint64(newReportedTcb)
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.WarnOnly,
			verifier:             skipVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "is lower than the policy minimum TCB")
			},
		},
		"current tcb < committed tcb": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				r.CurrentTcb = r.CommittedTcb - 1
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.WarnOnly,
			verifier:             skipVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "is lower than the report's COMMITTED_TCB")
			},
		},
		"current tcb < tcb in vcek": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				currentTcb := kds.DecomposeTCBVersion(kds.TCBVersion(r.CurrentTcb))
				currentTcb.UcodeSpl = 0x5c // testdata.AzureThimVCEK has ucode version 0x5d
				newCurrentTcb, err := kds.ComposeTCBParts(currentTcb)
				require.NoError(err)
				r.CurrentTcb = uint64(newCurrentTcb)
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.WarnOnly,
			verifier:             skipVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "is lower than the TCB of the V[CL]EK certificate")
			},
		},
		"reported tcb != tcb in vcek": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				r.ReportedTcb = uint64(0)
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.WarnOnly,
			verifier:             skipVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "does not match the TCB of the V[CL]EK certificate")
			},
		},
		"vmpl != 0": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				r.Vmpl = 1
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             skipVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain:            testdata.CertChain,
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "report VMPL 1 is not 0")
			},
		},
		"invalid ASK": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			verifier:             defaultVerifier,
			validator:            defaultValidator,
			vcek:                 testdata.AzureThimVCEK,
			certChain: func() []byte {
				c := make([]byte, len(testdata.CertChain))
				copy(c, testdata.CertChain)
				c[1676] = 0x41 // somewhere in the ASK signature
				return c
			}(),
			getter: newStubHTTPSGetter(
				&urlResponseMatcher{},
				nil,
			),
			wantErr: true,
			assertion: func(assert *assert.Assertions, err error) {
				assert.ErrorContains(err, "crypto/rsa: verification error")
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			instanceInfo, err := newStubAzureInstanceInfo(tc.vcek, tc.certChain, tc.report, tc.runtimeData)
			assert.NoError(err)

			statement, err := json.Marshal(instanceInfo)
			if err != nil {
				assert.Error(err)
			}

			attDoc := vtpm.AttestationDocument{
				InstanceInfo: statement,
				Attestation: &attest.Attestation{
					AkPub: akPub,
				},
			}

			defaultCfg.FirmwareSignerConfig = config.SNPFirmwareSignerConfig{
				AcceptedKeyDigests: tc.acceptedIDKeyDigests,
				EnforcementPolicy:  tc.enforcementPolicy,
			}

			validator := &Validator{
				hclValidator:         &instanceInfo,
				config:               defaultCfg,
				log:                  logger.NewTest(t),
				getter:               tc.getter,
				attestationVerifier:  tc.verifier,
				attestationValidator: tc.validator,
			}

			key, err := validator.getTrustedKey(context.Background(), attDoc, nil)
			if tc.wantErr {
				require.Error(err)
				if tc.assertion != nil {
					tc.assertion(assert, err)
				}
			} else {
				require.NoError(err)
				assert.NotNil(key)
			}
		})
	}
}

type stubAttestationVerifier struct {
	skipCheck bool // whether the verification function should be called
}

// SNPAttestation verifies the VCEK certificate as well as the certificate chain of the attestation report.
func (v *stubAttestationVerifier) SNPAttestation(attestation *spb.Attestation, options *verify.Options) error {
	if v.skipCheck {
		return nil
	}
	return verify.SnpAttestation(attestation, options)
}

type stubAttestationValidator struct {
	skipCheck bool // whether the verification function should be called
}

// SNPAttestation validates the attestation report against the given set of constraints.
func (v *stubAttestationValidator) SNPAttestation(attestation *spb.Attestation, options *validate.Options) error {
	if v.skipCheck {
		return nil
	}
	return validate.SnpAttestation(attestation, options)
}

type stubAzureInstanceInfo struct {
	AttestationReport []byte
	RuntimeData       []byte
	VCEK              []byte
	CertChain         []byte
}

func newStubAzureInstanceInfo(vcek, certChain []byte, report, runtimeData string) (stubAzureInstanceInfo, error) {
	validReport, err := hex.DecodeString(report)
	if err != nil {
		return stubAzureInstanceInfo{}, fmt.Errorf("invalid hex string report: %s", err)
	}

	decodedRuntime, err := hex.DecodeString(runtimeData)
	if err != nil {
		return stubAzureInstanceInfo{}, fmt.Errorf("invalid hex string runtimeData: %s", err)
	}

	return stubAzureInstanceInfo{
		AttestationReport: validReport,
		RuntimeData:       decodedRuntime,
		VCEK:              vcek,
		CertChain:         certChain,
	}, nil
}

func (s *stubAzureInstanceInfo) validateAk(runtimeDataRaw []byte, reportData []byte, _ *tpm2.RSAParams) error {
	var runtimeData runtimeData
	if err := json.Unmarshal(runtimeDataRaw, &runtimeData); err != nil {
		return fmt.Errorf("unmarshalling json: %w", err)
	}

	sum := sha256.Sum256(runtimeDataRaw)
	if !bytes.Equal(sum[:], reportData[:32]) {
		return fmt.Errorf("unexpected runtimeData digest in TPM")
	}

	return nil
}

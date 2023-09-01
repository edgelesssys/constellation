/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp/testdata"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	spb "github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstanceInfoAttestation tests the basic unmarshalling of the attestation report.
func TestInstanceInfoAttestation(t *testing.T) {
	defaultReport := testdata.AttestationReport
	cfg := config.DefaultForAzureSEVSNP()

	testCases := map[string]struct {
		report  []byte
		getter  *stubHTTPSGetter
		wantErr bool
	}{
		"report too short": {
			report: defaultReport[:len(defaultReport)-100],
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
		"corrupted report": {
			report: defaultReport[10 : len(defaultReport)-10],
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
		"certificate fetch error": {
			report: defaultReport[10 : len(defaultReport)-10],
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				assert.AnError,
			),
			wantErr: true,
		},
		"success": {
			report: defaultReport,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			instanceInfo := azureInstanceInfo{
				AttestationReport: tc.report,
			}

			report, err := instanceInfo.attestationWithCerts(tc.getter)
			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.NotNil(report)

				assert.Equal(hex.EncodeToString(report.Report.IdKeyDigest[:]), "57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1")

				// This is a canary for us: If this fails in the future we possibly downgraded a SVN.
				// See https://github.com/google/go-sev-guest/blob/14ac50e9ffcc05cd1d12247b710c65093beedb58/validate/validate.go#L336 for decomposition of the values.
				tcbValues := kds.DecomposeTCBVersion(kds.TCBVersion(report.Report.GetLaunchTcb()))
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
	return s.urlResponseMatcher.match(url), nil
}

type urlResponseMatcher struct {
	certChainResponse []byte
	vcekResponse      []byte
}

func newUrlResponseMatcher(certChainResponse []byte, vcekResponse []byte) *urlResponseMatcher {
	return &urlResponseMatcher{
		certChainResponse: certChainResponse,
		vcekResponse:      vcekResponse,
	}
}

func (m *urlResponseMatcher) match(url string) []byte {
	switch {
	case url == "https://kdsintf.amd.com/vcek/v1/Milan/cert_chain":
		return m.certChainResponse
	case regexp.MustCompile(`https:\/\/kdsintf.amd.com\/vcek\/v1\/Milan\/.*`).MatchString(url):
		return m.vcekResponse
	}
	return nil
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
	reportWithIdKeyDigest := func(idKeyDigest string) *spb.Attestation {
		report := &spb.Attestation{}
		report.Report = &spb.Report{}
		report.Report.IdKeyDigest = []byte(idKeyDigest)
		return report
	}
	newTestValidator := func(cfg *config.AzureSEVSNP, log *logger.Logger, validateTokenErr error) *Validator {
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
			report := reportWithIdKeyDigest(tc.idKeyDigest)
			validator := newTestValidator(cfg, logger.NewTest(t), tc.validateMaaTokenErr)

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
		acceptedIDKeyDigests idkeydigest.List
		enforcementPolicy    idkeydigest.Enforcement
		getter               *stubHTTPSGetter
		wantErr              bool
	}{
		"success": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
		},
		"certificate fetch error": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				assert.AnError,
			),
			wantErr: true,
		},
		"invalid report signature": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				r.Signature = make([]byte, 512)
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
		"invalid vcek": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, make([]byte, 0)),
				nil,
			),
			wantErr: true,
		},
		// TODO: Find out why this doesn't error.
		"invalid certchain": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(make([]byte, 0), testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
		"invalid runtime data": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData[:len(defaultRuntimeData)-10],
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.Equal,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
		"inacceptable idkeydigest, enforce": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: idkeydigest.List{[]byte{0x00}},
			enforcementPolicy:    idkeydigest.Equal,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
		"inacceptable idkeydigest, warn only": {
			report:               defaultReport,
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: idkeydigest.List{[]byte{0x00}},
			enforcementPolicy:    idkeydigest.WarnOnly,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
		},
		"launch tcb < minimum launch tcb": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				launchTcb := kds.DecomposeTCBVersion(kds.TCBVersion(r.LaunchTcb))
				launchTcb.UcodeSpl = defaultCfg.MicrocodeVersion.Value - 1
				newLaunchTcb, err := kds.ComposeTCBParts(launchTcb)
				require.NoError(err)
				r.LaunchTcb = uint64(newLaunchTcb)
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.WarnOnly,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
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
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
		"current tcb < committed tcb": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				r.CurrentTcb = r.CommittedTcb - 1
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.WarnOnly,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
		"current tcb < tcb in vcek": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				currentTcb := kds.DecomposeTCBVersion(kds.TCBVersion(r.CurrentTcb))
				currentTcb.UcodeSpl = 0x5c // testdata.VCEK has ucode version 0x5d
				newCurrentTcb, err := kds.ComposeTCBParts(currentTcb)
				require.NoError(err)
				r.CurrentTcb = uint64(newCurrentTcb)
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.WarnOnly,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
		"reported tcb != tcb in vcek": {
			report: reportTransformer(defaultReport, func(r *spb.Report) {
				r.ReportedTcb = uint64(0)
			}),
			runtimeData:          defaultRuntimeData,
			acceptedIDKeyDigests: defaultIDKeyDigest,
			enforcementPolicy:    idkeydigest.WarnOnly,
			getter: newStubHTTPSGetter(
				newUrlResponseMatcher(testdata.CertChain, testdata.VCEK),
				nil,
			),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			instanceInfo, err := newStubAzureInstanceInfo(tc.report, tc.runtimeData)
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
				hclValidator: &instanceInfo,
				config:       defaultCfg,
				log:          logger.NewTest(t),
				getter:       tc.getter,
			}

			key, err := validator.getTrustedKey(context.Background(), attDoc, nil)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				assert.NotNil(key)
			}
		})
	}
}

type stubAzureInstanceInfo struct {
	AttestationReport []byte
	RuntimeData       []byte
}

func newStubAzureInstanceInfo(report string, runtimeData string) (stubAzureInstanceInfo, error) {
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

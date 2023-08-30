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
	"github.com/google/go-sev-guest/kds"
	spb "github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/verify/trust"
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
		wantErr bool
	}{
		"report too short": {
			report:  defaultReport[:len(defaultReport)-10],
			wantErr: true,
		},
		"corrupted report": {
			report:  defaultReport[10 : len(defaultReport)-10],
			wantErr: true,
		},
		"success": {
			report:  defaultReport,
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

			report, err := instanceInfo.attestationWithCerts(trust.DefaultHTTPSGetter())
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

	defaultReport := hex.EncodeToString(testdata.AttestationReport)
	defaultRuntimeData := "7b226b657973223a5b7b226b6964223a2248434c416b507562222c226b65795f6f7073223a5b22656e6372797074225d2c226b7479223a22525341222c2265223a2241514142222c226e223a22747946717641414166324746656c6b5737566352684a6e4132597659364c6a427a65554e3276614d5a6e5a74685f74466e574d6b4b35415874757379434e656c337569703356475a7a54617a3558327447566a4772732d4d56486361703951647771555856573367394f515f74456269786378372d78626c554a516b474551666e626253646e5049326c764c7a4f73315a5f30766a65444178765351726d616773366e592d634a4157482d706744564a79487470735553735f5142576b6c617a44736f3557486d6e4d743973394d75696c57586f7830525379586e55656151796859316a753752545363526e5658754e7936377a5f454a6e774d393264727746623841556430534a5f396f687645596c34615a52444543476f3056726a635348552d4a474a6575574335566844425235454f6f4356424267716539653833765f6c4a784933574c65326f7653495a49497a416d625351227d5d2c22766d2d636f6e66696775726174696f6e223a7b22636f6e736f6c652d656e61626c6564223a747275652c2263757272656e742d74696d65223a313636313435353339312c227365637572652d626f6f74223a66616c73652c2274706d2d656e61626c6564223a747275652c22766d556e697175654964223a2242364339384333422d344543372d344441362d424432462d374439384432304437423735227d7d"
	defaultVcek := testdata.VCEK
	defaultCertChain := testdata.CertChain
	defaultIDKeyDigestOld, err := hex.DecodeString("57e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc1")
	require.NoError(err)
	defaultIDKeyDigest := idkeydigest.NewList([][]byte{defaultIDKeyDigestOld})
	defaultUrlResponseMatcher := func(url string) []byte {
		switch {
		case url == "https://kdsintf.amd.com/vcek/v1/Milan/cert_chain":
			return []byte(defaultCertChain)
		case regexp.MustCompile(`https:\/\/kdsintf.amd.com\/vcek\/v1\/Milan\/.*`).MatchString(url):
			return []byte(defaultVcek)
		default:
			t.Fatalf("unexpected url: %s", url)
		}
		return nil
	}
	httpsGetter := func(urlResponseMatcher func(string) []byte, err error) *stubHTTPSGetter {
		return &stubHTTPSGetter{
			urlResponseMatcher: urlResponseMatcher,
			err:                err,
		}
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
			getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		},
		// "invalid report signature": {
		// 	report:               "02000000020000001f0003000000000001000000000000000000000000000000020000000000000000000000000000000000000001000000020000000000065d010000000000000000000000000000000ccc0895ef2f2c3b8c8568f5a2bb65ff5bf9387a09359742ad41e686cacfd38b00000000000000000000000000000000000000000000000000000000000000005677f1de87289e7ad2c7e99c805d0468b1a9ccd83f0d245afa5242d405da4d5725852f8c6550564870e5f3206dfb1841000000000000000000000000000000000000000000000000000000000000000057e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005f7240b24a1babe2ece844c4f792bcd9844bf6907d14aeea00156310b9538daffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000065d0000000000000000000000000000000000000000000000009e44aaef02cfca6fddbaca669c6cfd29e1ab8d97ebc939857128acbb13b8740df31436d34e86e5f8ae0cdfeb3a0e185db46decac176cc77d761c22a1b9dcf25b020000000000065d0133010001330100020000000000065d000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ccb7dc15abff884802e774b39adba8e6ff7efcf05e115c91588e657065151056a320f70c788d0e3619391052922e422b000000000000000000000000000000000000000000000000e8dbf581140443bbc681c50eca8639a76ef6cab34e0780cbca977e2e2a03f8b864fd4e9774b0f8055511567e031e59bf00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005c02000001000000020000000100000048020000",
		// 	runtimeData:          defaultRuntimeData,
		// 	acceptedIDKeyDigests: defaultIDKeyDigest,
		// 	enforcementPolicy:    idkeydigest.Equal,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// 	wantErr:              true,
		// },
		// "invalid vcek": {
		// 	report:               defaultReport,
		// 	runtimeData:          defaultRuntimeData,
		// 	acceptedIDKeyDigests: defaultIDKeyDigest,
		// 	enforcementPolicy:    idkeydigest.Equal,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// 	wantErr:              true,
		// },
		// "malformed ask": {
		// 	report:               defaultReport,
		// 	runtimeData:          defaultRuntimeData,
		// 	acceptedIDKeyDigests: defaultIDKeyDigest,
		// 	enforcementPolicy:    idkeydigest.Equal,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// 	wantErr:              true,
		// },
		// "invalid runtime data digest": {
		// 	report:               defaultReport,
		// 	runtimeData:          "7b226b657973223a5b7b226b6964223a2248434c426b507562222c226b65795f6f7073223a5b22656e6372797074225d2c226b7479223a22525341222c2265223a2241514142222c226e223a22747946717641414166324746656c6b5737566352684a6e4132597659364c6a427a65554e3276614d5a6e5a74685f74466e574d6b4b35415874757379434e656c337569703356475a7a54617a3558327447566a4772732d4d56486361703951647771555856573367394f515f74456269786378372d78626c554a516b474551666e626253646e5049326c764c7a4f73315a5f30766a65444178765351726d616773366e592d634a4157482d706744564a79487470735553735f5142576b6c617a44736f3557486d6e4d743973394d75696c57586f7830525379586e55656151796859316a753752545363526e5658754e7936377a5f454a6e774d393264727746623841556430534a5f396f687645596c34615a52444543476f3056726a635348552d4a474a6575574335566844425235454f6f4356424267716539653833765f6c4a784933574c65326f7653495a49497a416d625351227d5d2c22766d2d636f6e66696775726174696f6e223a7b22636f6e736f6c652d656e61626c6564223a747275652c2263757272656e742d74696d65223a313636313435353339312c227365637572652d626f6f74223a66616c73652c2274706d2d656e61626c6564223a747275652c22766d556e697175654964223a2242364339384333422d344543372d344441362d424432462d374439384432304437423735227d7d",
		// 	acceptedIDKeyDigests: defaultIDKeyDigest,
		// 	enforcementPolicy:    idkeydigest.Equal,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// 	wantErr:              true,
		// },
		// "invalid idkeydigest": {
		// 	report:               defaultReport,
		// 	runtimeData:          defaultRuntimeData,
		// 	acceptedIDKeyDigests: idkeydigest.List{[]byte{0x00}},
		// 	enforcementPolicy:    idkeydigest.Equal,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// 	wantErr:              true,
		// },
		// "don't enforce idkeydigest": {
		// 	report:               defaultReport,
		// 	runtimeData:          defaultRuntimeData,
		// 	acceptedIDKeyDigests: idkeydigest.List{[]byte{0x00}},
		// 	enforcementPolicy:    idkeydigest.WarnOnly,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// },
		// "unsupported microcode version": {
		// 	report:               "02000000020000001f0003000000000001000000000000000000000000000000020000000000000000000000000000000000000001000000020000000000065d010000000000000000000000000000000ccc0895ef2f2c3b8c8568f5a2bb65ff5bf9387a09359742ad41e686cacfd38b00000000000000000000000000000000000000000000000000000000000000005677f1de87289e7ad2c7e99c805d0468b1a9ccd83f0d245afa5242d405da4d5725852f8c6550564870e5f3206dfb1841000000000000000000000000000000000000000000000000000000000000000057e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005f7240b24a1babe2ece844c4f792bcd9844bf6907d14aeea00156310b9538daffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000065d0000000000000000000000000000000000000000000000009e44aaef02cfca6fddbaca669c6cfd29e1ab8d97ebc939857128acbb13b8740df31436d34e86e5f8ae0cdfeb3a0e185db46decac176cc77d761c22a1b9dcf25b020000000000065c0133010001330100020000000000065d000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bcb7dc15abff884802e774b39adba8e6ff7efcf05e115c91588e657065151056a320f70c788d0e3619391052922e422b000000000000000000000000000000000000000000000000e8dbf581140443bbc681c50eca8639a76ef6cab34e0780cbca977e2e2a03f8b864fd4e9774b0f8055511567e031e59bf00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005c02000001000000020000000100000048020000",
		// 	runtimeData:          defaultRuntimeData,
		// 	acceptedIDKeyDigests: defaultIDKeyDigest,
		// 	enforcementPolicy:    idkeydigest.Equal,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// 	wantErr:              true,
		// },
		// "current tcb older than committed tcb": {
		// 	report:               "02000000020000001f0003000000000001000000000000000000000000000000020000000000000000000000000000000000000001000000020000000000065c010000000000000000000000000000000ccc0895ef2f2c3b8c8568f5a2bb65ff5bf9387a09359742ad41e686cacfd38b00000000000000000000000000000000000000000000000000000000000000005677f1de87289e7ad2c7e99c805d0468b1a9ccd83f0d245afa5242d405da4d5725852f8c6550564870e5f3206dfb1841000000000000000000000000000000000000000000000000000000000000000057e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005f7240b24a1babe2ece844c4f792bcd9844bf6907d14aeea00156310b9538daffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000065d0000000000000000000000000000000000000000000000009e44aaef02cfca6fddbaca669c6cfd29e1ab8d97ebc939857128acbb13b8740df31436d34e86e5f8ae0cdfeb3a0e185db46decac176cc77d761c22a1b9dcf25b020000000000065d0133010001330100020000000000065d000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bcb7dc15abff884802e774b39adba8e6ff7efcf05e115c91588e657065151056a320f70c788d0e3619391052922e422b000000000000000000000000000000000000000000000000e8dbf581140443bbc681c50eca8639a76ef6cab34e0780cbca977e2e2a03f8b864fd4e9774b0f8055511567e031e59bf00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005c02000001000000020000000100000048020000",
		// 	runtimeData:          defaultRuntimeData,
		// 	acceptedIDKeyDigests: defaultIDKeyDigest,
		// 	enforcementPolicy:    idkeydigest.Equal,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// 	wantErr:              true,
		// },
		// "launch tcb != committed tcb": {
		// 	report:               "02000000020000001f0003000000000001000000000000000000000000000000020000000000000000000000000000000000000001000000020000000000065d010000000000000000000000000000000ccc0895ef2f2c3b8c8568f5a2bb65ff5bf9387a09359742ad41e686cacfd38b00000000000000000000000000000000000000000000000000000000000000005677f1de87289e7ad2c7e99c805d0468b1a9ccd83f0d245afa5242d405da4d5725852f8c6550564870e5f3206dfb1841000000000000000000000000000000000000000000000000000000000000000057e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005f7240b24a1babe2ece844c4f792bcd9844bf6907d14aeea00156310b9538daffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000065d0000000000000000000000000000000000000000000000009e44aaef02cfca6fddbaca669c6cfd29e1ab8d97ebc939857128acbb13b8740df31436d34e86e5f8ae0cdfeb3a0e185db46decac176cc77d761c22a1b9dcf25b020000000000065d0133010001330100020000000000065c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bcb7dc15abff884802e774b39adba8e6ff7efcf05e115c91588e657065151056a320f70c788d0e3619391052922e422b000000000000000000000000000000000000000000000000e8dbf581140443bbc681c50eca8639a76ef6cab34e0780cbca977e2e2a03f8b864fd4e9774b0f8055511567e031e59bf00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005c02000001000000020000000100000048020000",
		// 	runtimeData:          defaultRuntimeData,
		// 	acceptedIDKeyDigests: defaultIDKeyDigest,
		// 	enforcementPolicy:    idkeydigest.Equal,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// 	wantErr:              true,
		// },
		// "debugging allowed": {
		// 	report:               "02000000020000001f000b000000000001000000000000000000000000000000020000000000000000000000000000000000000001000000020000000000065d010000000000000000000000000000000ccc0895ef2f2c3b8c8568f5a2bb65ff5bf9387a09359742ad41e686cacfd38b00000000000000000000000000000000000000000000000000000000000000005677f1de87289e7ad2c7e99c805d0468b1a9ccd83f0d245afa5242d405da4d5725852f8c6550564870e5f3206dfb1841000000000000000000000000000000000000000000000000000000000000000057e229e0ffe5fa92d0faddff6cae0e61c926fc9ef9afd20a8b8cfcf7129db9338cbe5bf3f6987733a2bf65d06dc38fc100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005f7240b24a1babe2ece844c4f792bcd9844bf6907d14aeea00156310b9538daffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000065d0000000000000000000000000000000000000000000000009e44aaef02cfca6fddbaca669c6cfd29e1ab8d97ebc939857128acbb13b8740df31436d34e86e5f8ae0cdfeb3a0e185db46decac176cc77d761c22a1b9dcf25b020000000000065d0133010001330100020000000000065d000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bcb7dc15abff884802e774b39adba8e6ff7efcf05e115c91588e657065151056a320f70c788d0e3619391052922e422b000000000000000000000000000000000000000000000000e8dbf581140443bbc681c50eca8639a76ef6cab34e0780cbca977e2e2a03f8b864fd4e9774b0f8055511567e031e59bf00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005c02000001000000020000000100000048020000",
		// 	runtimeData:          defaultRuntimeData,
		// 	acceptedIDKeyDigests: defaultIDKeyDigest,
		// 	enforcementPolicy:    idkeydigest.Equal,
		// 	getter:               httpsGetter(defaultUrlResponseMatcher, nil),
		// 	wantErr:              true,
		// },
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

			cfg := config.DefaultForAzureSEVSNP()
			cfg.FirmwareSignerConfig = config.SNPFirmwareSignerConfig{
				AcceptedKeyDigests: tc.acceptedIDKeyDigests,
				EnforcementPolicy:  tc.enforcementPolicy,
			}

			validator := &Validator{
				hclValidator: &instanceInfo,
				config:       cfg,
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

type stubHTTPSGetter struct {
	urlResponseMatcher func(string) []byte // maps responses to requested URLs
	err                error
}

func (s *stubHTTPSGetter) Get(url string) ([]byte, error) {
	fmt.Println(url)
	if s.err != nil {
		return nil, s.err
	}
	return s.urlResponseMatcher(url), nil
}

func TestValidateAzureCVM(t *testing.T) {
	assert.NoError(t, validateCVM(vtpm.AttestationDocument{}, nil))
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

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package vtpm

import (
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	tpmsim "github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm-tools/proto/tpm"
	"github.com/google/go-tpm-tools/simulator"
	"github.com/google/go-tpm/tpm2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type simTPMWithEventLog struct {
	*simulator.Simulator
}

func newSimTPMWithEventLog() (io.ReadWriteCloser, error) {
	tpmSim, err := simulator.Get()
	if err != nil {
		return nil, err
	}
	return &simTPMWithEventLog{tpmSim}, nil
}

// EventLog overrides the default event log getter.
func (s simTPMWithEventLog) EventLog() ([]byte, error) {
	// event log header for successful parsing of event log
	header := []byte{
		0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x25, 0x00, 0x00, 0x00, 0x53, 0x70, 0x65,
		0x63, 0x20, 0x49, 0x44, 0x20, 0x45, 0x76, 0x65, 0x6E, 0x74, 0x30, 0x33, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x02, 0x00, 0x02, 0x02, 0x00, 0x00, 0x00, 0x04, 0x00, 0x14, 0x00, 0x0B, 0x00, 0x20, 0x00, 0x00,
	}
	return header, nil
}

func fakeGetInstanceInfo(tpm io.ReadWriteCloser) ([]byte, error) {
	return []byte("unit-test"), nil
}

func TestValidate(t *testing.T) {
	require := require.New(t)

	fakeValidateCVM := func(AttestationDocument) error { return nil }
	fakeGetTrustedKey := func(aKPub, instanceInfo []byte) (crypto.PublicKey, error) {
		pubArea, err := tpm2.DecodePublic(aKPub)
		if err != nil {
			return nil, err
		}
		return pubArea.Key()
	}

	testExpectedPCRs := map[uint32][]byte{
		0: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		1: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	warnLog := &testAttestationLogger{}

	issuer := NewIssuer(newSimTPMWithEventLog, tpmclient.AttestationKeyRSA, fakeGetInstanceInfo)
	validator := NewValidator(testExpectedPCRs, []uint32{0, 1}, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15, warnLog)

	nonce := []byte{1, 2, 3, 4}
	challenge := []byte("Constellation")

	attDocRaw, err := issuer.Issue(challenge, nonce)
	require.NoError(err)

	var attDoc AttestationDocument
	err = json.Unmarshal(attDocRaw, &attDoc)
	require.NoError(err)
	require.Equal(challenge, attDoc.UserData)

	// valid test
	out, err := validator.Validate(attDocRaw, nonce)
	require.NoError(err)
	require.Equal(challenge, out)

	enforcedPCRs := []uint32{0, 1}
	expectedPCRs := map[uint32][]byte{
		0: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		1: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		2: {0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
		3: {0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f, 0x40},
		4: {0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f, 0x60},
		5: {0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f, 0x80},
	}
	warningValidator := NewValidator(
		expectedPCRs,
		enforcedPCRs,
		fakeGetTrustedKey,
		fakeValidateCVM,
		VerifyPKCS1v15,
		warnLog,
	)
	out, err = warningValidator.Validate(attDocRaw, nonce)
	require.NoError(err)
	assert.Equal(t, challenge, out)
	assert.Len(t, warnLog.warnings, len(expectedPCRs)-len(enforcedPCRs))

	testCases := map[string]struct {
		validator *Validator
		attDoc    []byte
		nonce     []byte
		wantErr   bool
	}{
		"invalid nonce": {
			validator: NewValidator(testExpectedPCRs, []uint32{0, 1}, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15, warnLog),
			attDoc:    mustMarshalAttestation(attDoc, require),
			nonce:     []byte{4, 3, 2, 1},
			wantErr:   true,
		},
		"invalid signature": {
			validator: NewValidator(testExpectedPCRs, []uint32{0, 1}, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15, warnLog),
			attDoc: mustMarshalAttestation(AttestationDocument{
				Attestation:       attDoc.Attestation,
				InstanceInfo:      attDoc.InstanceInfo,
				UserData:          []byte("wrong data"),
				UserDataSignature: attDoc.UserDataSignature,
			}, require),
			nonce:   nonce,
			wantErr: true,
		},
		"untrusted attestation public key": {
			validator: NewValidator(
				testExpectedPCRs,
				[]uint32{0, 1},
				func(akPub, instanceInfo []byte) (crypto.PublicKey, error) {
					return nil, errors.New("untrusted")
				},
				fakeValidateCVM, VerifyPKCS1v15, warnLog),
			attDoc:  mustMarshalAttestation(attDoc, require),
			nonce:   nonce,
			wantErr: true,
		},
		"not a CVM": {
			validator: NewValidator(
				testExpectedPCRs,
				[]uint32{0, 1},
				fakeGetTrustedKey,
				func(attestation AttestationDocument) error {
					return errors.New("untrusted")
				},
				VerifyPKCS1v15, warnLog),
			attDoc:  mustMarshalAttestation(attDoc, require),
			nonce:   nonce,
			wantErr: true,
		},
		"untrusted PCRs": {
			validator: NewValidator(
				map[uint32][]byte{
					0: {0xFF},
				},
				[]uint32{0},
				fakeGetTrustedKey,
				fakeValidateCVM,
				VerifyPKCS1v15, warnLog),
			attDoc:  mustMarshalAttestation(attDoc, require),
			nonce:   nonce,
			wantErr: true,
		},
		"no sha256 quote": {
			validator: NewValidator(testExpectedPCRs, []uint32{0, 1}, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15, warnLog),
			attDoc: mustMarshalAttestation(AttestationDocument{
				Attestation: &attest.Attestation{
					AkPub: attDoc.Attestation.AkPub,
					Quotes: []*tpm.Quote{
						attDoc.Attestation.Quotes[2],
					},
					EventLog:     attDoc.Attestation.EventLog,
					InstanceInfo: attDoc.Attestation.InstanceInfo,
				},
				InstanceInfo:      attDoc.InstanceInfo,
				UserData:          attDoc.UserData,
				UserDataSignature: attDoc.UserDataSignature,
			}, require),
			nonce:   nonce,
			wantErr: true,
		},
		"invalid attestation document": {
			validator: NewValidator(testExpectedPCRs, []uint32{0, 1}, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15, warnLog),
			attDoc:    []byte("invalid attestation"),
			nonce:     nonce,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err = tc.validator.Validate(tc.attDoc, tc.nonce)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func mustMarshalAttestation(attDoc AttestationDocument, require *require.Assertions) []byte {
	out, err := json.Marshal(attDoc)
	require.NoError(err)
	return out
}

func TestFailIssuer(t *testing.T) {
	testCases := map[string]struct {
		issuer   *Issuer
		userData []byte
		nonce    []byte
	}{
		"fail openTPM": {
			issuer: NewIssuer(
				func() (io.ReadWriteCloser, error) {
					return nil, errors.New("failure")
				},
				tpmclient.AttestationKeyRSA,
				fakeGetInstanceInfo,
			),
			userData: []byte("Constellation"),
			nonce:    []byte{1, 2, 3, 4},
		},
		"fail getAttestationKey": {
			issuer: NewIssuer(
				newSimTPMWithEventLog,
				func(tpm io.ReadWriter) (*tpmclient.Key, error) {
					return nil, errors.New("failure")
				},
				fakeGetInstanceInfo,
			),
			userData: []byte("Constellation"),
			nonce:    []byte{1, 2, 3, 4},
		},
		"fail Attest": {
			issuer: NewIssuer(
				newSimTPMWithEventLog,
				func(tpm io.ReadWriter) (*tpmclient.Key, error) {
					return &tpmclient.Key{}, nil
				},
				fakeGetInstanceInfo,
			),
			userData: []byte("Constellation"),
			nonce:    []byte{1, 2, 3, 4},
		},
		"fail getInstanceInfo": {
			issuer: NewIssuer(
				newSimTPMWithEventLog,
				tpmclient.AttestationKeyRSA,
				func(io.ReadWriteCloser) ([]byte, error) { return nil, errors.New("failure") },
			),
			userData: []byte("Constellation"),
			nonce:    []byte{1, 2, 3, 4},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err := tc.issuer.Issue(tc.userData, tc.nonce)
			assert.Error(err)
		})
	}
}

func TestGetSHA256QuoteIndex(t *testing.T) {
	testCases := map[string]struct {
		quotes  []*tpm.Quote
		wantIdx int
		wantErr bool
	}{
		"idx 0 is valid": {
			quotes: []*tpm.Quote{
				{
					Pcrs: &tpm.PCRs{
						Hash: tpm.HashAlgo_SHA256,
					},
				},
			},
			wantIdx: 0,
			wantErr: false,
		},
		"idx 1 is valid": {
			quotes: []*tpm.Quote{
				{
					Pcrs: &tpm.PCRs{
						Hash: tpm.HashAlgo_SHA1,
					},
				},
				{
					Pcrs: &tpm.PCRs{
						Hash: tpm.HashAlgo_SHA256,
					},
				},
			},
			wantIdx: 1,
			wantErr: false,
		},
		"no quotes": {
			quotes:  nil,
			wantErr: true,
		},
		"quotes is nil": {
			quotes:  make([]*tpm.Quote, 2),
			wantErr: true,
		},
		"pcrs is nil": {
			quotes: []*tpm.Quote{
				{
					Pcrs: &tpm.PCRs{
						Hash: tpm.HashAlgo_SHA1,
					},
				},
				{
					Pcrs: nil,
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			idx, err := GetSHA256QuoteIndex(tc.quotes)
			if tc.wantErr {
				assert.Error(err)
				assert.Equal(0, idx)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantIdx, idx)
			}
		})
	}
}

func TestGetSelectedPCRs(t *testing.T) {
	testCases := map[string]struct {
		openFunc     TPMOpenFunc
		pcrSelection tpm2.PCRSelection
		wantErr      bool
	}{
		"error": {
			openFunc: func() (io.ReadWriteCloser, error) { return nil, errors.New("error") },
			pcrSelection: tpm2.PCRSelection{
				Hash: tpm2.AlgSHA256,
				PCRs: []int{0, 1, 2},
			},
			wantErr: true,
		},
		"3 PCRs": {
			openFunc: tpmsim.OpenSimulatedTPM,
			pcrSelection: tpm2.PCRSelection{
				Hash: tpm2.AlgSHA256,
				PCRs: []int{0, 1, 2},
			},
		},
		"Azure PCRS": {
			openFunc:     tpmsim.OpenSimulatedTPM,
			pcrSelection: AzurePCRSelection,
		},
		"GCP PCRs": {
			openFunc:     tpmsim.OpenSimulatedTPM,
			pcrSelection: GCPPCRSelection,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			pcrs, err := GetSelectedPCRs(tc.openFunc, tc.pcrSelection)
			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)

				assert.Equal(len(pcrs), len(tc.pcrSelection.PCRs))
				for _, pcr := range pcrs {
					assert.Len(pcr, 32)
				}
			}
		})
	}
}

type testAttestationLogger struct {
	infos    []string
	warnings []string
}

func (w *testAttestationLogger) Infof(format string, args ...interface{}) {
	w.infos = append(w.infos, fmt.Sprintf(format, args...))
}

func (w *testAttestationLogger) Warnf(format string, args ...interface{}) {
	w.warnings = append(w.warnings, fmt.Sprintf(format, args...))
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package vtpm

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm-tools/proto/tpm"
	"github.com/google/go-tpm/tpm2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgelesssys/constellation/v2/internal/attestation/initialize"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	tpmsim "github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

type simTPMWithEventLog struct {
	io.ReadWriteCloser
}

func newSimTPMWithEventLog() (io.ReadWriteCloser, error) {
	tpmSim, err := simulator.OpenSimulatedTPM()
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

func fakeGetInstanceInfo(_ context.Context, _ io.ReadWriteCloser, _ []byte) ([]byte, error) {
	return []byte("unit-test"), nil
}

func TestValidate(t *testing.T) {
	require := require.New(t)

	fakeValidateCVM := func(AttestationDocument, *attest.MachineState) error { return nil }
	fakeGetTrustedKey := func(_ context.Context, attDoc AttestationDocument, _ []byte) (crypto.PublicKey, error) {
		pubArea, err := tpm2.DecodePublic(attDoc.Attestation.AkPub)
		if err != nil {
			return nil, err
		}
		return pubArea.Key()
	}

	testExpectedPCRs := measurements.M{
		0:                                      measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
		1:                                      measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
		uint32(measurements.PCRIndexClusterID): measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
	}
	warnLog := &testAttestationLogger{}

	tpmOpen, tpmCloser := tpmsim.NewSimulatedTPMOpenFunc()
	defer tpmCloser.Close()

	issuer := NewIssuer(tpmOpen, tpmclient.AttestationKeyRSA, fakeGetInstanceInfo, logger.NewTest(t))
	validator := NewValidator(testExpectedPCRs, fakeGetTrustedKey, fakeValidateCVM, nil)

	nonce := []byte{1, 2, 3, 4}
	challenge := []byte("Constellation")

	ctx := context.Background()

	attDocRaw, err := issuer.Issue(ctx, challenge, nonce)
	require.NoError(err)

	var attDoc AttestationDocument
	err = json.Unmarshal(attDocRaw, &attDoc)
	require.NoError(err)
	require.Equal(challenge, attDoc.UserData)

	// valid test
	out, err := validator.Validate(ctx, attDocRaw, nonce)
	require.NoError(err)
	require.Equal(challenge, out)

	// validation must fail after bootstrapping (change of enforced PCR)
	require.NoError(initialize.MarkNodeAsBootstrapped(tpmOpen, []byte{2}))
	attDocBootstrappedRaw, err := issuer.Issue(ctx, challenge, nonce)
	require.NoError(err)
	_, err = validator.Validate(ctx, attDocBootstrappedRaw, nonce)
	require.Error(err)

	// userData must be bound to PCR state
	attDocBootstrappedRaw, err = issuer.Issue(ctx, []byte{2, 3}, nonce)
	require.NoError(err)
	var attDocBootstrapped AttestationDocument
	require.NoError(json.Unmarshal(attDocBootstrappedRaw, &attDocBootstrapped))
	attDocBootstrapped.Attestation = attDoc.Attestation
	attDocBootstrappedRaw, err = json.Marshal(attDocBootstrapped)
	require.NoError(err)
	_, err = validator.Validate(ctx, attDocBootstrappedRaw, nonce)
	require.Error(err)

	expectedPCRs := measurements.M{
		0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
		1: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
		2: measurements.Measurement{
			Expected:      []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
			ValidationOpt: measurements.WarnOnly,
		},
		3: measurements.Measurement{
			Expected:      []byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f, 0x40},
			ValidationOpt: measurements.WarnOnly,
		},
		4: measurements.Measurement{
			Expected:      []byte{0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f, 0x60},
			ValidationOpt: measurements.WarnOnly,
		},
		5: measurements.Measurement{
			Expected:      []byte{0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f, 0x80},
			ValidationOpt: measurements.WarnOnly,
		},
	}
	warningValidator := NewValidator(
		expectedPCRs,
		fakeGetTrustedKey,
		fakeValidateCVM,
		warnLog,
	)
	out, err = warningValidator.Validate(ctx, attDocRaw, nonce)
	require.NoError(err)
	assert.Equal(t, challenge, out)
	assert.Len(t, warnLog.warnings, 4)

	testCases := map[string]struct {
		validator *Validator
		attDoc    []byte
		nonce     []byte
		wantErr   bool
	}{
		"valid": {
			validator: NewValidator(testExpectedPCRs, fakeGetTrustedKey, fakeValidateCVM, warnLog),
			attDoc:    mustMarshalAttestation(attDoc, require),
			nonce:     nonce,
		},
		"invalid nonce": {
			validator: NewValidator(testExpectedPCRs, fakeGetTrustedKey, fakeValidateCVM, warnLog),
			attDoc:    mustMarshalAttestation(attDoc, require),
			nonce:     []byte{4, 3, 2, 1},
			wantErr:   true,
		},
		"invalid signature": {
			validator: NewValidator(testExpectedPCRs, fakeGetTrustedKey, fakeValidateCVM, warnLog),
			attDoc: mustMarshalAttestation(AttestationDocument{
				Attestation:  attDoc.Attestation,
				InstanceInfo: attDoc.InstanceInfo,
				UserData:     []byte("wrong data"),
			}, require),
			nonce:   nonce,
			wantErr: true,
		},
		"untrusted attestation public key": {
			validator: NewValidator(
				testExpectedPCRs,
				func(context.Context, AttestationDocument, []byte) (crypto.PublicKey, error) {
					return nil, errors.New("untrusted")
				},
				fakeValidateCVM, warnLog),
			attDoc:  mustMarshalAttestation(attDoc, require),
			nonce:   nonce,
			wantErr: true,
		},
		"not a CVM": {
			validator: NewValidator(
				testExpectedPCRs,
				fakeGetTrustedKey,
				func(AttestationDocument, *attest.MachineState) error {
					return errors.New("untrusted")
				},
				warnLog),
			attDoc:  mustMarshalAttestation(attDoc, require),
			nonce:   nonce,
			wantErr: true,
		},
		"untrusted PCRs": {
			validator: NewValidator(
				measurements.M{
					0: measurements.Measurement{
						Expected:      []byte{0xFF},
						ValidationOpt: measurements.Enforce,
					},
				},
				fakeGetTrustedKey,
				fakeValidateCVM,
				warnLog),
			attDoc:  mustMarshalAttestation(attDoc, require),
			nonce:   nonce,
			wantErr: true,
		},
		"no sha256 quote": {
			validator: NewValidator(testExpectedPCRs, fakeGetTrustedKey, fakeValidateCVM, warnLog),
			attDoc: mustMarshalAttestation(AttestationDocument{
				Attestation: &attest.Attestation{
					AkPub: attDoc.Attestation.AkPub,
					Quotes: []*tpm.Quote{
						attDoc.Attestation.Quotes[2],
					},
					EventLog:     attDoc.Attestation.EventLog,
					InstanceInfo: attDoc.Attestation.InstanceInfo,
				},
				InstanceInfo: attDoc.InstanceInfo,
				UserData:     attDoc.UserData,
			}, require),
			nonce:   nonce,
			wantErr: true,
		},
		"invalid attestation document": {
			validator: NewValidator(testExpectedPCRs, fakeGetTrustedKey, fakeValidateCVM, warnLog),
			attDoc:    []byte("invalid attestation"),
			nonce:     nonce,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err = tc.validator.Validate(ctx, tc.attDoc, tc.nonce)
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
				nil,
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
				nil,
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
				nil,
			),
			userData: []byte("Constellation"),
			nonce:    []byte{1, 2, 3, 4},
		},
		"fail getInstanceInfo": {
			issuer: NewIssuer(
				newSimTPMWithEventLog,
				tpmclient.AttestationKeyRSA,
				func(context.Context, io.ReadWriteCloser, []byte) ([]byte, error) { return nil, errors.New("failure") },
				nil,
			),
			userData: []byte("Constellation"),
			nonce:    []byte{1, 2, 3, 4},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			tc.issuer.log = logger.NewTest(t)

			_, err := tc.issuer.Issue(context.Background(), tc.userData, tc.nonce)
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

func TestGetSelectedMeasurements(t *testing.T) {
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
			openFunc: simulator.OpenSimulatedTPM,
			pcrSelection: tpm2.PCRSelection{
				Hash: tpm2.AlgSHA256,
				PCRs: []int{0, 1, 2},
			},
		},
		"Azure PCRS": {
			openFunc:     simulator.OpenSimulatedTPM,
			pcrSelection: AzurePCRSelection,
		},
		"GCP PCRs": {
			openFunc:     simulator.OpenSimulatedTPM,
			pcrSelection: GCPPCRSelection,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			pcrs, err := GetSelectedMeasurements(tc.openFunc, tc.pcrSelection)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Len(pcrs, len(tc.pcrSelection.PCRs))
		})
	}
}

type testAttestationLogger struct {
	infos    []string
	warnings []string
}

func (w *testAttestationLogger) Infof(format string, args ...any) {
	w.infos = append(w.infos, fmt.Sprintf(format, args...))
}

func (w *testAttestationLogger) Warnf(format string, args ...any) {
	w.warnings = append(w.warnings, fmt.Sprintf(format, args...))
}

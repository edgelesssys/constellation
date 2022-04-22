package vtpm

import (
	"crypto"
	"encoding/json"
	"errors"
	"io"
	"testing"

	tpmsim "github.com/edgelesssys/constellation/coordinator/attestation/simulator"
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

func fakeGetTrustedKey(aKPub, instanceInfo []byte) (crypto.PublicKey, error) {
	pubArea, err := tpm2.DecodePublic(aKPub)
	if err != nil {
		return nil, err
	}
	return pubArea.Key()
}

func fakeValidateCVM(AttestationDocument) error { return nil }

var fakeTrustedPcrs = map[uint32][]byte{
	0: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	1: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
}

func TestValidate(t *testing.T) {
	require := require.New(t)

	issuer := NewIssuer(newSimTPMWithEventLog, tpmclient.AttestationKeyRSA, fakeGetInstanceInfo)
	validator := NewValidator(fakeTrustedPcrs, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15)

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

	testCases := map[string]struct {
		validator   *Validator
		attDoc      []byte
		nonce       []byte
		errExpected bool
	}{
		"invalid nonce": {
			validator:   NewValidator(fakeTrustedPcrs, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15),
			attDoc:      mustMarshalAttestation(attDoc, require),
			nonce:       []byte{4, 3, 2, 1},
			errExpected: true,
		},
		"invalid signature": {
			validator: NewValidator(fakeTrustedPcrs, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15),
			attDoc: mustMarshalAttestation(AttestationDocument{
				Attestation:       attDoc.Attestation,
				InstanceInfo:      attDoc.InstanceInfo,
				UserData:          []byte("wrong data"),
				UserDataSignature: attDoc.UserDataSignature,
			}, require),
			nonce:       nonce,
			errExpected: true,
		},
		"untrusted attestation public key": {
			validator: NewValidator(
				fakeTrustedPcrs,
				func(akPub, instanceInfo []byte) (crypto.PublicKey, error) {
					return nil, errors.New("untrusted")
				},
				fakeValidateCVM, VerifyPKCS1v15),
			attDoc:      mustMarshalAttestation(attDoc, require),
			nonce:       nonce,
			errExpected: true,
		},
		"not a CVM": {
			validator: NewValidator(
				fakeTrustedPcrs,
				fakeGetTrustedKey,
				func(attestation AttestationDocument) error {
					return errors.New("untrusted")
				},
				VerifyPKCS1v15),
			attDoc:      mustMarshalAttestation(attDoc, require),
			nonce:       nonce,
			errExpected: true,
		},
		"untrusted PCRs": {
			validator: NewValidator(
				map[uint32][]byte{
					0: {0xFF},
				},
				fakeGetTrustedKey,
				fakeValidateCVM,
				VerifyPKCS1v15),
			attDoc:      mustMarshalAttestation(attDoc, require),
			nonce:       nonce,
			errExpected: true,
		},
		"no sha256 quote": {
			validator: NewValidator(fakeTrustedPcrs, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15),
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
			nonce:       nonce,
			errExpected: true,
		},
		"invalid attestation document": {
			validator:   NewValidator(fakeTrustedPcrs, fakeGetTrustedKey, fakeValidateCVM, VerifyPKCS1v15),
			attDoc:      []byte("invalid attestation"),
			nonce:       nonce,
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err = tc.validator.Validate(tc.attDoc, tc.nonce)
			if tc.errExpected {
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
		quotes      []*tpm.Quote
		idxExpected int
		errExpected bool
	}{
		"idx 0 is valid": {
			quotes: []*tpm.Quote{
				{
					Pcrs: &tpm.PCRs{
						Hash: tpm.HashAlgo_SHA256,
					},
				},
			},
			idxExpected: 0,
			errExpected: false,
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
			idxExpected: 1,
			errExpected: false,
		},
		"no quotes": {
			quotes:      nil,
			errExpected: true,
		},
		"quotes is nil": {
			quotes:      make([]*tpm.Quote, 2),
			errExpected: true,
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
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			idx, err := GetSHA256QuoteIndex(tc.quotes)
			if tc.errExpected {
				assert.Error(err)
				assert.Equal(0, idx)
			} else {
				assert.NoError(err)
				assert.Equal(tc.idxExpected, idx)
			}
		})
	}
}

func TestGetSelectedPCRs(t *testing.T) {
	testCases := map[string]struct {
		openFunc     TPMOpenFunc
		pcrSelection tpm2.PCRSelection
		errExpected  bool
	}{
		"error": {
			openFunc: func() (io.ReadWriteCloser, error) { return nil, errors.New("error") },
			pcrSelection: tpm2.PCRSelection{
				Hash: tpm2.AlgSHA256,
				PCRs: []int{0, 1, 2},
			},
			errExpected: true,
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
			if tc.errExpected {
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

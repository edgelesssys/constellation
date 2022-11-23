/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package vtpm

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	tpmClient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/proto/attest"
	tpmProto "github.com/google/go-tpm-tools/proto/tpm"
	tpmServer "github.com/google/go-tpm-tools/server"
	"github.com/google/go-tpm/tpm2"
)

var (
	// AzurePCRSelection are the PCR values verified for Azure Constellations.
	// PCR[0] is excluded due to changing rarely, but unpredictably.
	// PCR[6] is excluded due to being different for any 2 VMs. See: https://trustedcomputinggroup.org/wp-content/uploads/TCG_PCClient_PFP_r1p05_v23_pub.pdf#%5B%7B%22num%22%3A157%2C%22gen%22%3A0%7D%2C%7B%22name%22%3A%22XYZ%22%7D%2C33%2C400%2C0%5D
	// PCR[10] is excluded since its value is derived from a digest of PCR[0-7]. See: https://sourceforge.net/p/linux-ima/wiki/Home/#ima-measurement-list
	AzurePCRSelection = tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{1, 2, 3, 4, 5, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	}

	// GCPPCRSelection are the PCR values verified for GCP Constellations.
	// On GCP firmware and other host controlled systems are static. This results in the same PCRs for any 2 VMs using the same image.
	GCPPCRSelection = tpmClient.FullPcrSel(tpm2.AlgSHA256)

	// AWSPCRSelection are the PCR values verified for AWS based Constellations.
	// PCR[1] is excluded. See: https://trustedcomputinggroup.org/wp-content/uploads/TCG_PCClient_PFP_r1p05_v23_pub.pdf#%5B%7B%22num%22:157,%22gen%22:0%7D,%7B%22name%22:%22XYZ%22%7D,33,400,0%5D
	// PCR[10] is excluded since its value is derived from a digest of PCR[0-7]. See: https://sourceforge.net/p/linux-ima/wiki/Home/#ima-measurement-list
	AWSPCRSelection = tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{0, 2, 3, 4, 5, 6, 7, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	}

	// QEMUPCRSelection are the PCR values verified for QEMU based Constellations.
	// PCR[1] is excluded. See: https://trustedcomputinggroup.org/wp-content/uploads/TCG_PCClient_PFP_r1p05_v23_pub.pdf#%5B%7B%22num%22:157,%22gen%22:0%7D,%7B%22name%22:%22XYZ%22%7D,33,400,0%5D
	// PCR[10] is excluded since its value is derived from a digest of PCR[0-7]. See: https://sourceforge.net/p/linux-ima/wiki/Home/#ima-measurement-list
	QEMUPCRSelection = tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{0, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	}
)

type (
	// GetTPMAttestationKey loads a TPM key to perform attestation.
	GetTPMAttestationKey func(tpm io.ReadWriter) (*tpmClient.Key, error)
	// GetTPMTrustedAttestationPublicKey verifies and returns the attestation public key.
	GetTPMTrustedAttestationPublicKey func(akPub []byte, instanceInfo []byte) (crypto.PublicKey, error)
	// GetInstanceInfo returns VM metdata.
	GetInstanceInfo func(tpm io.ReadWriteCloser) ([]byte, error)
	// ValidateCVM validates confidential computing capabilities of the instance issuing the attestation.
	ValidateCVM func(attestation AttestationDocument) error
	// VerifyUserData verifies signed user data.
	VerifyUserData func(pub crypto.PublicKey, hash crypto.Hash, hashed, sig []byte) error
)

// AttestationLogger is a logger used to print warnings and infos during attestation validation.
type AttestationLogger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
}

// AttestationDocument contains the TPM attestation with signed user data.
type AttestationDocument struct {
	// Attestation contains the TPM event log, PCR values and quotes, and public key of the key used to sign the attestation.
	Attestation *attest.Attestation
	// InstanceInfo is used to verify the provided public key.
	InstanceInfo []byte
	// arbitrary data, signed by the TPM.
	UserData          []byte
	UserDataSignature []byte
}

// Issuer handles issuing of TPM based attestation documents.
type Issuer struct {
	openTPM           TPMOpenFunc
	getAttestationKey GetTPMAttestationKey
	getInstanceInfo   GetInstanceInfo
}

// NewIssuer returns a new Issuer.
func NewIssuer(openTPM TPMOpenFunc, getAttestationKey GetTPMAttestationKey, getInstanceInfo GetInstanceInfo) *Issuer {
	return &Issuer{
		openTPM:           openTPM,
		getAttestationKey: getAttestationKey,
		getInstanceInfo:   getInstanceInfo,
	}
}

// Issue generates an attestation document using a TPM.
func (i *Issuer) Issue(userData []byte, nonce []byte) ([]byte, error) {
	tpm, err := i.openTPM()
	if err != nil {
		return nil, fmt.Errorf("opening TPM: %w", err)
	}
	defer tpm.Close()

	// Load the TPM's attestation key
	aK, err := i.getAttestationKey(tpm)
	if err != nil {
		return nil, fmt.Errorf("loading attestation key: %w", err)
	}
	defer aK.Close()

	// Create an attestation using the loaded key
	attestation, err := aK.Attest(tpmClient.AttestOpts{Nonce: nonce})
	if err != nil {
		return nil, fmt.Errorf("creating attestation: %w", err)
	}

	// Fetch instance info of the VM
	instanceInfo, err := i.getInstanceInfo(tpm)
	if err != nil {
		return nil, fmt.Errorf("fetching instance info: %w", err)
	}

	// Sign user provided data using the loaded key
	userDataSigned, err := aK.SignData(userData)
	if err != nil {
		return nil, fmt.Errorf("signing user data: %w", err)
	}

	attDoc := AttestationDocument{
		Attestation:       attestation,
		InstanceInfo:      instanceInfo,
		UserData:          userData,
		UserDataSignature: userDataSigned,
	}
	return json.Marshal(attDoc)
}

// Validator handles validation of TPM based attestation.
type Validator struct {
	expectedPCRs   map[uint32][]byte
	enforcedPCRs   map[uint32]struct{}
	getTrustedKey  GetTPMTrustedAttestationPublicKey
	validateCVM    ValidateCVM
	verifyUserData VerifyUserData

	log AttestationLogger
}

// NewValidator returns a new Validator.
func NewValidator(expectedPCRs map[uint32][]byte, enforcedPCRs []uint32, getTrustedKey GetTPMTrustedAttestationPublicKey,
	validateCVM ValidateCVM, verifyUserData VerifyUserData, log AttestationLogger,
) *Validator {
	// Convert the enforced PCR list to a map for convenient and fast lookup
	enforcedMap := make(map[uint32]struct{})
	for _, pcr := range enforcedPCRs {
		enforcedMap[pcr] = struct{}{}
	}

	return &Validator{
		expectedPCRs:   expectedPCRs,
		enforcedPCRs:   enforcedMap,
		getTrustedKey:  getTrustedKey,
		validateCVM:    validateCVM,
		verifyUserData: verifyUserData,
		log:            log,
	}
}

// Validate a TPM based attestation.
func (v *Validator) Validate(attDocRaw []byte, nonce []byte) ([]byte, error) {
	if v.log != nil {
		v.log.Infof("Validating attestation document")
	}

	var attDoc AttestationDocument
	if err := json.Unmarshal(attDocRaw, &attDoc); err != nil {
		return nil, fmt.Errorf("unmarshaling TPM attestation document: %w", err)
	}

	// Verify and retrieve the trusted attestation public key using the provided instance info
	aKP, err := v.getTrustedKey(attDoc.Attestation.AkPub, attDoc.InstanceInfo)
	if err != nil {
		return nil, fmt.Errorf("validating attestation public key: %w", err)
	}

	// Validate confidential computing capabilities of the VM
	if err := v.validateCVM(attDoc); err != nil {
		return nil, fmt.Errorf("verifying VM confidential computing capabilities: %w", err)
	}

	// Verify the TPM attestation
	if _, err := tpmServer.VerifyAttestation(
		attDoc.Attestation,
		tpmServer.VerifyOpts{
			Nonce:      nonce,
			TrustedAKs: []crypto.PublicKey{aKP},
			AllowSHA1:  false,
		},
	); err != nil {
		return nil, fmt.Errorf("verifying attestation document: %w", err)
	}

	// Verify PCRs
	quoteIdx, err := GetSHA256QuoteIndex(attDoc.Attestation.Quotes)
	if err != nil {
		return nil, err
	}
	for idx, pcr := range v.expectedPCRs {
		if !bytes.Equal(pcr, attDoc.Attestation.Quotes[quoteIdx].Pcrs.Pcrs[idx]) {
			if _, ok := v.enforcedPCRs[idx]; ok {
				return nil, fmt.Errorf("untrusted PCR value at PCR index %d", idx)
			}
			if v.log != nil {
				v.log.Warnf("Encountered untrusted PCR value at index %d", idx)
			}
		}
	}

	// Verify signed user data
	digest := sha256.Sum256(attDoc.UserData)
	if err = v.verifyUserData(aKP, crypto.SHA256, digest[:], attDoc.UserDataSignature); err != nil {
		return nil, fmt.Errorf("verifying signed user data: %w", err)
	}

	if v.log != nil {
		v.log.Infof("Successfully validated attestation document")
	}
	return attDoc.UserData, nil
}

// GetSHA256QuoteIndex performs safety checks and returns the index for SHA256 PCR quotes.
func GetSHA256QuoteIndex(quotes []*tpmProto.Quote) (int, error) {
	if len(quotes) == 0 {
		return 0, fmt.Errorf("attestation is missing quotes")
	}
	for idx, quote := range quotes {
		if quote == nil {
			return 0, fmt.Errorf("quote is nil")
		}
		if quote.Pcrs == nil {
			return 0, fmt.Errorf("no PCR data in attestation")
		}
		if quote.Pcrs.Hash == tpmProto.HashAlgo_SHA256 {
			return idx, nil
		}
	}
	return 0, fmt.Errorf("attestation did not include SHA256 hashed PCRs")
}

// VerifyPKCS1v15 is a convenience function to call rsa.VerifyPKCS1v15.
func VerifyPKCS1v15(pub crypto.PublicKey, hash crypto.Hash, hashed, sig []byte) error {
	key, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("key is not an RSA public key")
	}
	return rsa.VerifyPKCS1v15(key, hash, hashed, sig)
}

// GetSelectedPCRs returns a map of the selected PCR hashes.
func GetSelectedPCRs(open TPMOpenFunc, selection tpm2.PCRSelection) (map[uint32][]byte, error) {
	tpm, err := open()
	if err != nil {
		return nil, err
	}
	defer tpm.Close()

	pcrList, err := tpmClient.ReadPCRs(tpm, selection)
	if err != nil {
		return nil, err
	}

	return pcrList.Pcrs, nil
}

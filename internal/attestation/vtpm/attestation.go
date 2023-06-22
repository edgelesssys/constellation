/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package vtpm

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"io"

	tpmClient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/proto/attest"
	tpmProto "github.com/google/go-tpm-tools/proto/tpm"
	tpmServer "github.com/google/go-tpm-tools/server"
	"github.com/google/go-tpm/tpm2"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
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
		PCRs: []int{0, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
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
	GetTPMTrustedAttestationPublicKey func(context.Context, AttestationDocument, []byte) (crypto.PublicKey, error)
	// GetInstanceInfo returns VM metdata.
	GetInstanceInfo func(ctx context.Context, tpm io.ReadWriteCloser, extraData []byte) ([]byte, error)
	// ValidateCVM validates confidential computing capabilities of the instance issuing the attestation.
	ValidateCVM func(attestation AttestationDocument, state *attest.MachineState) error
)

// AttestationDocument contains the TPM attestation with signed user data.
type AttestationDocument struct {
	// Attestation contains the TPM event log, PCR values and quotes, and public key of the key used to sign the attestation.
	Attestation *attest.Attestation
	// InstanceInfo is used to verify the provided public key.
	InstanceInfo []byte
	// arbitrary data, quoted by the TPM.
	UserData []byte
}

// Issuer handles issuing of TPM based attestation documents.
type Issuer struct {
	openTPM           TPMOpenFunc
	getAttestationKey GetTPMAttestationKey
	getInstanceInfo   GetInstanceInfo
	log               attestation.Logger
}

// NewIssuer returns a new Issuer.
func NewIssuer(
	openTPM TPMOpenFunc, getAttestationKey GetTPMAttestationKey,
	getInstanceInfo GetInstanceInfo, log attestation.Logger,
) *Issuer {
	if log == nil {
		log = &attestation.NOPLogger{}
	}
	return &Issuer{
		openTPM:           openTPM,
		getAttestationKey: getAttestationKey,
		getInstanceInfo:   getInstanceInfo,
		log:               log,
	}
}

// Issue generates an attestation document using a TPM.
func (i *Issuer) Issue(ctx context.Context, userData []byte, nonce []byte) (res []byte, err error) {
	i.log.Infof("Issuing attestation statement")
	defer func() {
		if err != nil {
			i.log.Warnf("Failed to issue attestation statement: %s", err)
		}
	}()

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
	extraData := attestation.MakeExtraData(userData, nonce)
	tpmAttestation, err := aK.Attest(tpmClient.AttestOpts{Nonce: extraData})
	if err != nil {
		return nil, fmt.Errorf("creating attestation: %w", err)
	}

	// Fetch instance info of the VM
	instanceInfo, err := i.getInstanceInfo(ctx, tpm, extraData)
	if err != nil {
		return nil, fmt.Errorf("fetching instance info: %w", err)
	}

	attDoc := AttestationDocument{
		Attestation:  tpmAttestation,
		InstanceInfo: instanceInfo,
		UserData:     userData,
	}

	rawAttDoc, err := json.Marshal(attDoc)
	if err != nil {
		return nil, fmt.Errorf("marshaling attestation document: %w", err)
	}

	i.log.Infof("Successfully issued attestation statement")
	return rawAttDoc, nil
}

// Validator handles validation of TPM based attestation.
type Validator struct {
	expected      measurements.M
	getTrustedKey GetTPMTrustedAttestationPublicKey
	validateCVM   ValidateCVM

	log attestation.Logger
}

// NewValidator returns a new Validator.
func NewValidator(expected measurements.M, getTrustedKey GetTPMTrustedAttestationPublicKey,
	validateCVM ValidateCVM, log attestation.Logger,
) *Validator {
	if log == nil {
		log = &attestation.NOPLogger{}
	}
	return &Validator{
		expected:      expected,
		getTrustedKey: getTrustedKey,
		validateCVM:   validateCVM,
		log:           log,
	}
}

// Validate a TPM based attestation.
func (v *Validator) Validate(ctx context.Context, attDocRaw []byte, nonce []byte) (userData []byte, err error) {
	v.log.Infof("Validating attestation document")
	defer func() {
		if err != nil {
			v.log.Warnf("Failed to validate attestation document: %s", err)
		}
	}()

	var attDoc AttestationDocument
	if err := json.Unmarshal(attDocRaw, &attDoc); err != nil {
		return nil, fmt.Errorf("unmarshaling TPM attestation document: %w", err)
	}

	extraData := attestation.MakeExtraData(attDoc.UserData, nonce)

	// Verify and retrieve the trusted attestation public key using the provided instance info
	aKP, err := v.getTrustedKey(ctx, attDoc, extraData)
	if err != nil {
		return nil, fmt.Errorf("validating attestation public key: %w", err)
	}

	// Verify the TPM attestation
	state, err := tpmServer.VerifyAttestation(
		attDoc.Attestation,
		tpmServer.VerifyOpts{
			Nonce:      extraData,
			TrustedAKs: []crypto.PublicKey{aKP},
			AllowSHA1:  false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("verifying attestation document: %w", err)
	}

	// Validate confidential computing capabilities of the VM
	if err := v.validateCVM(attDoc, state); err != nil {
		return nil, fmt.Errorf("verifying VM confidential computing capabilities: %w", err)
	}

	// Verify PCRs
	quoteIdx, err := GetSHA256QuoteIndex(attDoc.Attestation.Quotes)
	if err != nil {
		return nil, err
	}
	for idx, pcr := range v.expected {
		if !bytes.Equal(pcr.Expected[:], attDoc.Attestation.Quotes[quoteIdx].Pcrs.Pcrs[idx]) {
			if pcr.ValidationOpt == measurements.Enforce {
				return nil, fmt.Errorf("untrusted PCR value at PCR index %d", idx)
			}
			v.log.Warnf("Encountered untrusted PCR value at index %d", idx)
		}
	}

	v.log.Infof("Successfully validated attestation document")
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

// GetSelectedMeasurements returns a map of Measurments for the PCRs in selection.
func GetSelectedMeasurements(open TPMOpenFunc, selection tpm2.PCRSelection) (measurements.M, error) {
	tpm, err := open()
	if err != nil {
		return nil, err
	}
	defer tpm.Close()

	pcrList, err := tpmClient.ReadPCRs(tpm, selection)
	if err != nil {
		return nil, err
	}

	m := make(measurements.M)
	for i, pcr := range pcrList.Pcrs {
		if len(pcr) != 32 {
			return nil, fmt.Errorf("invalid measurement: invalid length: %d", len(pcr))
		}
		m[i] = measurements.Measurement{
			Expected: pcr,
		}
	}

	return m, nil
}

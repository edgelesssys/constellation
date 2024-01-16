/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# Azure attestation

Constellation supports multiple attestation technologies on Azure.

  - SEV - Secure Nested Paging (SEV-SNP)

    TPM attestation verified using an SEV-SNP attestation statement.

  - Trusted Launch

    Basic TPM attestation.
*/
package azure

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/legacy/tpm2"
)

const (
	// tpmAkIdx is the NV index of the attestation key used by Azure VMs.
	tpmAkIdx = 0x81000003
)

// GetAttestationKey reads the attestation key put into the TPM during early boot.
func GetAttestationKey(tpm io.ReadWriter) (*tpmclient.Key, error) {
	ak, err := tpmclient.LoadCachedKey(tpm, tpmAkIdx, tpmclient.NullSession{})
	if err != nil {
		return nil, fmt.Errorf("reading HCL attestation key from TPM: %w", err)
	}

	return ak, nil
}

// HCLAkValidator validates an attestation key issued by the Host Compatibility Layer (HCL).
// The HCL is written by Azure, and sits between the Hypervisor and CVM OS.
// The HCL runs in the protected context of the CVM.
type HCLAkValidator struct{}

// Validate validates that the attestation key from the TPM is trustworthy. The steps are:
// 1. runtime data read from the TPM has the same sha256 digest as reported in `report_data` of the SNP report or `TdQuoteBody.ReportData` of the TDX report.
// 2. modulus reported in runtime data matches modulus from key at idx 0x81000003.
// 3. exponent reported in runtime data matches exponent from key at idx 0x81000003.
// The function is currently tested manually on a Azure Ubuntu CVM.
func (a *HCLAkValidator) Validate(runtimeDataRaw []byte, reportData []byte, rsaParameters *tpm2.RSAParams) error {
	var rtData runtimeData
	if err := json.Unmarshal(runtimeDataRaw, &rtData); err != nil {
		return fmt.Errorf("unmarshalling json: %w", err)
	}

	sum := sha256.Sum256(runtimeDataRaw)
	if len(reportData) < len(sum) {
		return fmt.Errorf("reportData has unexpected size: %d", len(reportData))
	}
	if !bytes.Equal(sum[:], reportData[:len(sum)]) {
		return errors.New("unexpected runtimeData digest in TPM")
	}

	if len(rtData.PublicPart) < 1 {
		return errors.New("did not receive any keys in runtime data")
	}
	rawN, err := base64.RawURLEncoding.DecodeString(rtData.PublicPart[0].N)
	if err != nil {
		return fmt.Errorf("decoding modulus string: %w", err)
	}
	if !bytes.Equal(rawN, rsaParameters.ModulusRaw) {
		return fmt.Errorf("unexpected modulus value in TPM")
	}

	rawE, err := base64.RawURLEncoding.DecodeString(rtData.PublicPart[0].E)
	if err != nil {
		return fmt.Errorf("decoding exponent string: %w", err)
	}
	paddedRawE := make([]byte, 4)
	copy(paddedRawE, rawE)
	exponent := binary.LittleEndian.Uint32(paddedRawE)

	// According to this comment [1] the TPM uses "0" to represent the default exponent "65537".
	// The go tpm library also reports the exponent as 0. Thus we have to handle it specially.
	// [1] https://github.com/tpm2-software/tpm2-tools/pull/1973#issue-596685005
	if !((exponent == 65537 && rsaParameters.ExponentRaw == 0) || exponent == rsaParameters.ExponentRaw) {
		return fmt.Errorf("unexpected N value in TPM")
	}

	return nil
}

type runtimeData struct {
	PublicPart []akPub `json:"keys"`
}

// akPub are the public parameters of an RSA attestation key.
type akPub struct {
	E string `json:"e"`
	N string `json:"n"`
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"crypto/sha512"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-tpm-tools/client"
	tpmclient "github.com/google/go-tpm-tools/client"
)

// Issuer for AWS SNP attestation.
type Issuer struct {
	variant.AWSSEVSNP
	*vtpm.Issuer
}

// NewIssuer creates a SEV-SNP based issuer for AWS.
func NewIssuer(log attestation.Logger) *Issuer {
	return &Issuer{
		Issuer: vtpm.NewIssuer(
			vtpm.OpenVTPM,
			getAttestationKey,
			getInstanceInfo,
			log,
		),
	}
}

// getAttestationKey returns a new attestation key.
func getAttestationKey(tpm io.ReadWriter) (*tpmclient.Key, error) {
	tpmAk, err := client.AttestationKeyRSA(tpm)
	if err != nil {
		return nil, fmt.Errorf("creating RSA Endorsement key: %w", err)
	}

	return tpmAk, nil
}

// getInstanceInfo generates an extended SNP report, i.e. the report and any loaded certificates.
// Report generation is triggered by sending ioctl syscalls to the SNP guest device, the AMD PSP generates the report.
// The returned bytes will be written into the attestation document.
func getInstanceInfo(_ context.Context, tpm io.ReadWriteCloser, _ []byte) ([]byte, error) {
	tpmAk, err := client.AttestationKeyRSA(tpm)
	if err != nil {
		return nil, fmt.Errorf("creating RSA Endorsement key: %w", err)
	}

	encoded, err := x509.MarshalPKIXPublicKey(tpmAk.PublicKey())
	if err != nil {
		return nil, fmt.Errorf("marshalling public key: %w", err)
	}

	akDigest := sha512.Sum512(encoded)

	report, certs, err := snp.GetExtendedReport(akDigest)
	if err != nil {
		return nil, fmt.Errorf("getting extended report: %w", err)
	}

	vlek, err := pemEncodedVLEK(certs)
	if err != nil {
		return nil, fmt.Errorf("parsing vlek: %w", err)
	}

	raw, err := json.Marshal(snp.InstanceInfo{AttestationReport: report, ReportSigner: vlek})
	if err != nil {
		return nil, fmt.Errorf("marshalling instance info: %w", err)
	}

	return raw, nil
}

// pemEncodedVLEK takes a marshalled SNP certificate table and returns the PEM-encoded VLEK certificate.
// AMD documentation on certificate tables can be found in section 4.1.8.1, revision 2.03 "SEV-ES Guest-Hypervisor Communication Block Standardization".
// https://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56421.pdf
func pemEncodedVLEK(certs []byte) ([]byte, error) {
	certTable := abi.CertTable{}
	if err := certTable.Unmarshal(certs); err != nil {
		return nil, fmt.Errorf("unmarshalling SNP certificate table: %w", err)
	}

	vlekRaw, err := certTable.GetByGUIDString(abi.VlekGUID)
	if err != nil {
		return nil, fmt.Errorf("getting VLEK certificate: %w", err)
	}

	// An optional check for certificate well-formedness. vlekRaw == cert.Raw.
	cert, err := x509.ParseCertificate(vlekRaw)
	if err != nil {
		return nil, fmt.Errorf("parsing certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	return certPEM, nil
}

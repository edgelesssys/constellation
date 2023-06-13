/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"

	sevclient "github.com/google/go-sev-guest/client"
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
		return nil, fmt.Errorf("error creating RSA Endorsement key: %w", err)
	}

	return tpmAk, nil
}

// getInstanceInfo generates an extended SNP report, i.e. the report and any loaded certificates.
// Report generation is triggered by sending ioctl syscalls to the SNP guest device, the AMD PSP generates the report.
// The returned bytes will be written into the attestation document.
func getInstanceInfo(context.Context, io.ReadWriteCloser, []byte) ([]byte, error) {
	device, err := sevclient.OpenDevice()
	if err != nil {
		return nil, fmt.Errorf("opening sev device: %w", err)
	}
	defer device.Close()

	report, certs, err := sevclient.GetRawExtendedReportAtVmpl(device, [64]byte{}, 0)
	if err != nil {
		return nil, fmt.Errorf("getting extended report: %w", err)
	}

	raw, err := json.Marshal(instanceInfo{Report: report, Certs: certs})
	if err != nil {
		return nil, fmt.Errorf("marshalling instance info: %w", err)
	}

	return raw, nil
}

type instanceInfo struct {
	// Report contains the marshalled AMD SEV-SNP Report.
	Report []byte
	// Certs contains the PEM encoded VLEK and ASK certificates, queried from the AMD PSP of the issuing party.
	Certs []byte
}

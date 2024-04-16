/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"

	"github.com/google/go-sev-guest/abi"
	sevclient "github.com/google/go-sev-guest/client"
	"github.com/google/go-tpm-tools/client"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/proto/attest"
)

// Issuer issues SEV-SNP attestations.
type Issuer struct {
	variant.GCPSEVSNP
	*vtpm.Issuer
}

// NewIssuer creates a SEV-SNP based issuer for GCP.
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
	tpmAk, err := client.GceAttestationKeyRSA(tpm)
	if err != nil {
		return nil, fmt.Errorf("creating RSA Endorsement key: %w", err)
	}

	return tpmAk, nil
}

// getInstanceInfo generates an extended SNP report, i.e. the report and any loaded certificates.
// Report generation is triggered by sending ioctl syscalls to the SNP guest device, the AMD PSP generates the report.
// The returned bytes will be written into the attestation document.
func getInstanceInfo(_ context.Context, _ io.ReadWriteCloser, extraData []byte) ([]byte, error) {
	if len(extraData) > 64 {
		return nil, fmt.Errorf("extra data too long: %d, should be 64 bytes at most", len(extraData))
	}
	var extraData64 [64]byte
	copy(extraData64[:], extraData)

	device, err := sevclient.OpenDevice()
	if err != nil {
		return nil, fmt.Errorf("opening sev device: %w", err)
	}
	defer device.Close()

	report, certs, err := sevclient.GetRawExtendedReportAtVmpl(device, extraData64, 0)
	if err != nil {
		return nil, fmt.Errorf("getting extended report: %w", err)
	}

	vcek, certChain, err := parseSNPCertTable(certs)
	if err != nil {
		return nil, fmt.Errorf("parsing vcek: %w", err)
	}

	gceInstanceInfo, err := gceInstanceInfo()
	if err != nil {
		return nil, fmt.Errorf("getting GCE instance info: %w", err)
	}

	raw, err := json.Marshal(snp.InstanceInfo{
		AttestationReport: report,
		ReportSigner:      vcek,
		CertChain:         certChain,
		GCP:               gceInstanceInfo,
	})
	if err != nil {
		return nil, fmt.Errorf("marshalling instance info: %w", err)
	}

	return raw, nil
}

// gceInstanceInfo returns the instance info for a GCE instance from the metadata API.
func gceInstanceInfo() (*attest.GCEInstanceInfo, error) {
	c := gcp.MetadataClient{}

	instanceName, err := c.InstanceName()
	if err != nil {
		return nil, fmt.Errorf("getting instance name: %w", err)
	}

	projectID, err := c.ProjectID()
	if err != nil {
		return nil, fmt.Errorf("getting project ID: %w", err)
	}

	zone, err := c.Zone()
	if err != nil {
		return nil, fmt.Errorf("getting zone: %w", err)
	}

	return &attest.GCEInstanceInfo{
		InstanceName: instanceName,
		ProjectId:    projectID,
		Zone:         zone,
	}, nil
}

// parseSNPCertTable takes a marshalled SNP certificate table and returns the PEM-encoded VCEK certificate and,
// if present, the ASK of the SNP certificate chain.
// AMD documentation on certificate tables can be found in section 4.1.8.1, revision 2.03 "SEV-ES Guest-Hypervisor Communication Block Standardization".
// https://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56421.pdf
func parseSNPCertTable(certs []byte) (vcekPEM []byte, certChain []byte, err error) {
	certTable := abi.CertTable{}
	if err := certTable.Unmarshal(certs); err != nil {
		return nil, nil, fmt.Errorf("unmarshalling SNP certificate table: %w", err)
	}

	vcekRaw, err := certTable.GetByGUIDString(abi.VcekGUID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting VCEK certificate: %w", err)
	}

	// An optional check for certificate well-formedness. vcekRaw == cert.Raw.
	vcek, err := x509.ParseCertificate(vcekRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing certificate: %w", err)
	}

	vcekPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: vcek.Raw,
	})

	var askPEM []byte
	if askRaw, err := certTable.GetByGUIDString(abi.AskGUID); err == nil {
		ask, err := x509.ParseCertificate(askRaw)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing ASK certificate: %w", err)
		}

		askPEM = pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: ask.Raw,
		})
	}

	return vcekPEM, askPEM, nil
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/tpm2"
)

const (
	lenHclHeader                   = 0x20
	lenSnpReport                   = 0x4a0
	lenSnpReportRuntimeDataPadding = 0x14
	tpmReportIdx                   = 0x01400001
	tpmAkCertIdx                   = 0x1C101D0
	tpmAkIdx                       = 0x81000003
)

// Issuer for Azure TPM attestation.
type Issuer struct {
	oid.AzureSEVSNP
	*vtpm.Issuer

	imds         imdsAPI
	reportGetter tpmReportGetter
	maa          maaTokenCreator
}

// NewIssuer initializes a new Azure Issuer.
func NewIssuer(log vtpm.AttestationLogger) *Issuer {
	i := &Issuer{
		imds:         newIMDSClient(),
		reportGetter: &tpmReport{},
		maa:          newMAAClient(),
	}

	i.Issuer = vtpm.NewIssuer(
		vtpm.OpenVTPM,
		getAttestationKey,
		i.getInstanceInfo,
		log,
	)
	return i
}

// getInstanceInfo loads and returns the SEV-SNP attestation report [1] and the
// AMD VCEK certificate chain.
// The attestation report is loaded from the TPM, the certificate chain is queried
// from the cloud metadata API.
// [1] https://github.com/AMDESE/sev-guest/blob/main/include/attestation.h
func (i *Issuer) getInstanceInfo(ctx context.Context, tpm io.ReadWriteCloser, userData []byte) ([]byte, error) {
	hclReport, err := i.reportGetter.get(tpm)
	if err != nil {
		return nil, fmt.Errorf("reading report from TPM: %w", err)
	}
	if len(hclReport) < lenHclHeader+lenSnpReport+lenSnpReportRuntimeDataPadding {
		return nil, fmt.Errorf("report read from TPM is shorter than expected: %x", hclReport)
	}
	hclReport = hclReport[lenHclHeader:]

	vcekResponse, err := i.imds.getVcek(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving VCEK certificate from IMDS API: %w", err)
	}

	vcekCert := []byte(vcekResponse.VcekCert)
	vcekChain := []byte(vcekResponse.CertificateChain)

	var maaToken string

	maaURL, err := i.imds.getMAAURL(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving MAA URL from IMDS API: %w", err)
	}
	if maaURL != "" {
		maaToken, err = i.maa.createToken(ctx, tpm, maaURL, userData, hclReport, vcekCert, vcekChain)
		if err != nil {
			return nil, fmt.Errorf("creating MAA token: %w", err)
		}
	}

	instanceInfo := azureInstanceInfo{
		Vcek:              vcekCert,
		CertChain:         vcekChain,
		AttestationReport: cutSNPReport(hclReport),
		RuntimeData:       cutRuntimeData(hclReport),
		MAAToken:          maaToken,
	}
	statement, err := json.Marshal(instanceInfo)
	if err != nil {
		return nil, fmt.Errorf("marshalling AzureInstanceInfo: %w", err)
	}

	return statement, nil
}

// getAttestationKey reads the attestation key put into the TPM during early boot.
func getAttestationKey(tpm io.ReadWriter) (*tpmclient.Key, error) {
	ak, err := tpmclient.LoadCachedKey(tpm, tpmAkIdx)
	if err != nil {
		return nil, fmt.Errorf("reading HCL attestation key from TPM: %w", err)
	}

	return ak, nil
}

func cutRuntimeData(hclReport []byte) []byte {
	runtimeData, _, _ := bytes.Cut(hclReport[lenSnpReport+lenSnpReportRuntimeDataPadding:], []byte{0})
	return runtimeData
}

func cutSNPReport(hclReport []byte) []byte {
	return hclReport[:lenSnpReport]
}

type tpmReport struct{}

func (s *tpmReport) get(tpm io.ReadWriteCloser) ([]byte, error) {
	return tpm2.NVReadEx(tpm, tpmReportIdx, tpm2.HandleOwner, "", 0)
}

type tpmReportGetter interface {
	get(tpm io.ReadWriteCloser) ([]byte, error)
}

type imdsAPI interface {
	getVcek(ctx context.Context) (vcekResponse, error)
	getMAAURL(ctx context.Context) (string, error)
}

type maaTokenCreator interface {
	createToken(context.Context, io.ReadWriter, string, []byte, []byte, []byte, []byte) (string, error)
}

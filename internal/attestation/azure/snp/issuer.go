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
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/go-azguestattestation/maa"
	tpmclient "github.com/google/go-tpm-tools/client"
)

const tpmAkIdx = 0x81000003

// Issuer for Azure TPM attestation.
type Issuer struct {
	variant.AzureSEVSNP
	*vtpm.Issuer

	imds imdsAPI
	maa  maaTokenCreator
}

// NewIssuer initializes a new Azure Issuer.
func NewIssuer(log attestation.Logger) *Issuer {
	i := &Issuer{
		imds: newIMDSClient(),
		maa:  newMAAClient(),
	}

	i.Issuer = vtpm.NewIssuer(
		vtpm.OpenVTPM,
		getAttestationKey,
		i.getInstanceInfo,
		log,
	)
	return i
}

func (i *Issuer) getInstanceInfo(ctx context.Context, tpm io.ReadWriteCloser, userData []byte) ([]byte, error) {
	params, err := i.maa.newParameters(ctx, userData, tpm)
	if err != nil {
		return nil, fmt.Errorf("getting system parameters: %w", err)
	}

	var maaToken string

	maaURL, err := i.imds.getMAAURL(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving MAA URL from IMDS API: %w", err)
	}
	if maaURL != "" {
		maaToken, err = i.maa.createToken(ctx, tpm, maaURL, userData, params)
		if err != nil {
			return nil, fmt.Errorf("creating MAA token: %w", err)
		}
	}

	instanceInfo := azureInstanceInfo{
		Vcek:              params.VcekCert,
		CertChain:         params.VcekChain,
		AttestationReport: params.SNPReport,
		RuntimeData:       params.RuntimeData,
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

type imdsAPI interface {
	getMAAURL(ctx context.Context) (string, error)
}

type maaTokenCreator interface {
	newParameters(context.Context, []byte, io.ReadWriter) (maa.Parameters, error)
	createToken(context.Context, io.ReadWriter, string, []byte, maa.Parameters) (string, error)
}

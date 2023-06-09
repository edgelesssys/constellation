/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package nitrotpm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"

	"github.com/google/go-tpm-tools/client"
	tpmclient "github.com/google/go-tpm-tools/client"
)

// Issuer for AWS TPM attestation.
type Issuer struct {
	variant.AWSNitroTPM
	*vtpm.Issuer
}

// NewIssuer creates a TPM based issuer for AWS.
func NewIssuer(log attestation.Logger) *Issuer {
	return &Issuer{
		Issuer: vtpm.NewIssuer(
			vtpm.OpenVTPM,
			getAttestationKey,
			getInstanceInfo(imds.New(imds.Options{})),
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

// getInstanceInfo returns information about the current instance using the aws Metadata SDK.
// The returned bytes will be written into the attestation document.
func getInstanceInfo(client awsMetaData) func(context.Context, io.ReadWriteCloser, []byte) ([]byte, error) {
	return func(ctx context.Context, _ io.ReadWriteCloser, _ []byte) ([]byte, error) {
		ec2InstanceIdentityOutput, err := client.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
		if err != nil {
			return nil, fmt.Errorf("fetching instance identity document: %w", err)
		}
		return json.Marshal(ec2InstanceIdentityOutput.InstanceIdentityDocument)
	}
}

type awsMetaData interface {
	GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error)
}

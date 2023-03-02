/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"

	"github.com/google/go-tpm-tools/client"
	tpmclient "github.com/google/go-tpm-tools/client"
)

// Issuer for AWS TPM attestation.
type Issuer struct {
	oid.AWSNitroTPM
	*vtpm.Issuer
}

// NewIssuer creates a new OpenVTPM based issuer for AWS.
func NewIssuer(log vtpm.AttestationLogger) *Issuer {
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
		log.Fatalf("error creating RSA Endorsement key!")
		return nil, err
	}

	return tpmAk, nil
}

// getInstanceInfo returns information about the current instance using the aws Metadata SDK.
// The returned bytes will be written into the attestation document.
func getInstanceInfo(client awsMetaData) func(tpm io.ReadWriteCloser) ([]byte, error) {
	return func(io.ReadWriteCloser) ([]byte, error) {
		ec2InstanceIdentityOutput, err := client.GetInstanceIdentityDocument(context.Background(), &imds.GetInstanceIdentityDocumentInput{})
		if err != nil {
			return nil, errors.New("unable to fetch instance identity document")
		}
		return json.Marshal(ec2InstanceIdentityOutput.InstanceIdentityDocument)
	}
}

type awsMetaData interface {
	GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error)
}

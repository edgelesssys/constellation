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

type Issuer struct {
	oid.AWS
	*vtpm.Issuer
}

func NewIssuer() *Issuer {
	awsIMDS := imds.New(imds.Options{})
	GetInstanceInfo := getInstanceInfo(awsIMDS)

	return &Issuer{
		Issuer: vtpm.NewIssuer(
			vtpm.OpenVTPM,
			getAttestationKey,
			GetInstanceInfo,
		),
	}
}

func getAttestationKey(tpm io.ReadWriter) (*tpmclient.Key, error) {
	tpmAk, err := client.AttestationKeyRSA(tpm)
	if err != nil {
		log.Fatalf("error creating RSA Endorsement key!")
	}

	return tpmAk, nil
}

// Get information about the current instance using the aws Metadata SDK
// The returned bytes will be written into the attestation document
func getInstanceInfo(client awsMetaData) func(tpm io.ReadWriteCloser) ([]byte, error) {
	return func(io.ReadWriteCloser) ([]byte, error) {
		ctx := context.TODO()
		ec2InstanceIdentityOutput, err := client.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
		ec2InstanceIdentityDocument := ec2InstanceIdentityOutput.InstanceIdentityDocument

		if err != nil {
			return nil, errors.New("unable to fetch instance identity document")
		}

		statement, err := json.Marshal(ec2InstanceIdentityDocument)
		if err != nil {
			return nil, errors.New("unable to marshal aws instance info")
		}

		return statement, nil
	}
}

type awsMetaData interface {
	GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error)
}

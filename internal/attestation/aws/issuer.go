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

	"github.com/google/go-tpm-tools/client"
	tpmclient "github.com/google/go-tpm-tools/client"
)

type Issuer struct {
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
		log.Fatalf("error getting RSA Endorsement key!")
	}

	return tpmAk, nil
}

// get information about the current instance using the aws Metadata SDK
func getInstanceInfo(client awsMetaData) func(tpm io.ReadWriteCloser) ([]byte, error) {
	return func(io.ReadWriteCloser) ([]byte, error) {
		ctx := context.TODO()
		ec2InstanceIdentityOutput, err := client.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
		ec2InstanceIdentityDocument := ec2InstanceIdentityOutput.InstanceIdentityDocument

		if err != nil {
			return nil, errors.New("unable to fetch instance identity document")
		}

		instanceInfo := awsInstanceInfo{
			ec2InstanceIdentityDocument.Region,
			ec2InstanceIdentityDocument.AccountID,
			ec2InstanceIdentityDocument.InstanceID,
		}

		statement, err := json.Marshal(instanceInfo)
		if err != nil {
			return nil, errors.New("unable to marshal aws instance info")
		}

		return statement, nil
	}
}

type awsMetaData interface {
	GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error)
}

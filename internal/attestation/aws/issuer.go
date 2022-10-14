/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"

	"github.com/google/go-tpm-tools/client"
	tpmclient "github.com/google/go-tpm-tools/client"
)

type Issuer struct {
	*vtpm.Issuer
}

func NewIssuer() *Issuer {
	return &Issuer{
		Issuer: vtpm.NewIssuer(
			vtpm.OpenVTPM,
			getAttestationKey,
			getInstanceInfo(ec2metadata.New(nil)),
		),
	}
}

func getAttestationKey(tpm io.ReadWriter) (*tpmclient.Key, error) {
	key, err := client.AttestationKeyRSA(tpm)
	if err != nil {
		return nil, errors.New("unable to retrieve attestation key")
	}

	return key, nil
}

// Get the metadta infos from the AWS Instance Document (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html)
func getInstanceInfo(client awsMetadataClient) func(tpm io.ReadWriteCloser) ([]byte, error) {
	return func(tpm io.ReadWriteCloser) ([]byte, error) {
		identityDocument, err := client.GetInstanceIdentityDocument()
		if err != nil {
			return nil, errors.New("unable to fetch instance identity document")
		}

		//
		instanceInfo := awsInstanceInfo{
			identityDocument.Region,
			identityDocument.AccountID,
			identityDocument.InstanceID,
		}

		statement, err := json.Marshal(instanceInfo)
		if err != nil {
			return nil, errors.New("unable to marshal aws instance info")
		}

		return statement, nil
	}
}

type awsMetadataClient interface {
	GetInstanceIdentityDocument() (ec2metadata.EC2InstanceIdentityDocument, error)
}

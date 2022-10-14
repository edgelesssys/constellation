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

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"

	"github.com/google/go-tpm-tools/client"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
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
	akIndex := 0x81000001
	keyTemplate := tpm2.Public{
		Type:       tpm2.AlgRSA,
		NameAlg:    tpm2.AlgSHA256,
		Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth | tpm2.FlagNoDA | tpm2.FlagRestricted | tpm2.FlagSign,
		RSAParameters: &tpm2.RSAParams{
			Sign: &tpm2.SigScheme{
				Alg:  tpm2.AlgRSASSA,
				Hash: tpm2.AlgSHA256,
			},
			KeyBits: 2048,
		},
	}

	tpmAk, err := client.NewCachedKey(tpm, tpm2.HandleOwner, keyTemplate, tpmutil.Handle(akIndex))

	if err != nil {
		return nil, errors.New("cannot get cached key")
	}

	return tpmAk, nil
}

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

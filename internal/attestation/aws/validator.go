/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"crypto"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
)

type Validator struct {
	oid.AWS
	*vtpm.Validator
}

func NewValidator(pcrs map[uint32][]byte, enforcedPCRs []uint32, log vtpm.WarnLogger) *Validator {
	return &Validator{
		Validator: vtpm.NewValidator(
			pcrs,
			enforcedPCRs,
			getTrustedKey,
			validateCVM,
			vtpm.VerifyPKCS1v15,
			log,
		),
	}
}

func getTrustedKey(akPub []byte, instanceInfo []byte) (crypto.PublicKey, error) {
	return nil, fmt.Errorf("for now you have to trust aws on this")
}

// verify if the virtual machine has the tpm2.0 featiure enabled
func tpmEnabled(idDocument imds.InstanceIdentityDocument) error {
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/verify-nitrotpm-support-on-ami.html
	// 1. get the vm's ami (from IdentidyDocument.imageId)
	// 2. check the value of key "TpmSupport": {"Value": "v2.0"}"

	vmRegion := idDocument.Region
	imageId := idDocument.ImageID

	// create session for ec2 requests
	session, err := session.NewSession()
	if err != nil {
		return err
	}

	svc := ec2.New(session,
		aws.NewConfig().
			WithRegion(vmRegion))

	imageAttributeOutput, err := svc.DescribeImageAttribute(&ec2.DescribeImageAttributeInput{
		ImageId:   aws.String(imageId),
		Attribute: aws.String("tpmSupport"),
	})

	if err != nil {
		fmt.Printf("Got error %v", err)
		return err
	}

	if *imageAttributeOutput.TpmSupport.Value == "v2.0" {
		return nil
	}

	return fmt.Errorf("iam image %s does not support tpm2.0", imageId)
}

// Validate if the current instance is a CVM instance.
// This information is retreived by the helper function tpmEnabled
func validateCVM(attestation vtpm.AttestationDocument) error {
	if attestation.Attestation == nil {
		return errors.New("missing attestation document")
	}

	// retrieve instanceIdentityDocument from attestation document
	idDocument := imds.InstanceIdentityDocument{}
	err := json.Unmarshal(attestation.UserData, &idDocument)
	if err != nil {
		return err
	}

	return tpmEnabled(idDocument)
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

	conf, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(vmRegion))
	if err != nil {
		return err
	}

	client := ec2.NewFromConfig(conf)

	// Currently, there seems to be a problem with retrieving image attributes directly.
	// Alternatively, parse it from the general output.
	imageOutput, err := client.DescribeImages(context.TODO(), &ec2.DescribeImagesInput{ImageIds: []string{imageId}})
	if err != nil {
		return err
	}

	if imageOutput.Images[0].TpmSupport == "v2.0" {
		return nil
	}

	return fmt.Errorf("iam image %s does not support TPM v2.0", imageId)
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

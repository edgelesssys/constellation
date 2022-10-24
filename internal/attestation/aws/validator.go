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
	"github.com/google/go-tpm/tpm2"
)

// Issuer for AWS TPM attestation
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

// getTrustedKeys return the public area of the provides attestation key.
// Normally, here the trust of this key should be verified, but currently AWS does not provide this feature.
func getTrustedKey(akPub []byte, instanceInfo []byte) (crypto.PublicKey, error) {
	// Copied from https://github.com/edgelesssys/constellation/blob/main/internal/attestation/qemu/validator.go
	pubArea, err := tpm2.DecodePublic(akPub)
	if err != nil {
		return nil, err
	}

	return pubArea.Key()
}

// tpmEnabled verifies if the virtual machine has the tpm2.0 feature enabled
func tpmEnabled(idDocument imds.InstanceIdentityDocument) error {
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/verify-nitrotpm-support-on-ami.html
	// 1. get the vm's ami (from IdentiTyDocument.imageId)
	// 2. check the value of key "TpmSupport": {"Value": "v2.0"}"

	imageId := idDocument.ImageID
	ctx := context.Background()

	conf, err := config.LoadDefaultConfig(ctx, config.WithEC2IMDSRegion())
	if err != nil {
		return err
	}

	client := ec2.NewFromConfig(conf)

	// Currently, there seems to be a problem with retrieving image attributes directly.
	// Alternatively, parse it from the general output.
	imageOutput, err := client.DescribeImages(ctx, &ec2.DescribeImagesInput{ImageIds: []string{imageId}})
	if err != nil {
		return err
	}

	if imageOutput.Images[0].TpmSupport == "v2.0" {
		return nil
	}

	return fmt.Errorf("iam image %s does not support TPM v2.0", imageId)
}

// validateVCM validates if the current instance is a CVM instance.
// This information is retrieved by the helper function tpmEnabled
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

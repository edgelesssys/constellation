/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/google/go-tpm/tpm2"
)

// Validator for AWS TPM attestation.
type Validator struct {
	oid.AWS
	*vtpm.Validator
	getDescribeClient func(context.Context) (awsMetadataAPI, error)
}

// NewValidator create a new Validator structure and returns it.
func NewValidator(pcrs map[uint32][]byte, enforcedPCRs []uint32, log vtpm.AttestationLogger) *Validator {
	v := &Validator{}
	v.Validator = vtpm.NewValidator(
		pcrs,
		enforcedPCRs,
		getTrustedKey,
		v.tpmEnabled,
		vtpm.VerifyPKCS1v15,
		log,
	)
	v.getDescribeClient = getEC2Client
	return v
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

// tpmEnabled verifies if the virtual machine has the tpm2.0 feature enabled.
func (v *Validator) tpmEnabled(attestation vtpm.AttestationDocument) error {
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/verify-nitrotpm-support-on-ami.html
	// 1. Get the vm's ami (from IdentiTyDocument.imageId)
	// 2. Check the value of key "TpmSupport": {"Value": "v2.0"}"
	ctx := context.Background()

	idDocument := imds.InstanceIdentityDocument{}
	err := json.Unmarshal(attestation.UserData, &idDocument)
	if err != nil {
		return err
	}

	imageId := idDocument.ImageID

	client, err := v.getDescribeClient(ctx)
	if err != nil {
		return err
	}
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

func getEC2Client(ctx context.Context) (awsMetadataAPI, error) {
	client, err := config.LoadDefaultConfig(ctx, config.WithEC2IMDSRegion())
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(client), nil
}

type awsMetadataAPI interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
}

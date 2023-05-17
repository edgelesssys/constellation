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

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/tpm2"
)

// Validator for AWS TPM attestation.
type Validator struct {
	variant.AWSNitroTPM
	*vtpm.Validator
	getDescribeClient func(context.Context, string) (awsMetadataAPI, error)
}

// NewValidator create a new Validator structure and returns it.
func NewValidator(cfg *config.AWSNitroTPM, log attestation.Logger) *Validator {
	v := &Validator{}
	v.Validator = vtpm.NewValidator(
		cfg.Measurements,
		getTrustedKey,
		v.tpmEnabled,
		log,
	)
	v.getDescribeClient = getEC2Client
	return v
}

// getTrustedKeys return the public area of the provides attestation key.
// Normally, here the trust of this key should be verified, but currently AWS does not provide this feature.
func getTrustedKey(_ context.Context, attDoc vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error) {
	// Copied from https://github.com/edgelesssys/constellation/blob/main/internal/attestation/qemu/validator.go
	pubArea, err := tpm2.DecodePublic(attDoc.Attestation.AkPub)
	if err != nil {
		return nil, err
	}

	return pubArea.Key()
}

// tpmEnabled verifies if the virtual machine has the tpm2.0 feature enabled.
func (v *Validator) tpmEnabled(attestation vtpm.AttestationDocument, _ *attest.MachineState) error {
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/verify-nitrotpm-support-on-ami.html
	// 1. Get the vm's ami (from IdentiTyDocument.imageId)
	// 2. Check the value of key "TpmSupport": {"Value": "v2.0"}"
	ctx := context.Background()

	idDocument := imds.InstanceIdentityDocument{}
	err := json.Unmarshal(attestation.InstanceInfo, &idDocument)
	if err != nil {
		return err
	}

	imageID := idDocument.ImageID

	client, err := v.getDescribeClient(ctx, idDocument.Region)
	if err != nil {
		return err
	}
	// Currently, there seems to be a problem with retrieving image attributes directly.
	// Alternatively, parse it from the general output.
	imageOutput, err := client.DescribeImages(ctx, &ec2.DescribeImagesInput{ImageIds: []string{imageID}})
	if err != nil {
		return err
	}

	if imageOutput.Images[0].TpmSupport == "v2.0" {
		return nil
	}

	return fmt.Errorf("iam image %s does not support TPM v2.0", imageID)
}

func getEC2Client(ctx context.Context, region string) (awsMetadataAPI, error) {
	client, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(client), nil
}

type awsMetadataAPI interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/stretchr/testify/assert"
)

func TestGeTrustedKey(t *testing.T) {
	testCases := map[string]struct {
		attDoc  []byte
		nonce   []byte
		wantErr bool
	}{
		"nul byte docs": {
			attDoc:  []byte{0x00, 0x00, 0x00, 0x00},
			nonce:   []byte{0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		"nil": {
			attDoc:  nil,
			nonce:   nil,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			out, err := getTrustedKey(tc.attDoc, tc.nonce)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			assert.Nil(out)
		})
	}
}

func TestTpmEnabled(t *testing.T) {
	mockDocBase := imds.InstanceIdentityDocument{
		DevpayProductCodes:      []string{"devpayProductCodes"},
		MarketplaceProductCodes: []string{"marketplaceProductCodes"},
		AvailabilityZone:        "availabilityZone",
		PrivateIP:               "privateIp",
		Version:                 "version",
		Region:                  "region",
		InstanceID:              "instanceId",
		BillingProducts:         []string{"billingProducts"},
		InstanceType:            "instanceType",
		AccountID:               "accountId",
		PendingTime:             time.Now(),
		ImageID:                 "imageId",
		KernelID:                "kernelId",
		RamdiskID:               "ramdiskId",
		Architecture:            "architecture",
	}

	// The pinned images are valid at the time of this commit.
	// When we finally support AWS, we should make these values part of the version bump.
	// Make sure, that the image is visible to all required users/clients. Verify using
	// aws ec2 describe-image-attribute  --region <REGION> --image-id <IMAGE_ID> --attribute tpmSupport

	// Since Amazon does not offer Linux stock AMI's with TPM support enabled, we use Windows images in this test.
	// https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/enable-nitrotpm-prerequisites.html
	mockDocTpm := mockDocBase
	mockDocTpm.ImageID = "ami-0e6b1588acc5a055f"
	mockDocTpm.Region = "us-east-2"

	// Stock aws linux (per default no TPM enabled), here could be the ami of ANY stock linux.
	mockDocNoTpm := mockDocBase
	mockDocNoTpm.ImageID = "ami-0b59bfac6be064b78"
	mockDocTpm.Region = "us-east-2"

	testCases := map[string]struct {
		idDoc   imds.InstanceIdentityDocument
		wantErr bool
	}{
		"ami with tpm": {
			idDoc:   mockDocTpm,
			wantErr: false,
		},
		"ami without tpm": {
			idDoc:   mockDocNoTpm,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := tpmEnabled(tc.idDoc)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.Nil(err)
			}
		})
	}
}

func TestValidateCVM(t *testing.T) {
	mockDocBase := imds.InstanceIdentityDocument{
		DevpayProductCodes:      []string{"devpayProductCodes"},
		MarketplaceProductCodes: []string{"marketplaceProductCodes"},
		AvailabilityZone:        "availabilityZone",
		PrivateIP:               "privateIp",
		Version:                 "version",
		Region:                  "region",
		InstanceID:              "instanceId",
		BillingProducts:         []string{"billingProducts"},
		InstanceType:            "instanceType",
		AccountID:               "accountId",
		PendingTime:             time.Now(),
		ImageID:                 "imageId",
		KernelID:                "kernelId",
		RamdiskID:               "ramdiskId",
		Architecture:            "architecture",
	}

	mockDocTpm := mockDocBase
	mockDocTpm.ImageID = "ami-0e6b1588acc5a055f"
	mockDocTpm.Region = "us-east-2"

	mockDocNoTpm := mockDocBase
	mockDocNoTpm.ImageID = "ami-0b59bfac6be064b78"
	mockDocTpm.Region = "us-east-2"

	attDocTpm := vtpm.AttestationDocument{}
	attDocTpm.UserData, _ = json.Marshal(mockDocTpm)
	attDocTpm.Attestation = &attest.Attestation{}

	attDocNoTpm := vtpm.AttestationDocument{}
	attDocNoTpm.UserData, _ = json.Marshal(mockDocNoTpm)
	attDocNoTpm.Attestation = &attest.Attestation{}

	testCases := map[string]struct {
		attestationDoc vtpm.AttestationDocument
		wantErr        bool
	}{
		"CVM enabled": {
			attestationDoc: attDocTpm,
			wantErr:        false,
		},
		"CVM disabled": {
			attestationDoc: attDocNoTpm,
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := validateCVM(tc.attestationDoc)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.Nil(err)
			}
		})
	}
}

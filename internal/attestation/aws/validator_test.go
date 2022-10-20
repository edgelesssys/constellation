/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"testing"
	"time"

	"github.com/aws/smithy-go/middleware"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
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
	mockDocBase := imds.GetInstanceIdentityDocumentOutput{
		InstanceIdentityDocument: imds.InstanceIdentityDocument{
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
		},
		ResultMetadata: middleware.Metadata{},
	}

	// is this a good idea? The images might not last forever -> also mock this
	mockDocTpm := mockDocBase
	mockDocTpm.ImageID = "ami-0a6cfb4290afa8e26"
	mockDocTpm.AvailabilityZone = "us-east-2"

	mockDocNoTpm := mockDocBase
	mockDocNoTpm.ImageID = "ami-0a6cfb4290afa8e26"
	mockDocNoTpm.AvailabilityZone = "us-east-2"

	testCases := map[string]struct {
		idDoc   imds.GetInstanceIdentityDocumentOutput
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

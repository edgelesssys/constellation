/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/smithy-go/middleware"
	"github.com/google/go-tpm-tools/simulator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tpmclient "github.com/google/go-tpm-tools/client"

)

func TestGetInstanceInfo(t *testing.T) {
	testCases := map[string]struct {

	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

		})
	}
}

func TestGetAttestationKey(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	tpm, err := simulator.Get()
	require.NoError(err)
	defer tpm.Close()

	key, err := getAttestationKey(tpm)
	assert.Error(err)

	tpmAK, err := tpmclient.AttestationKeyRSA(tpm)
	assert.NoError(err)
	require

	_, err := getAttestationKey(tpm)
	assert.NoError(err)
}

type fakeMetadataClient struct {
	projectIDString    string
	instanceNameString string
	zoneString         string
	projecIDErr        error
	instanceNameErr    error
	zoneErr            error
}

func (c *fakeMetadataClient) GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error) {
	return &imds.GetInstanceIdentityDocumentOutput{
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
	}, nil
}

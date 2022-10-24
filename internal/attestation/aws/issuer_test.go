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
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/simulator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAttestationKey(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	tpm, err := simulator.Get()
	require.NoError(err)
	defer tpm.Close()

	// create the attestation ket in RSA format
	tpmAk, err := tpmclient.AttestationKeyRSA(tpm)
	assert.NoError(err)
	assert.NotNil(tpmAk)

	// get the cached, already created key
	getAk, err := getAttestationKey(tpm)
	assert.NoError(err)
	assert.NotNil(getAk)

	// if everything worked fine, tpmAk and getAk are the same key
	assert.Equal(tpmAk, getAk)
}

func TestGetInstanceInfo(t *testing.T) {
	testCases := map[string]struct {
		client  fakeMetadataClient
		wantErr bool
	}{
		"test 1": {
			client: fakeMetadataClient{
				"1",
				"test",
				"us-east-2",
				nil,
				nil,
				nil,
			},
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			tpm, err := simulator.Get()
			assert.NoError(err)
			defer tpm.Close()

			instanceInfoFunc := getInstanceInfo(&tc.client)
			assert.NotNil(instanceInfoFunc)

			info, err := instanceInfoFunc(tpm)
			if tc.wantErr {
				assert.Error(err)
				assert.Nil(info)
			} else {
				assert.Nil(err)
				assert.NotNil(info)
			}
		})
	}
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

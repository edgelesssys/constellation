/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"errors"
	"testing"

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
		client  stubMetadataAPI
		wantErr bool
	}{
		"invalid region": {
			client: stubMetadataAPI{
				instanceDoc: imds.InstanceIdentityDocument{
					Region: "invalid-region",
				},
				instanceErr: errors.New("failed"),
			},
			wantErr: true,
		},
		"valid region": {
			client: stubMetadataAPI{
				instanceDoc: imds.InstanceIdentityDocument{
					Region: "us-east-2",
				},
			},
		},
		"invalid imageID": {
			client: stubMetadataAPI{
				instanceDoc: imds.InstanceIdentityDocument{
					ImageID: "ami-fail",
				},
				instanceErr: errors.New("failed"),
			},
			wantErr: true,
		},
		"valid imageID": {
			client: stubMetadataAPI{
				instanceDoc: imds.InstanceIdentityDocument{
					ImageID: "ami-09e7c7f5617a47830",
				},
			},
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

type stubMetadataAPI struct {
	instanceDoc imds.InstanceIdentityDocument
	instanceErr error
}

func (c *stubMetadataAPI) GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error) {
	output := &imds.InstanceIdentityDocument{}

	return &imds.GetInstanceIdentityDocumentOutput{
		InstanceIdentityDocument: *output,
		ResultMetadata:           middleware.Metadata{},
	}, c.instanceErr
}

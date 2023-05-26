/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package nitrotpm

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/stretchr/testify/assert"
)

func TestGeTrustedKey(t *testing.T) {
	testCases := map[string]struct {
		akPub   []byte
		info    []byte
		wantErr bool
	}{
		"nul byte docs": {
			akPub:   []byte{0x00, 0x00, 0x00, 0x00},
			info:    []byte{0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		"nil": {
			akPub:   nil,
			info:    nil,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			out, err := getTrustedKey(
				context.Background(),
				vtpm.AttestationDocument{
					Attestation: &attest.Attestation{
						AkPub: tc.akPub,
					},
					InstanceInfo: tc.info,
				},
				nil,
			)

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
	idDocNoTPM := imds.InstanceIdentityDocument{
		ImageID: "ami-tpm-disabled",
	}
	userDataNoTPM, _ := json.Marshal(idDocNoTPM)
	attDocNoTPM := vtpm.AttestationDocument{
		InstanceInfo: userDataNoTPM,
	}

	idDocTPM := imds.InstanceIdentityDocument{
		ImageID: "ami-tpm-enabled",
	}
	userDataTPM, _ := json.Marshal(idDocTPM)
	attDocTPM := vtpm.AttestationDocument{
		InstanceInfo: userDataTPM,
	}

	testCases := map[string]struct {
		attDoc  vtpm.AttestationDocument
		awsAPI  awsMetadataAPI
		wantErr bool
	}{
		"ami with tpm": {
			attDoc: attDocNoTPM,
			awsAPI: &stubDescribeAPI{describeImagesTPMSupport: "v2.0"},
		},
		"ami without tpm": {
			attDoc:  attDocTPM,
			awsAPI:  &stubDescribeAPI{describeImagesTPMSupport: "v1.0"},
			wantErr: true,
		},
		"ami undefined": {
			attDoc:  vtpm.AttestationDocument{},
			awsAPI:  &stubDescribeAPI{describeImagesErr: errors.New("failed")},
			wantErr: true,
		},
		"invalid json instanceIdentityDocument": {
			attDoc: vtpm.AttestationDocument{
				UserData: []byte("{invalid}"),
			},
			awsAPI:  &stubDescribeAPI{describeImagesErr: errors.New("failed")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			v := Validator{
				getDescribeClient: func(context.Context, string) (awsMetadataAPI, error) {
					return tc.awsAPI, nil
				},
			}

			err := v.tpmEnabled(tc.attDoc, nil)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.Nil(err)
			}
		})
	}
}

type stubDescribeAPI struct {
	describeImagesErr        error
	describeImagesTPMSupport string
}

func (a *stubDescribeAPI) DescribeImages(
	_ context.Context, _ *ec2.DescribeImagesInput, _ ...func(*ec2.Options),
) (*ec2.DescribeImagesOutput, error) {
	output := &ec2.DescribeImagesOutput{
		Images: []types.Image{
			{TpmSupport: types.TpmSupportValues(a.describeImagesTPMSupport)},
		},
	}

	return output, a.describeImagesErr
}

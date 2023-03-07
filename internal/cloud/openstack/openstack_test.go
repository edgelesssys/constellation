/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"context"
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
)

func TestSelf(t *testing.T) {
	someErr := fmt.Errorf("failed")

	testCases := map[string]struct {
		imds    imdsAPI
		want    metadata.InstanceMetadata
		wantErr bool
	}{
		"success": {
			imds: &stubIMDSClient{
				nameResult:       "name",
				providerIDResult: "providerID",
				roleResult:       role.ControlPlane,
				vpcIPResult:      "192.0.2.1",
			},
			want: metadata.InstanceMetadata{
				Name:       "name",
				ProviderID: "providerID",
				Role:       role.ControlPlane,
				VPCIP:      "192.0.2.1",
			},
		},
		"fail to get name": {
			imds: &stubIMDSClient{
				nameErr:          someErr,
				providerIDResult: "providerID",
				roleResult:       role.ControlPlane,
				vpcIPResult:      "192.0.2.1",
			},
			wantErr: true,
		},
		"fail to get provider ID": {
			imds: &stubIMDSClient{
				nameResult:    "name",
				providerIDErr: someErr,
				roleResult:    role.ControlPlane,
				vpcIPResult:   "192.0.2.1",
			},
			wantErr: true,
		},
		"fail to get role": {
			imds: &stubIMDSClient{
				nameResult:       "name",
				providerIDResult: "providerID",
				roleErr:          someErr,
				vpcIPResult:      "192.0.2.1",
			},
			wantErr: true,
		},
		"fail to get VPC IP": {
			imds: &stubIMDSClient{
				nameResult:       "name",
				providerIDResult: "providerID",
				roleResult:       role.ControlPlane,
				vpcIPErr:         someErr,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &Cloud{imds: tc.imds}

			got, err := c.Self(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, got)
			}
		})
	}
}

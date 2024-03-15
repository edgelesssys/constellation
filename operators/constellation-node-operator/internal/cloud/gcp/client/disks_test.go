/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskSourceToDiskReq(t *testing.T) {
	testCases := map[string]struct {
		diskSource  string
		wantRequest *computepb.GetDiskRequest
		wantErr     bool
	}{
		"valid request": {
			diskSource: "https://www.googleapis.com/compute/v1/projects/project/zones/zone/disks/disk",
			wantRequest: &computepb.GetDiskRequest{
				Disk:    "disk",
				Project: "project",
				Zone:    "zone",
			},
		},
		"invalid host": {
			diskSource: "https://hostname/compute/v1/projects/project/zones/zone/disks/disk",
			wantErr:    true,
		},
		"invalid scheme": {
			diskSource: "invalid://www.googleapis.com/compute/v1/projects/project/zones/zone/disks/disk",
			wantErr:    true,
		},
		"url dots in regex are escaped": {
			diskSource: "https://wwwAgoogleapisAcom/compute/v1/projects/project/zones/zone/disks/disk",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			request, err := diskSourceToDiskReq(tc.diskSource)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantRequest, request)
		})
	}
}

func TestURINormalize(t *testing.T) {
	testCases := map[string]struct {
		imageURI       string
		wantNormalized string
	}{
		"URI with scheme and host": {
			imageURI:       "https://www.googleapis.com/compute/v1/projects/project/global/images/image",
			wantNormalized: "projects/project/global/images/image",
		},
		"normalized": {
			imageURI:       "projects/project/global/images/image",
			wantNormalized: "projects/project/global/images/image",
		},
		"url dots in regex are escaped": {
			imageURI:       "https://wwwAgoogleapisAcom/compute/v1/projects/project/global/images/image",
			wantNormalized: "https://wwwAgoogleapisAcom/compute/v1/projects/project/global/images/image",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			normalized := uriNormalize(tc.imageURI)
			assert.Equal(tc.wantNormalized, normalized)
		})
	}
}

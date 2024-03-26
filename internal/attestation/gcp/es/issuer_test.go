/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package es

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGCEInstanceInfo(t *testing.T) {
	testCases := map[string]struct {
		client  fakeMetadataClient
		wantErr bool
	}{
		"success": {
			client: fakeMetadataClient{
				projectIDString:    "projectID",
				instanceNameString: "instanceName",
				zoneString:         "zone",
			},
			wantErr: false,
		},
		"projectID error": {
			client: fakeMetadataClient{
				projectIDString:    "projectID",
				instanceNameString: "instanceName",
				zoneString:         "zone",
				projecIDErr:        errors.New("error"),
			},
			wantErr: true,
		},
		"instanceName error": {
			client: fakeMetadataClient{
				projectIDString:    "projectID",
				instanceNameString: "instanceName",
				zoneString:         "zone",
				instanceNameErr:    errors.New("error"),
			},
			wantErr: true,
		},
		"zone error": {
			client: fakeMetadataClient{
				projectIDString:    "projectID",
				instanceNameString: "instanceName",
				zoneString:         "zone",
				zoneErr:            errors.New("error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			var tpm io.ReadWriteCloser

			out, err := gcp.GCEInstanceInfo(tc.client)(context.Background(), tpm, nil)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				var info attest.GCEInstanceInfo
				require.NoError(json.Unmarshal(out, &info))
				assert.Equal(tc.client.projectIDString, info.ProjectId)
				assert.Equal(tc.client.instanceNameString, info.InstanceName)
				assert.Equal(tc.client.zoneString, info.Zone)
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

func (c fakeMetadataClient) ProjectID() (string, error) {
	return c.projectIDString, c.projecIDErr
}

func (c fakeMetadataClient) InstanceName() (string, error) {
	return c.instanceNameString, c.instanceNameErr
}

func (c fakeMetadataClient) Zone() (string, error) {
	return c.zoneString, c.zoneErr
}

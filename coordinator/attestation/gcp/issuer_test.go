package gcp

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGCEInstanceInfo(t *testing.T) {
	testCases := map[string]struct {
		client      fakeMetadataClient
		errExpected bool
	}{
		"success": {
			client: fakeMetadataClient{
				projectIDString:    "projectID",
				instanceNameString: "instanceName",
				zoneString:         "zone",
			},
			errExpected: false,
		},
		"projectID error": {
			client: fakeMetadataClient{
				projectIDString:    "projectID",
				instanceNameString: "instanceName",
				zoneString:         "zone",
				projecIdErr:        errors.New("error"),
			},
			errExpected: true,
		},
		"instanceName error": {
			client: fakeMetadataClient{
				projectIDString:    "projectID",
				instanceNameString: "instanceName",
				zoneString:         "zone",
				instanceNameErr:    errors.New("error"),
			},
			errExpected: true,
		},
		"zone error": {
			client: fakeMetadataClient{
				projectIDString:    "projectID",
				instanceNameString: "instanceName",
				zoneString:         "zone",
				zoneErr:            errors.New("error"),
			},
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			var tpm io.ReadWriteCloser

			out, err := getGCEInstanceInfo(tc.client)(tpm)
			if tc.errExpected {
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
	projecIdErr        error
	instanceNameErr    error
	zoneErr            error
}

func (c fakeMetadataClient) projectID() (string, error) {
	return c.projectIDString, c.projecIdErr
}

func (c fakeMetadataClient) instanceName() (string, error) {
	return c.instanceNameString, c.instanceNameErr
}

func (c fakeMetadataClient) zone() (string, error) {
	return c.zoneString, c.zoneErr
}

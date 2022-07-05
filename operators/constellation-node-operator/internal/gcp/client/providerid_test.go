package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitProviderID(t *testing.T) {
	testCases := map[string]struct {
		providerID    string
		wantProjectID string
		wantZone      string
		wantInstance  string
		wantErr       bool
	}{
		"simple id": {
			providerID:    "gce://someProject/someZone/someInstance",
			wantProjectID: "someProject",
			wantZone:      "someZone",
			wantInstance:  "someInstance",
		},
		"incomplete id": {
			providerID: "gce://someProject/someZone",
			wantErr:    true,
		},
		"wrong provider": {
			providerID: "azure://someProject/someZone/someInstance",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			projectID, zone, instance, err := splitProviderID(tc.providerID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantProjectID, projectID)
			assert.Equal(tc.wantZone, zone)
			assert.Equal(tc.wantInstance, instance)
		})
	}
}

func TestJoinProviderID(t *testing.T) {
	testCases := map[string]struct {
		projectID      string
		zone           string
		instance       string
		wantProviderID string
	}{
		"simple id": {
			projectID:      "someProject",
			zone:           "someZone",
			instance:       "someInstance",
			wantProviderID: "gce://someProject/someZone/someInstance",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			providerID := joinProviderID(tc.projectID, tc.zone, tc.instance)

			assert.Equal(tc.wantProviderID, providerID)
		})
	}
}

func TestJoinnstanceID(t *testing.T) {
	testCases := map[string]struct {
		zone           string
		instanceName   string
		wantInstanceID string
	}{
		"simple id": {
			zone:           "someZone",
			instanceName:   "someInstance",
			wantInstanceID: "zones/someZone/instances/someInstance",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			providerID := joinInstanceID(tc.zone, tc.instanceName)

			assert.Equal(tc.wantInstanceID, providerID)
		})
	}
}

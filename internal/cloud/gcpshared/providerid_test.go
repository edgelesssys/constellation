/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcpshared

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

			projectID, zone, instance, err := SplitProviderID(tc.providerID)

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

			providerID := JoinProviderID(tc.projectID, tc.zone, tc.instance)

			assert.Equal(tc.wantProviderID, providerID)
		})
	}
}

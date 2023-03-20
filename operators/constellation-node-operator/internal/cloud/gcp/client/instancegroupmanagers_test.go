/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitInstanceGroupID(t *testing.T) {
	testCases := map[string]struct {
		instanceGroupID string

		wantProject       string
		wantZone          string
		wantInstanceGroup string
		wantErr           bool
	}{
		"valid request": {
			instanceGroupID:   "projects/project/zones/zone/instanceGroupManagers/instanceGroup",
			wantProject:       "project",
			wantZone:          "zone",
			wantInstanceGroup: "instanceGroup",
		},
		"wrong format": {
			instanceGroupID: "wrong-format",
			wantErr:         true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			gotProject, gotZone, gotInstanceGroup, err := splitInstanceGroupID(tc.instanceGroupID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantProject, gotProject)
			assert.Equal(tc.wantZone, gotZone)
			assert.Equal(tc.wantInstanceGroup, gotInstanceGroup)
		})
	}
}

func TestGenerateInstanceName(t *testing.T) {
	assert := assert.New(t)
	baseInstanceName := "base"
	gotInstanceName := generateInstanceName(baseInstanceName, &stubRng{result: 0})
	assert.Equal("base-aaaa", gotInstanceName)
}

func TestGenerateInstanceNameRandomTest(t *testing.T) {
	assert := assert.New(t)
	instanceNameRegexp := regexp.MustCompile(`^base-[0-9a-z]{4}$`)
	baseInstanceName := "base"
	random := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	gotInstanceName := generateInstanceName(baseInstanceName, random)
	assert.Regexp(instanceNameRegexp, gotInstanceName)
}

type stubRng struct {
	result int
}

func (r *stubRng) Intn(_ int) int {
	return r.result
}

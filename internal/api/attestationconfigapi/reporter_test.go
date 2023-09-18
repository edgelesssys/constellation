/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/
package attestationconfigapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterDatesWithinTime(t *testing.T) {
	dates := []string{
		"2022-01-01-00-00",
		"2022-01-02-00-00",
		"2022-01-03-00-00",
		"2022-01-04-00-00",
		"2022-01-05-00-00",
		"2022-01-06-00-00",
		"2022-01-07-00-00",
		"2022-01-08-00-00",
	}
	now := time.Date(2022, 1, 9, 0, 0, 0, 0, time.UTC)
	testCases := map[string]struct {
		timeFrame     time.Duration
		expectedDates []string
	}{
		"all dates within 3 days": {
			timeFrame:     time.Hour * 24 * 3,
			expectedDates: []string{"2022-01-06-00-00", "2022-01-07-00-00", "2022-01-08-00-00"},
		},
		"no dates within time frame": {
			timeFrame:     time.Hour,
			expectedDates: nil,
		},
		"some dates within time frame": {
			timeFrame:     time.Hour * 24 * 4,
			expectedDates: []string{"2022-01-05-00-00", "2022-01-06-00-00", "2022-01-07-00-00", "2022-01-08-00-00"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			filteredDates := filterDatesWithinTime(dates, now, tc.timeFrame)
			assert.Equal(t, tc.expectedDates, filteredDates)
		})
	}
}

func TestIsInputNewerThanLatestAPI(t *testing.T) {
	newTestCfg := func() AzureSEVSNPVersion {
		return AzureSEVSNPVersion{
			Microcode:  93,
			TEE:        0,
			SNP:        6,
			Bootloader: 2,
		}
	}

	testCases := map[string]struct {
		latest AzureSEVSNPVersion
		input  AzureSEVSNPVersion
		expect bool
		errMsg string
	}{
		"input is older than latest": {
			input: func(c AzureSEVSNPVersion) AzureSEVSNPVersion {
				c.Microcode--
				return c
			}(newTestCfg()),
			latest: newTestCfg(),
			expect: false,
			errMsg: "input Microcode version: 92 is older than latest API version: 93",
		},
		"input has greater and smaller version field than latest": {
			input: func(c AzureSEVSNPVersion) AzureSEVSNPVersion {
				c.Microcode++
				c.Bootloader--
				return c
			}(newTestCfg()),
			latest: newTestCfg(),
			expect: false,
			errMsg: "input Bootloader version: 1 is older than latest API version: 2",
		},
		"input is newer than latest": {
			input: func(c AzureSEVSNPVersion) AzureSEVSNPVersion {
				c.TEE++
				return c
			}(newTestCfg()),
			latest: newTestCfg(),
			expect: true,
		},
		"input is equal to latest": {
			input:  newTestCfg(),
			latest: newTestCfg(),
			expect: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			isNewer, err := isInputNewerThanLatestAPI(tc.input, tc.latest)
			assert := assert.New(t)
			if tc.errMsg != "" {
				assert.EqualError(err, tc.errMsg)
			} else {
				assert.NoError(err)
				assert.Equal(tc.expect, isNewer)
			}
		})
	}
}

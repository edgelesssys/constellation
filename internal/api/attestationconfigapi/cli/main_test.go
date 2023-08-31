/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/stretchr/testify/assert"
)

func TestIsInputNewerThanLatestAPI(t *testing.T) {
	newTestCfg := func() attestationconfigapi.AzureSEVSNPVersion {
		return attestationconfigapi.AzureSEVSNPVersion{
			Microcode:  93,
			TEE:        0,
			SNP:        6,
			Bootloader: 2,
		}
	}

	testCases := map[string]struct {
		latest attestationconfigapi.AzureSEVSNPVersion
		input  attestationconfigapi.AzureSEVSNPVersion
		expect bool
		errMsg string
	}{
		"input is older than latest": {
			input: func(c attestationconfigapi.AzureSEVSNPVersion) attestationconfigapi.AzureSEVSNPVersion {
				c.Microcode--
				return c
			}(newTestCfg()),
			latest: newTestCfg(),
			expect: false,
			errMsg: "input Microcode version: 92 is older than latest API version: 93",
		},
		"input has greater and smaller version field than latest": {
			input: func(c attestationconfigapi.AzureSEVSNPVersion) attestationconfigapi.AzureSEVSNPVersion {
				c.Microcode++
				c.Bootloader--
				return c
			}(newTestCfg()),
			latest: newTestCfg(),
			expect: false,
			errMsg: "input Bootloader version: 1 is older than latest API version: 2",
		},
		"input is newer than latest": {
			input: func(c attestationconfigapi.AzureSEVSNPVersion) attestationconfigapi.AzureSEVSNPVersion {
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

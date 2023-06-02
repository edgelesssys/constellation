/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/stretchr/testify/assert"
)

var testCfg = attestationconfig.AzureSEVSNPVersion{
	Microcode:  93,
	TEE:        0,
	SNP:        6,
	Bootloader: 2,
}

func TestIsInputNewerThanLatestAPI(t *testing.T) {
	testCases := map[string]struct {
		latest attestationconfig.AzureSEVSNPVersion
		input  attestationconfig.AzureSEVSNPVersion
		expect bool
		errMsg string
	}{
		"input is older than latest": {
			input: func(c attestationconfig.AzureSEVSNPVersion) attestationconfig.AzureSEVSNPVersion {
				c.Microcode--
				return c
			}(testCfg),
			latest: testCfg,
			expect: false,
			errMsg: "input Microcode version: 92 is older than latest API version: 93",
		},
		"input has greater and smaller version field than latest": {
			input: func(c attestationconfig.AzureSEVSNPVersion) attestationconfig.AzureSEVSNPVersion {
				c.Microcode++
				c.Bootloader--
				return c
			}(testCfg),
			latest: testCfg,
			expect: false,
			errMsg: "input Bootloader version: 1 is older than latest API version: 2",
		},
		"input is newer than latest": {
			input: func(c attestationconfig.AzureSEVSNPVersion) attestationconfig.AzureSEVSNPVersion {
				c.TEE++
				return c
			}(testCfg),
			latest: testCfg,
			expect: true,
		},
		"input is equal to latest": {
			input:  testCfg,
			latest: testCfg,
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

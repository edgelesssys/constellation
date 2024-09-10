/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/
package client

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/api/attestationconfig"
	"github.com/stretchr/testify/assert"
)

func TestIsInputNewerThanOtherSEVSNPVersion(t *testing.T) {
	newTestCfg := func() attestationconfig.SEVSNPVersion {
		return attestationconfig.SEVSNPVersion{
			Microcode:  93,
			TEE:        0,
			SNP:        6,
			Bootloader: 2,
		}
	}

	testCases := map[string]struct {
		latest attestationconfig.SEVSNPVersion
		input  attestationconfig.SEVSNPVersion
		expect bool
	}{
		"input is older than latest": {
			input: func(c attestationconfig.SEVSNPVersion) attestationconfig.SEVSNPVersion {
				c.Microcode--
				return c
			}(newTestCfg()),
			latest: newTestCfg(),
			expect: false,
		},
		"input has greater and smaller version field than latest": {
			input: func(c attestationconfig.SEVSNPVersion) attestationconfig.SEVSNPVersion {
				c.Microcode++
				c.Bootloader--
				return c
			}(newTestCfg()),
			latest: newTestCfg(),
			expect: false,
		},
		"input is newer than latest": {
			input: func(c attestationconfig.SEVSNPVersion) attestationconfig.SEVSNPVersion {
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
			isNewer := isInputNewerThanOtherSEVSNPVersion(tc.input, tc.latest)
			assert.Equal(t, tc.expect, isNewer)
		})
	}
}

func TestIsInputNewerThanOtherTDXVersion(t *testing.T) {
	newTestVersion := func() attestationconfig.TDXVersion {
		return attestationconfig.TDXVersion{
			QESVN:      1,
			PCESVN:     2,
			TEETCBSVN:  [16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			QEVendorID: [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			XFAM:       [8]byte{0, 1, 2, 3, 4, 5, 6, 7},
		}
	}

	testCases := map[string]struct {
		latest attestationconfig.TDXVersion
		input  attestationconfig.TDXVersion
		expect bool
	}{
		"input is older than latest": {
			input: func(c attestationconfig.TDXVersion) attestationconfig.TDXVersion {
				c.QESVN--
				return c
			}(newTestVersion()),
			latest: newTestVersion(),
			expect: false,
		},
		"input has greater and smaller version field than latest": {
			input: func(c attestationconfig.TDXVersion) attestationconfig.TDXVersion {
				c.QESVN++
				c.PCESVN--
				return c
			}(newTestVersion()),
			latest: newTestVersion(),
			expect: false,
		},
		"input is newer than latest": {
			input: func(c attestationconfig.TDXVersion) attestationconfig.TDXVersion {
				c.QESVN++
				return c
			}(newTestVersion()),
			latest: newTestVersion(),
			expect: true,
		},
		"input is equal to latest": {
			input:  newTestVersion(),
			latest: newTestVersion(),
			expect: false,
		},
		"tee tcb svn is newer": {
			input: func(c attestationconfig.TDXVersion) attestationconfig.TDXVersion {
				c.TEETCBSVN[4]++
				return c
			}(newTestVersion()),
			latest: newTestVersion(),
			expect: true,
		},
		"xfam is different": {
			input: func(c attestationconfig.TDXVersion) attestationconfig.TDXVersion {
				c.XFAM[3]++
				return c
			}(newTestVersion()),
			latest: newTestVersion(),
			expect: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			isNewer := isInputNewerThanOtherTDXVersion(tc.input, tc.latest)
			assert.Equal(t, tc.expect, isNewer)
		})
	}
}

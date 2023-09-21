/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measure

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPredictPCR9(t *testing.T) {
	assert := assert.New(t)

	sim := NewDefaultSimulator()

	cmdline := []byte("console=tty0\x00")
	initrdDigest := [32]byte{}

	out := bytes.NewBuffer(nil)
	assert.NoError(DescribeLinuxLoad2(out, cmdline, initrdDigest))
	assert.Equal("Linux LOAD_FILE2 protocol:\n"+
		"  cmdline: \"console=tty0\\x00\"\n"+
		"  initrd (digest 0000000000000000000000000000000000000000000000000000000000000000)\n",
		out.String())

	assert.NoError(PredictPCR9(sim, cmdline, initrdDigest))
	assert.Equal(PCR256{
		0xeb, 0x4f, 0x7b, 0xca, 0x86, 0x58, 0x07, 0xd3,
		0x16, 0x3b, 0x95, 0x17, 0x4d, 0x6e, 0x66, 0xcf,
		0xc7, 0x4a, 0xcf, 0x8b, 0x93, 0x0a, 0x55, 0x3e,
		0x95, 0xec, 0x94, 0x66, 0x2c, 0xb6, 0xfa, 0xcd,
	}, sim.Bank[9])
}

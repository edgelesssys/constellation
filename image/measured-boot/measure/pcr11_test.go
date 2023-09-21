/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measure

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/v2/image/measured-boot/pesection"
	"github.com/stretchr/testify/assert"
)

func TestPredictPCR11(t *testing.T) {
	assert := assert.New(t)

	sim := NewDefaultSimulator()

	peSections := []pesection.PESection{
		{
			Name:   ".text",
			Size:   100,
			Digest: [32]byte{},
		},
		{
			Name:    ".linux",
			Size:    100,
			Digest:  [32]byte{},
			Measure: true,
		},
		{
			Name: ".initrd",
			Digest: [32]byte{
				1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1,
			},
			Measure: true,
		},
	}

	out := bytes.NewBuffer(nil)
	assert.NoError(DescribeUKISections(out, peSections))
	assert.Equal("UKI sections:\n"+
		"  Section  1 - .text  :\tnot measured\n"+
		"  Section  2 - .linux  (       100 bytes):\t0da293e37ad5511c59be47993769aacb91b243f7d010288e118dc90e95aaef5a, 0000000000000000000000000000000000000000000000000000000000000000\n"+
		"  Section  3 - .initrd (         0 bytes):\t15ee37e75f1e8d42080e91fdbbd2560780918c81fe3687ae6d15c472bbdaac75, 0101010101010101010101010101010101010101010101010101010101010101\n",
		out.String())

	assert.NoError(PredictPCR11(sim, peSections))
	assert.Equal(PCR256{
		0x9d, 0xfe, 0x39, 0x9f, 0xcd, 0x44, 0x32, 0x63,
		0x9f, 0x0e, 0x20, 0xf4, 0x9d, 0xf8, 0x23, 0xaa,
		0x66, 0xb0, 0x95, 0xf0, 0x66, 0x4f, 0x0a, 0x4b,
		0x9f, 0xbd, 0xc1, 0x1e, 0xa6, 0x46, 0x83, 0xe2,
	}, sim.Bank[11])
}

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

func TestPredictPCR4(t *testing.T) {
	assert := assert.New(t)

	sim := NewDefaultSimulator()

	bootstages := []EFIBootStage{
		{
			Name:   "stage0",
			Digest: [32]byte{},
		},
		{
			Name: "stage1",
			Digest: [32]byte{
				1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1,
			},
		},
	}

	out := bytes.NewBuffer(nil)
	assert.NoError(DescribeBootStages(out, bootstages))
	assert.Equal("EFI Boot Stages:\n"+
		"  Stage 1 - stage0:\t0000000000000000000000000000000000000000000000000000000000000000\n"+
		"  Stage 2 - stage1:\t0101010101010101010101010101010101010101010101010101010101010101\n",
		out.String())

	assert.NoError(PredictPCR4(sim, bootstages))
	assert.Equal(PCR256{
		0x22, 0x11, 0x6d, 0xee, 0x86, 0x1a, 0xa6, 0xb4,
		0x42, 0x42, 0xac, 0x46, 0x9e, 0xab, 0x24, 0xce,
		0xad, 0x34, 0x4d, 0x52, 0xc7, 0x71, 0x31, 0xf5,
		0x4a, 0xc1, 0xca, 0xc9, 0xd6, 0xa2, 0x40, 0x8e,
	}, sim.Bank[4])
}

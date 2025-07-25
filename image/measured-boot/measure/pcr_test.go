/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package measure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtendPCR(t *testing.T) {
	assert := assert.New(t)

	sim := NewDefaultSimulator()
	assert.Equal(ZeroPCR256(), sim.Bank[4])

	assert.NoError(sim.ExtendPCR(4, EVSeparatorPCR256(), []byte{0x00, 0x00, 0x00, 0x00}, "EV_SEPARATOR"))
	assert.Equal(PCR256{
		0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
		0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
		0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
		0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
	}, sim.Bank[4])

	assert.NoError(sim.ExtendPCR(4, EVEFIActionPCR256(), nil, "EV_EFI_ACTION: Calling EFI Application from Boot Option"))
	assert.Equal(PCR256{
		0xdd, 0x50, 0xc8, 0xda, 0x0f, 0x89, 0x9f, 0x65,
		0x5b, 0x43, 0x05, 0xd2, 0x43, 0x86, 0x63, 0xc1,
		0xb3, 0xda, 0x6d, 0x19, 0x22, 0xa0, 0xc8, 0x22,
		0x65, 0x33, 0xac, 0x41, 0x7a, 0xbc, 0xd5, 0x23,
	}, sim.Bank[4])

	assert.Equal([]Event{
		{
			PCRIndex: 0x4, Digest: Digest256(EVSeparatorPCR256()),
			Data:        []uint8{0x0, 0x0, 0x0, 0x0},
			Description: "EV_SEPARATOR",
		},
		{
			PCRIndex: 0x4, Digest: Digest256(EVEFIActionPCR256()),
			Data:        []uint8(nil),
			Description: "EV_EFI_ACTION: Calling EFI Application from Boot Option",
		},
	}, sim.EventLog.Events)
}

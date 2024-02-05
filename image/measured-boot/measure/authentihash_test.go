/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measure

import (
	"bytes"
	"crypto"
	"testing"

	"github.com/edgelesssys/constellation/v2/image/measured-boot/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestPeSectionReader(t *testing.T) {
	assert := assert.New(t)

	peReader := bytes.NewReader(fixtures.UKI())
	digest, err := Authentihash(peReader, crypto.SHA256)
	assert.NoError(err)
	assert.Equal(
		[]byte{
			0xd3, 0x43, 0xbe, 0x62, 0x65, 0xeb, 0x3e, 0x23,
			0xf7, 0x8b, 0x0a, 0xe0, 0x96, 0xbf, 0xf3, 0x34,
			0xe3, 0x7a, 0x76, 0x0a, 0xe8, 0x30, 0x73, 0x62,
			0x83, 0xf9, 0xb0, 0x26, 0x8e, 0xce, 0xdc, 0xf2,
		},
		digest,
	)
}

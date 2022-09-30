/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sigstore

import (
	"testing"

	"github.com/sigstore/rekor/pkg/generated/models"
	hashedrekord "github.com/sigstore/rekor/pkg/types/hashedrekord/v0.0.1"
	"github.com/stretchr/testify/assert"
)

func TestIsEntrySignedBy(t *testing.T) {
	assert := assert.New(t)

	entry := &hashedrekord.V001Entry{
		HashedRekordObj: models.HashedrekordV001Schema{
			Signature: &models.HashedrekordV001SchemaSignature{
				PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
					Content: []byte("my key"),
				},
			},
		},
	}

	assert.True(IsEntrySignedBy(entry, "bXkga2V5")) // "my key" in base64
}

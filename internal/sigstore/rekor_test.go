/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sigstore

import (
	"context"
	"testing"

	"github.com/sigstore/rekor/pkg/generated/models"
	hashedrekord "github.com/sigstore/rekor/pkg/types/hashedrekord/v0.0.1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewRekor(t *testing.T) {
	assert := assert.New(t)
	rekor, err := NewRekor()
	assert.NoError(err)
	assert.NotNil(rekor)
}

func TestRekor_SearchByHash(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	rekor, err := NewRekor()
	require.NoError(err)

	uuids, err := rekor.SearchByHash(context.Background(), "40e137b9b9b8204d672642fd1e181c6d5ccb50cfc5cc7fcbb06a8c2c78f44afe")
	assert.NoError(err)
	assert.Empty(uuids)
}

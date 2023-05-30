/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package sigstore

import (
	"testing"

	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/stretchr/testify/assert"
)

func TestSignSignature(t *testing.T) {
	assert := assert.New(t)
	priv, err := cosign.GenerateKeyPair(nil)
	if err != nil {
		t.Fatalf("cosign.GeneratePrivateKey() failed: %v", err)
	}

	content := []byte(`["2021-01-01-01-01.json"]\n`)
	signature, err := SignContent(priv.Password(), priv.PrivateBytes, content)
	assert.NoError(err)
	assert.NoError(VerifySignature(content, signature, priv.PublicBytes))
}

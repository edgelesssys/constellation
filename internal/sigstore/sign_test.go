/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package sigstore_test

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/stretchr/testify/assert"
)

func TestSignSignature(t *testing.T) {
	assert := assert.New(t)
	priv, err := cosign.GenerateKeyPair(nil)
	if err != nil {
		t.Fatalf("cosign.GeneratePrivateKey() failed: %v", err)
	}
	testCases := map[string]struct {
		content    []byte
		password   []byte
		privateKey []byte
		wantErr    bool
	}{
		"sign content with matching password and private key": {
			content:    []byte(`["2021-01-01-01-01.json"]\n`),
			password:   priv.Password(),
			privateKey: priv.PrivateBytes,
		},
		"fail with wrong password": {
			content:    []byte(`["2021-01-01-01-01.json"]\n`),
			password:   []byte("wrong"),
			privateKey: priv.PrivateBytes,
			wantErr:    true,
		},
		"fail with wrong private key": {
			content:    []byte(`["2021-01-01-01-01.json"]\n`),
			password:   priv.Password(),
			privateKey: []byte("wrong"),
			wantErr:    true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			signature, err := sigstore.SignContent(tc.password, tc.privateKey, tc.content)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NoError(sigstore.VerifySignature(tc.content, signature, priv.PublicBytes))
			}
		})
	}
}

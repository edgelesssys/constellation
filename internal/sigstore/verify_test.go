/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sigstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifySignature(t *testing.T) {
	testCases := map[string]struct {
		content   []byte
		signature []byte
		publicKey []byte
		wantErr   bool
	}{
		"good verification": {
			content:   []byte("This is some content to be signed!\n"),
			signature: []byte("MEUCIQDzMN3yaiO9sxLGAaSA9YD8rLwzvOaZKWa/bzkcjImUFAIgXLLGzClYUd1dGbuEiY3O/g/eiwQYlyxqLQalxjFmz+8="),
			publicKey: []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAElWUhon39eAqzEC+/GP03oY4/MQg+
gCDlEzkuOCybCHf+q766bve799L7Y5y5oRsHY1MrUCUwYF/tL7Sg7EYMsA==
-----END PUBLIC KEY-----`),
		},
		"mismatching content": {
			content:   []byte("This is some completely different content!\n"),
			signature: []byte("MEUCIQDzMN3yaiO9sxLGAaSA9YD8rLwzvOaZKWa/bzkcjImUFAIgXLLGzClYUd1dGbuEiY3O/g/eiwQYlyxqLQalxjFmz+8="),
			publicKey: []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAElWUhon39eAqzEC+/GP03oY4/MQg+
gCDlEzkuOCybCHf+q766bve799L7Y5y5oRsHY1MrUCUwYF/tL7Sg7EYMsA==
-----END PUBLIC KEY-----`),
			wantErr: true,
		},
		"broken public key": {
			content:   []byte("This is some content to be signed!\n"),
			signature: []byte("MEUCIQDzMN3yaiO9sxLGAaSA9YD8rLwzvOaZKWa/bzkcjImUFAIgXLLGzClYUd1dGbuEiY3O/g/eiwQYlyxqLQalxjFmz+8="),
			publicKey: []byte(`-----BEGIN PUBLIC KEY-----
			MFkwEwYHKoZIthisIsNotAValidPublicAtAllUhon39eAqzEC+/GP03oY4/MQg+
			gCDlEzkuOCybCHf+q766bve799L7Y5y5oRsHY1MrUCUwYF/tL7Sg7EYMsA==
			-----END PUBLIC KEY-----`),
			wantErr: true,
		},
		"valid content and sig, but mismatching public key": {
			content:   []byte("This is some content to be signed!\n"),
			signature: []byte("MEUCIQDzMN3yaiO9sxLGAaSA9YD8rLwzvOaZKWa/bzkcjImUFAIgXLLGzClYUd1dGbuEiY3O/g/eiwQYlyxqLQalxjFmz+8="),
			publicKey: []byte(`-----BEGIN PUBLIC KEY-----
		MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFARL653CK4xicoxqwr4M9A2A/3hz
		hQaKKRsnjc2LITnxKYmQ4CYqTOAMfZ3agxpW/ndillUox4eDYcidZSXvWw==
		-----END PUBLIC KEY-----`),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cosign := CosignVerifier{}
			err := cosign.VerifySignature(tc.content, tc.signature, tc.publicKey)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

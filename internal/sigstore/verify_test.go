/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sigstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCosignVerifier(t *testing.T) {
	testCases := map[string]struct {
		publicKey []byte
		wantErr   bool
	}{
		"success": {
			publicKey: []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAElWUhon39eAqzEC+/GP03oY4/MQg+
gCDlEzkuOCybCHf+q766bve799L7Y5y5oRsHY1MrUCUwYF/tL7Sg7EYMsA==
-----END PUBLIC KEY-----`),
		},
		"broken public key": {
			publicKey: []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIthisIsNotAValidPublicAtAllUhon39eAqzEC+/GP03oY4/MQg+
gCDlEzkuOCybCHf+q766bve799L7Y5y5oRsHY1MrUCUwYF/tL7Sg7EYMsA==
-----END PUBLIC KEY-----`),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			verifier, err := NewCosignVerifier(tc.publicKey)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.NotEqual(verifier, CosignVerifier{})
		})
	}
}

func TestVerifySignature(t *testing.T) {
	testCases := map[string]struct {
		content   []byte
		signature []byte
		publicKey []byte
		wantErr   bool
	}{
		"success": {
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
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cosign, err := NewCosignVerifier(tc.publicKey)
			require.NoError(t, err)
			err = cosign.VerifySignature(tc.content, tc.signature)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestIsBase64(t *testing.T) {
	tests := map[string]struct {
		signature []byte
		wantErr   bool
	}{
		"valid base64": {
			signature: []byte("SGVsbG8gV29ybGQ="),
			wantErr:   false,
		},
		"invalid base64": {
			signature: []byte("not base64"),
			wantErr:   true,
		},
		"empty input": {
			signature: []byte{},
			wantErr:   false,
		},
	}

	for tc, tt := range tests {
		t.Run(tc, func(t *testing.T) {
			err := IsBase64(tt.signature)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

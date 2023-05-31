/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package sigstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignSignature(t *testing.T) {
	assert := assert.New(t)
	// Generated with: cosign generate-key-pair
	privateKey := []byte("-----BEGIN ENCRYPTED COSIGN PRIVATE KEY-----\neyJrZGYiOnsibmFtZSI6InNjcnlwdCIsInBhcmFtcyI6eyJOIjozMjc2OCwiciI6\nOCwicCI6MX0sInNhbHQiOiJlRHVYMWRQMGtIWVRnK0xkbjcxM0tjbFVJaU92eFVX\nVXgvNi9BbitFVk5BPSJ9LCJjaXBoZXIiOnsibmFtZSI6Im5hY2wvc2VjcmV0Ym94\nIiwibm9uY2UiOiJwaWhLL2txNmFXa2hqSVVHR3RVUzhTVkdHTDNIWWp4TCJ9LCJj\naXBoZXJ0ZXh0Ijoidm81SHVWRVFWcUZ2WFlQTTVPaTVaWHM5a255bndZU2dvcyth\nVklIeHcrOGFPamNZNEtvVjVmL3lHRHR0K3BHV2toanJPR1FLOWdBbmtsazFpQ0c5\na2czUXpPQTZsU2JRaHgvZlowRVRZQ0hLeElncEdPRVRyTDlDenZDemhPZXVSOXJ6\nTDcvRjBBVy9vUDVqZXR3dmJMNmQxOEhjck9kWE8yVmYxY2w0YzNLZjVRcnFSZzlN\ndlRxQWFsNXJCNHNpY1JaMVhpUUJjb0YwNHc9PSJ9\n-----END ENCRYPTED COSIGN PRIVATE KEY-----")
	password := []byte("")
	content := []byte(`["2021-01-01-01-01.json"]\n`)
	publicKey := []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----")

	testCases := map[string]struct {
		content    []byte
		password   []byte
		privateKey []byte
		wantErr    bool
	}{
		"sign content with matching password and private key": {
			content:    content,
			password:   password,
			privateKey: privateKey,
		},
		"fail with wrong password": {
			content:    content,
			password:   []byte("wrong"),
			privateKey: privateKey,
			wantErr:    true,
		},
		"fail with wrong private key": {
			content:    content,
			password:   password,
			privateKey: []byte("wrong"),
			wantErr:    true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			signature, err := SignContent(tc.password, tc.privateKey, tc.content)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NoError(VerifySignature(tc.content, signature, publicKey))

			}
		})
	}
}

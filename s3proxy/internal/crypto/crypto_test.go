/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package crypto

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := map[string]struct {
		plaintext []byte
	}{
		"simple": {
			plaintext: []byte("hello, world"),
		},
		"long": {
			plaintext: []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed non risus. Suspendisse lectus tortor, dignissim sit amet, adipiscing nec, ultricies sed, dolor."),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			kek := [32]byte{}
			_, err := rand.Read(kek[:])
			require.NoError(t, err)

			ciphertext, encryptedDEK, err := Encrypt(tt.plaintext, kek)
			require.NoError(t, err)

			assert.NotContains(t, ciphertext, tt.plaintext)

			// Decrypt the ciphertext using the KEK and encrypted DEK
			decrypted, err := Decrypt(ciphertext, encryptedDEK, kek)
			require.NoError(t, err)

			// Verify that the decrypted plaintext matches the original plaintext
			assert.Equal(t, tt.plaintext, decrypted, fmt.Sprintf("expected plaintext %s, got %s", tt.plaintext, decrypted))
		})
	}
}

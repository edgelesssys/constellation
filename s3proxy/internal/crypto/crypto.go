/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package crypto provides encryption and decryption functions for the s3proxy.
It uses AES-256-GCM to encrypt and decrypt data.
*/
package crypto

import (
	"fmt"

	aeadsubtle "github.com/tink-crypto/tink-go/v2/aead/subtle"
	kwpsubtle "github.com/tink-crypto/tink-go/v2/kwp/subtle"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

// Encrypt generates a random key to encrypt a plaintext using AES-256-GCM.
// The generated key is encrypted using the supplied key encryption key (KEK).
// The ciphertext and encrypted data encryption key (DEK) are returned.
func Encrypt(plaintext []byte, kek [32]byte) (ciphertext []byte, encryptedDEK []byte, err error) {
	dek := random.GetRandomBytes(32)
	aesgcm, err := aeadsubtle.NewAESGCMSIV(dek)
	if err != nil {
		return nil, nil, fmt.Errorf("getting aesgcm: %w", err)
	}

	ciphertext, err = aesgcm.Encrypt(plaintext, []byte(""))
	if err != nil {
		return nil, nil, fmt.Errorf("encrypting plaintext: %w", err)
	}

	keywrapper, err := kwpsubtle.NewKWP(kek[:])
	if err != nil {
		return nil, nil, fmt.Errorf("getting kwp: %w", err)
	}

	encryptedDEK, err = keywrapper.Wrap(dek)
	if err != nil {
		return nil, nil, fmt.Errorf("wrapping dek: %w", err)
	}

	return ciphertext, encryptedDEK, nil
}

// Decrypt decrypts a ciphertext using AES-256-GCM.
// The encrypted DEK is decrypted using the supplied KEK.
func Decrypt(ciphertext, encryptedDEK []byte, kek [32]byte) ([]byte, error) {
	keywrapper, err := kwpsubtle.NewKWP(kek[:])
	if err != nil {
		return nil, fmt.Errorf("getting kwp: %w", err)
	}

	dek, err := keywrapper.Unwrap(encryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("unwrapping dek: %w", err)
	}

	aesgcm, err := aeadsubtle.NewAESGCMSIV(dek)
	if err != nil {
		return nil, fmt.Errorf("getting aesgcm: %w", err)
	}

	plaintext, err := aesgcm.Decrypt(ciphertext, []byte(""))
	if err != nil {
		return nil, fmt.Errorf("decrypting ciphertext: %w", err)
	}

	return plaintext, nil
}

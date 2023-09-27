/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package crypto provides encryption and decryption functions for the s3proxy.
It uses AES-256-GCM to encrypt and decrypt data.
A new nonce is generated for each encryption operation.
*/
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// Encrypt takes a 32 byte key and encrypts a plaintext using AES-256-GCM.
// Output format is 12 byte nonce + ciphertext.
func Encrypt(plaintext, key []byte) ([]byte, error) {
	// Enforce AES-256
	if len(key) != 32 {
		return nil, aes.KeySizeError(len(key))
	}

	// None should not be reused more often that 2^32 times:
	// https://pkg.go.dev/crypto/cipher#NewGCM
	// Assuming n encryption operations per second, the key has to be rotated every:
	// n=1: 2^32 / (60*60*24*365*10) = 135 years.
	// n=10: 2^32 / (60*60*24*365*10) = 13.5 years.
	// n=100: 2^32 / (60*60*24*365*10) = 1.3 years.
	// n=1000: 2^32 / (60*60*24*365*10) = 50 days.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// Prepend the nonce to the ciphertext.
	ciphertext = append(nonce, ciphertext...)

	return ciphertext, nil
}

// Decrypt takes a 32 byte key and decrypts a ciphertext using AES-256-GCM.
// ciphertext is formatted as 12 byte nonce + ciphertext.
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Enforce AES-256
	if len(key) != 32 {
		return nil, aes.KeySizeError(len(key))
	}

	// Extract the nonce from the ciphertext.
	nonce := ciphertext[:12]
	ciphertext = ciphertext[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonce, ciphertext, nil)
}

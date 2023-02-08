/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package internal implements the CloudKMS interface using go-kms-wrapping.

Adding support for a new KMS that is supported by go-kms-wrapping,
simply requires implementing a New function that initializes a KMSClient.
*/
package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/crypto"
	kmsInterface "github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
)

type kmsWrapper interface {
	Decrypt(context.Context, *wrapping.BlobInfo, ...wrapping.Option) ([]byte, error)
	Encrypt(context.Context, []byte, ...wrapping.Option) (*wrapping.BlobInfo, error)
}

// KMSClient implements the CloudKMS interface using go-kms-wrapping.
type KMSClient struct {
	Storage kmsInterface.Storage
	Wrapper kmsWrapper
}

// GetDEK fetches an encrypted Data Encryption Key from storage.
// If the key exists, it is decrypted and returned.
// If no such key exists, a new one is generated, encrypted and saved to storage.
func (c *KMSClient) GetDEK(ctx context.Context, keyID string, dekSize int) ([]byte, error) {
	encryptedDEK, err := c.Storage.Get(ctx, keyID)
	if err != nil {
		if !errors.Is(err, storage.ErrDEKUnset) {
			return nil, fmt.Errorf("loading encrypted DEK from storage: %w", err)
		}

		// If the DEK does not exist we generate a new random DEK and save it to storage
		newDEK, err := crypto.GenerateRandomBytes(dekSize)
		if err != nil {
			return nil, fmt.Errorf("key generation: %w", err)
		}
		return newDEK, c.putDEK(ctx, keyID, newDEK)
	}

	wrappedKey := &wrapping.BlobInfo{}
	if err := json.Unmarshal(encryptedDEK, wrappedKey); err != nil {
		return nil, fmt.Errorf("unmarshaling wrapped DEK: %w", err)
	}

	dek, err := c.Wrapper.Decrypt(ctx, wrappedKey)
	if err != nil {
		return nil, fmt.Errorf("decrypting DEK: %w", err)
	}

	return dek, nil
}

// putDEK encrypts a Data Encryption Key and saves it to storage.
func (c *KMSClient) putDEK(ctx context.Context, keyID string, plainDEK []byte) error {
	wrappedKey, err := c.Wrapper.Encrypt(ctx, plainDEK)
	if err != nil {
		return fmt.Errorf("encrypting DEK: %w", err)
	}

	encryptedDEK, err := json.Marshal(wrappedKey)
	if err != nil {
		return fmt.Errorf("marshaling wrapped DEK: %w", err)
	}

	return c.Storage.Put(ctx, keyID, encryptedDEK)
}

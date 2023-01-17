/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package cluster implements a KMS backend for in cluster key management.

The cluster backend holds a master key, and corresponding salt.
Data Encryption Keys (DEK) are derived from master key and salt using HKDF.

This backend does not require a storage backend, as keys are derived on demand and not stored anywhere.
For that purpose the special NoStoreURI can be used during KMS initialization.
*/
package cluster

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation/v2/internal/crypto"
)

// KMS implements the kms.CloudKMS interface for in cluster key management.
type KMS struct {
	masterKey []byte
	salt      []byte
}

// New creates a new ClusterKMS.
func New(salt []byte) *KMS {
	return &KMS{salt: salt}
}

// CreateKEK sets the ClusterKMS masterKey.
func (c *KMS) CreateKEK(ctx context.Context, keyID string, kek []byte) error {
	c.masterKey = kek
	return nil
}

// GetDEK derives a key from the KMS masterKey.
func (c *KMS) GetDEK(ctx context.Context, kekID string, dekID string, dekSize int) ([]byte, error) {
	if len(c.masterKey) == 0 {
		return nil, errors.New("master key not set for Constellation KMS")
	}
	return crypto.DeriveKey(c.masterKey, c.salt, []byte(dekID), uint(dekSize))
}

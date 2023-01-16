/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
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
func New(key []byte, salt []byte) (*KMS, error) {
	if len(key) == 0 {
		return nil, errors.New("missing master key")
	}
	if len(salt) == 0 {
		return nil, errors.New("missing salt")
	}

	return &KMS{masterKey: key, salt: salt}, nil
}

// CreateKEK sets the ClusterKMS masterKey.
func (c *KMS) CreateKEK(ctx context.Context, keyID string, kek []byte) error {
	c.masterKey = kek
	return nil
}

// GetDEK derives a key from the KMS masterKey.
func (c *KMS) GetDEK(ctx context.Context, dekID string, dekSize int) ([]byte, error) {
	if len(c.masterKey) == 0 {
		return nil, errors.New("master key not set for Constellation KMS")
	}
	return crypto.DeriveKey(c.masterKey, c.salt, []byte(dekID), uint(dekSize))
}

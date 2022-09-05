/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cluster

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation/internal/crypto"
)

// ClusterKMS implements the kms.CloudKMS interface for in cluster key management.
type ClusterKMS struct {
	masterKey []byte
	salt      []byte
}

// New creates a new ClusterKMS.
func New(salt []byte) *ClusterKMS {
	return &ClusterKMS{salt: salt}
}

// CreateKEK sets the ClusterKMS masterKey.
func (c *ClusterKMS) CreateKEK(ctx context.Context, keyID string, kek []byte) error {
	c.masterKey = kek
	return nil
}

// GetDEK derives a key from the KMS masterKey.
func (c *ClusterKMS) GetDEK(ctx context.Context, kekID string, dekID string, dekSize int) ([]byte, error) {
	if len(c.masterKey) == 0 {
		return nil, errors.New("master key not set for Constellation KMS")
	}
	return crypto.DeriveKey(c.masterKey, c.salt, []byte(dekID), uint(dekSize))
}

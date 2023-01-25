/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package memfs implements a storage backend for the KMS that stores keys in memory only.
// This package should be used for testing only.
package memfs

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
)

// Storage is the standard implementation of the Storage interface, storing keys in memory only.
type Storage struct {
	dekPool map[string][]byte
}

// New creates and initializes a new Storage object.
func New() *Storage {
	s := &Storage{
		dekPool: make(map[string][]byte),
	}

	return s
}

// Get returns a DEK from Storage by key ID.
func (s *Storage) Get(ctx context.Context, keyID string) ([]byte, error) {
	encDEK, ok := s.dekPool[keyID]
	if ok {
		return encDEK, nil
	}
	return nil, storage.ErrDEKUnset
}

// Put saves a DEK to Storage by key ID.
func (s *Storage) Put(ctx context.Context, keyID string, encDEK []byte) error {
	s.dekPool[keyID] = encDEK
	return nil
}

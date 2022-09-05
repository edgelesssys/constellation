/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package storage

import "context"

// MemMapStorage is the standard implementation of the Storage interface, storing keys in memory only.
type MemMapStorage struct {
	dekPool map[string][]byte
}

// NewMemMapStorage creates and initialises a new MemMapStorage object.
func NewMemMapStorage() *MemMapStorage {
	s := &MemMapStorage{
		dekPool: make(map[string][]byte),
	}

	return s
}

// Get returns a DEK from MemMapStorage by key ID.
func (s *MemMapStorage) Get(ctx context.Context, keyID string) ([]byte, error) {
	encDEK, ok := s.dekPool[keyID]
	if ok {
		return encDEK, nil
	}
	return nil, ErrDEKUnset
}

// Put saves a DEK to MemMapStorage by key ID.
func (s *MemMapStorage) Put(ctx context.Context, keyID string, encDEK []byte) error {
	s.dekPool[keyID] = encDEK
	return nil
}

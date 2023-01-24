/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package kms provides an abstract interface for Key Management Services.
package kms

import (
	"context"
)

// CloudKMS enables using cloud base Key Management Services.
type CloudKMS interface {
	// GetDEK returns the DEK for dekID and kekID from the KMS.
	// If the DEK does not exist, a new one is created and saved to storage.
	GetDEK(ctx context.Context, dekID string, dekSize int) ([]byte, error)
	// Close closes any open connection on the KMS client.
	Close()
}

// Storage provides an abstract interface for the storage backend used for DEKs.
type Storage interface {
	// Get returns a DEK from the storage by key ID. If the DEK does not exist, returns storage.ErrDEKUnset.
	Get(context.Context, string) ([]byte, error)
	// Put saves a DEK to the storage by key ID.
	Put(context.Context, string, []byte) error
}

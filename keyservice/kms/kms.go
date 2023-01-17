/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package kms provides an abstract interface for Key Management Services.
package kms

import (
	"context"
	"errors"
)

// CloudKMS enables using cloud base Key Management Services.
type CloudKMS interface {
	// CreateKEK creates a new KEK with the given key material, if provided. If successful, the key can be referenced by keyID in the KMS in accordance to the policy.
	CreateKEK(ctx context.Context, keyID string, kek []byte) error
	// GetDEK returns the DEK for dekID and kekID from the KMS.
	// If the DEK does not exist, a new one is created and saved to storage.
	GetDEK(ctx context.Context, kekID string, dekID string, dekSize int) ([]byte, error)
}

// Storage provides an abstract interface for the storage backend used for DEKs.
type Storage interface {
	// Get returns a DEK from the storage by key ID. If the DEK does not exist, returns storage.ErrDEKUnset.
	Get(context.Context, string) ([]byte, error)
	// Put saves a DEK to the storage by key ID.
	Put(context.Context, string, []byte) error
}

// ErrKEKUnknown is an error raised by unknown KEK in the KMS.
var ErrKEKUnknown = errors.New("requested KEK not found")

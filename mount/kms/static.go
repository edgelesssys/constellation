package kms

import (
	"context"
)

// staticKMS is a KMS only returning keys containing of 0x41 bytes for every request.
// Use for testing ONLY.
type staticKMS struct{}

// NewStaticKMS creates a new StaticKMS.
// Use for testing ONLY.
func NewStaticKMS() *staticKMS {
	return &staticKMS{}
}

// GetDEK returns the key of staticKMS.
func (k *staticKMS) GetDEK(ctx context.Context, kekID, dekID string, dekSize int) ([]byte, error) {
	key := make([]byte, dekSize)
	for i := range key {
		key[i] = 0x41
	}
	return key, nil
}

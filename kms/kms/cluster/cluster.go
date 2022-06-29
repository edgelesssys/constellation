package cluster

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation/bootstrapper/util"
)

// ClusterKMS implements the kms.CloudKMS interface for in cluster key management.
type ClusterKMS struct {
	masterKey []byte
}

// CreateKEK sets the CoordinatorKMS masterKey.
func (c *ClusterKMS) CreateKEK(ctx context.Context, keyID string, kek []byte) error {
	c.masterKey = kek
	return nil
}

// GetDEK derives a key from the KMS masterKey.
func (c *ClusterKMS) GetDEK(ctx context.Context, kekID string, dekID string, dekSize int) ([]byte, error) {
	if len(c.masterKey) == 0 {
		return nil, errors.New("master key not set for Constellation KMS")
	}
	// TODO: Choose a way to salt key derivation
	return util.DeriveKey(c.masterKey, []byte("Constellation"), []byte("key"+dekID), uint(dekSize))
}

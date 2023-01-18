/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package gcp implements a KMS backend for Google Cloud KMS.

The following permissions are required for the service account used to authenticate with GCP:

  - cloudkms.cryptoKeyVersions.create

  - cloudkms.cryptoKeyVersions.update

  - cloudkms.cryptoKeyVersions.useToDecrypt

  - cloudkms.cryptoKeyVersions.useToEncrypt

  - cloudkms.importJobs.create

  - cloudkms.importJobs.get

  - cloudkms.importJobs.useToImport
*/
package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	cloudkms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	kmsInterface "github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/util"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
	gcpckms "github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms/v2"
)

type kmsWrapper interface {
	Decrypt(context.Context, *wrapping.BlobInfo, ...wrapping.Option) ([]byte, error)
	Encrypt(context.Context, []byte, ...wrapping.Option) (*wrapping.BlobInfo, error)
	Client() *cloudkms.KeyManagementClient
}

// KMSClient implements the CloudKMS interface for Google Cloud Platform.
type KMSClient struct {
	storage kmsInterface.Storage
	wrapper kmsWrapper
}

// New initializes a KMS client for Google Cloud Platform.
func New(ctx context.Context, kekID string, store kmsInterface.Storage, projectID, locationID, keyRingID string) (*KMSClient, error) {
	if store == nil {
		store = storage.NewMemMapStorage()
	}

	wrapper := gcpckms.NewWrapper()
	if _, err := wrapper.SetConfig(
		context.Background(),
		gcpckms.WithProject(projectID),
		gcpckms.WithRegion(locationID),
		gcpckms.WithKeyRing(keyRingID),
		gcpckms.WithCryptoKey(kekID),
	); err != nil {
		return nil, fmt.Errorf("setting GCP KMS wrapper config: %w", err)
	}

	c := &KMSClient{
		storage: store,
		wrapper: wrapper,
	}

	// test if the KMS can be reached with the given configuration
	client := wrapper.Client()
	if _, err := client.GetKeyRing(ctx, &kmspb.GetKeyRingRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", projectID, locationID, keyRingID),
	}); err != nil {
		return nil, fmt.Errorf("testing connection to GCP KMS: %w", err)
	}

	return c, nil
}

// GetDEK fetches an encrypted Data Encryption Key from storage and decrypts it using a KEK stored in Google's KMS.
func (c *KMSClient) GetDEK(ctx context.Context, keyID string, dekSize int) ([]byte, error) {
	encryptedDEK, err := c.storage.Get(ctx, keyID)
	if err != nil {
		if !errors.Is(err, storage.ErrDEKUnset) {
			return nil, fmt.Errorf("loading encrypted DEK from storage: %w", err)
		}

		// If the DEK does not exist we generate a new random DEK and save it to storage
		newDEK, err := util.GetRandomKey(dekSize)
		if err != nil {
			return nil, fmt.Errorf("key generation: %w", err)
		}
		return newDEK, c.putDEK(ctx, keyID, newDEK)
	}

	wrappedKey := &wrapping.BlobInfo{}
	if err := json.Unmarshal(encryptedDEK, wrappedKey); err != nil {
		return nil, fmt.Errorf("unmarshaling wrapped DEK: %w", err)
	}

	dek, err := c.wrapper.Decrypt(ctx, wrappedKey)
	if err != nil {
		return nil, fmt.Errorf("decrypting DEK: %w", err)
	}

	return dek, nil
}

// Close closes the KMS client.
func (c *KMSClient) Close() {
	_ = c.wrapper.Client().Close()
}

// putDEK saves an encrypted Data Encryption Key to storage.
func (c *KMSClient) putDEK(ctx context.Context, keyID string, plainDEK []byte) error {
	wrappedKey, err := c.wrapper.Encrypt(ctx, plainDEK)
	if err != nil {
		return fmt.Errorf("encrypting DEK: %w", err)
	}

	encryptedDEK, err := json.Marshal(wrappedKey)
	if err != nil {
		return fmt.Errorf("marshaling wrapped DEK: %w", err)
	}

	return c.storage.Put(ctx, keyID, encryptedDEK)
}

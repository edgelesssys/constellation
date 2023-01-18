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
	"fmt"

	cloudkms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	kmsInterface "github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/internal"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	gcpckms "github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms/v2"
)

// KMSClient implements the CloudKMS interface for Google Cloud Platform.
type KMSClient struct {
	kms    *internal.KMSClient
	client *cloudkms.KeyManagementClient
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

	// test if the KMS can be reached with the given configuration
	client := wrapper.Client()
	if _, err := client.GetKeyRing(ctx, &kmspb.GetKeyRingRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", projectID, locationID, keyRingID),
	}); err != nil {
		return nil, fmt.Errorf("testing connection to GCP KMS: %w", err)
	}

	return &KMSClient{
		kms: &internal.KMSClient{
			Storage: store,
			Wrapper: wrapper,
		},
		client: client,
	}, nil
}

// GetDEK fetches an encrypted Data Encryption Key from storage and decrypts it using a KEK stored in Google's KMS.
func (c *KMSClient) GetDEK(ctx context.Context, keyID string, dekSize int) ([]byte, error) {
	return c.kms.GetDEK(ctx, keyID, dekSize)
}

// Close closes the KMS client.
func (c *KMSClient) Close() {
	_ = c.client.Close()
}

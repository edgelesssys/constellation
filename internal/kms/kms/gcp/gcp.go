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
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/edgelesssys/constellation/v2/internal/kms/config"
	kmsInterface "github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/util"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type clientAPI interface {
	io.Closer
	CreateCryptoKey(context.Context, *kmspb.CreateCryptoKeyRequest, ...gax.CallOption) (*kmspb.CryptoKey, error)
	CreateImportJob(context.Context, *kmspb.CreateImportJobRequest, ...gax.CallOption) (*kmspb.ImportJob, error)
	Decrypt(context.Context, *kmspb.DecryptRequest, ...gax.CallOption) (*kmspb.DecryptResponse, error)
	Encrypt(context.Context, *kmspb.EncryptRequest, ...gax.CallOption) (*kmspb.EncryptResponse, error)
	GetKeyRing(context.Context, *kmspb.GetKeyRingRequest, ...gax.CallOption) (*kmspb.KeyRing, error)
	ImportCryptoKeyVersion(context.Context, *kmspb.ImportCryptoKeyVersionRequest, ...gax.CallOption) (*kmspb.CryptoKeyVersion, error)
	UpdateCryptoKeyPrimaryVersion(context.Context, *kmspb.UpdateCryptoKeyPrimaryVersionRequest, ...gax.CallOption) (*kmspb.CryptoKey, error)
	GetImportJob(context.Context, *kmspb.GetImportJobRequest, ...gax.CallOption) (*kmspb.ImportJob, error)
}

// KMSClient implements the CloudKMS interface for Google Cloud Platform.
type KMSClient struct {
	projectID        string
	locationID       string
	keyRingID        string
	newClient        func(ctx context.Context, opts ...option.ClientOption) (clientAPI, error)
	waitBackoffLimit int
	storage          kmsInterface.Storage
	protectionLevel  kmspb.ProtectionLevel
	kekID            string
	opts             []gax.CallOption
}

// New initializes a KMS client for Google Cloud Platform.
func New(ctx context.Context, projectID, locationID, keyRingID string, store kmsInterface.Storage, protectionLvl kmspb.ProtectionLevel, kekID string, opts ...gax.CallOption) (*KMSClient, error) {
	if store == nil {
		store = storage.NewMemMapStorage()
	}

	if protectionLvl != kmspb.ProtectionLevel_SOFTWARE && protectionLvl != kmspb.ProtectionLevel_HSM {
		protectionLvl = kmspb.ProtectionLevel_SOFTWARE
	}

	c := &KMSClient{
		projectID:        projectID,
		locationID:       locationID,
		keyRingID:        keyRingID,
		newClient:        keyManagementClientFactory,
		waitBackoffLimit: 10,
		storage:          store,
		protectionLevel:  protectionLvl,
		kekID:            kekID,
		opts:             opts,
	}

	// test if the KMS can be reached with the given configuration
	if err := c.testConnection(ctx); err != nil {
		return nil, fmt.Errorf("testing connection to GCP KMS: %w", err)
	}

	return c, nil
}

// CreateKEK creates a new Key Encryption Key using Google Key Management System.
//
// If no key material is provided, a new key is generated by Google's KMS, otherwise the key material is used to import the key.
func (c *KMSClient) CreateKEK(ctx context.Context, keyID string, key []byte) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	if len(key) == 0 {
		_, err := c.createNewKEK(ctx, keyID, client, false)
		if err != nil {
			return fmt.Errorf("creating new KEK in Google KMS: %w", err)
		}
		return nil
	}

	_, err = c.importKEK(ctx, keyID, key, client)
	if err != nil {
		return fmt.Errorf("importing KEK to Google KMS: %w", err)
	}
	return nil
}

// GetDEK fetches an encrypted Data Encryption Key from storage and decrypts it using a KEK stored in Google's KMS.
func (c *KMSClient) GetDEK(ctx context.Context, keyID string, dekSize int) ([]byte, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

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
		return newDEK, c.putDEK(ctx, client, c.kekID, keyID, newDEK)
	}

	request := &kmspb.DecryptRequest{
		Name:       fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", c.projectID, c.locationID, c.keyRingID, c.kekID),
		Ciphertext: encryptedDEK,
	}

	res, err := client.Decrypt(ctx, request, c.opts...)
	if err != nil {
		if status.Convert(err).Code() == codes.NotFound {
			return nil, kmsInterface.ErrKEKUnknown
		}
		return nil, fmt.Errorf("decrypting DEK: %w", err)
	}

	return res.GetPlaintext(), nil
}

// putDEK encrypts a Data Encryption Key using a KEK stored in Google's KMS and saves it to storage.
func (c *KMSClient) putDEK(ctx context.Context, client clientAPI, kekID, keyID string, plainDEK []byte) error {
	request := &kmspb.EncryptRequest{
		Name:      fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", c.projectID, c.locationID, c.keyRingID, kekID),
		Plaintext: plainDEK,
	}

	res, err := client.Encrypt(ctx, request, c.opts...)
	if err != nil {
		if status.Convert(err).Code() == codes.NotFound {
			return kmsInterface.ErrKEKUnknown
		}
		return fmt.Errorf("encrypting DEK: %w", err)
	}

	return c.storage.Put(ctx, keyID, res.Ciphertext)
}

// createNewKEK creates a new symmetric Crypto Key in Google's KMS.
func (c *KMSClient) createNewKEK(ctx context.Context, keyID string, client clientAPI, importOnly bool) (*kmspb.CryptoKey, error) {
	request := &kmspb.CreateCryptoKeyRequest{
		Parent:      fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", c.projectID, c.locationID, c.keyRingID),
		CryptoKeyId: keyID,
		CryptoKey: &kmspb.CryptoKey{
			Purpose: kmspb.CryptoKey_ENCRYPT_DECRYPT,
			Labels: map[string]string{
				"created-by": "constellation-kms-client",
				"component":  "constellation-kek",
			},
			VersionTemplate: &kmspb.CryptoKeyVersionTemplate{
				ProtectionLevel: c.protectionLevel,
				Algorithm:       kmspb.CryptoKeyVersion_GOOGLE_SYMMETRIC_ENCRYPTION,
			},
			ImportOnly: importOnly,
		},
		SkipInitialVersionCreation: importOnly,
	}

	return client.CreateCryptoKey(ctx, request, c.opts...)
}

// importKEK imports a symmetric Crypto Key to Google's KMS-
//
// Keys in the Google KMS can not be removed, only disabled and/or key material destroyed.
// Since we create the initial key with `SkipInitialVersionCreation=true`, no key material is created and we do not perform any cleanup on failure.
func (c *KMSClient) importKEK(ctx context.Context, keyID string, key []byte, client clientAPI) (*kmspb.CryptoKey, error) {
	// we need an empty crypto key to import into
	parentKey, err := c.createNewKEK(ctx, keyID, client, true)
	if err != nil {
		return nil, err
	}

	// Create import job
	jobName := fmt.Sprintf("import-job-%s", keyID)
	request := &kmspb.CreateImportJobRequest{
		Parent:      fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", c.projectID, c.locationID, c.keyRingID),
		ImportJobId: jobName,
		ImportJob: &kmspb.ImportJob{
			ImportMethod:    kmspb.ImportJob_RSA_OAEP_4096_SHA1_AES_256,
			ProtectionLevel: c.protectionLevel,
		},
	}
	_, err = client.CreateImportJob(ctx, request, c.opts...)
	if err != nil {
		return nil, err
	}
	impRes, ok := c.waitBackoff(ctx, jobName, client)
	if !ok {
		return nil, fmt.Errorf("import job was not active after %d tries, giving up", c.waitBackoffLimit)
	}

	// Wrap the to be imported key using a public RSA key from the created import job and an ephemeral AES as specified here: https://cloud.google.com/kms/docs/wrapping-a-key
	wrappingPublicKey, err := util.ParsePEMtoPublicKeyRSA([]byte(impRes.PublicKey.GetPem()))
	if err != nil {
		return nil, err
	}
	wrappedKey, err := wrapCryptoKey(key, wrappingPublicKey)
	if err != nil {
		return nil, fmt.Errorf("wrapping public key: %w", err)
	}

	// Perform the actual key import
	importReq := &kmspb.ImportCryptoKeyVersionRequest{
		Parent:    parentKey.GetName(),
		Algorithm: kmspb.CryptoKeyVersion_GOOGLE_SYMMETRIC_ENCRYPTION,
		ImportJob: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/importJobs/%s", c.projectID, c.locationID, c.keyRingID, jobName),
		WrappedKeyMaterial: &kmspb.ImportCryptoKeyVersionRequest_RsaAesWrappedKey{
			RsaAesWrappedKey: wrappedKey,
		},
	}
	res, err := client.ImportCryptoKeyVersion(ctx, importReq, c.opts...)
	if err != nil {
		return nil, err
	}

	// Set the imported key as the primary key version
	newVersion := strings.Split(res.GetName(), "/")
	updateRequest := &kmspb.UpdateCryptoKeyPrimaryVersionRequest{
		Name:               parentKey.GetName(),
		CryptoKeyVersionId: newVersion[len(newVersion)-1], // We only need the Version ID of the imported key, not the full resource name
	}
	return client.UpdateCryptoKeyPrimaryVersion(ctx, updateRequest, c.opts...)
}

// waitBackoff is a utility function to wait for the creation of an import job.
func (c *KMSClient) waitBackoff(ctx context.Context, jobName string, client clientAPI) (*kmspb.ImportJob, bool) {
	for i := 0; i < c.waitBackoffLimit; i++ {
		res, err := client.GetImportJob(ctx, &kmspb.GetImportJobRequest{
			Name: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/importJobs/%s", c.projectID, c.locationID, c.keyRingID, jobName),
		}, c.opts...)
		if (err == nil) && (res.State == kmspb.ImportJob_ACTIVE) {
			return res, true
		}

		// wait for increasingly longer time until we either reach the preset limit or get an active job
		time.Sleep(time.Second * 5 * time.Duration(i))
	}
	return nil, false
}

// testConnection checks if the KMS is reachable with the given configuration.
func (c *KMSClient) testConnection(ctx context.Context) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	if _, err := client.GetKeyRing(ctx, &kmspb.GetKeyRingRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", c.projectID, c.locationID, c.keyRingID),
	}); err != nil {
		return fmt.Errorf("GCP KMS not reachable: %w", err)
	}
	return nil
}

func keyManagementClientFactory(ctx context.Context, opts ...option.ClientOption) (clientAPI, error) {
	return kms.NewKeyManagementClient(ctx, opts...)
}

// wrapCryptoKey wraps a key for import using a public RSA key, see: https://cloud.google.com/kms/docs/wrapping-a-key
func wrapCryptoKey(key []byte, wrapKeyRSA *rsa.PublicKey) ([]byte, error) {
	// Enforce 256bit key length
	if len(key) != config.SymmetricKeyLength {
		return nil, fmt.Errorf("invalid key size: want [%d], got [%d]", config.SymmetricKeyLength, len(key))
	}
	// create random 256bit AES wrapping key
	wrapKeyAES := make([]byte, config.SymmetricKeyLength)
	if _, err := rand.Read(wrapKeyAES); err != nil {
		return nil, err
	}

	// Perform CKM_AES_KEY_WRAP_PAD to wrap the key
	wrappedKey, err := util.WrapAES(key, wrapKeyAES)
	if err != nil {
		return nil, err
	}

	// Encrypt the ephemeral AES key with the KMS provided wrapping key
	// Google KMS requires RSAES-OAEP with SHA-1 and an empty label
	encWrapKeyAES, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, wrapKeyRSA, wrapKeyAES, nil)
	if err != nil {
		return nil, err
	}

	return append(encWrapKeyAES, wrappedKey...), nil
}

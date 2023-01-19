/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package setup provides functions to create a KMS and key store from a given URI.

This package does not provide any functionality to interact with the KMS or key store,
but only to create them.

Adding support for a new KMS or storage backend requires adding a new URI for that backend,
and implementing the corresponding get*Config function.
*/
package setup

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/aws"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/azure"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/cluster"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/gcp"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
)

// Well known endpoints for KMS services.
const (
	AWSKMSURI     = "kms://aws?keyPolicy=%s&kekID=%s"
	AzureKMSURI   = "kms://azure-kms?name=%s&type=%s&kekID=%s"
	AzureHSMURI   = "kms://azure-hsm?name=%s&kekID=%s"
	GCPKMSURI     = "kms://gcp?project=%s&location=%s&keyRing=%s&protectionLvl=%s&kekID=%s"
	ClusterKMSURI = "kms://cluster-kms?key=%s&salt=%s"
	AWSS3URI      = "storage://aws?bucket=%s"
	AzureBlobURI  = "storage://azure?container=%s&connectionString=%s"
	GCPStorageURI = "storage://gcp?projects=%s&bucket=%s"
	NoStoreURI    = "storage://no-store"
)

// MasterSecret holds the master key and salt for deriving keys.
type MasterSecret struct {
	Key  []byte `json:"key"`
	Salt []byte `json:"salt"`
}

// EncodeToURI returns an URI encoding the master secret.
func (m *MasterSecret) EncodeToURI() string {
	return fmt.Sprintf(
		ClusterKMSURI,
		base64.URLEncoding.EncodeToString(m.Key),
		base64.URLEncoding.EncodeToString(m.Salt),
	)
}

// KMSInformation about an existing KMS.
type KMSInformation struct {
	KMSURI             string
	StorageURI         string
	KeyEncryptionKeyID string
}

// KMS creates a KMS and key store from the given parameters.
func KMS(ctx context.Context, storageURI, kmsURI string) (kms.CloudKMS, error) {
	store, err := getStore(ctx, storageURI)
	if err != nil {
		return nil, err
	}

	return getKMS(ctx, kmsURI, store)
}

// getStore creates a key store depending on the given parameters.
func getStore(ctx context.Context, storageURI string) (kms.Storage, error) {
	uri, err := url.Parse(storageURI)
	if err != nil {
		return nil, err
	}
	if uri.Scheme != "storage" {
		return nil, fmt.Errorf("invalid storage URI: invalid scheme: %s", uri.Scheme)
	}

	switch uri.Host {
	case "aws":
		bucket, err := getAWSS3Config(uri)
		if err != nil {
			return nil, err
		}
		return storage.NewAWSS3Storage(ctx, bucket, nil)

	case "azure":
		container, connString, err := getAzureBlobConfig(uri)
		if err != nil {
			return nil, err
		}
		return storage.NewAzureStorage(ctx, connString, container, nil)

	case "gcp":
		project, bucket, err := getGCPStorageConfig(uri)
		if err != nil {
			return nil, err
		}
		return storage.NewGoogleCloudStorage(ctx, project, bucket, nil)

	case "no-store":
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown storage type: %s", uri.Host)
	}
}

// getKMS creates a KMS client with the given key store and depending on the given parameters.
func getKMS(ctx context.Context, kmsURI string, store kms.Storage) (kms.CloudKMS, error) {
	uri, err := url.Parse(kmsURI)
	if err != nil {
		return nil, err
	}
	if uri.Scheme != "kms" {
		return nil, fmt.Errorf("invalid KMS URI: invalid scheme: %s", uri.Scheme)
	}

	switch uri.Host {
	case "aws":
		poliyProducer, kekID, err := getAWSKMSConfig(uri)
		if err != nil {
			return nil, err
		}
		return aws.New(ctx, poliyProducer, store, kekID)

	case "azure-kms":
		vaultName, vaultType, kekID, err := getAzureKMSConfig(uri)
		if err != nil {
			return nil, err
		}
		return azure.New(ctx, vaultName, azure.VaultSuffix(vaultType), store, kekID, nil)

	case "azure-hsm":
		vaultName, kekID, err := getAzureHSMConfig(uri)
		if err != nil {
			return nil, err
		}
		return azure.NewHSM(ctx, vaultName, store, kekID, nil)

	case "gcp":
		project, location, keyRing, protectionLvl, kekID, err := getGCPKMSConfig(uri)
		if err != nil {
			return nil, err
		}
		return gcp.New(ctx, project, location, keyRing, store, kmspb.ProtectionLevel(protectionLvl), kekID)

	case "cluster-kms":
		masterSecret, err := getClusterKMSConfig(uri)
		if err != nil {
			return nil, err
		}
		return cluster.New(masterSecret.Key, masterSecret.Salt)

	default:
		return nil, fmt.Errorf("unknown KMS type: %s", uri.Host)
	}
}

type defaultPolicyProducer struct {
	policy string
}

func (p *defaultPolicyProducer) CreateKeyPolicy(keyID string) (string, error) {
	return p.policy, nil
}

func getAWSS3Config(uri *url.URL) (string, error) {
	r, err := getConfig(uri.Query(), []string{"bucket"})
	return r[0], err
}

func getAWSKMSConfig(uri *url.URL) (*defaultPolicyProducer, string, error) {
	r, err := getConfig(uri.Query(), []string{"keyPolicy", "kekID"})
	if err != nil {
		return nil, "", err
	}

	if len(r) != 2 {
		return nil, "", fmt.Errorf("expected 2 KmsURI args, got %d", len(r))
	}

	kekID, err := base64.URLEncoding.DecodeString(r[1])
	if err != nil {
		return nil, "", fmt.Errorf("parsing kekID from kmsUri: %w", err)
	}

	return &defaultPolicyProducer{policy: r[0]}, string(kekID), err
}

func getAzureKMSConfig(uri *url.URL) (string, string, string, error) {
	r, err := getConfig(uri.Query(), []string{"name", "type", "kekID"})
	if err != nil {
		return "", "", "", fmt.Errorf("getting config: %w", err)
	}
	if len(r) != 3 {
		return "", "", "", fmt.Errorf("expected 3 KmsURI args, got %d", len(r))
	}

	kekID, err := base64.URLEncoding.DecodeString(r[2])
	if err != nil {
		return "", "", "", fmt.Errorf("parsing kekID from kmsUri: %w", err)
	}

	return r[0], r[1], string(kekID), err
}

func getAzureHSMConfig(uri *url.URL) (string, string, error) {
	r, err := getConfig(uri.Query(), []string{"name", "kekID"})
	if err != nil {
		return "", "", fmt.Errorf("getting config: %w", err)
	}
	if len(r) != 2 {
		return "", "", fmt.Errorf("expected 2 KmsURI args, got %d", len(r))
	}

	kekID, err := base64.URLEncoding.DecodeString(r[1])
	if err != nil {
		return "", "", fmt.Errorf("parsing kekID from kmsUri: %w", err)
	}

	return r[0], string(kekID), err
}

func getAzureBlobConfig(uri *url.URL) (string, string, error) {
	r, err := getConfig(uri.Query(), []string{"container", "connectionString"})
	if err != nil {
		return "", "", err
	}
	return r[0], r[1], nil
}

func getGCPKMSConfig(uri *url.URL) (project string, location string, keyRing string, protectionLvl int32, kekID string, err error) {
	r, err := getConfig(uri.Query(), []string{"project", "location", "keyRing", "protectionLvl", "kekID"})
	if err != nil {
		return "", "", "", 0, "", err
	}

	if len(r) != 5 {
		return "", "", "", 0, "", fmt.Errorf("expected 5 KmsURI args, got %d", len(r))
	}

	kekIDByte, err := base64.URLEncoding.DecodeString(r[4])
	if err != nil {
		return "", "", "", 0, "", fmt.Errorf("parsing kekID from kmsUri: %w", err)
	}

	protectionLvl32, err := strconv.ParseInt(r[3], 10, 32)
	if err != nil {
		return "", "", "", 0, "", err
	}
	return r[0], r[1], r[2], int32(protectionLvl32), string(kekIDByte), nil
}

func getGCPStorageConfig(uri *url.URL) (string, string, error) {
	r, err := getConfig(uri.Query(), []string{"project", "bucket"})
	return r[0], r[1], err
}

func getClusterKMSConfig(uri *url.URL) (MasterSecret, error) {
	r, err := getConfig(uri.Query(), []string{"key", "salt"})
	if err != nil {
		return MasterSecret{}, err
	}

	if len(r) != 2 {
		return MasterSecret{}, fmt.Errorf("expected 2 KmsURI args, got %d", len(r))
	}

	key, err := base64.URLEncoding.DecodeString(r[0])
	if err != nil {
		return MasterSecret{}, fmt.Errorf("parsing key from kmsUri: %w", err)
	}
	salt, err := base64.URLEncoding.DecodeString(r[1])
	if err != nil {
		return MasterSecret{}, fmt.Errorf("parsing salt from kmsUri: %w", err)
	}

	return MasterSecret{Key: key, Salt: salt}, nil
}

// getConfig parses url query values, returning a map of the requested values.
// Returns an error if a key has no value.
// This function MUST always return a slice of the same length as len(keys).
func getConfig(values url.Values, keys []string) ([]string, error) {
	res := make([]string, len(keys))

	for idx, key := range keys {
		val := values.Get(key)
		if val == "" {
			return res, fmt.Errorf("missing value for key: %q", key)
		}
		val, err := url.QueryUnescape(val)
		if err != nil {
			return res, fmt.Errorf("failed to unescape value for key: %q", key)
		}
		res[idx] = val
	}

	return res, nil
}

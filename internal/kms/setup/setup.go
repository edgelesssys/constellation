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

	"github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/aws"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/azure"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/cluster"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/gcp"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/awss3"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/azureblob"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/gcs"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
)

// Well known endpoints for KMS services.
const (
	AWSKMSURI     = "kms://aws?keyPolicy=%s&kekID=%s"
	AzureKMSURI   = "kms://azure?tenantID=%s&clientID=%s&clientSecret=%s&name=%s&type=%s&kekID=%s"
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
	url, err := url.Parse(storageURI)
	if err != nil {
		return nil, err
	}
	if url.Scheme != "storage" {
		return nil, fmt.Errorf("invalid storage URI: invalid scheme: %s", url.Scheme)
	}

	switch url.Host {
	case "aws":
		cfg, err := uri.DecodeAWSS3ConfigFromURI(storageURI)
		if err != nil {
			return nil, err
		}
		return awss3.New(ctx, cfg)

	case "azure":
		cfg, err := uri.DecodeAzureBlobConfigFromURI(storageURI)
		if err != nil {
			return nil, err
		}
		return azureblob.New(ctx, cfg)

	case "gcp":
		project, bucket, err := getGCPStorageConfig(url)
		if err != nil {
			return nil, err
		}
		return gcs.New(ctx, project, bucket, nil)

	case "no-store":
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown storage type: %s", url.Host)
	}
}

// getKMS creates a KMS client with the given key store and depending on the given parameters.
func getKMS(ctx context.Context, kmsURI string, store kms.Storage) (kms.CloudKMS, error) {
	url, err := url.Parse(kmsURI)
	if err != nil {
		return nil, err
	}
	if url.Scheme != "kms" {
		return nil, fmt.Errorf("invalid KMS URI: invalid scheme: %s", url.Scheme)
	}

	switch url.Host {
	case "aws":
		cfg, err := uri.DecodeAWSConfigFromURI(kmsURI)
		if err != nil {
			return nil, fmt.Errorf("invalid AWS KMS URI: %w", err)
		}
		return aws.New(ctx, store, cfg)

	case "azure":
		cfg, err := uri.DecodeAzureConfigFromURI(kmsURI)
		if err != nil {
			return nil, fmt.Errorf("invalid Azure Key Vault URI: %w", err)
		}
		return azure.New(ctx, store, cfg)

	case "gcp":
		cfg, err := uri.DecodeGCPConfigFromURI(kmsURI)
		if err != nil {
			return nil, fmt.Errorf("invalid GCP KMS URI: %w", err)
		}
		return gcp.New(ctx, store, cfg)

	case "cluster-kms":
		cfg, err := uri.DecodeMasterSecretFromURI(kmsURI)
		if err != nil {
			return nil, err
		}
		return cluster.New(cfg.Key, cfg.Salt)

	default:
		return nil, fmt.Errorf("unknown KMS type: %s", url.Host)
	}
}

func getGCPStorageConfig(uri *url.URL) (string, string, error) {
	r, err := getConfig(uri.Query(), []string{"project", "bucket"})
	return r[0], r[1], err
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

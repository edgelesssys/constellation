/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package setup

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/edgelesssys/constellation/v2/kms/internal/storage"
	"github.com/edgelesssys/constellation/v2/kms/kms"
	"github.com/edgelesssys/constellation/v2/kms/kms/aws"
	"github.com/edgelesssys/constellation/v2/kms/kms/azure"
	"github.com/edgelesssys/constellation/v2/kms/kms/cluster"
	"github.com/edgelesssys/constellation/v2/kms/kms/gcp"
)

const (
	AWSKMSURI     = "kms://aws?keyPolicy=%s"
	AzureKMSURI   = "kms://azure-kms?name=%s&type=%s"
	AzureHSMURI   = "kms://azure-hsm?name=%s"
	GCPKMSURI     = "kms://gcp?project=%s&location=%s&keyRing=%s&protectionLvl=%s"
	ClusterKMSURI = "kms://cluster-kms"
	AWSS3URI      = "storage://aws?bucket=%s"
	AzureBlobURI  = "storage://azure?container=%s&connectionString=%s"
	GCPStorageURI = "storage://gcp?projects=%s&bucket=%s"
	NoStoreURI    = "storage://no-store"
)

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
		poliyProducer, err := getAWSKMSConfig(uri)
		if err != nil {
			return nil, err
		}
		return aws.New(ctx, poliyProducer, store)

	case "azure-kms":
		vaultName, vaultType, err := getAzureKMSConfig(uri)
		if err != nil {
			return nil, err
		}
		return azure.New(ctx, vaultName, azure.VaultSuffix(vaultType), store, nil)

	case "azure-hsm":
		vaultName, err := getAzureHSMConfig(uri)
		if err != nil {
			return nil, err
		}
		return azure.NewHSM(ctx, vaultName, store, nil)

	case "gcp":
		project, location, keyRing, protectionLvl, err := getGCPKMSConfig(uri)
		if err != nil {
			return nil, err
		}
		return gcp.New(ctx, project, location, keyRing, store, kmspb.ProtectionLevel(protectionLvl))

	case "cluster-kms":
		salt, err := getClusterKMSConfig(uri)
		if err != nil {
			return nil, err
		}
		return cluster.New(salt), nil

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

func getAWSKMSConfig(uri *url.URL) (*defaultPolicyProducer, error) {
	r, err := getConfig(uri.Query(), []string{"keyPolicy"})
	if err != nil {
		return nil, err
	}
	return &defaultPolicyProducer{policy: r[0]}, err
}

func getAzureKMSConfig(uri *url.URL) (string, string, error) {
	r, err := getConfig(uri.Query(), []string{"name", "type"})
	return r[0], r[1], err
}

func getAzureHSMConfig(uri *url.URL) (string, error) {
	r, err := getConfig(uri.Query(), []string{"name"})
	return r[0], err
}

func getAzureBlobConfig(uri *url.URL) (string, string, error) {
	r, err := getConfig(uri.Query(), []string{"container", "connectionString"})
	if err != nil {
		return "", "", err
	}
	return r[0], r[1], nil
}

func getGCPKMSConfig(uri *url.URL) (string, string, string, int, error) {
	r, err := getConfig(uri.Query(), []string{"project", "location", "keyRing", "protectionLvl"})
	if err != nil {
		return "", "", "", 0, err
	}
	protectionLvl, err := strconv.Atoi(r[3])
	if err != nil {
		return "", "", "", 0, err
	}
	return r[0], r[1], r[2], protectionLvl, nil
}

func getGCPStorageConfig(uri *url.URL) (string, string, error) {
	r, err := getConfig(uri.Query(), []string{"project", "bucket"})
	return r[0], r[1], err
}

func getClusterKMSConfig(uri *url.URL) ([]byte, error) {
	r, err := getConfig(uri.Query(), []string{"salt"})
	if err != nil {
		return nil, err
	}
	return base64.URLEncoding.DecodeString(r[0])
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

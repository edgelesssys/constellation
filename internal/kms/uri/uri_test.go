/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package uri

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestMasterSecretURI(t *testing.T) {
	cfg := MasterSecret{
		Key:  []byte("key"),
		Salt: []byte("salt"),
	}

	checkURI(t, cfg, DecodeMasterSecretFromURI)
}

func TestAWSURI(t *testing.T) {
	cfg := AWSConfig{
		KeyName:     "key",
		Region:      "region",
		AccessKeyID: "accessKeyID",
		AccessKey:   "accessKey",
	}

	checkURI(t, cfg, DecodeAWSConfigFromURI)
}

func TestAWSS3URI(t *testing.T) {
	cfg := AWSS3Config{
		Bucket:      "bucket",
		Region:      "region",
		AccessKeyID: "accessKeyID",
		AccessKey:   "accessKey",
	}

	checkURI(t, cfg, DecodeAWSS3ConfigFromURI)
}

func TestAzureURI(t *testing.T) {
	cfg := AzureConfig{
		KeyName:      "key",
		TenantID:     "tenantID",
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
		VaultName:    "vaultName",
		VaultType:    DefaultCloud,
	}

	checkURI(t, cfg, DecodeAzureConfigFromURI)
}

func TestAzureBlobURI(t *testing.T) {
	cfg := AzureBlobConfig{
		StorageAccount: "accountName",
		Container:      "containerName",
		TenantID:       "tenantID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
	}

	checkURI(t, cfg, DecodeAzureBlobConfigFromURI)
}

func TestGCPURI(t *testing.T) {
	cfg := GCPConfig{
		KeyName:         "key",
		ProjectID:       "project",
		Location:        "location",
		KeyRing:         "keyRing",
		CredentialsPath: "/path/to/credentials",
	}

	checkURI(t, cfg, DecodeGCPConfigFromURI)
}

func TestGoogleCloudStorageURI(t *testing.T) {
	cfg := GoogleCloudStorageConfig{
		ProjectID:       "project",
		Bucket:          "bucket",
		CredentialsPath: "/path/to/credentials",
	}

	checkURI(t, cfg, DecodeGoogleCloudStorageConfigFromURI)
}

type cfgStruct interface {
	EncodeToURI() string
}

func checkURI[T any](t *testing.T, cfg cfgStruct, decodeFunc func(string) (T, error)) {
	t.Helper()
	require := require.New(t)
	assert := assert.New(t)

	uri := cfg.EncodeToURI()
	decoded, err := decodeFunc(uri)
	require.NoError(err, "failed to decode URI to config: %s", uri)

	assert.Equal(cfg, decoded, "decoded config does not match original config")
}

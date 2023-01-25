//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package test

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/kms/kms/azure"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/azureblob"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/memfs"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureStorage(t *testing.T) {
	if !*runAzStorage {
		t.Skip("Skipping Azure storage test")
	}
	if *azConnectionString == "" || *azContainer == "" {
		flag.Usage()
		t.Fatal("Required flags not set: --az-connection-string, --az-container")
	}

	assert := assert.New(t)
	require := require.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	store, err := azureblob.New(ctx, *azConnectionString, *azContainer, nil)
	require.NoError(err)

	testData := []byte("Constellation test data")
	testName := "constellation-test"

	err = store.Put(ctx, testName, testData)
	assert.NoError(err)

	got, err := store.Get(ctx, testName)
	assert.NoError(err)
	assert.Equal(testData, got)

	_, err = store.Get(ctx, addSuffix("does-not-exist"))
	assert.ErrorIs(err, storage.ErrDEKUnset)
}

func TestAzureKeyKMS(t *testing.T) {
	if !*runAzKms {
		t.Skip("Skipping Azure Key Vault test")
	}

	if *kekID == "" || *azClientID == "" || *azClientSecret == "" || *azTenantID == "" || *azVaultName == "" {
		flag.Usage()
		t.Fatal("Required flags not set: --az-tenant-id, --az-client-id, --az-client-secret, --az-vault-name, --kek-id")
	}
	require := require.New(t)

	store := memfs.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cfg := uri.AzureConfig{
		TenantID:     *azTenantID,
		ClientID:     *azClientID,
		ClientSecret: *azClientSecret,
		VaultName:    *azVaultName,
		VaultType:    uri.DefaultCloud,
		KeyName:      *kekID,
	}
	kmsClient, err := azure.New(ctx, store, cfg)
	require.NoError(err)

	runKMSTest(t, kmsClient)
}

func TestAzureKeyHSM(t *testing.T) {
	if !*runAzHsm {
		t.Skip("Skipping Azure HSM test")
	}

	if *kekID == "" || *azClientID == "" || *azClientSecret == "" || *azTenantID == "" || *azVaultName == "" {
		flag.Usage()
		t.Fatal("Required flags not set: --az-tenant-id, --az-client-id, --az-client-secret, --az-vault-name, --kek-id")
	}
	require := require.New(t)

	store := memfs.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cfg := uri.AzureConfig{
		TenantID:     *azTenantID,
		ClientID:     *azClientID,
		ClientSecret: *azClientSecret,
		VaultName:    *azVaultName,
		VaultType:    uri.HSMDefaultCloud,
		KeyName:      *kekID,
	}
	kmsClient, err := azure.New(ctx, store, cfg)
	require.NoError(err)

	runKMSTest(t, kmsClient)
}

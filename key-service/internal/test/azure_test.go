//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package test

import (
	"context"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/kms/config"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/azure"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	azVaultName     = "edgeless-constellation-t"
	azHSMName       = "edgeless-constellation-h"
	azContainerName = "constellation-test-storage"
)

func TestAzureStorage(t *testing.T) {
	if !*runAzStorage {
		t.Skip("Skipping Azure storage test")
	}
	if *azConnectionString == "" {
		t.Fatal("Connection string for Azure storage must be set using the '-azStorageConn' flag")
	}
	assert := assert.New(t)
	require := require.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	store, err := storage.NewAzureStorage(ctx, *azConnectionString, azContainerName, nil)
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

func TestAzureKeyVault(t *testing.T) {
	if !*runAzKms {
		t.Skip("Skipping Azure Key Vault test")
	}

	assert := assert.New(t)
	require := require.New(t)

	store := storage.NewMemMapStorage()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	kekName := "test-kek"
	client, err := azure.New(ctx, azVaultName, azure.DefaultCloud, store, kekName, nil)
	require.NoError(err)

	dekName := "test-dek"

	assert.NoError(client.CreateKEK(ctx, kekName, nil))

	res, err := client.GetDEK(ctx, dekName, config.SymmetricKeyLength)
	assert.NoError(err)

	res2, err := client.GetDEK(ctx, dekName, config.SymmetricKeyLength)
	assert.NoError(err)
	assert.Equal(res, res2)

	res3, err := client.GetDEK(ctx, addSuffix(dekName), config.SymmetricKeyLength)
	assert.NoError(err)
	assert.Len(res3, config.SymmetricKeyLength)
	assert.NotEqual(res, res3)
}

func TestAzureHSM(t *testing.T) {
	if !*runAzHsm {
		t.Skip("Skipping Azure HSM test")
	}

	// This test requires an actively running Azure HSM
	// Since the HSMs are quiet expensive, you will have to create one manually to run this test
	// See: https://docs.microsoft.com/en-us/azure/key-vault/managed-hsm/quick-create-cli
	// Don't forget to remove the HSM after testing: az keyvault purge --hsm-name <HSM>
	assert := assert.New(t)
	require := require.New(t)

	store := storage.NewMemMapStorage()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	kekName := "test-kek"
	client, err := azure.NewHSM(ctx, azHSMName, store, kekName, nil)
	require.NoError(err)

	dekName := "test-dek"
	importedKek := "test-kek-import"
	kekData := []byte{0x52, 0xFD, 0xFC, 0x07, 0x21, 0x82, 0x65, 0x4F, 0x16, 0x3F, 0x5F, 0x0F, 0x9A, 0x62, 0x1D, 0x72, 0x95, 0x66, 0xC7, 0x4D, 0x10, 0x03, 0x7C, 0x4D, 0x7B, 0xBB, 0x04, 0x07, 0xD1, 0xE2, 0xC6, 0x49}

	assert.NoError(client.CreateKEK(ctx, importedKek, kekData))

	assert.NoError(client.CreateKEK(ctx, kekName, nil))

	res, err := client.GetDEK(ctx, dekName, config.SymmetricKeyLength)
	require.NoError(err)
	assert.NotNil(res)

	res2, err := client.GetDEK(ctx, dekName, config.SymmetricKeyLength)
	require.NoError(err)
	assert.Equal(res, res2)

	res3, err := client.GetDEK(ctx, addSuffix(dekName), config.SymmetricKeyLength)
	require.NoError(err)
	assert.Len(res3, config.SymmetricKeyLength)
	assert.NotEqual(res, res3)
}

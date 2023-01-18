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

	"github.com/edgelesssys/constellation/v2/internal/kms/config"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/gcp"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGCPKMS(t *testing.T) {
	if !*runGcpKms {
		t.Skip("Skipping Google KMS key creation test")
	}
	assert := assert.New(t)
	require := require.New(t)
	store := storage.NewMemMapStorage()

	if *gcpProjectID == "" || *gcpLocation == "" || *gcpKeyRing == "" || *gcpKEKID == "" {
		flag.Usage()
		t.Fatal("Required flags not set")
	}

	dekName := "test-dek"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	kmsClient, err := gcp.New(ctx, *gcpKEKID, store, *gcpProjectID, *gcpLocation, *gcpKeyRing)
	require.NoError(err)
	defer kmsClient.Close()

	res, err := kmsClient.GetDEK(ctx, dekName, config.SymmetricKeyLength)
	require.NoError(err)
	t.Logf("DEK 1: %x\n", res)

	res2, err := kmsClient.GetDEK(ctx, dekName, config.SymmetricKeyLength)
	require.NoError(err)
	assert.Equal(res, res2)
	t.Logf("DEK 2: %x\n", res2)

	res3, err := kmsClient.GetDEK(ctx, addSuffix(dekName), config.SymmetricKeyLength)
	require.NoError(err)
	assert.Len(res3, config.SymmetricKeyLength)
	assert.NotEqual(res, res3)
	t.Logf("DEK 3: %x\n", res3)
}

func TestGcpStorage(t *testing.T) {
	if !*runGcpStorage {
		t.Skip("Skipping Google Storage test")
	}

	if *gcpProjectID == "" || *gcpBucket == "" {
		flag.Usage()
		t.Fatal("Required flags not set")
	}

	assert := assert.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	store, err := storage.NewGoogleCloudStorage(ctx, *gcpProjectID, *gcpBucket, nil)
	assert.NoError(err)

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

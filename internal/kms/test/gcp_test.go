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

	"github.com/edgelesssys/constellation/v2/internal/kms/kms/gcp"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/gcs"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/memfs"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGCPKMS(t *testing.T) {
	if !*runGcpKms {
		t.Skip("Skipping Google KMS test")
	}
	if *gcpProjectID == "" || *gcpLocation == "" || *gcpKeyRing == "" || *kekID == "" {
		flag.Usage()
		t.Fatal("Required flags not set: --gcp-project, --gcp-location, --gcp-keyring, --kek-id")
	}
	require := require.New(t)

	store := memfs.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cfg := uri.GCPConfig{
		CredentialsPath: *gcpCredentialsPath,
		ProjectID:       *gcpProjectID,
		Location:        *gcpLocation,
		KeyRing:         *gcpKeyRing,
		KeyName:         *kekID,
	}
	kmsClient, err := gcp.New(ctx, store, cfg)
	require.NoError(err)
	defer kmsClient.Close()

	runKMSTest(t, kmsClient)
}

func TestGcpStorage(t *testing.T) {
	if !*runGcpStorage {
		t.Skip("Skipping Google Storage test")
	}

	if *gcpProjectID == "" || *gcpBucket == "" {
		flag.Usage()
		t.Fatal("Required flags not set: --gcp-project, --gcp-bucket ")
	}

	assert := assert.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	store, err := gcs.New(ctx, *gcpProjectID, *gcpBucket, nil)
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

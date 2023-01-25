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
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/gcs"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/memfs"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
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
	if *gcpProjectID == "" || *gcpBucket == "" || *gcpCredentialsPath == "" {
		flag.Usage()
		t.Fatal("Required flags not set: --gcp-project, --gcp-bucket ")
	}
	require := require.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cfg := uri.GoogleCloudStorageConfig{
		CredentialsPath: *gcpCredentialsPath,
		ProjectID:       *gcpProjectID,
		Bucket:          *gcpBucket,
	}
	store, err := gcs.New(ctx, cfg)
	require.NoError(err)

	runStorageTest(t, store)
}

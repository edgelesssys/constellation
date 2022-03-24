//go:build integration
// +build integration

package test

import (
	"context"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/kms/config"
	"github.com/edgelesssys/constellation/kms/kms/gcp"
	"github.com/edgelesssys/constellation/kms/storage"
	"github.com/stretchr/testify/assert"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

const (
	gcpBucket    = "constellation-test-bucket"
	gcpProjectID = "constellation-kms-integration-test"
	gcpKeyRing   = "test-ring"
	gcpLocation  = "global"
)

func TestCreateGcpKEK(t *testing.T) {
	if !*runGcpKms {
		t.Skip("Skipping Google KMS key creation test")
	}
	assert := assert.New(t)
	store := storage.NewMemMapStorage()

	kekName := addSuffix("test-kek")
	dekName := "test-dek"

	kmsClient := gcp.New(gcpProjectID, gcpLocation, gcpKeyRing, store, kmspb.ProtectionLevel_SOFTWARE)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	// Key name is random, but there is a chance we try to create a key that already exists, in that case the test fails
	assert.NoError(kmsClient.CreateKEK(ctx, kekName, nil))

	res, err := kmsClient.GetDEK(ctx, kekName, dekName, config.SymmetricKeyLength)
	assert.NoError(err)

	res2, err := kmsClient.GetDEK(ctx, kekName, dekName, config.SymmetricKeyLength)
	assert.NoError(err)
	assert.Equal(res, res2)

	res3, err := kmsClient.GetDEK(ctx, kekName, addSuffix(dekName), config.SymmetricKeyLength)
	assert.NoError(err)
	assert.Len(res3, config.SymmetricKeyLength)
	assert.NotEqual(res, res3)
}

func TestImportGcpKEK(t *testing.T) {
	if !*runGcpKms {
		t.Skip("Skipping Google KMS key import test")
	}
	assert := assert.New(t)
	store := storage.NewMemMapStorage()

	kekName := addSuffix("test-kek")
	kekData := []byte{0x52, 0xFD, 0xFC, 0x07, 0x21, 0x82, 0x65, 0x4F, 0x16, 0x3F, 0x5F, 0x0F, 0x9A, 0x62, 0x1D, 0x72, 0x95, 0x66, 0xC7, 0x4D, 0x10, 0x03, 0x7C, 0x4D, 0x7B, 0xBB, 0x04, 0x07, 0xD1, 0xE2, 0xC6, 0x49}
	dekName := "test-dek"

	kmsClient := gcp.New(gcpProjectID, gcpLocation, gcpKeyRing, store, kmspb.ProtectionLevel_SOFTWARE)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	assert.NoError(kmsClient.CreateKEK(ctx, kekName, kekData))

	res, err := kmsClient.GetDEK(ctx, kekName, dekName, config.SymmetricKeyLength)
	assert.NoError(err)

	res2, err := kmsClient.GetDEK(ctx, kekName, dekName, config.SymmetricKeyLength)
	assert.NoError(err)
	assert.Equal(res, res2)
}

func TestGcpStorage(t *testing.T) {
	if !*runGcpStorage {
		t.Skip("Skipping Google Storage test")
	}

	assert := assert.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	store, err := storage.NewGoogleCloudStorage(ctx, gcpProjectID, gcpBucket, nil)
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

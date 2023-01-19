//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package test provides integration tests for KMS and storage backends.
package test

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/kms/config"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	runAwsStorage = flag.Bool("awsStorage", false, "set to run AWS S3 Bucket Storage test")
	runAwsKms     = flag.Bool("awsKms", false, "set to run AWS KMS test")

	azConnectionString = flag.String("azStorageConn", "", "Connection string for Azure storage account. Required for Azure storage test.")
	azContainer        = flag.String("azContainer", "constellation-test-storage", "Container to save test data to. Required for Azure storage test.")
	runAzStorage       = flag.Bool("azStorage", false, "set to run Azure Storage test")
	runAzKms           = flag.Bool("azKms", false, "set to run Azure KMS test")
	azVaultName        = flag.String("azVaultName", "", "Name of the Azure Key Vault to use. Required for Azure KMS/HSM test.")
	azKEKID            = flag.String("azKEKID", "", "ID of the key to use for Azure KMS test. Required for Azure KMS/HSM test.")
	azTenantID         = flag.String("azTenantID", "", "Tenant ID to use for Azure tests. Required for Azure KMS/HSM test.")
	azClientID         = flag.String("azClientID", "", "Client ID to use for Azure tests. Required for Azure KMS/HSM test.")
	azClientSecret     = flag.String("azClientSecret", "", "Client secret to use for Azure tests. Required for Azure KMS/HSM test.")

	runGcpKms          = flag.Bool("gcpKms", false, "set to run Google KMS test")
	runGcpStorage      = flag.Bool("gcpStorage", false, "set to run Google Storage test")
	gcpCredentialsPath = flag.String("gcpCredentialsPath", "", "Path to a credentials file. Optional for Google KMS and Google storage test.")
	gcpBucket          = flag.String("gcpBucket", "", "Bucket to save test data to. Required for Google Storage test.")
	gcpProjectID       = flag.String("gcpProjectID", "", "Project ID to use for Google tests. Required for Google KMS and Google storage test.")
	gcpKeyRing         = flag.String("gcpKeyRing", "", "Key ring to use for Google KMS test. Required for Google KMS test.")
	gcpLocation        = flag.String("gcpLocation", "global", "Location of the keyring. Required for Google KMS test.")
	gcpKEKID           = flag.String("gcpKEKID", "", "ID of the key to use for Google KMS test. Required for Google KMS test.")
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	flag.Parse()
	os.Exit(m.Run())
}

func runKMSTest(t *testing.T, kms kms.CloudKMS) {
	assert := assert.New(t)
	require := require.New(t)

	dekName := "test-dek"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	res, err := kms.GetDEK(ctx, dekName, config.SymmetricKeyLength)
	require.NoError(err)
	t.Logf("DEK 1: %x\n", res)

	res2, err := kms.GetDEK(ctx, dekName, config.SymmetricKeyLength)
	require.NoError(err)
	assert.Equal(res, res2)
	t.Logf("DEK 2: %x\n", res2)

	res3, err := kms.GetDEK(ctx, addSuffix(dekName), config.SymmetricKeyLength)
	require.NoError(err)
	assert.Len(res3, config.SymmetricKeyLength)
	assert.NotEqual(res, res3)
	t.Logf("DEK 3: %x\n", res3)
}

func addSuffix(s string) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 5)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return s + "-" + string(b)
}

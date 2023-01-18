//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package test provides integration tests for KMS and storage backends.
package test

import (
	"flag"
	"math/rand"
	"os"
	"testing"
	"time"
)

var (
	runAwsStorage = flag.Bool("awsStorage", false, "set to run AWS S3 Bucket Storage test")
	runAwsKms     = flag.Bool("awsKms", false, "set to run AWS KMS test")

	azConnectionString = flag.String("azStorageConn", "", "Connection string for Azure storage account. Required for Azure storage test.")
	runAzStorage       = flag.Bool("azStorage", false, "set to run Azure Storage test")
	runAzKms           = flag.Bool("azKms", false, "set to run Azure KMS test")
	runAzHsm           = flag.Bool("azHsm", false, "set to run Azure HSM test")

	runGcpKms     = flag.Bool("gcpKms", false, "set to run Google KMS test")
	runGcpStorage = flag.Bool("gcpStorage", false, "set to run Google Storage test")
	gcpBucket     = flag.String("gcpBucket", "", "Bucket to save test data to. Required for Google Storage test.")
	gcpProjectID  = flag.String("gcpProjectID", "", "Project ID to use for Google tests. Required for Google KMS and Google storage test.")
	gcpKeyRing    = flag.String("gcpKeyRing", "", "Key ring to use for Google KMS test. Required for Google KMS test.")
	gcpLocation   = flag.String("gcpLocation", "global", "Location of the keyring. Required for Google KMS test.")
	gcpKEKID      = flag.String("gcpKEKID", "", "ID of the key to use for Google KMS test. Required for Google KMS test.")
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	flag.Parse()
	os.Exit(m.Run())
}

func addSuffix(s string) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 5)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return s + "-" + string(b)
}

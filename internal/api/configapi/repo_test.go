//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package configapi_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/stretchr/testify/require"
)

const (
	awsBucket = "cdn-constellation-backend"
	awsRegion = "eu-central-1"
)

var cfg uri.AWSS3Config

var (
	awsAccessKeyID  = flag.String("aws-access-key-id", "", "ID of the Access key to use for AWS tests. Required for AWS KMS and storage test.")
	awsAccessKey    = flag.String("aws-access-key", "", "Access key to use for AWS tests. Required for AWS KMS and storage test.")
	cosignPwd       = flag.String("cosign-pwd", "", "Password to decrypt the cosign private key. Required for signing.")
	priviateKeyPath = flag.String("private-key", "", "Path to the private key used for signing. Required for signing.")
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *awsAccessKey == "" || *awsAccessKeyID == "" || *cosignPwd == "" || *priviateKeyPath == "" {
		flag.Usage()
		fmt.Println("Required flags not set: --aws-access-key, --aws-access-key-id, --aws-bucket, --aws-region. Skipping tests.")
		os.Exit(0)
	}
	cfg = uri.AWSS3Config{
		AccessKeyID: *awsAccessKeyID,
		AccessKey:   *awsAccessKey,
		Bucket:      awsBucket,
		Region:      awsRegion,
	}
	os.Exit(m.Run())
}

var versionValues = configapi.AzureSEVSNPVersion{
	Bootloader: 2,
	TEE:        0,
	SNP:        6,
	Microcode:  93,
}

func TestUploadAzureSEVSNPVersions(t *testing.T) {
	ctx := context.Background()
	sut, err := configapi.NewAttestationVersionRepo(ctx, cfg, []byte(*cosignPwd), []byte(*priviateKeyPath))
	require.NoError(t, err)
	d := time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC)
	require.NoError(t, sut.UploadAzureSEVSNP(ctx, versionValues, d))
}

func TestListVersions(t *testing.T) {
	ctx := context.Background()

	sut, err := configapi.NewAttestationVersionRepo(ctx, cfg, []byte(*cosignPwd), []byte(*priviateKeyPath))
	require.NoError(t, err)

	err = sut.DeleteList(ctx, variant.AzureSEVSNP{})
	require.NoError(t, err)

	res, err := sut.List(ctx, variant.AzureSEVSNP{})
	require.NoError(t, err)
	require.Equal(t, []string{}, res)

	d := time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC)
	err = sut.UploadAzureSEVSNP(ctx, versionValues, d)
	require.NoError(t, err)
	res, err = sut.List(ctx, variant.AzureSEVSNP{})
	require.NoError(t, err)
	require.Equal(t, []string{"2021-01-01-01-01.json"}, res)

	err = sut.DeleteList(ctx, variant.AzureSEVSNP{})
	require.NoError(t, err)
}

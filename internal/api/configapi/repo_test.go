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

var (
	awsRegion      = flag.String("aws-region", "us-east-1", "Region to use for AWS tests. Required for AWS KMS test.")
	awsAccessKeyID = flag.String("aws-access-key-id", "", "ID of the Access key to use for AWS tests. Required for AWS KMS and storage test.")
	awsAccessKey   = flag.String("aws-access-key", "", "Access key to use for AWS tests. Required for AWS KMS and storage test.")
	awsBucket      = flag.String("aws-bucket", "", "Name of the S3 bucket to use for AWS storage test. Required for AWS storage test.")
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *awsAccessKey == "" || *awsAccessKeyID == "" || *awsBucket == "" || *awsRegion == "" {
		flag.Usage()
		fmt.Println("Required flags not set: --aws-access-key, --aws-access-key-id, --aws-bucket, --aws-region. Skipping tests.")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

var cfg = uri.AWSS3Config{
	Bucket:      *awsBucket,
	AccessKeyID: *awsAccessKeyID,
	AccessKey:   *awsAccessKey,
	Region:      *awsRegion,
}

func newVersion(version uint8) configapi.AttestationVersion {
	return configapi.AttestationVersion{
		Value:    version,
		IsLatest: false,
	}
}

var versionValues = configapi.AzureSEVSNPVersion{
	Bootloader: newVersion(2),
	TEE:        newVersion(0),
	SNP:        newVersion(6),
	Microcode:  newVersion(93),
}

func TestUploadAzureSEVSNPVersions(t *testing.T) {
	ctx := context.Background()
	sut, err := configapi.NewAttestationVersionRepo(ctx, cfg)
	require.NoError(t, err)
	d := time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC)
	require.NoError(t, sut.UploadAzureSEVSNP(ctx, versionValues, d))
}

func TestListVersions(t *testing.T) {
	ctx := context.Background()

	sut, err := configapi.NewAttestationVersionRepo(ctx, cfg)
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

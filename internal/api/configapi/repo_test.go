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
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/stretchr/testify/require"
)

const (
	awsBucket   = "cdn-constellation-backend"
	awsRegion   = "eu-central-1"
	envAwsKeyID = "AWS_ACCESS_KEY_ID"
	envAwsKey   = "AWS_ACCESS_KEY"
)

var cfg staticupload.Config

var (
	cosignPwd      = flag.String("cosign-pwd", "", "Password to decrypt the cosign private key. Required for signing.")
	privateKeyPath = flag.String("private-key", "", "Path to the private key used for signing. Required for signing.")
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *cosignPwd == "" || *privateKeyPath == "" {
		flag.Usage()
		fmt.Println("Required flags not set: --cosign-pwd, --private-key. Skipping tests.")
		os.Exit(1)
	}
	if _, present := os.LookupEnv(envAwsKey); !present {
		fmt.Printf("%s not set. Skipping tests.\n", envAwsKey)
		os.Exit(1)
	}
	if _, present := os.LookupEnv(envAwsKeyID); !present {
		fmt.Printf("%s not set. Skipping tests.\n", envAwsKeyID)
		os.Exit(1)
	}
	cfg = staticupload.Config{
		Bucket: awsBucket,
		Region: awsRegion,
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
	sut, err := configapi.NewAttestationVersionRepo(ctx, cfg, []byte(*cosignPwd), []byte(*privateKeyPath))
	require.NoError(t, err)
	d := time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC)
	require.NoError(t, sut.UploadAzureSEVSNP(ctx, versionValues, d))
}

func TestListVersions(t *testing.T) {
	ctx := context.Background()

	sut, err := configapi.NewAttestationVersionRepo(ctx, cfg, []byte(*cosignPwd), []byte(*privateKeyPath))
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

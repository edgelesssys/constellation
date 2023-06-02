//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package client_test

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig/client"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
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
	privateKey     []byte
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
	file, _ := os.Open(*privateKeyPath)
	var err error
	privateKey, err = io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

var versionValues = attestationconfig.AzureSEVSNPVersion{
	Bootloader: 2,
	TEE:        0,
	SNP:        6,
	Microcode:  93,
}

func TestUploadAzureSEVSNPVersions(t *testing.T) {
	ctx := context.Background()
	client, err := client.New(ctx, cfg, []byte(*cosignPwd), privateKey)
	require.NoError(t, err)
	d := time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC)
	require.NoError(t, client.UploadAzureSEVSNP(ctx, versionValues, d))
}

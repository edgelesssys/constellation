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

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/aws"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/awss3"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/memfs"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/stretchr/testify/require"
)

func TestAwsStorage(t *testing.T) {
	if !*runAwsStorage {
		t.Skip("Skipping AWS storage test")
	}
	if *awsAccessKey == "" || *awsAccessKeyID == "" || *awsBucket == "" || *awsRegion == "" {
		flag.Usage()
		t.Fatal("Required flags not set: --aws-access-key, --aws-access-key-id, --aws-bucket, --aws-region")
	}
	require := require.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// create bucket
	cfg := uri.AWSS3Config{
		Bucket:      *awsBucket,
		AccessKeyID: *awsAccessKeyID,
		AccessKey:   *awsAccessKey,
		Region:      *awsRegion,
	}
	store, err := awss3.New(ctx, cfg)
	require.NoError(err)

	runStorageTest(t, store)

	cleanUpBucket(ctx, require, *awsBucket, *awsRegion)
}

func cleanUpBucket(ctx context.Context, require *require.Assertions, bucketID, awsRegion string) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	require.NoError(err)
	client := s3.NewFromConfig(cfg)

	// List all objects of the bucket
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: &bucketID,
	}
	output, err := client.ListObjectsV2(ctx, listObjectsInput)
	require.NoError(err)
	var objects []string
	var i int32
	for i = 0; i < *output.KeyCount; i++ {
		objects = append(objects, *output.Contents[i].Key)
	}
	// Delete all objects of the bucket
	require.NoError(cleanUpObjects(ctx, client, bucketID, objects))

	// Delete the bucket
	deleteBucketInput := &s3.DeleteBucketInput{
		Bucket: &bucketID,
	}
	_, err = client.DeleteBucket(ctx, deleteBucketInput)
	require.NoError(err)
}

func cleanUpObjects(ctx context.Context, client *s3.Client, bucketID string, objectsToDelete []string) error {
	var objectsIdentifier []types.ObjectIdentifier
	for _, object := range objectsToDelete {
		objectsIdentifier = append(objectsIdentifier, types.ObjectIdentifier{Key: func(s string) *string { return &s }(object)})
	}
	deleteObjectsInput := &s3.DeleteObjectsInput{
		Bucket: &bucketID,
		Delete: &types.Delete{Objects: objectsIdentifier},
	}
	_, err := client.DeleteObjects(ctx, deleteObjectsInput)
	return err
}

func TestAwsKms(t *testing.T) {
	if !*runAwsKms {
		t.Skip("Skipping AWS KMS test")
	}

	if *kekID == "" || *awsAccessKeyID == "" || *awsAccessKey == "" || *awsRegion == "" {
		flag.Usage()
		t.Fatal("Required flags not set: --aws-access-key-id, --aws-access-key, --aws-region, --kek-id")
	}

	require := require.New(t)

	store := memfs.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cfg := uri.AWSConfig{
		KeyName:     *kekID,
		Region:      *awsRegion,
		AccessKeyID: *awsAccessKeyID,
		AccessKey:   *awsAccessKey,
	}
	kmsClient, err := aws.New(ctx, store, cfg)
	require.NoError(err)

	runKMSTest(t, kmsClient)
}

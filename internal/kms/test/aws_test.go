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
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/awss3"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/memfs"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/stretchr/testify/assert"
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

	assert := assert.New(t)
	require := require.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// create bucket
	cfg := uri.AWSS3Config{
		Bucket:      *awsBucket,
		AccessKeyID: *awsAccessKey,
		AccessKey:   *awsAccessKeyID,
		Region:      *awsRegion,
	}
	store, err := awss3.New(ctx, cfg)
	require.NoError(err)

	testDEK1 := []byte("test DEK")
	testDEK2 := []byte("more test DEK")

	// request unset value
	_, err = store.Get(ctx, "test:input")
	assert.Error(err)

	// test Put method
	require.NoError(store.Put(ctx, "volume01", testDEK1))
	require.NoError(store.Put(ctx, "volume02", testDEK2))

	// make sure values have been set
	val, err := store.Get(ctx, "volume01")
	require.NoError(err)
	assert.Equal(testDEK1, val)
	val, err = store.Get(ctx, "volume02")
	require.NoError(err)
	assert.Equal(testDEK2, val)

	_, err = store.Get(ctx, "invalid:key")
	assert.Error(err)
	assert.ErrorIs(err, storage.ErrDEKUnset)

	cleanUpBucket(ctx, require, *awsBucket, func(*s3.Options) {})
}

func cleanUpBucket(ctx context.Context, require *require.Assertions, bucketID string, optFns ...func(*s3.Options)) {
	cfg, err := config.LoadDefaultConfig(ctx)
	require.NoError(err)
	client := s3.NewFromConfig(cfg, optFns...)

	// List all objects of the bucket
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: &bucketID,
	}
	output, _ := client.ListObjectsV2(ctx, listObjectsInput)
	var objects []string
	var i int32
	for i = 0; i < output.KeyCount; i++ {
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

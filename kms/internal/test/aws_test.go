//go:build integration
// +build integration

package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	kmsconfig "github.com/edgelesssys/constellation/kms/internal/config"
	"github.com/edgelesssys/constellation/kms/internal/storage"
	awsInterface "github.com/edgelesssys/constellation/kms/kms/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultPolicyWithDecryption = `{
	"Version" : "2012-10-17",
	"Id" : "updated-policy",
	"Statement" : [ {
	  "Sid" : "Enable IAM User Permissions",
	  "Effect" : "Allow",
	  "Principal" : {
		"AWS" : "*"
	  },
	  "Action" : [
		  "kms:CreateAlias",
		  "kms:CreateKey",
		  "kms:Decrypt",
		  "kms:DeleteAlias",
		  "kms:DescribeKey",
		  "kms:Encrypt",
		  "kms:GenerateDataKey",
		  "kms:GenerateDataKeyWithoutPlaintext",
		  "kms:GetKeyPolicy",
		  "kms:GetParametersForImport",
		  "kms:ImportKeyMaterial",
		  "kms:PutKeyPolicy",
		  "kms:ScheduleKeyDeletion"
		],
	  "Resource" : "*"
	} ]
  }`

func TestAwsStorage(t *testing.T) {
	if !*runAwsStorage {
		t.Skip("Skipping AWS storage test")
	}
	assert := assert.New(t)
	require := require.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	bucketID := strings.ToLower(addSuffix("test-bucket"))

	// create bucket
	store, err := storage.NewAWSS3Storage(ctx, bucketID, func(*s3.Options) {})
	require.NoError(err)

	testDEK1 := []byte("test DEK")
	testDEK2 := []byte("more test DEK")

	// request unset value
	_, err = store.Get(ctx, "test:input")
	assert.Error(err)

	// test Put method
	assert.NoError(store.Put(ctx, "volume01", testDEK1))
	assert.NoError(store.Put(ctx, "volume02", testDEK2))

	// make sure values have been set
	val, err := store.Get(ctx, "volume01")
	assert.NoError(err)
	assert.Equal(testDEK1, val)
	val, err = store.Get(ctx, "volume02")
	assert.NoError(err)
	assert.Equal(testDEK2, val)

	_, err = store.Get(ctx, "invalid:key")
	assert.Error(err)
	assert.ErrorIs(err, storage.ErrDEKUnset)

	cleanUpBucket(ctx, require, bucketID, func(*s3.Options) {})
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
		objectsIdentifier = append(objectsIdentifier, types.ObjectIdentifier{Key: aws.String(object)})
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
	require := require.New(t)
	assert := assert.New(t)

	newKEKId1 := addSuffix("test-key-")
	newKEKId2 := addSuffix("test-key-")
	require.NotEqual(newKEKId1, newKEKId2)
	var keyPolicyProducer createKeyPolicyFunc

	client, err := awsInterface.New(context.Background(), &keyPolicyProducer, nil)
	require.NoError(err)

	privateKEK1 := []byte(strings.Repeat("1234", 8))
	privateKEK2 := []byte(strings.Repeat("5678", 8))
	privateKEK3 := make([]byte, 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// test setting first KEK
	assert.NoError(client.CreateKEK(ctx, newKEKId1, privateKEK1))

	// make sure that CreateKEK is idempotent
	assert.NoError(client.CreateKEK(ctx, newKEKId1, privateKEK1))

	// make sure you can not overwrite KEK with different key material
	assert.NoError(client.CreateKEK(ctx, newKEKId1, privateKEK2))

	// make sure that GetDEK is idempotent
	volumeKey1, err := client.GetDEK(ctx, newKEKId1, "volume01", kmsconfig.SymmetricKeyLength)
	require.NoError(err)
	volumeKey1Copy, err := client.GetDEK(ctx, newKEKId1, "volume01", kmsconfig.SymmetricKeyLength)
	require.NoError(err)
	assert.Equal(volumeKey1, volumeKey1Copy)

	// test setting a second DEK
	volumeKey2, err := client.GetDEK(ctx, newKEKId1, "volume02", kmsconfig.SymmetricKeyLength)
	require.NoError(err)
	assert.NotEqual(volumeKey1, volumeKey2)

	// make sure AWS KMS generates KEK when calling CreateKEK with empty key
	assert.NoError(client.CreateKEK(ctx, newKEKId2, privateKEK3))

	// make sure that CreateKEK is idempotent
	assert.NoError(client.CreateKEK(ctx, newKEKId2, privateKEK3))

	// test setting a DEK with AWS KMS generated KEK
	volumeKey3, err := client.GetDEK(ctx, newKEKId2, "volume03", kmsconfig.SymmetricKeyLength)
	require.NoError(err)
	assert.NotEqual(volumeKey1, volumeKey3)

	cleanUp(ctx, assert, require, newKEKId1)
	cleanUp(ctx, assert, require, newKEKId2)
}

func cleanUp(ctx context.Context, assert *assert.Assertions, require *require.Assertions, alias string) {
	cfg, err := config.LoadDefaultConfig(ctx)
	require.NoError(err)
	awsClient := kms.NewFromConfig(cfg)
	require.NotNil(awsClient)

	describeKeyInput := &kms.DescribeKeyInput{
		KeyId: aws.String("alias/" + alias),
	}
	describeKeyOutput, err := awsClient.DescribeKey(ctx, describeKeyInput)
	assert.NoError(err)
	deleteAliasInput := &kms.DeleteAliasInput{
		AliasName: aws.String("alias/" + alias),
	}
	_, err = awsClient.DeleteAlias(ctx, deleteAliasInput)
	assert.NoError(err)
	scheduleKeyDeletionInput := &kms.ScheduleKeyDeletionInput{
		KeyId:               describeKeyOutput.KeyMetadata.KeyId,
		PendingWindowInDays: aws.Int32(7),
	}
	_, err = awsClient.ScheduleKeyDeletion(ctx, scheduleKeyDeletionInput)
	assert.NoError(err)
}

type createKeyPolicyFunc struct{}

func (fn *createKeyPolicyFunc) CreateKeyPolicy(keyID string) (string, error) {
	policy := defaultPolicyWithDecryption
	policy = strings.Replace(policy, "<AWSAccountId>", "795746500882", 2)
	return policy, nil
}

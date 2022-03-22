package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/kms/pkg/config"
)

// AWSS3Storage is an implementation of the Storage interface, storing keys in AWS S3 buckets.
type AWSS3Storage struct {
	bucketID string
	client   *s3.Client
	optFns   []func(*s3.Options)
}

// NewAWSS3Storage creates a Storage client for AWS S3: https://aws.amazon.com/s3/
//
// You need to provide credentials to authenticate to AWS using the cfg parameter.
func NewAWSS3Storage(ctx context.Context, bucketID string, optFns ...func(*s3.Options)) (*AWSS3Storage, error) {
	// Create S3 client
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, optFns...)

	// Try to create new bucket, continue if bucket already exists
	createBucketInput := &s3.CreateBucketInput{
		Bucket: &bucketID,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(cfg.Region),
		},
	}
	_, err = client.CreateBucket(ctx, createBucketInput, optFns...)
	if err != nil {
		var bne *types.BucketAlreadyExists
		var baowby *types.BucketAlreadyOwnedByYou
		if !(errors.As(err, &bne) || errors.As(err, &baowby)) {
			return nil, fmt.Errorf("creating storage container: %w", err)
		}
	}
	return &AWSS3Storage{client: client, bucketID: bucketID, optFns: optFns}, nil
}

// Get returns a DEK from from AWS S3 Storage by key ID.
func (s *AWSS3Storage) Get(ctx context.Context, keyID string) ([]byte, error) {
	getObjectInput := &s3.GetObjectInput{
		Bucket: &s.bucketID,
		Key:    &keyID,
	}
	output, err := s.client.GetObject(ctx, getObjectInput, s.optFns...)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, ErrDEKUnset
		}
		return nil, fmt.Errorf("downloading DEK from storage: %w", err)
	}
	return io.ReadAll(output.Body)
}

// Put saves a DEK to AWS S3 Storage by key ID.
func (s *AWSS3Storage) Put(ctx context.Context, keyID string, data []byte) error {
	putObjectInput := &s3.PutObjectInput{
		Bucket:  &s.bucketID,
		Key:     &keyID,
		Body:    bytes.NewReader(data),
		Tagging: &config.AWSS3Tag,
	}
	if _, err := s.client.PutObject(ctx, putObjectInput, s.optFns...); err != nil {
		return fmt.Errorf("uploading DEK to storage: %w", err)
	}
	return nil
}

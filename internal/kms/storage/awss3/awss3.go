/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package memfs implements a storage backend for the KMS using AWS S3: https://aws.amazon.com/s3/
package awss3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/kms/config"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
)

type awsS3ClientAPI interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
}

// Storage is an implementation of the Storage interface, storing keys in AWS S3 buckets.
type Storage struct {
	bucketID string
	client   awsS3ClientAPI
}

// New creates a Storage client for AWS S3: https://aws.amazon.com/s3/
//
// You need to provide credentials to authenticate to AWS using the cfg parameter.
func New(ctx context.Context, bucketID string, accessKey, accessKeyID string) (*Storage, error) {
	credProvider := credentials.NewStaticCredentialsProvider(accessKeyID, accessKey, "")
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithCredentialsProvider(credProvider))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	store := &Storage{client: client, bucketID: bucketID}

	// Try to create new bucket, continue if bucket already exists
	if err := store.createBucket(ctx, bucketID, cfg.Region); err != nil {
		return nil, err
	}
	return store, nil
}

// Get returns a DEK from from AWS S3 Storage by key ID.
func (s *Storage) Get(ctx context.Context, keyID string) ([]byte, error) {
	getObjectInput := &s3.GetObjectInput{
		Bucket: &s.bucketID,
		Key:    &keyID,
	}
	output, err := s.client.GetObject(ctx, getObjectInput)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, storage.ErrDEKUnset
		}
		return nil, fmt.Errorf("downloading DEK from storage: %w", err)
	}
	return io.ReadAll(output.Body)
}

// Put saves a DEK to AWS S3 Storage by key ID.
func (s *Storage) Put(ctx context.Context, keyID string, data []byte) error {
	putObjectInput := &s3.PutObjectInput{
		Bucket:  &s.bucketID,
		Key:     &keyID,
		Body:    bytes.NewReader(data),
		Tagging: &config.AWSS3Tag,
	}
	if _, err := s.client.PutObject(ctx, putObjectInput); err != nil {
		return fmt.Errorf("uploading DEK to storage: %w", err)
	}
	return nil
}

func (s *Storage) createBucket(ctx context.Context, bucketID, region string, optFns ...func(*s3.Options)) error {
	createBucketInput := &s3.CreateBucketInput{
		Bucket: &bucketID,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}
	if _, err := s.client.CreateBucket(ctx, createBucketInput, optFns...); err != nil {
		var bne *types.BucketAlreadyExists
		var baowby *types.BucketAlreadyOwnedByYou
		if !(errors.As(err, &bne) || errors.As(err, &baowby)) {
			return fmt.Errorf("creating storage container: %w", err)
		}
	}
	return nil
}

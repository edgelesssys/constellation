/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package awss3 implements a storage backend for the KMS using AWS S3: https://aws.amazon.com/s3/
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
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
)

type awsS3ClientAPI interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// Storage is an implementation of the Storage interface, storing keys in AWS S3 buckets.
type Storage struct {
	bucketID string
	client   awsS3ClientAPI
}

// New creates a Storage client for AWS S3 using the provided config.
//
// See the AWS docs for more information: https://aws.amazon.com/s3/
func New(ctx context.Context, cfg uri.AWSS3Config) (*Storage, error) {
	clientCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKey, "")),
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS S3 client config: %w", err)
	}

	client := s3.NewFromConfig(clientCfg)

	store := &Storage{client: client, bucketID: cfg.Bucket}

	// Try to create new bucket, continue if bucket already exists
	if err := store.createBucket(ctx, cfg.Bucket, cfg.Region); err != nil {
		return nil, fmt.Errorf("creating storage bucket: %w", err)
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

func (s *Storage) Delete(ctx context.Context, keyID string) error {
	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: &s.bucketID,
		Key:    &keyID,
	}
	if _, err := s.client.DeleteObject(ctx, deleteObjectInput); err != nil {
		return fmt.Errorf("deleting DEK from storage: %w", err)
	}
	return nil
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

func (s *Storage) createBucket(ctx context.Context, bucketID, region string) error {
	createBucketInput := &s3.CreateBucketInput{
		Bucket: &bucketID,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}

	if _, err := s.client.CreateBucket(ctx, createBucketInput); err != nil {
		var bne *types.BucketAlreadyExists
		var baowby *types.BucketAlreadyOwnedByYou
		if !(errors.As(err, &bne) || errors.As(err, &baowby)) {
			return fmt.Errorf("creating storage container: %w", err)
		}
	}
	return nil
}

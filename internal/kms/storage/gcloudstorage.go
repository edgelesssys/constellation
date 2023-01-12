/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package storage

import (
	"context"
	"errors"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type gcpStorageAPI interface {
	Attrs(ctx context.Context, bucketName string) (*storage.BucketAttrs, error)
	Close() error
	CreateBucket(ctx context.Context, bucketName, projectID string, attrs *storage.BucketAttrs) error
	NewWriter(ctx context.Context, bucketName, objectName string) io.WriteCloser
	NewReader(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error)
}

type wrappedGCPClient struct {
	*storage.Client
}

func (c *wrappedGCPClient) Attrs(ctx context.Context, bucketName string) (*storage.BucketAttrs, error) {
	return c.Client.Bucket(bucketName).Attrs(ctx)
}

func (c *wrappedGCPClient) CreateBucket(ctx context.Context, bucketName, projectID string, attrs *storage.BucketAttrs) error {
	return c.Client.Bucket(bucketName).Create(ctx, projectID, attrs)
}

func (c *wrappedGCPClient) NewWriter(ctx context.Context, bucketName, objectName string) io.WriteCloser {
	return c.Client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
}

func (c *wrappedGCPClient) NewReader(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	return c.Client.Bucket(bucketName).Object(objectName).NewReader(ctx)
}

// GoogleCloudStorage is an implementation of the Storage interface, storing keys in Google Cloud Storage buckets.
type GoogleCloudStorage struct {
	newClient  func(ctx context.Context, opts ...option.ClientOption) (gcpStorageAPI, error)
	projectID  string
	bucketName string
	opts       []option.ClientOption
}

// NewGoogleCloudStorage creates a Storage client for Google Cloud Storage: https://cloud.google.com/storage/docs/
//
// The parameter bucketOptions is optional, if not present default options will be created.
func NewGoogleCloudStorage(ctx context.Context, projectID, bucketName string, bucketOptions *storage.BucketAttrs, opts ...option.ClientOption) (*GoogleCloudStorage, error) {
	s := &GoogleCloudStorage{
		newClient:  gcpStorageClientFactory,
		projectID:  projectID,
		bucketName: bucketName,
		opts:       opts,
	}

	// Make sure the storage bucket exists, if not create it
	if err := s.createContainerOrContinue(ctx, bucketOptions); err != nil {
		return nil, err
	}

	return s, nil
}

// Get returns a DEK from Google Cloud Storage by key ID.
func (s *GoogleCloudStorage) Get(ctx context.Context, keyID string) ([]byte, error) {
	client, err := s.newClient(ctx, s.opts...)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reader, err := client.NewReader(ctx, s.bucketName, keyID)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, ErrDEKUnset
		}
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// Put saves a DEK to Google Cloud Storage by key ID.
func (s *GoogleCloudStorage) Put(ctx context.Context, keyID string, data []byte) error {
	client, err := s.newClient(ctx, s.opts...)
	if err != nil {
		return err
	}
	defer client.Close()

	writer := client.NewWriter(ctx, s.bucketName, keyID)
	defer writer.Close()

	_, err = writer.Write(data)
	return err
}

func (s *GoogleCloudStorage) createContainerOrContinue(ctx context.Context, bucketOptions *storage.BucketAttrs) error {
	client, err := s.newClient(ctx, s.opts...)
	if err != nil {
		return err
	}
	defer client.Close()

	if _, err := client.Attrs(ctx, s.bucketName); errors.Is(err, storage.ErrBucketNotExist) {
		return client.CreateBucket(ctx, s.bucketName, s.projectID, bucketOptions)
	} else if err != nil {
		return err
	}

	return nil
}

func gcpStorageClientFactory(ctx context.Context, opts ...option.ClientOption) (gcpStorageAPI, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &wrappedGCPClient{client}, nil
}

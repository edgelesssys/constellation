/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package gcs implements a storage backend for the KMS using Google Cloud Storage (GCS).
package gcs

import (
	"context"
	"errors"
	"io"

	gcstorage "cloud.google.com/go/storage"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"google.golang.org/api/option"
)

type gcpStorageAPI interface {
	Attrs(ctx context.Context, bucketName string) (*gcstorage.BucketAttrs, error)
	Close() error
	CreateBucket(ctx context.Context, bucketName, projectID string, attrs *gcstorage.BucketAttrs) error
	NewWriter(ctx context.Context, bucketName, objectName string) io.WriteCloser
	NewReader(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error)
}

type wrappedGCPClient struct {
	*gcstorage.Client
}

func (c *wrappedGCPClient) Attrs(ctx context.Context, bucketName string) (*gcstorage.BucketAttrs, error) {
	return c.Client.Bucket(bucketName).Attrs(ctx)
}

func (c *wrappedGCPClient) CreateBucket(ctx context.Context, bucketName, projectID string, attrs *gcstorage.BucketAttrs) error {
	return c.Client.Bucket(bucketName).Create(ctx, projectID, attrs)
}

func (c *wrappedGCPClient) NewWriter(ctx context.Context, bucketName, objectName string) io.WriteCloser {
	return c.Client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
}

func (c *wrappedGCPClient) NewReader(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	return c.Client.Bucket(bucketName).Object(objectName).NewReader(ctx)
}

// Storage is an implementation of the Storage interface, storing keys in Google Cloud Storage buckets.
type Storage struct {
	newClient  func(ctx context.Context) (gcpStorageAPI, error)
	bucketName string
}

// New creates a Storage client for Google Cloud Storage using the provided config.
//
// See the Google docs for more information: https://cloud.google.com/storage/docs/
func New(ctx context.Context, cfg uri.GoogleCloudStorageConfig) (*Storage, error) {
	s := &Storage{
		newClient:  newGCPStorageClientFactory(cfg.CredentialsPath),
		bucketName: cfg.Bucket,
	}

	// Make sure the storage bucket exists, if not create it
	if err := s.createContainerOrContinue(ctx, cfg.ProjectID); err != nil {
		return nil, err
	}

	return s, nil
}

// Get returns a DEK from Google Cloud Storage by key ID.
func (s *Storage) Get(ctx context.Context, keyID string) ([]byte, error) {
	client, err := s.newClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reader, err := client.NewReader(ctx, s.bucketName, keyID)
	if err != nil {
		if errors.Is(err, gcstorage.ErrObjectNotExist) {
			return nil, storage.ErrDEKUnset
		}
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// Put saves a DEK to Google Cloud Storage by key ID.
func (s *Storage) Put(ctx context.Context, keyID string, data []byte) error {
	client, err := s.newClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	writer := client.NewWriter(ctx, s.bucketName, keyID)
	defer writer.Close()

	_, err = writer.Write(data)
	return err
}

func (s *Storage) createContainerOrContinue(ctx context.Context, projectID string) error {
	client, err := s.newClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	if _, err := client.Attrs(ctx, s.bucketName); errors.Is(err, gcstorage.ErrBucketNotExist) {
		return client.CreateBucket(ctx, s.bucketName, projectID, nil)
	} else if err != nil {
		return err
	}

	return nil
}

func newGCPStorageClientFactory(credPath string) func(context.Context) (gcpStorageAPI, error) {
	return func(ctx context.Context) (gcpStorageAPI, error) {
		client, err := gcstorage.NewClient(ctx, option.WithCredentialsFile(credPath))
		if err != nil {
			return nil, err
		}
		return &wrappedGCPClient{client}, nil
	}
}

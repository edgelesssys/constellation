package storage

import (
	"context"
	"errors"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// GoogleCloudStorage is an implementation of the Storage interface, storing keys in Google Cloud Storage buckets.
type GoogleCloudStorage struct {
	projectID  string
	bucketName string
	opts       []option.ClientOption
}

// NewGoogleCloudStorage creates a Storage client for Google Cloud Storage: https://cloud.google.com/storage/docs/
//
// The parameter bucketOptions is optional, if not present default options will be created.
func NewGoogleCloudStorage(ctx context.Context, projectID, bucketName string, bucketOptions *storage.BucketAttrs, opts ...option.ClientOption) (*GoogleCloudStorage, error) {
	gcStorage := &GoogleCloudStorage{
		projectID:  projectID,
		bucketName: bucketName,
		opts:       opts,
	}

	// Make sure the storage bucket exists, if not create it
	client, err := storage.NewClient(ctx, gcStorage.opts...)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	_, err = client.Bucket(gcStorage.bucketName).Attrs(ctx)
	if err == nil {
		return gcStorage, nil
	}

	if errors.Is(err, storage.ErrBucketNotExist) {
		err = client.Bucket(gcStorage.bucketName).Create(ctx, gcStorage.projectID, bucketOptions)
	}

	return gcStorage, err
}

// Get returns a DEK from Google Cloud Storage by key ID.
func (s *GoogleCloudStorage) Get(ctx context.Context, keyID string) ([]byte, error) {
	client, err := storage.NewClient(ctx, s.opts...)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reader, err := client.Bucket(s.bucketName).Object(keyID).NewReader(ctx)
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
	client, err := storage.NewClient(ctx, s.opts...)
	if err != nil {
		return err
	}
	defer client.Close()

	writer := client.Bucket(s.bucketName).Object(keyID).NewWriter(ctx)

	if _, err := writer.Write(data); err != nil {
		return err
	}

	return writer.Close()
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package staticupload

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgelesssys/constellation/v2/internal/kms/config"
)

// Upload uploads the given object to S3 and invalidates the CDN cache.
// It returns the upload output or an error.
// The error will be of type InvalidationError if the CDN cache could not be invalidated.
func (c *Client) Upload(
	ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader),
) (*s3manager.UploadOutput, error) {
	if input == nil || input.Key == nil {
		return nil, errors.New("key is not set")
	}
	output, uploadErr := c.uploadClient.Upload(ctx, input, opts...)
	if uploadErr != nil {
		return nil, fmt.Errorf("uploading object: %w", uploadErr)
	}

	if err := c.invalidate(ctx, []string{*input.Key}); err != nil {
		return nil, err
	}
	return output, nil
}

// Put is a convenience method to save a DEK to AWS S3 Storage by key ID.
func (c *Client) Put(ctx context.Context, keyID string, data []byte) error {
	putObjectInput := &s3.PutObjectInput{
		Bucket:  &c.bucketID,
		Key:     &keyID,
		Body:    bytes.NewReader(data),
		Tagging: &config.AWSS3Tag,
	}
	_, err := c.Upload(ctx, putObjectInput)
	return err
}

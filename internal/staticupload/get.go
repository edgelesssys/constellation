/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package staticupload

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
)

// Get is a convenience method to return a DEK from from AWS S3 Storage by key ID.
func (s *Client) Get(ctx context.Context, keyID string) ([]byte, error) {
	getObjectInput := &s3.GetObjectInput{
		Bucket: &s.bucketID,
		Key:    &keyID,
	}
	output, err := s.s3Client.GetObject(ctx, getObjectInput)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, storage.ErrDEKUnset
		}
		return nil, fmt.Errorf("downloading DEK from storage: %w", err)
	}
	return io.ReadAll(output.Body)
}

// GetObject returns an object from from AWS S3 Storage.
func (s *Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return s.s3Client.GetObject(ctx, params, optFns...)
}

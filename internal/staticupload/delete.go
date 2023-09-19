/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package staticupload

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// DeleteObject deletes the given key from S3 and invalidates the CDN cache.
// It returns the delete output or an error.
// The error will be of type InvalidationError if the CDN cache could not be invalidated.
func (c *Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput,
	optFns ...func(*s3.Options),
) (*s3.DeleteObjectOutput, error) {
	if params == nil || params.Key == nil {
		return nil, errors.New("key is not set")
	}
	output, err := c.s3Client.DeleteObject(ctx, params, optFns...)
	if err != nil {
		return nil, fmt.Errorf("deleting object: %w", err)
	}

	if err := c.invalidate(ctx, []string{*params.Key}); err != nil {
		return nil, err
	}
	return output, nil
}

// DeleteObject deletes the given key from S3 and invalidates the CDN cache.
// It returns the delete output or an error.
// The error will be of type InvalidationError if the CDN cache could not be invalidated.
func (c *ClientWithoutCache) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput,
	optFns ...func(*s3.Options),
) (*s3.DeleteObjectOutput, error) {
	if params == nil || params.Key == nil {
		return nil, errors.New("key is not set")
	}
	output, err := c.s3Client.DeleteObject(ctx, params, optFns...)
	if err != nil {
		return nil, fmt.Errorf("deleting object: %w", err)
	}
	return output, nil
}

// DeleteObjects deletes the given objects from S3 and invalidates the CDN cache.
// It returns the delete output or an error.
// The error will be of type InvalidationError if the CDN cache could not be invalidated.
func (c *Client) DeleteObjects(
	ctx context.Context, params *s3.DeleteObjectsInput,
	optFns ...func(*s3.Options),
) (*s3.DeleteObjectsOutput, error) {
	if params == nil || params.Delete == nil || params.Delete.Objects == nil {
		return nil, errors.New("objects are not set")
	}
	for _, obj := range params.Delete.Objects {
		if obj.Key == nil {
			return nil, errors.New("key is not set")
		}
	}
	output, deleteErr := c.s3Client.DeleteObjects(ctx, params, optFns...)
	if deleteErr != nil {
		return nil, fmt.Errorf("deleting objects: %w", deleteErr)
	}

	keys := make([]string, len(params.Delete.Objects))
	for i, obj := range params.Delete.Objects {
		keys[i] = *obj.Key
	}

	if err := c.invalidate(ctx, keys); err != nil {
		return nil, err
	}
	return output, nil
}

// DeleteObjects deletes the given objects from S3 and invalidates the CDN cache.
// It returns the delete output or an error.
// The error will be of type InvalidationError if the CDN cache could not be invalidated.
func (c *ClientWithoutCache) DeleteObjects(
	ctx context.Context, params *s3.DeleteObjectsInput,
	optFns ...func(*s3.Options),
) (*s3.DeleteObjectsOutput, error) {
	if params == nil || params.Delete == nil || params.Delete.Objects == nil {
		return nil, errors.New("objects are not set")
	}
	for _, obj := range params.Delete.Objects {
		if obj.Key == nil {
			return nil, errors.New("key is not set")
		}
	}
	output, deleteErr := c.s3Client.DeleteObjects(ctx, params, optFns...)
	if deleteErr != nil {
		return nil, fmt.Errorf("deleting objects: %w", deleteErr)
	}

	keys := make([]string, len(params.Delete.Objects))
	for i, obj := range params.Delete.Objects {
		keys[i] = *obj.Key
	}
	return output, nil
}

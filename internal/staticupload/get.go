/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package staticupload

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// GetObject retrieves objects from Amazon S3.
func (c *Client) GetObject(
	ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options),
) (*s3.GetObjectOutput, error) {
	return c.s3Client.GetObject(ctx, params, optFns...)
}

// ListObjectsV2 returns some or all (up to 1,000) of the objects in a bucket.
func (c *Client) ListObjectsV2(
	ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	return c.s3Client.ListObjectsV2(ctx, params, optFns...)
}

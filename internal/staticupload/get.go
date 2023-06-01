/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package staticupload

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// GetObject returns an object from from AWS S3 Storage.
func (s *Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return s.s3Client.GetObject(ctx, params, optFns...)
}

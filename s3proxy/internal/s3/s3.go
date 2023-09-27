/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package s3 implements a very thin wrapper around the AWS S3 client.
It only exists to enable stubbing of the AWS S3 client in tests.
*/
package s3

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Client is a wrapper around the AWS S3 client.
type Client struct {
	s3client *s3.Client
}

// NewClient creates a new AWS S3 client.
func NewClient(region string) (*Client, error) {
	// Use context.Background here because this context will not influence the later operations of the client.
	// The context given here is used for http requests that are made during client construction.
	// Client construction happens once during proxy setup.
	clientCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS S3 client config: %w", err)
	}

	client := s3.NewFromConfig(clientCfg)

	return &Client{client}, nil
}

// GetObject returns the object with the given key from the given bucket.
// If a versionID is given, the specific version of the object is returned.
func (c Client) GetObject(ctx context.Context, bucket, key, versionID, sseCustomerAlgorithm, sseCustomerKey, sseCustomerKeyMD5 string) (*s3.GetObjectOutput, error) {
	getObjectInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	if versionID != "" {
		getObjectInput.VersionId = &versionID
	}
	if sseCustomerAlgorithm != "" {
		getObjectInput.SSECustomerAlgorithm = &sseCustomerAlgorithm
	}
	if sseCustomerKey != "" {
		getObjectInput.SSECustomerKey = &sseCustomerKey
	}
	if sseCustomerKeyMD5 != "" {
		getObjectInput.SSECustomerKeyMD5 = &sseCustomerKeyMD5
	}

	return c.s3client.GetObject(ctx, getObjectInput)
}

// PutObject creates a new object in the given bucket with the given key and body.
// Various optional parameters can be set.
func (c Client) PutObject(ctx context.Context, bucket, key, tags, contentType, objectLockLegalHoldStatus, objectLockMode, sseCustomerAlgorithm, sseCustomerKey, sseCustomerKeyMD5 string, objectLockRetainUntilDate time.Time, metadata map[string]string, body []byte) (*s3.PutObjectOutput, error) {
	// The AWS Go SDK has two versions. V1 does not set the Content-Type header.
	// V2 always sets the Content-Type header. We use V2.
	// The s3 API sets an object's content-type to binary/octet-stream if
	// it receives a request without a Content-Type header set.
	// Since a client using V1 may depend on the Content-Type binary/octet-stream
	// we have to explicitly emulate the S3 API behavior, if we receive a request
	// without a Content-Type.
	if contentType == "" {
		contentType = "binary/octet-stream"
	}

	contentMD5 := md5.Sum(body)
	encodedContentMD5 := base64.StdEncoding.EncodeToString(contentMD5[:])

	putObjectInput := &s3.PutObjectInput{
		Bucket:                    &bucket,
		Key:                       &key,
		Body:                      bytes.NewReader(body),
		Tagging:                   &tags,
		Metadata:                  metadata,
		ContentMD5:                &encodedContentMD5,
		ContentType:               &contentType,
		ObjectLockLegalHoldStatus: types.ObjectLockLegalHoldStatus(objectLockLegalHoldStatus),
	}
	if sseCustomerAlgorithm != "" {
		putObjectInput.SSECustomerAlgorithm = &sseCustomerAlgorithm
	}
	if sseCustomerKey != "" {
		putObjectInput.SSECustomerKey = &sseCustomerKey
	}
	if sseCustomerKeyMD5 != "" {
		putObjectInput.SSECustomerKeyMD5 = &sseCustomerKeyMD5
	}

	// It is not allowed to only set one of these two properties.
	if objectLockMode != "" && !objectLockRetainUntilDate.IsZero() {
		putObjectInput.ObjectLockMode = types.ObjectLockMode(objectLockMode)
		putObjectInput.ObjectLockRetainUntilDate = &objectLockRetainUntilDate
	}

	return c.s3client.PutObject(ctx, putObjectInput)
}

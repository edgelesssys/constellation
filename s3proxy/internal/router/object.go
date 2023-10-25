/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package router

import (
	"context"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/s3proxy/internal/crypto"
	"go.uber.org/zap"
)

const (
	// dekTag is the name of the header that holds the encrypted data encryption key for the attached object. Presence of the key implies the object needs to be decrypted.
	// Use lowercase only, as AWS automatically lowercases all metadata keys.
	dekTag = "constellation-dek"
)

// object bundles data to implement http.Handler methods that use data from incoming requests.
type object struct {
	kek                       [32]byte
	client                    s3Client
	key                       string
	bucket                    string
	data                      []byte
	query                     url.Values
	tags                      string
	contentType               string
	metadata                  map[string]string
	objectLockLegalHoldStatus string
	objectLockMode            string
	objectLockRetainUntilDate time.Time
	sseCustomerAlgorithm      string
	sseCustomerKey            string
	sseCustomerKeyMD5         string
	log                       *logger.Logger
}

// get is a http.HandlerFunc that implements the GET method for objects.
func (o object) get(w http.ResponseWriter, r *http.Request) {
	o.log.With(zap.String("key", o.key), zap.String("host", o.bucket)).Debugf("getObject")

	versionID, ok := o.query["versionId"]
	if !ok {
		versionID = []string{""}
	}

	output, err := o.client.GetObject(r.Context(), o.bucket, o.key, versionID[0], o.sseCustomerAlgorithm, o.sseCustomerKey, o.sseCustomerKeyMD5)
	if err != nil {
		// log with Info as it might be expected behavior (e.g. object not found).
		o.log.With(zap.Error(err)).Errorf("GetObject sending request to S3")

		// We want to forward error codes from the s3 API to clients as much as possible.
		code := parseErrorCode(err)
		if code != 0 {
			http.Error(w, err.Error(), code)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if output.ETag != nil {
		w.Header().Set("ETag", strings.Trim(*output.ETag, "\""))
	}
	if output.Expiration != nil {
		w.Header().Set("x-amz-expiration", *output.Expiration)
	}
	if output.ChecksumCRC32 != nil {
		w.Header().Set("x-amz-checksum-crc32", *output.ChecksumCRC32)
	}
	if output.ChecksumCRC32C != nil {
		w.Header().Set("x-amz-checksum-crc32c", *output.ChecksumCRC32C)
	}
	if output.ChecksumSHA1 != nil {
		w.Header().Set("x-amz-checksum-sha1", *output.ChecksumSHA1)
	}
	if output.ChecksumSHA256 != nil {
		w.Header().Set("x-amz-checksum-sha256", *output.ChecksumSHA256)
	}
	if output.SSECustomerAlgorithm != nil {
		w.Header().Set("x-amz-server-side-encryption-customer-algorithm", *output.SSECustomerAlgorithm)
	}
	if output.SSECustomerKeyMD5 != nil {
		w.Header().Set("x-amz-server-side-encryption-customer-key-MD5", *output.SSECustomerKeyMD5)
	}
	if output.SSEKMSKeyId != nil {
		w.Header().Set("x-amz-server-side-encryption-aws-kms-key-id", *output.SSEKMSKeyId)
	}
	if output.ServerSideEncryption != "" {
		w.Header().Set("x-amz-server-side-encryption-context", string(output.ServerSideEncryption))
	}

	body, err := io.ReadAll(output.Body)
	if err != nil {
		o.log.With(zap.Error(err)).Errorf("GetObject reading S3 response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	plaintext := body
	rawEncryptedDEK, ok := output.Metadata[dekTag]
	if ok {
		encryptedDEK, err := hex.DecodeString(rawEncryptedDEK)
		if err != nil {
			o.log.Error("GetObject decoding DEK", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		plaintext, err = crypto.Decrypt(body, encryptedDEK, o.kek)
		if err != nil {
			o.log.With(zap.Error(err)).Errorf("GetObject decrypting response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(plaintext); err != nil {
		o.log.With(zap.Error(err)).Errorf("GetObject sending response")
	}
}

// put is a http.HandlerFunc that implements the PUT method for objects.
func (o object) put(w http.ResponseWriter, r *http.Request) {
	ciphertext, encryptedDEK, err := crypto.Encrypt(o.data, o.kek)
	if err != nil {
		o.log.With(zap.Error(err)).Errorf("PutObject")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	o.metadata[dekTag] = hex.EncodeToString(encryptedDEK)

	output, err := o.client.PutObject(r.Context(), o.bucket, o.key, o.tags, o.contentType, o.objectLockLegalHoldStatus, o.objectLockMode, o.sseCustomerAlgorithm, o.sseCustomerKey, o.sseCustomerKeyMD5, o.objectLockRetainUntilDate, o.metadata, ciphertext)
	if err != nil {
		o.log.With(zap.Error(err)).Errorf("PutObject sending request to S3")

		// We want to forward error codes from the s3 API to clients whenever possible.
		code := parseErrorCode(err)
		if code != 0 {
			http.Error(w, err.Error(), code)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("x-amz-server-side-encryption", string(output.ServerSideEncryption))

	if output.VersionId != nil {
		w.Header().Set("x-amz-version-id", *output.VersionId)
	}
	if output.ETag != nil {
		w.Header().Set("ETag", strings.Trim(*output.ETag, "\""))
	}
	if output.Expiration != nil {
		w.Header().Set("x-amz-expiration", *output.Expiration)
	}
	if output.ChecksumCRC32 != nil {
		w.Header().Set("x-amz-checksum-crc32", *output.ChecksumCRC32)
	}
	if output.ChecksumCRC32C != nil {
		w.Header().Set("x-amz-checksum-crc32c", *output.ChecksumCRC32C)
	}
	if output.ChecksumSHA1 != nil {
		w.Header().Set("x-amz-checksum-sha1", *output.ChecksumSHA1)
	}
	if output.ChecksumSHA256 != nil {
		w.Header().Set("x-amz-checksum-sha256", *output.ChecksumSHA256)
	}
	if output.SSECustomerAlgorithm != nil {
		w.Header().Set("x-amz-server-side-encryption-customer-algorithm", *output.SSECustomerAlgorithm)
	}
	if output.SSECustomerKeyMD5 != nil {
		w.Header().Set("x-amz-server-side-encryption-customer-key-MD5", *output.SSECustomerKeyMD5)
	}
	if output.SSEKMSKeyId != nil {
		w.Header().Set("x-amz-server-side-encryption-aws-kms-key-id", *output.SSEKMSKeyId)
	}
	if output.SSEKMSEncryptionContext != nil {
		w.Header().Set("x-amz-server-side-encryption-context", *output.SSEKMSEncryptionContext)
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(nil); err != nil {
		o.log.With(zap.Error(err)).Errorf("PutObject sending response")
	}
}

func parseErrorCode(err error) int {
	regex := regexp.MustCompile(`https response error StatusCode: (\d+)`)
	matches := regex.FindStringSubmatch(err.Error())
	if len(matches) > 1 {
		code, _ := strconv.Atoi(matches[1])
		return code
	}

	return 0
}

type s3Client interface {
	GetObject(ctx context.Context, bucket, key, versionID, sseCustomerAlgorithm, sseCustomerKey, sseCustomerKeyMD5 string) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, bucket, key, tags, contentType, objectLockLegalHoldStatus, objectLockMode, sseCustomerAlgorithm, sseCustomerKey, sseCustomerKeyMD5 string, objectLockRetainUntilDate time.Time, metadata map[string]string, body []byte) (*s3.PutObjectOutput, error)
}

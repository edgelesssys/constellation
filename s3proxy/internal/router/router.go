/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package router implements the main interception logic of s3proxy.
It decides which packages to forward and which to intercept.

The routing logic in this file is taken from this blog post: https://benhoyt.com/writings/go-routing/#regex-switch.
We should be able to replace this once this is part of the stdlib: https://github.com/golang/go/issues/61410.

If the router intercepts a PutObject request it will encrypt the body before forwarding it to the S3 API.
The stored object will have a tag that holds an encrypted data encryption key (DEK).
That DEK is used to encrypt the object's body.
The DEK is generated randomly for each PutObject request.
The DEK is encrypted with a key encryption key (KEK) fetched from Constellation's keyservice.
*/
package router

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/s3proxy/internal/kms"
	"github.com/edgelesssys/constellation/v2/s3proxy/internal/s3"
)

const (
	// Use a 32*8 = 256 bit key for AES-256.
	kekSizeBytes = 32
	kekID        = "s3proxy-kek"
)

var (
	keyPattern          = regexp.MustCompile("/(.+)")
	bucketAndKeyPattern = regexp.MustCompile("/([^/?]+)/(.+)")
)

// Router implements the interception logic for the s3proxy.
type Router struct {
	region string
	kek    [32]byte
	// forwardMultipartReqs controls whether we forward the following requests: CreateMultipartUpload, UploadPart, CompleteMultipartUpload, AbortMultipartUpload.
	// s3proxy does not implement those yet.
	// Setting forwardMultipartReqs to true will forward those requests to the S3 API, otherwise we block them (secure defaults).
	forwardMultipartReqs bool
	log                  *logger.Logger
}

// New creates a new Router.
func New(region, endpoint string, forwardMultipartReqs bool, log *slog.Logger) (Router, error) {
	kms := kms.New(log, endpoint)

	// Get the key encryption key that encrypts all DEKs.
	kek, err := kms.GetDataKey(context.Background(), kekID, kekSizeBytes)
	if err != nil {
		return Router{}, fmt.Errorf("getting KEK: %w", err)
	}

	kekArray, err := byteSliceToByteArray(kek)
	if err != nil {
		return Router{}, fmt.Errorf("converting KEK to byte array: %w", err)
	}

	return Router{region: region, kek: kekArray, forwardMultipartReqs: forwardMultipartReqs, log: log}, nil
}

// Serve implements the routing logic for the s3 proxy.
// It intercepts GetObject and PutObject requests, encrypting/decrypting their bodies if necessary.
// All other requests are forwarded to the S3 API.
// Ideally we could separate routing logic, request handling and s3 interactions.
// Currently routing logic and request handling are integrated.
func (r Router) Serve(w http.ResponseWriter, req *http.Request) {
	client, err := s3.NewClient(r.region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var key string
	var bucket string
	var matchingPath bool
	if containsBucket(req.Host) {
		// BUCKET.s3.REGION.amazonaws.com
		parts := strings.Split(req.Host, ".")
		bucket = parts[0]

		matchingPath = match(req.URL.Path, keyPattern, &key)

	} else {
		matchingPath = match(req.URL.Path, bucketAndKeyPattern, &bucket, &key)
	}

	var h http.Handler

	switch {
	// intercept GetObject.
	case matchingPath && req.Method == "GET" && !isUnwantedGetEndpoint(req.URL.Query()):
		h = handleGetObject(client, key, bucket, r.log)
	// intercept PutObject.
	case matchingPath && req.Method == "PUT" && !isUnwantedPutEndpoint(req.Header, req.URL.Query()):
		h = handlePutObject(client, key, bucket, r.log)
	case !r.forwardMultipartReqs && matchingPath && isUploadPart(req.Method, req.URL.Query()):
		h = handleUploadPart(r.log)
	case !r.forwardMultipartReqs && matchingPath && isCreateMultipartUpload(req.Method, req.URL.Query()):
		h = handleCreateMultipartUpload(r.log)
	case !r.forwardMultipartReqs && matchingPath && isCompleteMultipartUpload(req.Method, req.URL.Query()):
		h = handleCompleteMultipartUpload(r.log)
	case !r.forwardMultipartReqs && matchingPath && isAbortMultipartUpload(req.Method, req.URL.Query()):
		h = handleAbortMultipartUpload(r.log)
	// Forward all other requests.
	default:
		h = handleForwards(r.log)
	}

	h.ServeHTTP(w, req)
}

func isAbortMultipartUpload(method string, query url.Values) bool {
	_, uploadID := query["uploadId"]

	return method == "DELETE" && uploadID
}

func isCompleteMultipartUpload(method string, query url.Values) bool {
	_, multipart := query["uploadId"]

	return method == "POST" && multipart
}

func isCreateMultipartUpload(method string, query url.Values) bool {
	_, multipart := query["uploads"]

	return method == "POST" && multipart
}

func isUploadPart(method string, query url.Values) bool {
	_, partNumber := query["partNumber"]
	_, uploadID := query["uploadId"]

	return method == "PUT" && partNumber && uploadID
}

// ContentSHA256MismatchError is a helper struct to create an XML formatted error message.
// s3 clients might try to parse error messages, so we need to serve correctly formatted messages.
type ContentSHA256MismatchError struct {
	XMLName                     xml.Name `xml:"Error"`
	Code                        string   `xml:"Code"`
	Message                     string   `xml:"Message"`
	ClientComputedContentSHA256 string   `xml:"ClientComputedContentSHA256"`
	S3ComputedContentSHA256     string   `xml:"S3ComputedContentSHA256"`
}

// NewContentSHA256MismatchError creates a new ContentSHA256MismatchError.
func NewContentSHA256MismatchError(clientComputedContentSHA256, s3ComputedContentSHA256 string) ContentSHA256MismatchError {
	return ContentSHA256MismatchError{
		Code:                        "XAmzContentSHA256Mismatch",
		Message:                     "The provided 'x-amz-content-sha256' header does not match what was computed.",
		ClientComputedContentSHA256: clientComputedContentSHA256,
		S3ComputedContentSHA256:     s3ComputedContentSHA256,
	}
}

// byteSliceToByteArray casts a byte slice to a byte array of length 32.
// It does a length check to prevent the cast from panic'ing.
func byteSliceToByteArray(input []byte) ([32]byte, error) {
	if len(input) != 32 {
		return [32]byte{}, fmt.Errorf("input length mismatch, got: %d", len(input))
	}

	return ([32]byte)(input), nil
}

// containsBucket is a helper to recognizes cases where the bucket name is sent as part of the host.
// In other cases the bucket name is sent as part of the path.
func containsBucket(host string) bool {
	parts := strings.Split(host, ".")
	return len(parts) > 4
}

// isUnwantedGetEndpoint returns true if the request is any of these requests: GetObjectAcl, GetObjectAttributes, GetObjectLegalHold, GetObjectRetention, GetObjectTagging, GetObjectTorrent, ListParts.
// These requests are all structured similarly: they all have a query param that is not present in GetObject.
// Otherwise those endpoints are similar to GetObject.
func isUnwantedGetEndpoint(query url.Values) bool {
	_, acl := query["acl"]
	_, attributes := query["attributes"]
	_, legalHold := query["legal-hold"]
	_, retention := query["retention"]
	_, tagging := query["tagging"]
	_, torrent := query["torrent"]
	_, uploadID := query["uploadId"]

	return acl || attributes || legalHold || retention || tagging || torrent || uploadID
}

// isUnwantedPutEndpoint returns true if the request is any of these requests: UploadPart, PutObjectTagging.
// These requests are all structured similarly: they all have a query param that is not present in PutObject.
// Otherwise those endpoints are similar to PutObject.
func isUnwantedPutEndpoint(header http.Header, query url.Values) bool {
	if header.Get("x-amz-copy-source") != "" {
		return true
	}

	_, partNumber := query["partNumber"]
	_, uploadID := query["uploadId"]
	_, tagging := query["tagging"]
	_, legalHold := query["legal-hold"]
	_, objectLock := query["object-lock"]
	_, retention := query["retention"]
	_, publicAccessBlock := query["publicAccessBlock"]
	_, acl := query["acl"]

	return partNumber || uploadID || tagging || legalHold || objectLock || retention || publicAccessBlock || acl
}

func sha256sum(data []byte) string {
	digest := sha256.Sum256(data)
	return fmt.Sprintf("%x", digest)
}

// getMetadataHeaders parses user-defined metadata headers from a
// http.Header object. Users can define custom headers by taking
// HEADERNAME and prefixing it with "x-amz-meta-".
func getMetadataHeaders(header http.Header) map[string]string {
	result := map[string]string{}

	for key := range header {
		key = strings.ToLower(key)

		if strings.HasPrefix(key, "x-amz-meta-") {
			name := strings.TrimPrefix(key, "x-amz-meta-")
			result[name] = strings.Join(header.Values(key), ",")
		}
	}

	return result
}

func parseRetentionTime(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, raw)
}

// repackage implements all modifications we need to do to an incoming request that we want to forward to the s3 API.
func repackage(r *http.Request) http.Request {
	req := r.Clone(r.Context())

	// HTTP clients are not supposed to set this field, however when we receive a request it is set.
	// So, we unset it.
	req.RequestURI = ""

	req.URL.Host = r.Host
	// We always want to use HTTPS when talking to S3.
	req.URL.Scheme = "https"

	return *req
}

// validateContentMD5 checks if the content-md5 header matches the body.
func validateContentMD5(contentMD5 string, body []byte) error {
	if contentMD5 == "" {
		return nil
	}

	expected, err := base64.StdEncoding.DecodeString(contentMD5)
	if err != nil {
		return fmt.Errorf("decoding base64: %w", err)
	}

	if len(expected) != 16 {
		return fmt.Errorf("content-md5 must be 16 bytes long, got %d bytes", len(expected))
	}

	actual := md5.Sum(body)

	if !bytes.Equal(actual[:], expected) {
		return fmt.Errorf("content-md5 mismatch, header is %x, body is %x", expected, actual)
	}

	return nil
}

// match reports whether path matches pattern, and if it matches,
// assigns any capture groups to the *string or *int vars.
func match(path string, pattern *regexp.Regexp, vars ...*string) bool {
	matches := pattern.FindStringSubmatch(path)
	if len(matches) <= 0 {
		return false
	}

	for i, match := range matches[1:] {
		// assign the value of 'match' to the i-th argument.
		*vars[i] = match
	}
	return true
}

// allowMethod takes a HandlerFunc and wraps it in a handler that only
// responds if the request method is the given method, otherwise it
// responds with HTTP 405 Method Not Allowed.
func allowMethod(h http.HandlerFunc, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if method != r.Method {
			w.Header().Set("Allow", method)
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}

// get takes a HandlerFunc and wraps it to only allow the GET method.
func get(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, "GET")
}

// put takes a HandlerFunc and wraps it to only allow the POST method.
func put(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, "PUT")
}

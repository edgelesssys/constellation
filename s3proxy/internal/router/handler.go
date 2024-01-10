/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package router

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/edgelesssys/constellation/v2/s3proxy/internal/s3"
)

func handleGetObject(client *s3.Client, key string, bucket string, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.With(slog.String("path", req.URL.Path), slog.String("method", req.Method), slog.String("host", req.Host)).Debug("intercepting")
		if req.Header.Get("Range") != "" {
			log.Error("GetObject Range header unsupported")
			http.Error(w, "s3proxy currently does not support Range headers", http.StatusNotImplemented)
			return
		}

		obj := object{
			client:               client,
			key:                  key,
			bucket:               bucket,
			query:                req.URL.Query(),
			sseCustomerAlgorithm: req.Header.Get("x-amz-server-side-encryption-customer-algorithm"),
			sseCustomerKey:       req.Header.Get("x-amz-server-side-encryption-customer-key"),
			sseCustomerKeyMD5:    req.Header.Get("x-amz-server-side-encryption-customer-key-MD5"),
			log:                  log,
		}
		get(obj.get)(w, req)
	}
}

func handlePutObject(client *s3.Client, key string, bucket string, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.With(slog.String("path", req.URL.Path), slog.String("method", req.Method), slog.String("host", req.Host)).Debug("intercepting")
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.With(slog.Any("error", err)).Error("PutObject")
			http.Error(w, fmt.Sprintf("reading body: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		clientDigest := req.Header.Get("x-amz-content-sha256")
		serverDigest := sha256sum(body)

		// There may be a client that wants to test that incorrect content digests result in API errors.
		// For encrypting the body we have to recalculate the content digest.
		// If the client intentionally sends a mismatching content digest, we would take the client request, rewrap it,
		// calculate the correct digest for the new body and NOT get an error.
		// Thus we have to check incoming requets for matching content digests.
		// UNSIGNED-PAYLOAD can be used to disabled payload signing. In that case we don't check the content digest.
		if clientDigest != "" && clientDigest != "UNSIGNED-PAYLOAD" && clientDigest != serverDigest {
			log.Debug("PutObject", "error", "x-amz-content-sha256 mismatch")
			// The S3 API responds with an XML formatted error message.
			mismatchErr := NewContentSHA256MismatchError(clientDigest, serverDigest)
			marshalled, err := xml.Marshal(mismatchErr)
			if err != nil {
				log.With(slog.Any("error", err)).Error("PutObject")
				http.Error(w, fmt.Sprintf("marshalling error: %s", err.Error()), http.StatusInternalServerError)
				return
			}

			http.Error(w, string(marshalled), http.StatusBadRequest)
			return
		}

		metadata := getMetadataHeaders(req.Header)

		raw := req.Header.Get("x-amz-object-lock-retain-until-date")
		retentionTime, err := parseRetentionTime(raw)
		if err != nil {
			log.With(slog.String("data", raw), slog.Any("error", err)).Error("parsing lock retention time")
			http.Error(w, fmt.Sprintf("parsing x-amz-object-lock-retain-until-date: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		err = validateContentMD5(req.Header.Get("content-md5"), body)
		if err != nil {
			log.With(slog.Any("error", err)).Error("validating content md5")
			http.Error(w, fmt.Sprintf("validating content md5: %s", err.Error()), http.StatusBadRequest)
			return
		}

		obj := object{
			client:                    client,
			key:                       key,
			bucket:                    bucket,
			data:                      body,
			query:                     req.URL.Query(),
			tags:                      req.Header.Get("x-amz-tagging"),
			contentType:               req.Header.Get("Content-Type"),
			metadata:                  metadata,
			objectLockLegalHoldStatus: req.Header.Get("x-amz-object-lock-legal-hold"),
			objectLockMode:            req.Header.Get("x-amz-object-lock-mode"),
			objectLockRetainUntilDate: retentionTime,
			sseCustomerAlgorithm:      req.Header.Get("x-amz-server-side-encryption-customer-algorithm"),
			sseCustomerKey:            req.Header.Get("x-amz-server-side-encryption-customer-key"),
			sseCustomerKeyMD5:         req.Header.Get("x-amz-server-side-encryption-customer-key-MD5"),
			log:                       log,
		}

		put(obj.put)(w, req)
	}
}

func handleForwards(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.With(slog.String("path", req.URL.Path), slog.String("method", req.Method), slog.String("host", req.Host)).Debug("forwarding")

		newReq := repackage(req)

		httpClient := http.DefaultClient
		resp, err := httpClient.Do(&newReq)
		if err != nil {
			log.With(slog.Any("error", err)).Error("do request")
			http.Error(w, fmt.Sprintf("do request: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for key := range resp.Header {
			w.Header().Set(key, resp.Header.Get(key))
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.With(slog.Any("error", err)).Error("ReadAll")
			http.Error(w, fmt.Sprintf("reading body: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(resp.StatusCode)
		if body == nil {
			return
		}

		if _, err := w.Write(body); err != nil {
			log.With(slog.Any("error", err)).Error("Write")
			http.Error(w, fmt.Sprintf("writing body: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

// handleCreateMultipartUpload logs the request and blocks with an error message.
func handleCreateMultipartUpload(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.With(slog.String("path", req.URL.Path), slog.String("method", req.Method), slog.String("host", req.Host)).Debug("intercepting CreateMultipartUpload")

		log.Error("Blocking CreateMultipartUpload request")
		http.Error(w, "s3proxy is configured to block CreateMultipartUpload requests", http.StatusNotImplemented)
	}
}

// handleUploadPart logs the request and blocks with an error message.
func handleUploadPart(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.With(slog.String("path", req.URL.Path), slog.String("method", req.Method), slog.String("host", req.Host)).Debug("intercepting UploadPart")

		log.Error("Blocking UploadPart request")
		http.Error(w, "s3proxy is configured to block UploadPart requests", http.StatusNotImplemented)
	}
}

// handleCompleteMultipartUpload logs the request and blocks with an error message.
func handleCompleteMultipartUpload(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.With(slog.String("path", req.URL.Path), slog.String("method", req.Method), slog.String("host", req.Host)).Debug("intercepting CompleteMultipartUpload")

		log.Error("Blocking CompleteMultipartUpload request")
		http.Error(w, "s3proxy is configured to block CompleteMultipartUpload requests", http.StatusNotImplemented)
	}
}

// handleAbortMultipartUpload logs the request and blocks with an error message.
func handleAbortMultipartUpload(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.With(slog.String("path", req.URL.Path), slog.String("method", req.Method), slog.String("host", req.Host)).Debug("intercepting AbortMultipartUpload")

		log.Error("Blocking AbortMultipartUpload request")
		http.Error(w, "s3proxy is configured to block AbortMultipartUpload requests", http.StatusNotImplemented)
	}
}

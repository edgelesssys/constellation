/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package retry provides functions to check if a gRPC error is retryable.
package retry

import (
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	authEOFErr                       = `connection error: desc = "transport: authentication handshake failed: EOF"`
	authReadTCPErr                   = `connection error: desc = "transport: authentication handshake failed: read tcp`
	authHandshakeErr                 = `connection error: desc = "transport: authentication handshake failed`
	authHandshakeDeadlineExceededErr = `connection error: desc = "transport: authentication handshake failed: context deadline exceeded`
)

// grpcErr is the error type that is returned by the grpc client.
// taken from google.golang.org/grpc/status.FromError.
type grpcErr interface {
	GRPCStatus() *status.Status
	Error() string
}

// ServiceIsUnavailable checks if the error is a grpc status with code Unavailable.
// In the special case of an authentication handshake failure, false is returned to prevent further retries.
// Since the GCP proxy loadbalancer may error with an authentication handshake failure if no available backends are ready,
// the special handshake errors caused by the GCP LB (e.g. "read tcp", "EOF") are retried.
func ServiceIsUnavailable(err error) bool {
	var targetErr grpcErr
	if !errors.As(err, &targetErr) {
		return false
	}

	statusErr, ok := status.FromError(targetErr)
	if !ok {
		return false
	}

	if statusErr.Code() != codes.Unavailable {
		return false
	}

	// retry if GCP proxy LB isn't available
	if strings.HasPrefix(statusErr.Message(), authEOFErr) {
		return true
	}

	// retry if GCP proxy LB isn't fully available yet
	if strings.HasPrefix(statusErr.Message(), authReadTCPErr) {
		return true
	}

	// retry if the handshake deadline was exceeded
	if strings.HasPrefix(statusErr.Message(), authHandshakeDeadlineExceededErr) {
		return true
	}

	return !strings.HasPrefix(statusErr.Message(), authHandshakeErr)
}

// LoadbalancerIsNotReady checks if the error was caused by a GCP LB not being ready yet.
func LoadbalancerIsNotReady(err error) bool {
	var targetErr grpcErr
	if !errors.As(err, &targetErr) {
		return false
	}

	statusErr, ok := status.FromError(targetErr)
	if !ok {
		return false
	}

	if statusErr.Code() != codes.Unavailable {
		return false
	}

	// retry if the handshake deadline was exceeded
	if strings.HasPrefix(statusErr.Message(), authHandshakeDeadlineExceededErr) {
		return true
	}

	// retry if GCP proxy LB isn't fully available yet
	return strings.HasPrefix(statusErr.Message(), authReadTCPErr)
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package retry

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceIsUnavailable(t *testing.T) {
	testCases := map[string]struct {
		err             error
		wantUnavailable bool
	}{
		"nil": {},
		"not status error": {
			err: errors.New("error"),
		},
		"not unavailable": {
			err: status.Error(codes.Internal, "error"),
		},
		"unavailable error with authentication handshake failure": {
			err: status.Error(codes.Unavailable, `connection error: desc = "transport: authentication handshake failed`),
		},
		"normal unavailable error": {
			err:             status.Error(codes.Unavailable, "error"),
			wantUnavailable: true,
		},
		"wrapped error": {
			err:             fmt.Errorf("some wrapping: %w", status.Error(codes.Unavailable, "error")),
			wantUnavailable: true,
		},
		"code unknown": {
			err: status.Error(codes.Unknown, "unknown"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.wantUnavailable, ServiceIsUnavailable(tc.err))
		})
	}
}

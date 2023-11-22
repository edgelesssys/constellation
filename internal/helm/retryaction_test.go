/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestRetryApply(t *testing.T) {
	testCases := map[string]struct {
		applier *stubRetriableApplier
		wantErr bool
	}{
		"success": {
			applier: &stubRetriableApplier{
				atomic: true,
			},
		},
		"two errors": {
			applier: &stubRetriableApplier{
				applyErrs: []error{
					assert.AnError,
					assert.AnError,
					nil,
				},
				atomic: true,
			},
		},
		"retries are aborted after maximumRetryAttempts": {
			applier: &stubRetriableApplier{
				applyErrs: []error{
					assert.AnError,
					assert.AnError,
					assert.AnError,
					assert.AnError,
				},
				atomic: true,
			},
			wantErr: true,
		},
		"non atomic actions are not retried": {
			applier: &stubRetriableApplier{
				atomic: false,
				applyErrs: []error{
					assert.AnError,
					assert.AnError,
					nil,
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := retryApply(context.Background(), tc.applier, time.Millisecond, logger.NewTest(t))
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubRetriableApplier struct {
	atomic    bool
	applyErrs []error
}

func (s *stubRetriableApplier) apply(context.Context) error {
	if len(s.applyErrs) == 0 {
		return nil
	}

	// return the first error in the list
	// and remove it from the list
	err := s.applyErrs[0]
	if len(s.applyErrs) > 1 {
		s.applyErrs = s.applyErrs[1:]
	} else {
		s.applyErrs = nil
	}
	return err
}

func (s *stubRetriableApplier) ReleaseName() string {
	return ""
}

func (s *stubRetriableApplier) IsAtomic() bool {
	return s.atomic
}

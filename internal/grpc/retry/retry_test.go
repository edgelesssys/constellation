package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	testclock "k8s.io/utils/clock/testing"
)

func TestDo(t *testing.T) {
	testCases := map[string]struct {
		cancel  bool
		errors  []error
		wantErr error
	}{
		"no error": {
			errors: []error{
				nil,
			},
		},
		"permanent error": {
			errors: []error{
				errors.New("error"),
			},
			wantErr: errors.New("error"),
		},
		"service unavailable then success": {
			errors: []error{
				status.Error(codes.Unavailable, "error"),
				nil,
			},
		},
		"service unavailable then permanent error": {
			errors: []error{
				status.Error(codes.Unavailable, "error"),
				errors.New("error"),
			},
			wantErr: errors.New("error"),
		},
		"cancellation works": {
			cancel: true,
			errors: []error{
				status.Error(codes.Unavailable, "error"),
			},
			wantErr: context.Canceled,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			doer := newStubDoer()
			clock := testclock.NewFakeClock(time.Now())
			retrier := IntervalRetrier{
				doer:  doer,
				clock: clock,
			}
			retrierResult := make(chan error, 1)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() { retrierResult <- retrier.Do(ctx) }()

			if tc.cancel {
				cancel()
			}

			for _, err := range tc.errors {
				doer.errC <- err
				clock.Step(retrier.interval)
			}

			assert.Equal(tc.wantErr, <-retrierResult)
		})
	}
}

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
			retrier := IntervalRetrier{}
			assert.Equal(tc.wantUnavailable, retrier.serviceIsUnavailable(tc.err))
		})
	}
}

type stubDoer struct {
	errC chan error
}

func newStubDoer() *stubDoer {
	return &stubDoer{
		errC: make(chan error),
	}
}

func (d *stubDoer) Do(_ context.Context) error {
	return <-d.errC
}

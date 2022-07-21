package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
				errors.New("retry me"),
				nil,
			},
		},
		"service unavailable then permanent error": {
			errors: []error{
				errors.New("retry me"),
				errors.New("error"),
			},
			wantErr: errors.New("error"),
		},
		"cancellation works": {
			cancel: true,
			errors: []error{
				errors.New("retry me"),
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
				doer:      doer,
				clock:     clock,
				retriable: isRetriable,
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

func isRetriable(err error) bool {
	return err.Error() == "retry me"
}

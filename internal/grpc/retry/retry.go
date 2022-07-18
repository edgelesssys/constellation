package retry

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/clock"
)

// IntervalRetrier is retries a grpc call with an interval.
type IntervalRetrier struct {
	interval time.Duration
	doer     Doer
	clock    clock.WithTicker
}

// NewIntervalRetrier returns a new IntervalRetrier.
func NewIntervalRetrier(doer Doer, interval time.Duration) *IntervalRetrier {
	return &IntervalRetrier{
		interval: interval,
		doer:     doer,
		clock:    clock.RealClock{},
	}
}

// Do retries performing a grpc call until it succeeds, returns a permanent error or the context is cancelled.
func (r *IntervalRetrier) Do(ctx context.Context) error {
	ticker := r.clock.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		err := r.doer.Do(ctx)
		if err == nil {
			return nil
		}

		if !r.serviceIsUnavailable(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C():
		}
	}
}

// serviceIsUnavailable checks if the error is a grpc status with code Unavailable.
// In the special case of an authentication handshake failure, false is returned to prevent further retries.
func (r *IntervalRetrier) serviceIsUnavailable(err error) bool {
	// taken from google.golang.org/grpc/status.FromError
	var targetErr interface {
		GRPCStatus() *status.Status
		Error() string
	}

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

	// ideally we would check the error type directly, but grpc only provides a string
	return !strings.HasPrefix(statusErr.Message(), `connection error: desc = "transport: authentication handshake failed`)
}

type Doer interface {
	// Do performs a grpc operation.
	//
	// It should return a grpc status with code Unavailable error to signal a transient fault.
	Do(ctx context.Context) error
}

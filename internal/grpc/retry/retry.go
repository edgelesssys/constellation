package retry

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/clock"
)

type IntervalRetryer struct {
	interval time.Duration
	doer     Doer
	clock    clock.WithTicker
}

func NewIntervalRetryer(doer Doer, interval time.Duration) *IntervalRetryer {
	return &IntervalRetryer{
		interval: interval,
		doer:     doer,
		clock:    clock.RealClock{},
	}
}

func (r *IntervalRetryer) Do(ctx context.Context) error {
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
		case <-ctx.Done(): // TODO(katexochen): is this necessary?
			return ctx.Err()
		case <-ticker.C():
		}
	}
}

func (r *IntervalRetryer) serviceIsUnavailable(err error) bool {
	statusErr, ok := status.FromError(err)
	if !ok {
		return false
	}
	if statusErr.Code() != codes.Unavailable {
		return false
	}
	// ideally we would check the error type directly, but grpc only provides a string
	return strings.HasPrefix(statusErr.Message(), `connection error: desc = "transport: authentication handshake failed`)
}

type Doer interface {
	Do(ctx context.Context) error
}

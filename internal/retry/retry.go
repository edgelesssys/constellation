/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package retry provides a simple interface for retrying operations.
package retry

import (
	"context"
	"time"

	"k8s.io/utils/clock"
)

// IntervalRetrier retries a call with an interval. The call is defined in the Doer property.
type IntervalRetrier struct {
	interval  time.Duration
	doer      Doer
	clock     clock.WithTicker
	retriable func(error) bool
}

// NewIntervalRetrier returns a new IntervalRetrier. The optional clock is used for testing.
func NewIntervalRetrier(doer Doer, interval time.Duration, retriable func(error) bool, optClock ...clock.WithTicker) *IntervalRetrier {
	var clock clock.WithTicker = clock.RealClock{}
	if len(optClock) > 0 {
		clock = optClock[0]
	}

	return &IntervalRetrier{
		interval:  interval,
		doer:      doer,
		clock:     clock,
		retriable: retriable,
	}
}

// Do retries performing a call until it succeeds, returns a permanent error or the context is cancelled.
func (r *IntervalRetrier) Do(ctx context.Context) error {
	ticker := r.clock.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		err := r.doer.Do(ctx)
		if err == nil {
			return nil
		}

		if !r.retriable(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C():
		}
	}
}

// Doer does something and returns an error.
type Doer interface {
	// Do performs an operation.
	//
	// It should return an error that can be checked for retriability.
	Do(ctx context.Context) error
}

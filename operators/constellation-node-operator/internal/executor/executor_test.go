/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestStartTriggersImmediateReconciliation(t *testing.T) {
	assert := assert.New(t)
	ctrl := newStubController(Result{}, nil)
	cfg := Config{
		PollingFrequency: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		RateLimiter:      &stubRateLimiter{},   // no rate limiting
	}
	exec := New(ctrl, cfg)
	// on start, the executor should trigger a reconciliation
	stopAndWait := exec.Start(context.Background())
	<-ctrl.waitUntilReconciled // makes sure to wait until reconcile was called
	ctrl.stop <- struct{}{}

	stopAndWait()

	// initial trigger
	assert.Equal(1, ctrl.reconciliationCounter)
}

func TestStartMultipleTimesIsCoalesced(t *testing.T) {
	assert := assert.New(t)
	ctrl := newStubController(Result{}, nil)
	cfg := Config{
		PollingFrequency: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		RateLimiter:      &stubRateLimiter{},   // no rate limiting
	}
	exec := New(ctrl, cfg)
	// start once
	stopAndWait := exec.Start(context.Background())
	// start again multiple times
	for i := 0; i < 10; i++ {
		_ = exec.Start(context.Background())
	}

	<-ctrl.waitUntilReconciled // makes sure to wait until reconcile was called
	ctrl.stop <- struct{}{}

	stopAndWait()

	// initial trigger. extra start calls should be coalesced
	assert.Equal(1, ctrl.reconciliationCounter)
}

func TestErrorTriggersImmediateReconciliation(t *testing.T) {
	assert := assert.New(t)
	// returning an error should trigger a reconciliation immediately
	ctrl := newStubController(Result{}, errors.New("reconciler error"))
	cfg := Config{
		PollingFrequency: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		RateLimiter:      &stubRateLimiter{},   // no rate limiting
	}
	exec := New(ctrl, cfg)
	stopAndWait := exec.Start(context.Background())
	for i := 0; i < 10; i++ {
		<-ctrl.waitUntilReconciled // makes sure to wait until reconcile was called
	}
	ctrl.stop <- struct{}{}

	stopAndWait()

	// we cannot assert the exact number of reconciliations here, because the executor might
	// select the stop case or the timer case first.
	assertBetween(assert, 10, 11, ctrl.reconciliationCounter)
}

func TestErrorTriggersRateLimiting(t *testing.T) {
	assert := assert.New(t)
	// returning an error should trigger a reconciliation immediately
	ctrl := newStubController(Result{}, errors.New("reconciler error"))
	cfg := Config{
		PollingFrequency: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		RateLimiter: &stubRateLimiter{
			whenRes: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		},
	}
	exec := New(ctrl, cfg)
	stopAndWait := exec.Start(context.Background())
	<-ctrl.waitUntilReconciled // makes sure to wait until reconcile was called once to trigger rate limiting
	ctrl.stop <- struct{}{}

	stopAndWait()

	// initial trigger. error triggers are rate limited to 1 per year
	assert.Equal(1, ctrl.reconciliationCounter)
}

func TestRequeueAfterResultRequeueInterval(t *testing.T) {
	assert := assert.New(t)
	// setting a requeue result should trigger a reconciliation after the specified delay
	ctrl := newStubController(Result{
		Requeue:      true,
		RequeueAfter: time.Microsecond,
	}, nil)
	cfg := Config{
		PollingFrequency: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		RateLimiter: &stubRateLimiter{
			whenRes: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		},
	}
	exec := New(ctrl, cfg)
	stopAndWait := exec.Start(context.Background())
	for i := 0; i < 10; i++ {
		<-ctrl.waitUntilReconciled // makes sure to wait until reconcile was called
	}
	ctrl.stop <- struct{}{}

	stopAndWait()

	// we cannot assert the exact number of reconciliations here, because the executor might
	// select the stop case or the timer case first.
	assertBetween(assert, 10, 11, ctrl.reconciliationCounter)
}

func TestExternalTrigger(t *testing.T) {
	assert := assert.New(t)
	ctrl := newStubController(Result{}, nil)
	cfg := Config{
		PollingFrequency: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		RateLimiter: &stubRateLimiter{
			whenRes: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		},
	}
	exec := New(ctrl, cfg)
	stopAndWait := exec.Start(context.Background())
	<-ctrl.waitUntilReconciled // initial trigger
	for i := 0; i < 10; i++ {
		exec.Trigger()
		<-ctrl.waitUntilReconciled // external trigger
	}
	ctrl.stop <- struct{}{}

	stopAndWait()

	// initial trigger + 10 external triggers
	assert.Equal(11, ctrl.reconciliationCounter)
}

func TestSimultaneousExternalTriggers(t *testing.T) {
	assert := assert.New(t)
	ctrl := newStubController(Result{}, nil)
	cfg := Config{
		PollingFrequency: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		RateLimiter: &stubRateLimiter{
			whenRes: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		},
	}
	exec := New(ctrl, cfg)
	stopAndWait := exec.Start(context.Background())
	<-ctrl.waitUntilReconciled // initial trigger
	for i := 0; i < 100; i++ {
		exec.Trigger() // extra trigger calls are coalesced
	}
	<-ctrl.waitUntilReconciled // external trigger
	ctrl.stop <- struct{}{}

	stopAndWait()

	// we cannot assert the exact number of reconciliations here, because the executor might
	// select the stop case or the next manual trigger case first.
	assertBetween(assert, 2, 3, ctrl.reconciliationCounter)
}

func TestContextCancel(t *testing.T) {
	assert := assert.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	ctrl := newStubController(Result{}, nil)
	cfg := Config{
		PollingFrequency: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		RateLimiter:      &stubRateLimiter{},   // no rate limiting
	}
	exec := New(ctrl, cfg)
	_ = exec.Start(ctx)        // no need to explicitly stop the executor, it will stop when the context is canceled
	<-ctrl.waitUntilReconciled // initial trigger

	// canceling the context should stop the executor without blocking
	cancel()
	ctrl.stop <- struct{}{}

	// poll for the executor stop running
	// this is necessary since the executor doesn't expose
	// a pure wait method
	assert.Eventually(func() bool {
		return !exec.Running()
	}, 5*time.Second, 10*time.Millisecond)

	// initial trigger
	assert.Equal(1, ctrl.reconciliationCounter)
}

func TestRequeueAfterPollingFrequency(t *testing.T) {
	assert := assert.New(t)
	ctrl := newStubController(Result{}, nil)
	cfg := Config{
		PollingFrequency: time.Microsecond, // basically no delay
		RateLimiter: &stubRateLimiter{
			whenRes: time.Hour * 24 * 365, // 1 year. Should be high enough to not trigger the timer in the test.
		},
	}
	exec := New(ctrl, cfg)
	stopAndWait := exec.Start(context.Background())
	for i := 0; i < 10; i++ {
		<-ctrl.waitUntilReconciled // makes sure to wait until reconcile was called
	}
	ctrl.stop <- struct{}{}

	stopAndWait()

	// we cannot assert the exact number of reconciliations here, because the executor might
	// select the stop case or the timer case first.
	assertBetween(assert, 10, 11, ctrl.reconciliationCounter)
}

type stubController struct {
	stopped               bool
	stop                  chan struct{}
	waitUntilReconciled   chan struct{}
	reconciliationCounter int
	result                Result
	err                   error
}

func newStubController(result Result, err error) *stubController {
	return &stubController{
		waitUntilReconciled: make(chan struct{}),
		stop:                make(chan struct{}, 1),
		result:              result,
		err:                 err,
	}
}

func (s *stubController) Reconcile(_ context.Context) (Result, error) {
	if s.stopped {
		return Result{}, errors.New("controller stopped")
	}
	s.reconciliationCounter++
	select {
	case <-s.stop:
		s.stopped = true
	case s.waitUntilReconciled <- struct{}{}:
	}

	return s.result, s.err
}

type stubRateLimiter struct {
	whenRes time.Duration
}

func (s *stubRateLimiter) When(_ any) time.Duration {
	return s.whenRes
}

func (s *stubRateLimiter) Forget(_ any) {}

func assertBetween(assert *assert.Assertions, minimum, maximum, actual int) {
	assert.GreaterOrEqual(actual, minimum)
	assert.LessOrEqual(actual, maximum)
}

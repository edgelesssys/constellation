/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package executor contains a task executor / scheduler for the constellation node operator.
// It is used to execute tasks (outside of the k8s specific operator controllers) with regular intervals and
// based of external triggers.
package executor

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defaultPollingFrequency = 15 * time.Minute
	// rateLimiterItem is the key used to rate limit the reconciliation loop.
	// since we don't have a reconcile request, we use a constant key.
	rateLimiterItem = "reconcile"
)

// Controller is a type with a reconcile method.
// It is modeled after the controller-runtime reconcile method,
// but reconciles on external resources instead of k8s resources.
type Controller interface {
	Reconcile(ctx context.Context) (Result, error)
}

// taskExecutor is a task executor / scheduler.
// It will call the reconcile method of the given controller with a regular interval
// or when triggered externally.
type taskExecutor struct {
	running atomic.Bool

	// controller is the controller to be reconciled.
	controller Controller

	// pollingFrequency is the default frequency with which the controller is reconciled
	// if no external trigger is received and no requeue is requested by the controller.
	pollingFrequency time.Duration
	// rateLimiter is used to rate limit the reconciliation loop.
	rateLimiter RateLimiter
	// externalTrigger is used to trigger a reconciliation immediately from the outside.
	// multiple triggers in a short time will be coalesced into one externalTrigger.
	externalTrigger chan struct{}
	// stop is used to stop the reconciliation loop.
	stop chan struct{}
}

// New creates a new Executor.
func New(controller Controller, cfg Config) Executor {
	cfg.applyDefaults()
	return &taskExecutor{
		controller:       controller,
		pollingFrequency: cfg.PollingFrequency,
		rateLimiter:      cfg.RateLimiter,
		externalTrigger:  make(chan struct{}, 1),
		stop:             make(chan struct{}, 1),
	}
}

// Executor is a task executor / scheduler.
type Executor interface {
	Start(ctx context.Context) StopWaitFn
	Running() bool
	Trigger()
}

// StopWaitFn is a function that can be called to stop the executor and wait for it to stop.
type StopWaitFn func()

// Start starts the executor in a separate go routine.
// Call Stop to stop the executor.
// IMPORTANT: The executor can only be started once.
func (e *taskExecutor) Start(ctx context.Context) StopWaitFn {
	wg := &sync.WaitGroup{}
	logr := log.FromContext(ctx)
	stopWait := func() {
		defer wg.Wait()
		e.Stop()
	}

	// this will return early if the executor is already running
	// if the executor is not running, set the running flag to true
	// and continue
	if !e.running.CompareAndSwap(false, true) {
		return stopWait
	}
	// execute is used by the go routines below to communicate
	// that a reconciliation should happen
	execute := make(chan struct{}, 1)
	// nextScheduledReconcile is used to communicate the next scheduled reconciliation time
	nextScheduledReconcile := make(chan time.Duration, 1)
	// trigger a reconciliation on startup
	nextScheduledReconcile <- 0

	wg.Add(2)

	// timer routine is responsible for triggering the reconciliation after the timer expires
	// or when triggered externally
	go func() {
		defer func() {
			e.running.Store(false)
		}()
		defer wg.Done()
		defer close(execute)
		defer logr.Info("Timer stopped")
		for {
			nextScheduledReconcileAfter := <-nextScheduledReconcile
			timer := *time.NewTimer(nextScheduledReconcileAfter)
			select {
			case <-e.stop:
				return
			case <-ctx.Done():
				return
			case <-e.externalTrigger:
			case <-timer.C:
			}
			execute <- struct{}{}
		}
	}()

	// executor routine is responsible for executing the reconciliation
	go func() {
		defer wg.Done()
		defer close(nextScheduledReconcile)
		defer logr.Info("Executor stopped")

		for {
			_, ok := <-execute
			// execute channel closed. executor should stop
			if !ok {
				return
			}
			res, err := e.controller.Reconcile(ctx)
			var requeueAfter time.Duration
			switch {
			case err != nil:
				logr.Error(err, "reconciliation failed")
				requeueAfter = e.rateLimiter.When(rateLimiterItem) // requeue with rate limiter
			case res.Requeue && res.RequeueAfter != 0:
				e.rateLimiter.Forget(rateLimiterItem) // reset the rate limiter
				requeueAfter = res.RequeueAfter       // requeue after the given duration
			case res.Requeue:
				requeueAfter = e.rateLimiter.When(rateLimiterItem) // requeue with rate limiter
			default:
				e.rateLimiter.Forget(rateLimiterItem) // reset the rate limiter
				requeueAfter = e.pollingFrequency     // default polling frequency
			}

			nextScheduledReconcile <- requeueAfter
		}
	}()

	return stopWait
}

// Stop stops the executor.
// It does not block until the executor is stopped.
func (e *taskExecutor) Stop() {
	select {
	case e.stop <- struct{}{}:
	default:
	}
	close(e.stop)
}

// Running returns true if the executor is running.
// When the executor is stopped, it is not running anymore.
func (e *taskExecutor) Running() bool {
	return e.running.Load()
}

// Trigger triggers a reconciliation.
// If a reconciliation is already pending, this call is a no-op.
func (e *taskExecutor) Trigger() {
	select {
	case e.externalTrigger <- struct{}{}:
	default:
	}
}

// Config is the configuration for the executor.
type Config struct {
	PollingFrequency time.Duration
	RateLimiter      RateLimiter
}

// NewDefaultConfig creates a new default configuration.
func NewDefaultConfig() Config {
	cfg := Config{}
	cfg.applyDefaults()
	return cfg
}

func (c *Config) applyDefaults() {
	if c.PollingFrequency == 0 {
		c.PollingFrequency = defaultPollingFrequency
	}
	if c.RateLimiter == nil {
		c.RateLimiter = workqueue.DefaultTypedControllerRateLimiter[any]()
	}
}

// Result is the result of a reconciliation.
type Result struct {
	Requeue      bool
	RequeueAfter time.Duration
}

// RateLimiter is a stripped down version of the controller-runtime ratelimiter.RateLimiter interface.
type RateLimiter interface {
	// When gets an item and gets to decide how long that item should wait
	When(item any) time.Duration
	// Forget indicates that an item is finished being retried.  Doesn't matter whether its for perm failing
	// or for success, we'll stop tracking it
	Forget(item any)
}

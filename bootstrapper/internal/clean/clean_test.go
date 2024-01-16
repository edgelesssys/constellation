/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package clean

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestNew(t *testing.T) {
	assert := assert.New(t)

	cleaner := New(&spyStopper{stopped: &atomic.Bool{}})
	assert.NotNil(cleaner)
	assert.NotEmpty(cleaner.stoppers)
}

func TestWith(t *testing.T) {
	assert := assert.New(t)

	cleaner := New().With(&spyStopper{stopped: &atomic.Bool{}})
	assert.NotEmpty(cleaner.stoppers)
}

func TestClean(t *testing.T) {
	assert := assert.New(t)

	stopper := &spyStopper{stopped: &atomic.Bool{}}
	cleaner := New(stopper)
	go cleaner.Start()
	cleaner.Clean()
	cleaner.Done()
	assert.True(stopper.stopped.Load())
	// call again to make sure it doesn't panic or block or clean up again
	cleaner.Clean()
	assert.True(stopper.stopped.Load())
}

func TestCleanBeforeStart(t *testing.T) {
	assert := assert.New(t)
	// calling Clean before Start should work
	stopper := &spyStopper{stopped: &atomic.Bool{}}
	cleaner := New(stopper)
	cleaner.Clean()
	cleaner.Start()
	cleaner.Done()
	assert.True(stopper.stopped.Load())
}

func TestConcurrent(t *testing.T) {
	assert := assert.New(t)
	// calling Clean concurrently should call Stop exactly once

	stopper := &spyStopper{stopped: &atomic.Bool{}}
	cleaner := New(stopper)

	parallelism := 10
	wg := sync.WaitGroup{}

	start := func() {
		defer wg.Done()
		cleaner.Start()
	}

	clean := func() {
		defer wg.Done()
		cleaner.Clean()
	}

	done := func() {
		defer wg.Done()
		cleaner.Done()
	}

	wg.Add(3 * parallelism)
	for i := 0; i < parallelism; i++ {
		go start()
		go clean()
		go done()
	}
	wg.Wait()
	cleaner.Done()
	assert.True(stopper.stopped.Load())
}

type spyStopper struct {
	stopped *atomic.Bool
}

func (s *spyStopper) Stop() {
	s.stopped.Store(true)
}

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
	goleak.VerifyTestMain(m)
}

func TestNew(t *testing.T) {
	assert := assert.New(t)

	cleaner := New(&spyStopper{})
	assert.NotNil(cleaner)
	assert.NotEmpty(cleaner.stoppers)
}

func TestWith(t *testing.T) {
	assert := assert.New(t)

	cleaner := New().With(&spyStopper{})
	assert.NotEmpty(cleaner.stoppers)
}

func TestClean(t *testing.T) {
	assert := assert.New(t)

	stopper := &spyStopper{}
	cleaner := New(stopper)
	go cleaner.Start()
	cleaner.Clean()
	cleaner.Done()
	assert.Equal(int64(1), atomic.LoadInt64(&stopper.stopped))
	// call again to make sure it doesn't panic or block or clean up again
	cleaner.Clean()
	assert.Equal(int64(1), atomic.LoadInt64(&stopper.stopped))
}

func TestCleanBeforeStart(t *testing.T) {
	assert := assert.New(t)
	// calling Clean before Start should work
	stopper := &spyStopper{}
	cleaner := New(stopper)
	cleaner.Clean()
	cleaner.Start()
	cleaner.Done()
	assert.Equal(int64(1), atomic.LoadInt64(&stopper.stopped))
}

func TestConcurrent(t *testing.T) {
	assert := assert.New(t)
	// calling Clean concurrently should call Stop exactly once

	stopper := &spyStopper{}
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
	assert.Equal(int64(1), atomic.LoadInt64(&stopper.stopped))
}

type spyStopper struct {
	stopped int64
}

func (s *spyStopper) Stop() {
	atomic.AddInt64(&s.stopped, 1)
}

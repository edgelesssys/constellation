/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package clean

import (
	"sync"
)

// Cleaner can be used to stop a list of services gracefully.
// To register an arbitrary amount of stoppers either use New or With.
// Start needs to be called to ready the Cleaner, then Clean will activate it.
// Done can be used to wait for Cleaner to run all registered stoppers.
type Cleaner struct {
	stoppers  []stopper
	stopC     chan struct{}
	startOnce sync.Once
	wg        sync.WaitGroup
}

// New creates a new cleaner.
func New(stoppers ...stopper) *Cleaner {
	res := &Cleaner{
		stoppers: stoppers,
		stopC:    make(chan struct{}, 1),
	}
	res.wg.Add(1) // for the Start goroutine
	return res
}

// With adds a new stopper to the cleaner.
func (c *Cleaner) With(stopper stopper) *Cleaner {
	c.stoppers = append(c.stoppers, stopper)
	return c
}

// Start blocks until it receives a stop message, stops all services gracefully and returns.
func (c *Cleaner) Start() {
	c.startOnce.Do(func() {
		defer c.wg.Done()
		// wait for the stop message
		<-c.stopC

		c.wg.Add(len(c.stoppers))
		for _, stopItem := range c.stoppers {
			go func(stopItem stopper) {
				defer c.wg.Done()
				stopItem.Stop()
			}(stopItem)
		}
	})
}

// Clean initiates the cleanup but does not wait for it to complete.
func (c *Cleaner) Clean() {
	// try to enqueue the stop message once
	// if the channel is full, the message is dropped
	select {
	case c.stopC <- struct{}{}:
	default:
	}
}

// Done waits for the cleanup to complete.
func (c *Cleaner) Done() {
	c.wg.Wait()
}

type stopper interface {
	Stop()
}

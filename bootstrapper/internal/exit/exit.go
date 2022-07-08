package exit

import (
	"sync"
)

type cleaner struct {
	stoppers []stopper

	cleanupDone bool
	wg          sync.WaitGroup
	mux         sync.Mutex
}

// New creates a new cleaner.
func New(stoppers ...stopper) *cleaner {
	return &cleaner{
		stoppers: stoppers,
	}
}

// With adds a new stopper to the cleaner.
func (c *cleaner) With(stopper stopper) *cleaner {
	c.stoppers = append(c.stoppers, stopper)
	return c
}

// Clean stops all services gracefully.
func (c *cleaner) Clean() {
	// only cleanup once
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.cleanupDone {
		return
	}

	c.wg.Add(len(c.stoppers))
	for _, stopItem := range c.stoppers {
		go func(stopItem stopper) {
			stopItem.Stop()
			c.wg.Done()
		}(stopItem)
	}
	c.wg.Wait()
	c.cleanupDone = true
}

type stopper interface {
	Stop()
}

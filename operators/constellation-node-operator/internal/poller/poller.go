// Package poller implements a poller that can be used to wait for a condition to be met.
// The poller is designed to be a replacement for the azure-sdk-for-go poller
// with exponential backoff and an injectable clock.
// reference: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azcore@v1.1.1/runtime#Poller .
package poller

import (
	"context"
	"errors"
	"time"

	"k8s.io/utils/clock"
)

// PollUntilDoneOptions provides options for the Poller.
// Used to specify backoff and clock options.
type PollUntilDoneOptions struct {
	StartingBackoff time.Duration
	MaxBackoff      time.Duration
	clock.Clock
}

// NewPollUntilDoneOptions creates a new PollUntilDoneOptions with the default values and a real clock.
func NewPollUntilDoneOptions() *PollUntilDoneOptions {
	return &PollUntilDoneOptions{
		Clock: clock.RealClock{},
	}
}

// Poller is a poller that can be used to wait for a condition to be met.
// The poller is designed to be a replacement for the azure-sdk-for-go poller
// with exponential backoff and an injectable clock.
// reference: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azcore@v1.1.1/runtime#Poller .
type Poller[T any] struct {
	handler PollingHandler[T]
	err     error
	result  *T
	done    bool
}

// New creates a new Poller.
func New[T any](handler PollingHandler[T]) *Poller[T] {
	return &Poller[T]{
		handler: handler,
		result:  new(T),
	}
}

// PollUntilDone polls the handler until the condition is met or the context is cancelled.
func (p *Poller[T]) PollUntilDone(ctx context.Context, options *PollUntilDoneOptions) (T, error) {
	if options == nil {
		options = NewPollUntilDoneOptions()
	}
	if options.MaxBackoff == 0 {
		options.MaxBackoff = time.Minute
	}
	if options.StartingBackoff < time.Second {
		options.StartingBackoff = time.Second
	}
	backoff := options.StartingBackoff
	for {
		timer := options.Clock.NewTimer(backoff)
		err := p.Poll(ctx)
		if err != nil {
			return *new(T), err
		}
		if p.Done() {
			return p.Result(ctx)
		}
		select {
		case <-ctx.Done():
			return *new(T), ctx.Err()
		case <-timer.C():
		}
		if backoff >= options.MaxBackoff/2 {
			backoff = options.MaxBackoff
		} else {
			backoff *= 2
		}
	}
}

// Poll polls the handler.
func (p *Poller[T]) Poll(ctx context.Context) error {
	return p.handler.Poll(ctx)
}

// Done returns true if the condition is met.
func (p *Poller[T]) Done() bool {
	return p.handler.Done()
}

// Result returns the result of the poller if the condition is met.
// If the condition is not met, an error is returned.
func (p *Poller[T]) Result(ctx context.Context) (T, error) {
	if !p.Done() {
		return *new(T), errors.New("poller is in a non-terminal state")
	}
	if p.done {
		// the result has already been retrieved, return the cached value
		if p.err != nil {
			return *new(T), p.err
		}
		return *p.result, nil
	}
	err := p.handler.Result(ctx, p.result)
	p.done = true
	if err != nil {
		p.err = err
		return *new(T), p.err
	}

	return *p.result, nil
}

// PollingHandler is a handler that can be used to poll for a condition to be met.
type PollingHandler[T any] interface {
	Done() bool
	Poll(ctx context.Context) error
	Result(ctx context.Context, out *T) error
}

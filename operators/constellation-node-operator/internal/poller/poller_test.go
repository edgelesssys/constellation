/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package poller

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/stretchr/testify/assert"
	testclock "k8s.io/utils/clock/testing"
)

func TestResult(t *testing.T) {
	testCases := map[string]struct {
		done       bool
		pollErr    error
		resultErr  error
		wantErr    bool
		wantResult int
	}{
		"result called before poller is done": {
			wantErr: true,
		},
		"result returns error": {
			done:      true,
			resultErr: errors.New("result error"),
			wantErr:   true,
		},
		"result succeeds": {
			done: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			poller := New[int](&stubPoller[int]{
				result:    &tc.wantResult,
				done:      tc.done,
				pollErr:   tc.pollErr,
				resultErr: tc.resultErr,
			})
			_, firstErr := poller.Result(context.Background())
			if tc.wantErr {
				assert.Error(firstErr)
				// calling Result again should return the same error
				_, secondErr := poller.Result(context.Background())
				assert.Equal(firstErr, secondErr)
				return
			}
			assert.NoError(firstErr)
			// calling Result again should still not return an error
			_, secondErr := poller.Result(context.Background())
			assert.NoError(secondErr)
		})
	}
}

func TestPollUntilDone(t *testing.T) {
	testCases := map[string]struct {
		messages   []message
		maxBackoff time.Duration
		resultErr  error
		wantErr    bool
		wantResult int
	}{
		"poll succeeds on first try": {
			messages: []message{
				{pollErr: to.Ptr[error](nil), done: to.Ptr(true)},
				{done: to.Ptr(true)}, // Result() will call Done() after the last poll
			},
			wantResult: 1,
		},
		"poll succeeds on fourth try": {
			messages: []message{
				{pollErr: to.Ptr[error](nil), done: to.Ptr(false), backoff: time.Second},
				{pollErr: to.Ptr[error](nil), done: to.Ptr(false), backoff: 2 * time.Second},
				{pollErr: to.Ptr[error](nil), done: to.Ptr(false), backoff: 4 * time.Second},
				{pollErr: to.Ptr[error](nil), done: to.Ptr(true)},
				{done: to.Ptr(true)}, // Result() will call Done() after the last poll
			},
			wantResult: 1,
		},
		"max backoff reached": {
			messages: []message{
				{pollErr: to.Ptr[error](nil), done: to.Ptr(false), backoff: time.Second},
				{pollErr: to.Ptr[error](nil), done: to.Ptr(false), backoff: time.Second},
				{pollErr: to.Ptr[error](nil), done: to.Ptr(false), backoff: time.Second},
				{pollErr: to.Ptr[error](nil), done: to.Ptr(true)},
				{done: to.Ptr(true)}, // Result() will call Done() after the last poll
			},
			maxBackoff: time.Second,
			wantResult: 1,
		},
		"poll errors": {
			messages: []message{
				{pollErr: to.Ptr(errors.New("poll error"))},
			},
			wantErr: true,
		},
		"result errors": {
			messages: []message{
				{pollErr: to.Ptr[error](nil), done: to.Ptr(true)},
				{done: to.Ptr(true)}, // Result() will call Done() after the last poll
			},
			resultErr: errors.New("result error"),
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			doneC := make(chan bool)
			pollC := make(chan error)
			poller := New[int](&fakePoller[int]{
				result:    &tc.wantResult,
				resultErr: tc.resultErr,
				doneC:     doneC,
				pollC:     pollC,
			})
			clock := testclock.NewFakeClock(time.Now())
			var gotResult int
			var gotErr error
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				gotResult, gotErr = poller.PollUntilDone(context.Background(), &PollUntilDoneOptions{
					MaxBackoff: tc.maxBackoff,
					Clock:      clock,
				})
			}()

			for _, msg := range tc.messages {
				if msg.pollErr != nil {
					pollC <- *msg.pollErr
				}
				if msg.done != nil {
					doneC <- *msg.done
				}
				clock.Step(msg.backoff)
			}

			wg.Wait()
			if tc.wantErr {
				assert.Error(gotErr)
				return
			}
			assert.NoError(gotErr)
			assert.Equal(tc.wantResult, gotResult)
		})
	}
}

type stubPoller[T any] struct {
	result    *T
	done      bool
	pollErr   error
	resultErr error
}

func (s *stubPoller[T]) Poll(_ context.Context) error {
	return s.pollErr
}

func (s *stubPoller[T]) Done() bool {
	return s.done
}

func (s *stubPoller[T]) Result(_ context.Context, out *T) error {
	*out = *s.result
	return s.resultErr
}

type message struct {
	pollErr *error
	done    *bool
	backoff time.Duration
}

type fakePoller[T any] struct {
	result    *T
	resultErr error

	doneC chan bool
	pollC chan error
}

func (s *fakePoller[T]) Poll(_ context.Context) error {
	return <-s.pollC
}

func (s *fakePoller[T]) Done() bool {
	return <-s.doneC
}

func (s *fakePoller[T]) Result(_ context.Context, out *T) error {
	*out = *s.result
	return s.resultErr
}

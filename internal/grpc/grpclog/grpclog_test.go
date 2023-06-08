/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// grpclog provides a logging utilities for gRPC.
package grpclog

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/connectivity"
)

func TestLogStateChanges(t *testing.T) {
	testCases := map[string]struct {
		name   string
		conn   getStater
		assert func(t *testing.T, lg *fakeLog, isReadyCallbackCalled bool)
	}{
		"state: connecting, ready": {
			conn: &fakeConn{
				states: []connectivity.State{
					connectivity.Connecting,
					connectivity.Ready,
					connectivity.Ready,
				},
			},
			assert: func(t *testing.T, lg *fakeLog, isReadyCallbackCalled bool) {
				require.Len(t, lg.msgs, 3)
				assert.Equal(t, "Connection state started as CONNECTING", lg.msgs[0])
				assert.Equal(t, "Connection state changed to CONNECTING", lg.msgs[1])
				assert.Equal(t, "Connection ready", lg.msgs[2])
			},
		},
		"state: ready": {
			conn: &fakeConn{
				states: []connectivity.State{
					connectivity.Ready,
					connectivity.Idle,
				},
				stopWaitForChange: false,
			},
			assert: func(t *testing.T, lg *fakeLog, isReadyCallbackCalledCallback bool) {
				require.Len(t, lg.msgs, 2)
				assert.Equal(t, "Connection state started as READY", lg.msgs[0])
				assert.Equal(t, "Connection ready", lg.msgs[1])
				assert.True(t, isReadyCallbackCalledCallback)
			},
		},
		"no WaitForStateChange (e.g. when context is canceled)": {
			conn: &fakeConn{
				states: []connectivity.State{
					connectivity.Connecting,
					connectivity.Idle,
				},
				stopWaitForChange: true,
			},
			assert: func(t *testing.T, lg *fakeLog, isReadyCallbackCalled bool) {
				require.Len(t, lg.msgs, 2)
				assert.Equal(t, "Connection state started as CONNECTING", lg.msgs[0])
				assert.Equal(t, "Connection state ended with CONNECTING", lg.msgs[1])
				assert.False(t, isReadyCallbackCalled)
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := &fakeLog{}

			var wg sync.WaitGroup
			isReadyCallbackCalled := false
			LogStateChangesUntilReady(context.Background(), tc.conn, logger, &wg, func() { isReadyCallbackCalled = true })
			wg.Wait()
			tc.assert(t, logger, isReadyCallbackCalled)
		})
	}
}

type fakeLog struct {
	msgs []string
}

func (f *fakeLog) Debugf(format string, args ...any) {
	f.msgs = append(f.msgs, fmt.Sprintf(format, args...))
}

type fakeConn struct {
	states            []connectivity.State
	idx               int
	stopWaitForChange bool
}

func (f *fakeConn) GetState() connectivity.State {
	if f.idx > len(f.states)-1 {
		return f.states[len(f.states)-1]
	}
	res := f.states[f.idx]
	f.idx++
	return res
}

func (f *fakeConn) WaitForStateChange(context.Context, connectivity.State) bool {
	if f.stopWaitForChange {
		return false
	}
	return f.idx < len(f.states)
}

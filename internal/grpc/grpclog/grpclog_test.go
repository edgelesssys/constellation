/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// grpclog provides a logging utilities for gRPC.
package grpclog

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/connectivity"
)

func TestLogStateChanges(t *testing.T) {
	testCases := map[string]struct {
		name   string
		conn   getStater
		assert func(t *testing.T, lg *spyLogger, isReady bool)
	}{
		"log state changes": {
			conn: &fakeConn{
				states: []connectivity.State{
					connectivity.Connecting,
					connectivity.Ready,
				},
			},
			assert: func(t *testing.T, spy *spyLogger, isReady bool) {
				require.Len(t, spy.msgs, 3)
				assert.Equal(t, "Connection state started as CONNECTING", spy.msgs[0])
				assert.Equal(t, "Connection state changed to CONNECTING", spy.msgs[1])
				assert.Equal(t, "Connection ready", spy.msgs[2])
			},
		},
		"WaitForStateChange returns false (e.g. when context is canceled)": {
			conn: &fakeConn{
				states: []connectivity.State{
					connectivity.Connecting,
					connectivity.Idle,
				},
				stopWaitForChange: true,
			},
			assert: func(t *testing.T, spy *spyLogger, isReady bool) {
				require.Len(t, spy.msgs, 2)
				assert.Equal(t, "Connection state started as CONNECTING", spy.msgs[0])
				assert.Equal(t, "Connection state ended with CONNECTING", spy.msgs[1])
				assert.False(t, isReady)
			},
		},
		"initial connection state is Ready": {
			conn: &fakeConn{
				states: []connectivity.State{
					connectivity.Ready,
					connectivity.Idle,
				},
				stopWaitForChange: false,
			},
			assert: func(t *testing.T, spy *spyLogger, isReady bool) {
				require.Len(t, spy.msgs, 2)
				assert.Equal(t, "Connection state started as READY", spy.msgs[0])
				assert.Equal(t, "Connection ready", spy.msgs[1])
				assert.True(t, isReady)
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := &spyLogger{}

			isReady := false
			waitFn := LogStateChangesUntilReady(context.Background(), tc.conn, logger, func() { isReady = true })
			waitFn()
			tc.assert(t, logger, isReady)
		})
	}
}

type spyLogger struct {
	msgs []string
}

func (f *spyLogger) Debugf(format string, args ...any) {
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

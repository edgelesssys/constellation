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
	"google.golang.org/grpc/connectivity"
)

func TestLogStateChanges(t *testing.T) {
	testCases := []struct {
		name    string
		conn    getStater
		isReady bool
		assert  func(t *testing.T, lg *fakeLog, isReady bool)
	}{
		{
			name: "log state changes",
			conn: &fakeConn{
				states: []connectivity.State{
					connectivity.Connecting,
					connectivity.Ready,
				},
			},
			isReady: true,
			assert: func(t *testing.T, lg *fakeLog, isReady bool) {
				assert.Equal(t, "Connection state started as CONNECTING", lg.msgs[0])
				assert.Equal(t, "Connection state changed to CONNECTING", lg.msgs[1])
				assert.Equal(t, "Connection ready", lg.msgs[2])
				assert.Len(t, lg.msgs, 3)
			},
		},
		{
			name: "WaitForStateChange returns false (e.g. when context is canceled)",
			conn: &fakeConn{
				states: []connectivity.State{
					connectivity.Connecting,
					connectivity.Idle,
				},
				stopWaitForChange: true,
			},
			isReady: false,
			assert: func(t *testing.T, lg *fakeLog, isReady bool) {
				assert.Equal(t, "Connection state started as CONNECTING", lg.msgs[0])
				assert.Equal(t, "Connection state ended with CONNECTING", lg.msgs[1])
				assert.Len(t, lg.msgs, 2)
				assert.False(t, isReady)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lg := &fakeLog{}

			var wg sync.WaitGroup
			isReady := false
			LogStateChanges(context.Background(), tc.conn, lg, &wg, func() { isReady = true })
			wg.Wait()
			tc.assert(t, lg, isReady)
		})
	}
}

type fakeLog struct {
	msgs []string
}

func (f *fakeLog) Debugf(format string, args ...interface{}) {
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

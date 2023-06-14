/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// grpclog provides a logging utilities for gRPC.
package grpclog

import (
	"context"
	"sync"

	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/peer"
)

// PeerAddrFromContext returns a peer's address from context, or "unknown" if not found.
func PeerAddrFromContext(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "unknown"
	}
	return p.Addr.String()
}

// LogStateChangesUntilReady logs the state changes of a gRPC connection.
func LogStateChangesUntilReady(ctx context.Context, conn getStater, log debugLog, wg *sync.WaitGroup, isReadyCallback func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		state := conn.GetState()
		log.Debugf("Connection state started as %s", state)
		for ; state != connectivity.Ready && conn.WaitForStateChange(ctx, state); state = conn.GetState() {
			log.Debugf("Connection state changed to %s", state)
		}
		if state == connectivity.Ready {
			log.Debugf("Connection ready")
			isReadyCallback()
		} else {
			log.Debugf("Connection state ended with %s", state)
		}
	}()
}

type getStater interface {
	GetState() connectivity.State
	WaitForStateChange(context.Context, connectivity.State) bool
}

type debugLog interface {
	Debugf(format string, args ...any)
}

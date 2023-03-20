/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package testdialer provides a fake dialer for testing.
package testdialer

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc/test/bufconn"
)

// BufconnDialer is a fake dialer based on gRPC bufconn package.
type BufconnDialer struct {
	mut       sync.Mutex
	listeners map[string]*bufconn.Listener
}

// NewBufconnDialer creates a new bufconn dialer for testing.
func NewBufconnDialer() *BufconnDialer {
	return &BufconnDialer{listeners: make(map[string]*bufconn.Listener)}
}

// DialContext implements the Dialer interface.
func (b *BufconnDialer) DialContext(ctx context.Context, _, address string) (net.Conn, error) {
	b.mut.Lock()
	listener, ok := b.listeners[address]
	b.mut.Unlock()
	if !ok {
		return nil, fmt.Errorf("could not connect to server on %v", address)
	}
	return listener.DialContext(ctx)
}

// GetListener returns a fake listener that is coupled with this dialer.
func (b *BufconnDialer) GetListener(endpoint string) net.Listener {
	listener := bufconn.Listen(1024)
	b.mut.Lock()
	b.listeners[endpoint] = listener
	b.mut.Unlock()
	return listener
}

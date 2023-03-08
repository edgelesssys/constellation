/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package nodelock handles locking operations on the node.
package nodelock

import (
	"sync"

	"github.com/edgelesssys/constellation/v2/internal/attestation/initialize"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
)

// Lock locks the node once there the join or the init is at a point
// where there is no turning back and the other operation does not need
// to continue.
//
// This can be viewed as a state machine with two states: unlocked and locked.
// There is no way to unlock, so the state changes only once from unlock to
// locked.
type Lock struct {
	tpm vtpm.TPMOpenFunc
	mux *sync.Mutex
}

// New creates a new NodeLock, which is unlocked.
func New(tpm vtpm.TPMOpenFunc) *Lock {
	return &Lock{
		tpm: tpm,
		mux: &sync.Mutex{},
	}
}

// TryLockOnce tries to lock the node. If the node is already locked, it
// returns false. If the node is unlocked, it locks it and returns true.
func (l *Lock) TryLockOnce(clusterID []byte) (bool, error) {
	if !l.mux.TryLock() {
		return false, nil
	}

	return true, initialize.MarkNodeAsBootstrapped(l.tpm, clusterID)
}

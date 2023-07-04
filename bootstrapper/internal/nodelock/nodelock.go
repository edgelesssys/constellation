/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package nodelock handles locking operations on the node.
package nodelock

import (
	"io"
	"sync/atomic"

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
	tpm    vtpm.TPMOpenFunc
	marker tpmMarker
	inner  atomic.Bool
}

// New creates a new NodeLock, which is unlocked.
func New(tpm vtpm.TPMOpenFunc) *Lock {
	return &Lock{
		tpm:    tpm,
		marker: initialize.MarkNodeAsBootstrapped,
	}
}

// TryLockOnce tries to lock the node. If the node is already locked, it
// returns false. If the node is unlocked, it locks it and returns true.
func (l *Lock) TryLockOnce(clusterID []byte) (bool, error) {
	// CompareAndSwap first checks if the node is currently unlocked.
	// If it was already locked, it returns early.
	// If it is unlocked, it swaps the value to locked atomically and continues.
	if !l.inner.CompareAndSwap(unlocked, locked) {
		return false, nil
	}

	return true, l.marker(l.tpm, clusterID)
}

// tpmMarker is a function that marks the node as bootstrapped in the TPM.
type tpmMarker func(openDevice func() (io.ReadWriteCloser, error), clusterID []byte) error

const (
	unlocked = false
	locked   = true
)

package nodelock

import (
	"sync"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
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
	locked bool
	state  *sync.Mutex
	mux    *sync.RWMutex
}

// New creates a new NodeLock, which is unlocked.
func New(tpm vtpm.TPMOpenFunc) *Lock {
	return &Lock{
		tpm:   tpm,
		state: &sync.Mutex{},
		mux:   &sync.RWMutex{},
	}
}

// TryLockOnce tries to lock the node. If the node is already locked, it
// returns false. If the node is unlocked, it locks it and returns true.
func (l *Lock) TryLockOnce(ownerID, clusterID []byte) (bool, error) {
	success := l.state.TryLock()
	if success {
		l.mux.Lock()
		defer l.mux.Unlock()
		l.locked = true
		if err := vtpm.MarkNodeAsBootstrapped(l.tpm, ownerID, clusterID); err != nil {
			return success, err
		}
	}
	return success, nil
}

// Locked returns true if the node is locked.
func (l *Lock) Locked() bool {
	l.mux.RLock()
	defer l.mux.RUnlock()
	return l.locked
}

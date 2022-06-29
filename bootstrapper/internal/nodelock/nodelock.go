package nodelock

import "sync"

// Lock locks the node once there the join or the init is at a point
// where there is no turning back and the other operation does not need
// to continue.
//
// This can be viewed as a state machine with two states: unlocked and locked.
// There is no way to unlock, so the state changes only once from unlock to
// locked.
type Lock struct {
	mux *sync.Mutex
}

// New creates a new NodeLock, which is unlocked.
func New() *Lock {
	return &Lock{mux: &sync.Mutex{}}
}

// TryLockOnce tries to lock the node. If the node is already locked, it
// returns false. If the node is unlocked, it locks it and returns true.
func (n *Lock) TryLockOnce() bool {
	return n.mux.TryLock()
}

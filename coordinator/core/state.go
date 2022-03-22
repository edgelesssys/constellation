package core

import (
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/state"
)

// GetState returns the current state.
func (c *Core) GetState() state.State {
	return c.state.Get()
}

// RequireState checks if the peer is in one of the desired states and returns an error otherwise.
func (c *Core) RequireState(states ...state.State) error {
	return c.state.Require(states...)
}

// AdvanceState advances the state. It also marks the peer as initialized for the corresponding state transition.
func (c *Core) AdvanceState(newState state.State, ownerID, clusterID []byte) error {
	if newState != state.Failed && c.state.Get() == state.AcceptingInit {
		if err := c.data().PutClusterID(clusterID); err != nil {
			return err
		}
		if err := vtpm.MarkNodeAsInitialized(c.openTPM, ownerID, clusterID); err != nil {
			return err
		}
	}
	c.state.Advance(newState)
	return nil
}

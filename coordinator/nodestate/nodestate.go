package nodestate

import (
	"fmt"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/role"
)

const nodeStatePath = "/run/state/constellation/node_state.json"

// NodeState is the state of a constellation node that is required to recover from a reboot.
// Can be persisted to disk and reloaded later.
type NodeState struct {
	Role       role.Role
	VPNPrivKey []byte
}

// FromFile reads a NodeState from disk.
func FromFile(fileHandler file.Handler) (*NodeState, error) {
	nodeState := &NodeState{}
	if err := fileHandler.ReadJSON(nodeStatePath, nodeState); err != nil {
		return nil, fmt.Errorf("could not load node state: %w", err)
	}
	return nodeState, nil
}

// ToFile writes a NodeState to disk.
func (nodeState *NodeState) ToFile(fileHandler file.Handler) error {
	return fileHandler.WriteJSON(nodeStatePath, nodeState, false)
}

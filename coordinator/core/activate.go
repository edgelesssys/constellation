package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// SetNodeActive activates as node and joins the cluster.
func (c *Core) SetNodeActive(diskKey, ownerID, clusterID []byte, kubeAPIendpoint, token, discoveryCACertHash string) (reterr error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	if err := c.RequireState(state.AcceptingInit); err != nil {
		return fmt.Errorf("node is not in required state for activation: %w", err)
	}

	if len(ownerID) == 0 || len(clusterID) == 0 {
		c.zaplogger.Error("Missing data to taint worker node as initialized")
		return errors.New("missing data to taint worker node as initialized")
	}

	// If any of the following actions fail, we cannot revert.
	// Thus, mark this peer as failed.
	defer func() {
		if reterr != nil {
			_ = c.AdvanceState(state.Failed, nil, nil)
		}
	}()

	// AdvanceState MUST be called before any other functions that are not sanity checks or otherwise required
	// This ensures the node is marked as initialzed before the node is in a state that allows code execution
	// Any new additions to ActivateAsNode MUST come after
	if err := c.AdvanceState(state.IsNode, ownerID, clusterID); err != nil {
		return fmt.Errorf("advancing node state: %w", err)
	}

	// TODO: SSH keys are currently not available from the Aaas, so we can't create user SSH keys here.

	if err := c.PersistNodeState(role.Node, "", ownerID, clusterID); err != nil {
		return fmt.Errorf("persisting node state: %w", err)
	}

	if err := c.UpdateDiskPassphrase(string(diskKey)); err != nil {
		return fmt.Errorf("updateing disk passphrase: %w", err)
	}

	btd := &kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: kubeAPIendpoint,
		Token:             token,
		CACertHashes:      []string{discoveryCACertHash},
	}
	if err := c.JoinCluster(context.TODO(), btd, "", role.Node); err != nil {
		return fmt.Errorf("joining Kubernetes cluster: %w", err)
	}

	return nil
}

// SetCoordinatorActive activates as coordinator.
func (c *Core) SetCoordinatorActive() error {
	panic("not implemented")
}

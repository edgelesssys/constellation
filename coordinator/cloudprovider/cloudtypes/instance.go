package cloudtypes

import "github.com/edgelesssys/constellation/coordinator/role"

// Instance describes metadata of a peer.
type Instance struct {
	Name          string
	ProviderID    string
	Role          role.Role
	PrivateIPs    []string
	PublicIPs     []string
	AliasIPRanges []string
	// SSHKeys maps usernames to ssh public keys.
	SSHKeys map[string][]string
}

package cloudprovider

import (
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/role"
)

// ExtractRole extracts role from cloud provider metadata.
func ExtractRole(metadata map[string]string) role.Role {
	switch metadata[core.RoleMetadataKey] {
	case role.Coordinator.String():
		return role.Coordinator
	case role.Node.String():
		return role.Node
	default:
		return role.Unknown
	}
}

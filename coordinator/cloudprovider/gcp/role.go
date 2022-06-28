package gcp

import (
	"github.com/edgelesssys/constellation/coordinator/role"
)

const roleMetadataKey = "constellation-role"

// extractRole extracts role from cloud provider metadata.
func extractRole(metadata map[string]string) role.Role {
	switch metadata[roleMetadataKey] {
	case role.Coordinator.String():
		return role.Coordinator
	case role.Node.String():
		return role.Node
	default:
		return role.Unknown
	}
}

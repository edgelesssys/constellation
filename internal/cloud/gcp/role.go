package gcp

import (
	"github.com/edgelesssys/constellation/bootstrapper/role"
)

const roleMetadataKey = "constellation-role"

// extractRole extracts role from cloud provider metadata.
func extractRole(metadata map[string]string) role.Role {
	switch metadata[roleMetadataKey] {
	case role.ControlPlane.String():
		return role.ControlPlane
	case role.Worker.String():
		return role.Worker
	default:
		return role.Unknown
	}
}

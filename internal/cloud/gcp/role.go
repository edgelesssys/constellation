/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"github.com/edgelesssys/constellation/v2/internal/role"
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

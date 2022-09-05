/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"testing"

	"github.com/edgelesssys/constellation/internal/role"
	"github.com/stretchr/testify/assert"
)

func TestExtractRole(t *testing.T) {
	testCases := map[string]struct {
		metadata map[string]string
		wantRole role.Role
	}{
		"bootstrapper role": {
			metadata: map[string]string{
				roleMetadataKey: role.ControlPlane.String(),
			},
			wantRole: role.ControlPlane,
		},
		"node role": {
			metadata: map[string]string{
				roleMetadataKey: role.Worker.String(),
			},
			wantRole: role.Worker,
		},
		"unknown role": {
			metadata: map[string]string{
				roleMetadataKey: "some-unknown-role",
			},
			wantRole: role.Unknown,
		},
		"no role": {
			wantRole: role.Unknown,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			role := extractRole(tc.metadata)

			assert.Equal(tc.wantRole, role)
		})
	}
}

package cloudprovider

import (
	"testing"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/stretchr/testify/assert"
)

func TestExtractRole(t *testing.T) {
	testCases := map[string]struct {
		metadata     map[string]string
		expectedRole role.Role
	}{
		"coordinator role": {
			metadata: map[string]string{
				core.RoleMetadataKey: role.Coordinator.String(),
			},
			expectedRole: role.Coordinator,
		},
		"node role": {
			metadata: map[string]string{
				core.RoleMetadataKey: role.Node.String(),
			},
			expectedRole: role.Node,
		},
		"unknown role": {
			metadata: map[string]string{
				core.RoleMetadataKey: "some-unknown-role",
			},
			expectedRole: role.Unknown,
		},
		"no role": {
			expectedRole: role.Unknown,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			role := ExtractRole(tc.metadata)

			assert.Equal(tc.expectedRole, role)
		})
	}
}

package role

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	testCases := map[string]struct {
		role         Role
		jsonExpected string
		errExpected  bool
	}{
		"coordinator role": {
			role:         Coordinator,
			jsonExpected: `"Coordinator"`,
		},
		"node role": {
			role:         Node,
			jsonExpected: `"Node"`,
		},
		"admin role": {
			role:         Admin,
			jsonExpected: `"Admin"`,
		},
		"unknown role": {
			role:         Unknown,
			jsonExpected: `"Unknown"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			jsonRole, err := tc.role.MarshalJSON()
			if tc.errExpected {
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.jsonExpected, string(jsonRole))
		})
	}
}

func TestUnmarshal(t *testing.T) {
	testCases := map[string]struct {
		json         string
		expectedRole Role
		errExpected  bool
	}{
		"Coordinator can be unmarshaled": {
			json:         `"Coordinator"`,
			expectedRole: Coordinator,
		},
		"lowercase coordinator can be unmarshaled": {
			json:         `"coordinator"`,
			expectedRole: Coordinator,
		},
		"Node can be unmarshaled": {
			json:         `"Node"`,
			expectedRole: Node,
		},
		"lowercase node can be unmarshaled": {
			json:         `"node"`,
			expectedRole: Node,
		},
		"Admin can be unmarshaled": {
			json:         `"Admin"`,
			expectedRole: Admin,
		},
		"lowercase admin can be unmarshaled": {
			json:         `"admin"`,
			expectedRole: Admin,
		},
		"other strings unmarshal to the unknown role": {
			json:         `"anything"`,
			expectedRole: Unknown,
		},
		"invalid json fails": {
			json:        `"unterminated string literal`,
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var role Role
			err := role.UnmarshalJSON([]byte(tc.json))

			if tc.errExpected {
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.expectedRole, role)
		})
	}
}

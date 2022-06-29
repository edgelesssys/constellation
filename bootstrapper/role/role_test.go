package role

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestMarshal(t *testing.T) {
	testCases := map[string]struct {
		role     Role
		wantJson string
		wantErr  bool
	}{
		"controlePlane role": {
			role:     ControlPlane,
			wantJson: `"ControlPlane"`,
		},
		"node role": {
			role:     Worker,
			wantJson: `"Worker"`,
		},
		"admin role": {
			role:     Admin,
			wantJson: `"Admin"`,
		},
		"unknown role": {
			role:     Unknown,
			wantJson: `"Unknown"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			jsonRole, err := tc.role.MarshalJSON()
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.wantJson, string(jsonRole))
		})
	}
}

func TestUnmarshal(t *testing.T) {
	testCases := map[string]struct {
		json     string
		wantRole Role
		wantErr  bool
	}{
		"ControlPlane can be unmarshaled": {
			json:     `"ControlPlane"`,
			wantRole: ControlPlane,
		},
		"lowercase controlPlane can be unmarshaled": {
			json:     `"controlPlane"`,
			wantRole: ControlPlane,
		},
		"Worker can be unmarshaled": {
			json:     `"Worker"`,
			wantRole: Worker,
		},
		"lowercase worker can be unmarshaled": {
			json:     `"worker"`,
			wantRole: Worker,
		},
		"Admin can be unmarshaled": {
			json:     `"Admin"`,
			wantRole: Admin,
		},
		"lowercase admin can be unmarshaled": {
			json:     `"admin"`,
			wantRole: Admin,
		},
		"other strings unmarshal to the unknown role": {
			json:     `"anything"`,
			wantRole: Unknown,
		},
		"invalid json fails": {
			json:    `"unterminated string literal`,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var role Role
			err := role.UnmarshalJSON([]byte(tc.json))

			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.wantRole, role)
		})
	}
}

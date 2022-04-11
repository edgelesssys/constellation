package nodestate

import (
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromFile(t *testing.T) {
	testCases := map[string]struct {
		fileContents  string
		expectedState *NodeState
		errExpected   bool
	}{
		"nodestate exists": {
			fileContents: `{	"Role": "Coordinator",	"VPNPrivKey": "dGVzdA=="	}`,
			expectedState: &NodeState{
				Role:       role.Coordinator,
				VPNPrivKey: []byte("test"),
			},
		},
		"nodestate file does not exist": {
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			if tc.fileContents != "" {
				require.NoError(fs.MkdirAll(filepath.Dir(nodeStatePath), 0o755))
				require.NoError(afero.WriteFile(fs, nodeStatePath, []byte(tc.fileContents), 0o644))
			}
			fileHandler := file.NewHandler(fs)
			state, err := FromFile(fileHandler)
			if tc.errExpected {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedState, state)
		})
	}
}

func TestToFile(t *testing.T) {
	testCases := map[string]struct {
		precreateFile bool
		state         *NodeState
		expectedFile  string
		errExpected   bool
	}{
		"writing works": {
			state: &NodeState{
				Role:       role.Coordinator,
				VPNPrivKey: []byte("test"),
			},
			expectedFile: `{
	"Role": "Coordinator",
	"VPNPrivKey": "dGVzdA=="
}`,
		},
		"file exists already": {
			precreateFile: true,
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			require.NoError(fs.MkdirAll(filepath.Dir(nodeStatePath), 0o755))
			if tc.precreateFile {
				require.NoError(afero.WriteFile(fs, nodeStatePath, []byte("pre-existing"), 0o644))
			}
			fileHandler := file.NewHandler(fs)
			err := tc.state.ToFile(fileHandler)
			if tc.errExpected {
				assert.Error(err)
				return
			}
			require.NoError(err)

			fileContents, err := afero.ReadFile(fs, nodeStatePath)
			require.NoError(err)
			assert.Equal(tc.expectedFile, string(fileContents))
		})
	}
}

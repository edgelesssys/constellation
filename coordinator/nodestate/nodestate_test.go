package nodestate

import (
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromFile(t *testing.T) {
	testCases := map[string]struct {
		fileContents string
		wantState    *NodeState
		wantErr      bool
	}{
		"nodestate exists": {
			fileContents: `{	"Role": "Coordinator", "VPNIP": "192.0.2.1", "VPNPrivKey": "dGVzdA==", "OwnerID": "T3duZXJJRA==", "ClusterID": "Q2x1c3RlcklE"	}`,
			wantState: &NodeState{
				Role:       role.Coordinator,
				VPNIP:      "192.0.2.1",
				VPNPrivKey: []byte("test"),
				OwnerID:    []byte("OwnerID"),
				ClusterID:  []byte("ClusterID"),
			},
		},
		"nodestate file does not exist": {
			wantErr: true,
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
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantState, state)
		})
	}
}

func TestToFile(t *testing.T) {
	testCases := map[string]struct {
		precreateFile bool
		state         *NodeState
		wantFile      string
		wantErr       bool
	}{
		"writing works": {
			state: &NodeState{
				Role:       role.Coordinator,
				VPNIP:      "192.0.2.1",
				VPNPrivKey: []byte("test"),
				OwnerID:    []byte("OwnerID"),
				ClusterID:  []byte("ClusterID"),
			},
			wantFile: `{
	"Role": "Coordinator",
	"VPNIP": "192.0.2.1",
	"VPNPrivKey": "dGVzdA==",
	"OwnerID": "T3duZXJJRA==",
	"ClusterID": "Q2x1c3RlcklE"
}`,
		},
		"file exists already": {
			precreateFile: true,
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			if tc.precreateFile {
				require.NoError(fs.MkdirAll(filepath.Dir(nodeStatePath), 0o755))
				require.NoError(afero.WriteFile(fs, nodeStatePath, []byte("pre-existing"), 0o644))
			}
			fileHandler := file.NewHandler(fs)
			err := tc.state.ToFile(fileHandler)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			fileContents, err := afero.ReadFile(fs, nodeStatePath)
			require.NoError(err)
			assert.Equal(tc.wantFile, string(fileContents))
		})
	}
}

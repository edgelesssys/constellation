/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package nodestate

import (
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestFromFile(t *testing.T) {
	testCases := map[string]struct {
		fileContents string
		wantState    *NodeState
		wantErr      bool
	}{
		"nodestate exists": {
			fileContents: `{	"Role": "ControlPlane", "MeasurementSalt": "U2FsdA=="	}`,
			wantState: &NodeState{
				Role:            role.ControlPlane,
				MeasurementSalt: []byte("Salt"),
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
				Role:            role.ControlPlane,
				MeasurementSalt: []byte("Salt"),
			},
			wantFile: `{
	"Role": "ControlPlane",
	"MeasurementSalt": "U2FsdA=="
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

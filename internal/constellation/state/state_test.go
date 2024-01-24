/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package state

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// defaultState returns a valid default state for testing.
func defaultState() *State {
	return &State{
		Version: "v1",
		Infrastructure: Infrastructure{
			UID:               "123",
			Name:              "test-cluster",
			ClusterEndpoint:   "0.0.0.0",
			InClusterEndpoint: "0.0.0.0",
			InitSecret:        []byte{0x41},
			APIServerCertSANs: []string{
				"127.0.0.1",
				"www.example.com",
			},
			IPCidrNode: "0.0.0.0/24",
			Azure: &Azure{
				ResourceGroup:            "test-rg",
				SubscriptionID:           "test-sub",
				NetworkSecurityGroupName: "test-nsg",
				LoadBalancerName:         "test-lb",
				UserAssignedIdentity:     "test-uami",
				AttestationURL:           "test-maaUrl",
			},
			GCP: &GCP{
				ProjectID: "test-project",
				IPCidrPod: "0.0.0.0/24",
			},
		},
		ClusterValues: ClusterValues{
			ClusterID:       "test-cluster-id",
			OwnerID:         "test-owner-id",
			MeasurementSalt: []byte{0x41},
		},
	}
}

func defaultAzureState() *State {
	s := defaultState()
	s.Infrastructure.GCP = nil
	return s
}

func defaultGCPState() *State {
	s := defaultState()
	s.Infrastructure.Azure = nil
	return s
}

func TestWriteToFile(t *testing.T) {
	testCases := map[string]struct {
		state   *State
		fh      file.Handler
		wantErr bool
	}{
		"success": {
			state: defaultState(),
			fh:    file.NewHandler(afero.NewMemMapFs()),
		},
		"overwrite": {
			state: defaultState(),
			fh: func() file.Handler {
				fs := file.NewHandler(afero.NewMemMapFs())
				require.NoError(t, fs.Write(constants.StateFilename, []byte{0x41}))
				return fs
			}(),
		},
		"empty state": {
			state: &State{},
			fh:    file.NewHandler(afero.NewMemMapFs()),
		},
		"rofs": {
			state:   defaultState(),
			fh:      file.NewHandler(afero.NewReadOnlyFs(afero.NewMemMapFs())),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			err := tc.state.WriteToFile(tc.fh, constants.StateFilename)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.YAMLEq(mustMarshalYaml(require, tc.state), mustReadFromFile(require, tc.fh))
			}
		})
	}
}

func TestReadFromFile(t *testing.T) {
	testCases := map[string]struct {
		fs        file.Handler
		wantState *State
		wantErr   bool
	}{
		"success": {
			fs: func() file.Handler {
				fh := file.NewHandler(afero.NewMemMapFs())
				require.NoError(t, fh.WriteYAML(constants.StateFilename, defaultState()))
				return fh
			}(),
			wantState: defaultState(),
		},
		"no state file present": {
			fs:      file.NewHandler(afero.NewMemMapFs()),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			state, err := ReadFromFile(tc.fs, constants.StateFilename)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.YAMLEq(mustMarshalYaml(require, tc.wantState), mustMarshalYaml(require, state))
			}
		})
	}
}

func mustMarshalYaml(require *require.Assertions, v any) string {
	b, err := encoder.NewEncoder(v).Encode()
	require.NoError(err)
	return string(b)
}

func mustReadFromFile(require *require.Assertions, fh file.Handler) string {
	b, err := fh.Read(constants.StateFilename)
	require.NoError(err)
	return string(b)
}

func TestMerge(t *testing.T) {
	testCases := map[string]struct {
		state    *State
		other    *State
		expected *State
		wantErr  bool
	}{
		"success": {
			state: &State{
				Infrastructure: Infrastructure{
					ClusterEndpoint: "test-cluster-endpoint",
					UID:             "123",
				},
			},
			other: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
			expected: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					ClusterEndpoint: "test-cluster-endpoint",
					UID:             "456",
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
		},
		"empty state": {
			state: &State{},
			other: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
			expected: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
		},
		"empty other": {
			state: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
			other: &State{},
			expected: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
		},
		"empty state and other": {
			state:    &State{},
			other:    &State{},
			expected: &State{},
		},
		"identical": {
			state: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
			other: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
			expected: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
		},
		"nested pointer": {
			state: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "123",
					Azure: &Azure{
						AttestationURL: "test-maaUrl",
					},
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
			other: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
					Azure: &Azure{
						AttestationURL: "test-maaUrl-2",
					},
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
			expected: &State{
				Version: "v1",
				Infrastructure: Infrastructure{
					UID: "456",
					Azure: &Azure{
						AttestationURL: "test-maaUrl-2",
					},
				},
				ClusterValues: ClusterValues{
					ClusterID: "test-cluster-id",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err := tc.state.Merge(tc.other)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.expected, tc.state)
			}
		})
	}
}

func TestCreateOrRead(t *testing.T) {
	testCases := map[string]struct {
		fs        file.Handler
		wantState *State
		wantErr   bool
	}{
		"file exists": {
			fs: func() file.Handler {
				fh := file.NewHandler(afero.NewMemMapFs())
				require.NoError(t, fh.WriteYAML(constants.StateFilename, defaultState()))
				return fh
			}(),
			wantState: defaultState(),
		},
		"file does not exist": {
			fs:        file.NewHandler(afero.NewMemMapFs()),
			wantState: New(),
		},
		"unable to write file": {
			fs:      file.NewHandler(afero.NewReadOnlyFs(afero.NewMemMapFs())),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			require := require.New(t)
			state, err := CreateOrRead(tc.fs, constants.StateFilename)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.YAMLEq(mustMarshalYaml(require, tc.wantState), mustMarshalYaml(require, state))
			assert.YAMLEq(mustMarshalYaml(require, tc.wantState), mustReadFromFile(require, tc.fs))
		})
	}
}

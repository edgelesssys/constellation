package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/crypto/testvector"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverCmdArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"no args":            {[]string{}, false},
		"too many arguments": {[]string{"abc"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := NewRecoverCmd()
			err := cmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestRecover(t *testing.T) {
	validState := state.ConstellationState{CloudProvider: "GCP"}
	invalidCSPState := state.ConstellationState{CloudProvider: "invalid"}

	testCases := map[string]struct {
		existingState    state.ConstellationState
		client           *stubRecoveryClient
		masterSecret     testvector.HKDF
		endpointFlag     string
		diskUUIDFlag     string
		masterSecretFlag string
		configFlag       string
		stateless        bool
		wantErr          bool
	}{
		"works": {
			existingState: validState,
			client:        &stubRecoveryClient{},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  testvector.HKDFZero.Info,
			masterSecret:  testvector.HKDFZero,
		},
		"uppercase disk uuid works": {
			existingState: validState,
			client:        &stubRecoveryClient{},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  strings.ToUpper(testvector.HKDF0xFF.Info),
			masterSecret:  testvector.HKDF0xFF,
		},
		"lowercase disk uuid results in same key": {
			existingState: validState,
			client:        &stubRecoveryClient{},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  strings.ToLower(testvector.HKDF0xFF.Info),
			masterSecret:  testvector.HKDF0xFF,
		},
		"missing flags": {
			wantErr: true,
		},
		"missing config": {
			endpointFlag: "192.0.2.1",
			diskUUIDFlag: testvector.HKDFZero.Info,
			masterSecret: testvector.HKDFZero,
			configFlag:   "nonexistent-config",
			wantErr:      true,
		},
		"missing state": {
			existingState: validState,
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  testvector.HKDFZero.Info,
			masterSecret:  testvector.HKDFZero,
			stateless:     true,
			wantErr:       true,
		},
		"invalid cloud provider": {
			existingState: invalidCSPState,
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  testvector.HKDFZero.Info,
			masterSecret:  testvector.HKDFZero,
			wantErr:       true,
		},
		"connect fails": {
			existingState: validState,
			client:        &stubRecoveryClient{connectErr: errors.New("connect failed")},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  testvector.HKDFZero.Info,
			masterSecret:  testvector.HKDFZero,
			wantErr:       true,
		},
		"pushing state key fails": {
			existingState: validState,
			client:        &stubRecoveryClient{pushStateDiskKeyErr: errors.New("pushing key failed")},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  testvector.HKDFZero.Info,
			masterSecret:  testvector.HKDFZero,
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewRecoverCmd()
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(&bytes.Buffer{})
			if tc.endpointFlag != "" {
				require.NoError(cmd.Flags().Set("endpoint", tc.endpointFlag))
			}
			if tc.diskUUIDFlag != "" {
				require.NoError(cmd.Flags().Set("disk-uuid", tc.diskUUIDFlag))
			}
			if tc.masterSecretFlag != "" {
				require.NoError(cmd.Flags().Set("master-secret", tc.masterSecretFlag))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}

			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)

			config := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.FromString(tc.existingState.CloudProvider))
			require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, config))

			require.NoError(fileHandler.WriteJSON("constellation-mastersecret.json", masterSecret{Key: tc.masterSecret.Secret, Salt: tc.masterSecret.Salt}, file.OptNone))
			if !tc.stateless {
				require.NoError(fileHandler.WriteJSON(constants.StateFilename, tc.existingState, file.OptNone))
			}

			err := recover(cmd, fileHandler, tc.client)

			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Contains(out.String(), "Pushed recovery key.")
			assert.Equal(tc.masterSecret.Output, tc.client.pushStateDiskKeyKey)
		})
	}
}

func TestParseRecoverFlags(t *testing.T) {
	testCases := map[string]struct {
		args      []string
		wantFlags recoverFlags
		wantErr   bool
	}{
		"no flags": {
			wantErr: true,
		},
		"invalid ip": {
			args:    []string{"-e", "192.0.2.1:2:2", "--disk-uuid", "12345678-1234-1234-1234-123456789012"},
			wantErr: true,
		},
		"invalid disk uuid": {
			args:    []string{"-e", "192.0.2.1:2", "--disk-uuid", "invalid"},
			wantErr: true,
		},
		"minimal args set": {
			args: []string{"-e", "192.0.2.1:2", "--disk-uuid", "12345678-1234-1234-1234-123456789012"},
			wantFlags: recoverFlags{
				endpoint:   "192.0.2.1:2",
				diskUUID:   "12345678-1234-1234-1234-123456789012",
				secretPath: "constellation-mastersecret.json",
			},
		},
		"all args set": {
			args: []string{"-e", "192.0.2.1:2", "--disk-uuid", "12345678-1234-1234-1234-123456789012", "--config", "config-path", "--master-secret", "/path/super-secret.json"},
			wantFlags: recoverFlags{
				endpoint:   "192.0.2.1:2",
				diskUUID:   "12345678-1234-1234-1234-123456789012",
				secretPath: "/path/super-secret.json",
				configPath: "config-path",
			},
		},
		"uppercase disk-uuid is converted to lowercase": {
			args: []string{"-e", "192.0.2.1:2", "--disk-uuid", "ABCDEFAB-CDEF-ABCD-ABCD-ABCDEFABCDEF"},
			wantFlags: recoverFlags{
				endpoint:   "192.0.2.1:2",
				diskUUID:   "abcdefab-cdef-abcd-abcd-abcdefabcdef",
				secretPath: "constellation-mastersecret.json",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewRecoverCmd()
			cmd.Flags().String("config", "", "") // register persistent flag manually
			require.NoError(cmd.ParseFlags(tc.args))
			flags, err := parseRecoverFlags(cmd)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantFlags, flags)
		})
	}
}

func TestDeriveStateDiskKey(t *testing.T) {
	testCases := map[string]struct {
		masterSecret testvector.HKDF
	}{
		"all zero": {
			masterSecret: testvector.HKDFZero,
		},
		"all 0xff": {
			masterSecret: testvector.HKDF0xFF,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			stateDiskKey, err := deriveStateDiskKey(tc.masterSecret.Secret, tc.masterSecret.Salt, tc.masterSecret.Info)

			assert.NoError(err)
			assert.Equal(tc.masterSecret.Output, stateDiskKey)
		})
	}
}

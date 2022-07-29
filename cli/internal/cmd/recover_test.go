package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/constants"
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

	writeMasterSecret := func(require *require.Assertions) afero.Fs {
		fs := afero.NewMemMapFs()
		handler := file.NewHandler(fs)
		require.NoError(handler.WriteJSON("constellation-mastersecret.json", masterSecret{Key: []byte("master-secret"), Salt: []byte("salt")}, file.OptNone))
		return fs
	}

	testCases := map[string]struct {
		setupFs          func(*require.Assertions) afero.Fs
		existingState    state.ConstellationState
		client           *stubRecoveryClient
		endpointFlag     string
		diskUUIDFlag     string
		masterSecretFlag string
		configFlag       string
		stateless        bool
		wantErr          bool
		wantKey          []byte
	}{
		"works": {
			setupFs:       writeMasterSecret,
			existingState: validState,
			client:        &stubRecoveryClient{},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "00000000-0000-0000-0000-000000000000",
			wantKey: []byte{
				0xcb, 0x31, 0x9d, 0x3d, 0xe0, 0xb7, 0x8a, 0xe9, 0x20, 0xc0, 0x62, 0x00, 0x8f, 0xe4, 0x58, 0xa1,
				0x87, 0x85, 0x5d, 0xa0, 0xca, 0xe2, 0x04, 0x5c, 0x80, 0xe8, 0xe6, 0xd5, 0x8a, 0x6c, 0xcb, 0x49,
			},
		},
		"uppercase disk uuid works": {
			setupFs:       writeMasterSecret,
			existingState: validState,
			client:        &stubRecoveryClient{},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "ABCDEFAB-CDEF-ABCD-ABCD-ABCDEFABCDEF",
			wantKey: []byte{
				0x9b, 0xb8, 0x4a, 0x62, 0x01, 0xb3, 0x32, 0xf6, 0xf2, 0x79, 0x43, 0x09, 0x86, 0xe7, 0x25, 0x0e,
				0xd2, 0x77, 0xcb, 0x14, 0xe8, 0x8f, 0x38, 0xab, 0xe7, 0xd6, 0x25, 0x14, 0xa5, 0xa1, 0xff, 0xda,
			},
		},
		"lowercase disk uuid results in same key": {
			setupFs:       writeMasterSecret,
			existingState: validState,
			client:        &stubRecoveryClient{},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "abcdefab-cdef-abcd-abcd-abcdefabcdef",
			wantKey: []byte{
				0x9b, 0xb8, 0x4a, 0x62, 0x01, 0xb3, 0x32, 0xf6, 0xf2, 0x79, 0x43, 0x09, 0x86, 0xe7, 0x25, 0x0e,
				0xd2, 0x77, 0xcb, 0x14, 0xe8, 0x8f, 0x38, 0xab, 0xe7, 0xd6, 0x25, 0x14, 0xa5, 0xa1, 0xff, 0xda,
			},
		},
		"missing flags": {
			setupFs: func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			wantErr: true,
		},
		"missing config": {
			setupFs:      writeMasterSecret,
			endpointFlag: "192.0.2.1",
			diskUUIDFlag: "00000000-0000-0000-0000-000000000000",
			configFlag:   "nonexistent-config",
			wantErr:      true,
		},
		"missing state": {
			setupFs:       writeMasterSecret,
			existingState: validState,
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "00000000-0000-0000-0000-000000000000",
			stateless:     true,
			wantErr:       true,
		},
		"invalid cloud provider": {
			setupFs:       writeMasterSecret,
			existingState: invalidCSPState,
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "00000000-0000-0000-0000-000000000000",
			wantErr:       true,
		},
		"connect fails": {
			setupFs:       writeMasterSecret,
			existingState: validState,
			client:        &stubRecoveryClient{connectErr: errors.New("connect failed")},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "00000000-0000-0000-0000-000000000000",
			wantErr:       true,
		},
		"pushing state key fails": {
			setupFs:       writeMasterSecret,
			existingState: validState,
			client:        &stubRecoveryClient{pushStateDiskKeyErr: errors.New("pushing key failed")},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "00000000-0000-0000-0000-000000000000",
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewRecoverCmd()
			cmd.Flags().String("config", "", "") // register persistent flag manually
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
			fileHandler := file.NewHandler(tc.setupFs(require))
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
			assert.Equal(tc.wantKey, tc.client.pushStateDiskKeyKey)
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
		masterKey        []byte
		salt             []byte
		diskUUID         string
		wantStateDiskKey []byte
	}{
		"all zero": {
			masterKey: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
			salt: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
			diskUUID: "00000000-0000-0000-0000-000000000000",
			wantStateDiskKey: []byte{
				0x2b, 0x5c, 0x82, 0x9f, 0x69, 0x1b, 0xfd, 0x42, 0x32, 0xf8, 0xaa, 0xc4, 0x64, 0xbc, 0xcc, 0x07,
				0xe6, 0x05, 0xc7, 0xc8, 0xe8, 0x2c, 0xbc, 0xa0, 0x98, 0x37, 0xe8, 0x6d, 0x0b, 0x6d, 0x06, 0x65,
			},
		},
		"all 0xff": {
			masterKey: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			},
			salt: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			},
			diskUUID: "ffffffff-ffff-ffff-ffff-ffffffffffff",
			wantStateDiskKey: []byte{
				0x14, 0x84, 0xaa, 0x39, 0xd3, 0x41, 0x5e, 0x90, 0x6e, 0x07, 0x94, 0x0f, 0xf2, 0x15, 0xd8, 0xb1,
				0xee, 0xe7, 0x05, 0xd3, 0x02, 0x7d, 0xba, 0x93, 0x30, 0x6a, 0xf4, 0xab, 0xff, 0x4f, 0x70, 0xbe,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			stateDiskKey, err := deriveStateDiskKey(tc.masterKey, tc.salt, tc.diskUUID)

			assert.NoError(err)
			assert.Equal(tc.wantStateDiskKey, stateDiskKey)
		})
	}
}

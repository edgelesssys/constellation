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
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="), 0o777))
				return fs
			},
			existingState: validState,
			client:        &stubRecoveryClient{},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "00000000-0000-0000-0000-000000000000",
			wantKey:       []byte{0x4d, 0x34, 0x19, 0x1a, 0xf9, 0x23, 0xb9, 0x61, 0x55, 0x9b, 0xb2, 0x6, 0x15, 0x1b, 0x5f, 0xe, 0x21, 0xc2, 0xe5, 0x18, 0x1c, 0xfa, 0x32, 0x79, 0xa4, 0x6b, 0x84, 0x86, 0x7e, 0xd7, 0xf6, 0x76},
		},
		"uppercase disk uuid works": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="), 0o777))
				return fs
			},
			existingState: validState,
			client:        &stubRecoveryClient{},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "ABCDEFAB-CDEF-ABCD-ABCD-ABCDEFABCDEF",
			wantKey:       []byte{0x7e, 0xc0, 0xa8, 0x84, 0xc4, 0x7, 0xda, 0x1, 0xed, 0xa9, 0xc8, 0x87, 0x77, 0xad, 0x86, 0x7c, 0x7d, 0x40, 0xa7, 0x28, 0x3d, 0xbd, 0x92, 0xea, 0xa1, 0x84, 0x67, 0x78, 0x58, 0x76, 0x13, 0x70},
		},
		"lowercase disk uuid results in same key": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="), 0o777))
				return fs
			},
			existingState: validState,
			client:        &stubRecoveryClient{},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "abcdefab-cdef-abcd-abcd-abcdefabcdef",
			wantKey:       []byte{0x7e, 0xc0, 0xa8, 0x84, 0xc4, 0x7, 0xda, 0x1, 0xed, 0xa9, 0xc8, 0x87, 0x77, 0xad, 0x86, 0x7c, 0x7d, 0x40, 0xa7, 0x28, 0x3d, 0xbd, 0x92, 0xea, 0xa1, 0x84, 0x67, 0x78, 0x58, 0x76, 0x13, 0x70},
		},
		"missing flags": {
			setupFs: func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			wantErr: true,
		},
		"missing config": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="), 0o777))
				return fs
			},
			endpointFlag: "192.0.2.1",
			diskUUIDFlag: "00000000-0000-0000-0000-000000000000",
			configFlag:   "nonexistent-config",
			wantErr:      true,
		},
		"missing state": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="), 0o777))
				return fs
			},
			existingState: validState,
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "00000000-0000-0000-0000-000000000000",
			stateless:     true,
			wantErr:       true,
		},
		"invalid cloud provider": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="), 0o777))
				return fs
			},
			existingState: invalidCSPState,
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "00000000-0000-0000-0000-000000000000",
			wantErr:       true,
		},
		"connect fails": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="), 0o777))
				return fs
			},
			existingState: validState,
			client:        &stubRecoveryClient{connectErr: errors.New("connect failed")},
			endpointFlag:  "192.0.2.1",
			diskUUIDFlag:  "00000000-0000-0000-0000-000000000000",
			wantErr:       true,
		},
		"pushing state key fails": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="), 0o777))
				return fs
			},
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
			cmd.Flags().String("config", "", "") // register persisten flag manually
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
		"invalid master secret path": {
			args:    []string{"-e", "192.0.2.1:2", "--disk-uuid", "12345678-1234-1234-1234-123456789012", "--master-secret", "invalid"},
			wantErr: true,
		},
		"minimal args set": {
			args: []string{"-e", "192.0.2.1:2", "--disk-uuid", "12345678-1234-1234-1234-123456789012"},
			wantFlags: recoverFlags{
				endpoint:     "192.0.2.1:2",
				diskUUID:     "12345678-1234-1234-1234-123456789012",
				masterSecret: []byte("constellation-master-secret-leng"),
			},
		},
		"all args set": {
			args: []string{
				"-e", "192.0.2.1:2", "--disk-uuid", "12345678-1234-1234-1234-123456789012",
				"--master-secret", "constellation-mastersecret.base64", "--config", "config-path",
			},
			wantFlags: recoverFlags{
				endpoint:     "192.0.2.1:2",
				diskUUID:     "12345678-1234-1234-1234-123456789012",
				masterSecret: []byte("constellation-master-secret-leng"),
				configPath:   "config-path",
			},
		},
		"uppercase disk-uuid is converted to lowercase": {
			args: []string{"-e", "192.0.2.1:2", "--disk-uuid", "ABCDEFAB-CDEF-ABCD-ABCD-ABCDEFABCDEF"},
			wantFlags: recoverFlags{
				endpoint:     "192.0.2.1:2",
				diskUUID:     "abcdefab-cdef-abcd-abcd-abcdefabcdef",
				masterSecret: []byte("constellation-master-secret-leng"),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="), 0o777))
			cmd := NewRecoverCmd()
			cmd.Flags().String("config", "", "") // register persistent flag manually
			require.NoError(cmd.ParseFlags(tc.args))
			flags, err := parseRecoverFlags(cmd, file.NewHandler(fs))

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantFlags, flags)
		})
	}
}

func TestReadMasterSecret(t *testing.T) {
	testCases := map[string]struct {
		fileContents     []byte
		filename         string
		wantMasterSecret []byte
		wantErr          bool
	}{
		"invalid base64": {
			fileContents: []byte("invalid"),
			filename:     "constellation-mastersecret.base64",
			wantErr:      true,
		},
		"invalid filename": {
			fileContents: []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="),
			filename:     "invalid",
			wantErr:      true,
		},
		"correct master secret": {
			fileContents:     []byte("Y29uc3RlbGxhdGlvbi1tYXN0ZXItc2VjcmV0LWxlbmc="),
			filename:         "constellation-mastersecret.base64",
			wantMasterSecret: []byte("constellation-master-secret-leng"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			require.NoError(afero.WriteFile(fs, "constellation-mastersecret.base64", tc.fileContents, 0o777))
			masterSecret, err := readMasterSecret(file.NewHandler(fs), tc.filename)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantMasterSecret, masterSecret)
		})
	}
}

func TestDeriveStateDiskKey(t *testing.T) {
	testCases := map[string]struct {
		masterKey        []byte
		diskUUID         string
		wantStateDiskKey []byte
	}{
		"all zero": {
			masterKey: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
			diskUUID: "00000000-0000-0000-0000-000000000000",
			wantStateDiskKey: []byte{
				0xc6, 0xe0, 0xae, 0xfc, 0xbe, 0x7b, 0x7e, 0x87, 0x7a, 0xdd, 0xb2, 0x87, 0xe0, 0xcd, 0x4c, 0xe4,
				0xde, 0xee, 0xb3, 0x57, 0xaa, 0x6c, 0xc9, 0x44, 0x90, 0xc4, 0x07, 0x72, 0x01, 0x7d, 0xd6, 0xb1,
			},
		},
		"all 0xff": {
			masterKey: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			},
			diskUUID: "ffffffff-ffff-ffff-ffff-ffffffffffff",
			wantStateDiskKey: []byte{
				0x00, 0x74, 0x4c, 0xb0, 0x92, 0x9d, 0x20, 0x08, 0xfa, 0x72, 0xac, 0xd2, 0xb6, 0xe4, 0xc6, 0x6f,
				0xa3, 0x53, 0x16, 0xb1, 0x9e, 0x77, 0x42, 0xe8, 0xd3, 0x66, 0xe8, 0x22, 0x33, 0xfc, 0x63, 0x4d,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			stateDiskKey, err := deriveStateDiskKey(tc.masterKey, tc.diskUUID)

			assert.NoError(err)
			assert.Equal(tc.wantStateDiskKey, stateDiskKey)
		})
	}
}

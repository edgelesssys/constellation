package cmd

import (
	"bytes"
	"context"
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

			cmd := newRecoverCmd()
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
			wantKey:       []byte{0x2e, 0x4d, 0x40, 0x3a, 0x90, 0x96, 0x6e, 0xd, 0x42, 0x3, 0x98, 0xd, 0xce, 0xc5, 0x73, 0x26, 0xf4, 0x87, 0xcf, 0x85, 0x73, 0xe1, 0xb7, 0xd6, 0xb2, 0x82, 0x4c, 0xd9, 0xbc, 0xa5, 0x7c, 0x32},
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
			wantKey:       []byte{0xa9, 0x4, 0x3a, 0x74, 0x53, 0xeb, 0x23, 0xb2, 0xbc, 0x88, 0xce, 0xa7, 0x4e, 0xa9, 0xda, 0x9f, 0x11, 0x85, 0xc4, 0x2f, 0x1f, 0x25, 0x10, 0xc9, 0xec, 0xfe, 0xa, 0x6c, 0xa2, 0x6f, 0x53, 0x34},
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
			wantKey:       []byte{0xa9, 0x4, 0x3a, 0x74, 0x53, 0xeb, 0x23, 0xb2, 0xbc, 0x88, 0xce, 0xa7, 0x4e, 0xa9, 0xda, 0x9f, 0x11, 0x85, 0xc4, 0x2f, 0x1f, 0x25, 0x10, 0xc9, 0xec, 0xfe, 0xa, 0x6c, 0xa2, 0x6f, 0x53, 0x34},
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

			cmd := newRecoverCmd()
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

			ctx := context.Background()
			err := recover(ctx, cmd, fileHandler, tc.client)

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
			cmd := newRecoverCmd()
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
				0xa8, 0xb0, 0x86, 0x83, 0x6f, 0x0b, 0x26, 0x04, 0x86, 0x22, 0x27, 0xcc, 0xa1, 0x1c, 0xaf, 0x6c,
				0x30, 0x4d, 0x90, 0x89, 0x82, 0x68, 0x53, 0x7f, 0x4f, 0x46, 0x7a, 0x65, 0xa2, 0x5d, 0x5e, 0x43,
			},
		},
		"all 0xff": {
			masterKey: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			},
			diskUUID: "ffffffff-ffff-ffff-ffff-ffffffffffff",
			wantStateDiskKey: []byte{
				0x24, 0x18, 0x84, 0x7f, 0xca, 0x86, 0x55, 0xb5, 0x45, 0xa6, 0xb3, 0xc4, 0x45, 0xbb, 0x08, 0x10,
				0x16, 0xb3, 0xde, 0x30, 0x30, 0x74, 0x0b, 0xd4, 0x1e, 0x22, 0x55, 0x45, 0x51, 0x91, 0xfb, 0xa9,
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

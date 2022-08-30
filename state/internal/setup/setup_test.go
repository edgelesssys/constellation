package setup

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/nodestate"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPrepareExistingDisk(t *testing.T) {
	someErr := errors.New("error")

	testCases := map[string]struct {
		keyWaiter       *stubKeyWaiter
		mapper          *stubMapper
		mounter         *stubMounter
		configGenerator *stubConfigurationGenerator
		openTPM         vtpm.TPMOpenFunc
		missingState    bool
		wantErr         bool
	}{
		"success": {
			keyWaiter:       &stubKeyWaiter{},
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{},
			openTPM:         vtpm.OpenNOPTPM,
		},
		"WaitForDecryptionKey fails": {
			keyWaiter:       &stubKeyWaiter{waitErr: someErr},
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{},
			openTPM:         vtpm.OpenNOPTPM,
			wantErr:         true,
		},
		"MapDisk fails causes a repeat": {
			keyWaiter: &stubKeyWaiter{},
			mapper: &stubMapper{
				uuid:                 "test",
				mapDiskErr:           someErr,
				mapDiskRepeatedCalls: 2,
			},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{},
			openTPM:         vtpm.OpenNOPTPM,
			wantErr:         false,
		},
		"MkdirAll fails": {
			keyWaiter:       &stubKeyWaiter{},
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{mkdirAllErr: someErr},
			configGenerator: &stubConfigurationGenerator{},
			openTPM:         vtpm.OpenNOPTPM,
			wantErr:         true,
		},
		"Mount fails": {
			keyWaiter:       &stubKeyWaiter{},
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{mountErr: someErr},
			configGenerator: &stubConfigurationGenerator{},
			openTPM:         vtpm.OpenNOPTPM,
			wantErr:         true,
		},
		"Unmount fails": {
			keyWaiter:       &stubKeyWaiter{},
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{unmountErr: someErr},
			configGenerator: &stubConfigurationGenerator{},
			openTPM:         vtpm.OpenNOPTPM,
			wantErr:         true,
		},
		"MarkNodeAsBootstrapped fails": {
			keyWaiter:       &stubKeyWaiter{},
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{unmountErr: someErr},
			configGenerator: &stubConfigurationGenerator{},
			openTPM:         failOpener,
			wantErr:         true,
		},
		"Generating config fails": {
			keyWaiter:       &stubKeyWaiter{},
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{generateErr: someErr},
			openTPM:         failOpener,
			wantErr:         true,
		},
		"no state file": {
			keyWaiter:       &stubKeyWaiter{},
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{},
			openTPM:         vtpm.OpenNOPTPM,
			missingState:    true,
			wantErr:         true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			salt := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
			if !tc.missingState {
				handler := file.NewHandler(fs)
				require.NoError(t, handler.WriteJSON(stateInfoPath, nodestate.NodeState{MeasurementSalt: salt}, file.OptMkdirAll))
			}

			setupManager := &SetupManager{
				log:       logger.NewTest(t),
				csp:       "test",
				diskPath:  "disk-path",
				fs:        fs,
				keyWaiter: tc.keyWaiter,
				mapper:    tc.mapper,
				mounter:   tc.mounter,
				config:    tc.configGenerator,
				openTPM:   tc.openTPM,
			}

			err := setupManager.PrepareExistingDisk()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.mapper.uuid, tc.keyWaiter.receivedUUID)
				assert.True(tc.mapper.mapDiskCalled)
				assert.True(tc.mounter.mountCalled)
				assert.True(tc.mounter.unmountCalled)
				assert.False(tc.mapper.formatDiskCalled)
			}
		})
	}
}

func failOpener() (io.ReadWriteCloser, error) {
	return nil, errors.New("error")
}

func TestPrepareNewDisk(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		fs              afero.Afero
		mapper          *stubMapper
		configGenerator *stubConfigurationGenerator
		wantErr         bool
	}{
		"success": {
			fs:              afero.Afero{Fs: afero.NewMemMapFs()},
			mapper:          &stubMapper{uuid: "test"},
			configGenerator: &stubConfigurationGenerator{},
		},
		"creating directory fails": {
			fs:              afero.Afero{Fs: afero.NewReadOnlyFs(afero.NewMemMapFs())},
			mapper:          &stubMapper{},
			configGenerator: &stubConfigurationGenerator{},
			wantErr:         true,
		},
		"FormatDisk fails": {
			fs: afero.Afero{Fs: afero.NewMemMapFs()},
			mapper: &stubMapper{
				uuid:          "test",
				formatDiskErr: someErr,
			},
			configGenerator: &stubConfigurationGenerator{},
			wantErr:         true,
		},
		"MapDisk fails": {
			fs: afero.Afero{Fs: afero.NewMemMapFs()},
			mapper: &stubMapper{
				uuid:                 "test",
				mapDiskErr:           someErr,
				mapDiskRepeatedCalls: 1,
			},
			configGenerator: &stubConfigurationGenerator{},
			wantErr:         true,
		},
		"Generating config fails": {
			fs:              afero.Afero{Fs: afero.NewMemMapFs()},
			mapper:          &stubMapper{uuid: "test"},
			configGenerator: &stubConfigurationGenerator{generateErr: someErr},
			wantErr:         true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			setupManager := &SetupManager{
				log:      logger.NewTest(t),
				csp:      "test",
				diskPath: "disk-path",
				fs:       tc.fs,
				mapper:   tc.mapper,
				config:   tc.configGenerator,
			}

			err := setupManager.PrepareNewDisk()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.True(tc.mapper.formatDiskCalled)
				assert.True(tc.mapper.mapDiskCalled)

				data, err := tc.fs.ReadFile(filepath.Join(keyPath, keyFile))
				require.NoError(t, err)
				assert.Len(data, crypto.RNGLengthDefault)
			}
		})
	}
}

func TestReadMeasurementSalt(t *testing.T) {
	salt := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	testCases := map[string]struct {
		salt      []byte
		writeFile bool
		wantErr   bool
	}{
		"success": {
			salt:      salt,
			writeFile: true,
		},
		"no state file": {
			wantErr: true,
		},
		"missing salt": {
			writeFile: true,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			if tc.writeFile {
				handler := file.NewHandler(fs)
				state := nodestate.NodeState{MeasurementSalt: tc.salt}
				require.NoError(handler.WriteJSON("test-state.json", state, file.OptMkdirAll))
			}

			setupManager := New(logger.NewTest(t), "test", "disk-path", fs, nil, nil, nil, nil)

			measurementSalt, err := setupManager.readMeasurementSalt("test-state.json")
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.salt, measurementSalt)
			}
		})
	}
}

type stubMapper struct {
	formatDiskCalled     bool
	formatDiskErr        error
	mapDiskRepeatedCalls int
	mapDiskCalled        bool
	mapDiskErr           error
	unmapDiskCalled      bool
	unmapDiskErr         error
	uuid                 string
}

func (s *stubMapper) DiskUUID() string {
	return s.uuid
}

func (s *stubMapper) FormatDisk(string) error {
	s.formatDiskCalled = true
	return s.formatDiskErr
}

func (s *stubMapper) MapDisk(string, string) error {
	if s.mapDiskRepeatedCalls == 0 {
		s.mapDiskErr = nil
	}
	s.mapDiskRepeatedCalls--
	s.mapDiskCalled = true
	return s.mapDiskErr
}

func (s *stubMapper) UnmapDisk(string) error {
	s.unmapDiskCalled = true
	return s.unmapDiskErr
}

type stubMounter struct {
	mountCalled   bool
	mountErr      error
	unmountCalled bool
	unmountErr    error
	mkdirAllErr   error
}

func (s *stubMounter) Mount(source string, target string, fstype string, flags uintptr, data string) error {
	s.mountCalled = true
	return s.mountErr
}

func (s *stubMounter) Unmount(target string, flags int) error {
	s.unmountCalled = true
	return s.unmountErr
}

func (s *stubMounter) MkdirAll(path string, perm fs.FileMode) error {
	return s.mkdirAllErr
}

type stubKeyWaiter struct {
	receivedUUID      string
	decryptionKey     []byte
	measurementSecret []byte
	waitErr           error
	waitCalled        bool
}

func (s *stubKeyWaiter) WaitForDecryptionKey(uuid, addr string) ([]byte, []byte, error) {
	if s.waitCalled {
		return nil, nil, errors.New("wait called before key was reset")
	}
	s.waitCalled = true
	s.receivedUUID = uuid
	return s.decryptionKey, s.measurementSecret, s.waitErr
}

func (s *stubKeyWaiter) ResetKey() {
	s.waitCalled = false
}

type stubConfigurationGenerator struct {
	generateErr error
}

func (s *stubConfigurationGenerator) Generate(volumeName, encryptedDevice, keyFile, options string) error {
	return s.generateErr
}

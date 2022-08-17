package setup

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/bootstrapper/nodestate"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
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
		fs           afero.Afero
		keyWaiter    *stubKeyWaiter
		mapper       *stubMapper
		mounter      *stubMounter
		openTPM      vtpm.TPMOpenFunc
		missingState bool
		wantErr      bool
	}{
		"success": {
			fs:        afero.Afero{Fs: afero.NewMemMapFs()},
			keyWaiter: &stubKeyWaiter{},
			mapper:    &stubMapper{uuid: "test"},
			mounter:   &stubMounter{},
			openTPM:   vtpm.OpenNOPTPM,
		},
		"WaitForDecryptionKey fails": {
			fs:        afero.Afero{Fs: afero.NewMemMapFs()},
			keyWaiter: &stubKeyWaiter{waitErr: someErr},
			mapper:    &stubMapper{uuid: "test"},
			mounter:   &stubMounter{},
			openTPM:   vtpm.OpenNOPTPM,
			wantErr:   true,
		},
		"MapDisk fails causes a repeat": {
			fs:        afero.Afero{Fs: afero.NewMemMapFs()},
			keyWaiter: &stubKeyWaiter{},
			mapper: &stubMapper{
				uuid:                 "test",
				mapDiskErr:           someErr,
				mapDiskRepeatedCalls: 2,
			},
			mounter: &stubMounter{},
			openTPM: vtpm.OpenNOPTPM,
			wantErr: false,
		},
		"MkdirAll fails": {
			fs:        afero.Afero{Fs: afero.NewMemMapFs()},
			keyWaiter: &stubKeyWaiter{},
			mapper:    &stubMapper{uuid: "test"},
			mounter:   &stubMounter{mkdirAllErr: someErr},
			openTPM:   vtpm.OpenNOPTPM,
			wantErr:   true,
		},
		"Mount fails": {
			fs:        afero.Afero{Fs: afero.NewMemMapFs()},
			keyWaiter: &stubKeyWaiter{},
			mapper:    &stubMapper{uuid: "test"},
			mounter:   &stubMounter{mountErr: someErr},
			openTPM:   vtpm.OpenNOPTPM,
			wantErr:   true,
		},
		"Unmount fails": {
			fs:        afero.Afero{Fs: afero.NewMemMapFs()},
			keyWaiter: &stubKeyWaiter{},
			mapper:    &stubMapper{uuid: "test"},
			mounter:   &stubMounter{unmountErr: someErr},
			openTPM:   vtpm.OpenNOPTPM,
			wantErr:   true,
		},
		"MarkNodeAsBootstrapped fails": {
			fs:        afero.Afero{Fs: afero.NewMemMapFs()},
			keyWaiter: &stubKeyWaiter{},
			mapper:    &stubMapper{uuid: "test"},
			mounter:   &stubMounter{unmountErr: someErr},
			openTPM:   failOpener,
			wantErr:   true,
		},
		"no state file": {
			fs:           afero.Afero{Fs: afero.NewMemMapFs()},
			keyWaiter:    &stubKeyWaiter{},
			mapper:       &stubMapper{uuid: "test"},
			mounter:      &stubMounter{},
			openTPM:      vtpm.OpenNOPTPM,
			missingState: true,
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			salt := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
			if !tc.missingState {
				handler := file.NewHandler(tc.fs)
				require.NoError(t, handler.WriteJSON(stateInfoPath, nodestate.NodeState{MeasurementSalt: salt}, file.OptMkdirAll))
			}

			setupManager := New(
				logger.NewTest(t),
				"test",
				tc.fs,
				tc.keyWaiter,
				tc.mapper,
				tc.mounter,
				tc.openTPM,
			)

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
		fs      afero.Afero
		mapper  *stubMapper
		wantErr bool
	}{
		"success": {
			fs:     afero.Afero{Fs: afero.NewMemMapFs()},
			mapper: &stubMapper{uuid: "test"},
		},
		"creating directory fails": {
			fs:      afero.Afero{Fs: afero.NewReadOnlyFs(afero.NewMemMapFs())},
			mapper:  &stubMapper{},
			wantErr: true,
		},
		"FormatDisk fails": {
			fs: afero.Afero{Fs: afero.NewMemMapFs()},
			mapper: &stubMapper{
				uuid:          "test",
				formatDiskErr: someErr,
			},
			wantErr: true,
		},
		"MapDisk fails": {
			fs: afero.Afero{Fs: afero.NewMemMapFs()},
			mapper: &stubMapper{
				uuid:                 "test",
				mapDiskErr:           someErr,
				mapDiskRepeatedCalls: 1,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			setupManager := New(logger.NewTest(t), "test", tc.fs, nil, tc.mapper, nil, nil)

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
		fs        afero.Afero
		salt      []byte
		writeFile bool
		wantErr   bool
	}{
		"success": {
			fs:        afero.Afero{Fs: afero.NewMemMapFs()},
			salt:      salt,
			writeFile: true,
		},
		"no state file": {
			fs:      afero.Afero{Fs: afero.NewMemMapFs()},
			wantErr: true,
		},
		"missing salt": {
			fs:        afero.Afero{Fs: afero.NewMemMapFs()},
			writeFile: true,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			if tc.writeFile {
				handler := file.NewHandler(tc.fs)
				state := nodestate.NodeState{MeasurementSalt: tc.salt}
				require.NoError(handler.WriteJSON("test-state.json", state, file.OptMkdirAll))
			}

			setupManager := New(logger.NewTest(t), "test", tc.fs, nil, nil, nil, nil)

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

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package setup

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net"
	"path/filepath"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/nodestate"
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
	testRecoveryDoer := &stubRecoveryDoer{
		passphrase: []byte("passphrase"),
		secret:     []byte("secret"),
	}

	testCases := map[string]struct {
		recoveryDoer    *stubRecoveryDoer
		mapper          *stubMapper
		mounter         *stubMounter
		configGenerator *stubConfigurationGenerator
		openDevice      vtpm.TPMOpenFunc
		missingState    bool
		wantErr         bool
	}{
		"success": {
			recoveryDoer:    testRecoveryDoer,
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{},
			openDevice:      vtpm.OpenNOPTPM,
		},
		"WaitForDecryptionKey fails": {
			recoveryDoer:    &stubRecoveryDoer{recoveryErr: someErr},
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{},
			openDevice:      vtpm.OpenNOPTPM,
			wantErr:         true,
		},
		"MapDisk fails": {
			recoveryDoer: testRecoveryDoer,
			mapper: &stubMapper{
				uuid:       "test",
				mapDiskErr: someErr,
			},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{},
			openDevice:      vtpm.OpenNOPTPM,
			wantErr:         true,
		},
		"MkdirAll fails": {
			recoveryDoer:    testRecoveryDoer,
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{mkdirAllErr: someErr},
			configGenerator: &stubConfigurationGenerator{},
			openDevice:      vtpm.OpenNOPTPM,
			wantErr:         true,
		},
		"Mount fails": {
			recoveryDoer:    testRecoveryDoer,
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{mountErr: someErr},
			configGenerator: &stubConfigurationGenerator{},
			openDevice:      vtpm.OpenNOPTPM,
			wantErr:         true,
		},
		"Unmount fails": {
			recoveryDoer:    testRecoveryDoer,
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{unmountErr: someErr},
			configGenerator: &stubConfigurationGenerator{},
			openDevice:      vtpm.OpenNOPTPM,
			wantErr:         true,
		},
		"MarkNodeAsBootstrapped fails": {
			recoveryDoer:    testRecoveryDoer,
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{unmountErr: someErr},
			configGenerator: &stubConfigurationGenerator{},
			openDevice:      failOpener,
			wantErr:         true,
		},
		"Generating config fails": {
			recoveryDoer:    testRecoveryDoer,
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{generateErr: someErr},
			openDevice:      failOpener,
			wantErr:         true,
		},
		"no state file": {
			recoveryDoer:    testRecoveryDoer,
			mapper:          &stubMapper{uuid: "test"},
			mounter:         &stubMounter{},
			configGenerator: &stubConfigurationGenerator{},
			openDevice:      vtpm.OpenNOPTPM,
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

			setupManager := &Manager{
				log:        logger.NewTest(t),
				csp:        "test",
				diskPath:   "disk-path",
				fs:         fs,
				mapper:     tc.mapper,
				mounter:    tc.mounter,
				config:     tc.configGenerator,
				openDevice: tc.openDevice,
			}

			err := setupManager.PrepareExistingDisk(tc.recoveryDoer)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
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
				uuid:       "test",
				mapDiskErr: someErr,
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

			setupManager := &Manager{
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

			setupManager := New(logger.NewTest(t), "test", "disk-path", fs, nil, nil, nil)

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

func TestRecoveryDoer(t *testing.T) {
	assert := assert.New(t)

	rejoinClientKey := []byte("rejoinClientKey")
	rejoinClientSecret := []byte("rejoinClientSecret")
	recoveryServerKey := []byte("recoveryServerKey")
	recoveryServerSecret := []byte("recoveryServerSecret")

	recoveryServerErr := errors.New("error")
	recoveryServer := &stubRecoveryServer{
		key:      recoveryServerKey,
		secret:   recoveryServerSecret,
		sendKeys: make(chan struct{}, 1),
		err:      recoveryServerErr,
	}
	rejoinClient := &stubRejoinClient{
		key:      rejoinClientKey,
		secret:   rejoinClientSecret,
		sendKeys: make(chan struct{}, 1),
	}
	recoverer := NewNodeRecoverer(recoveryServer, rejoinClient)

	var wg sync.WaitGroup
	var key, secret []byte
	var err error

	// error from recovery server
	wg.Add(1)
	go func() {
		defer wg.Done()
		key, secret, err = recoverer.Do("", "")
	}()
	recoveryServer.sendKeys <- struct{}{}
	wg.Wait()
	assert.ErrorIs(err, recoveryServerErr)

	recoveryServer.err = nil
	recoveryServer.sendKeys = make(chan struct{}, 1)

	// recovery server returns its key and secret
	wg.Add(1)
	go func() {
		defer wg.Done()
		key, secret, err = recoverer.Do("", "")
	}()
	recoveryServer.sendKeys <- struct{}{}
	wg.Wait()
	assert.NoError(err)
	assert.Equal(recoveryServerKey, key)
	assert.Equal(recoveryServerSecret, secret)

	recoveryServer.sendKeys = make(chan struct{}, 1)

	// rejoin client returns its key and secret
	wg.Add(1)
	go func() {
		defer wg.Done()
		key, secret, err = recoverer.Do("", "")
	}()
	rejoinClient.sendKeys <- struct{}{}
	wg.Wait()
	assert.NoError(err)
	assert.Equal(rejoinClientKey, key)
	assert.Equal(rejoinClientSecret, secret)
}

type stubRecoveryServer struct {
	key      []byte
	secret   []byte
	sendKeys chan struct{}
	err      error
}

func (s *stubRecoveryServer) Serve(ctx context.Context, _ net.Listener, _ string) ([]byte, []byte, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-s.sendKeys:
			return s.key, s.secret, s.err
		}
	}
}

type stubRejoinClient struct {
	key      []byte
	secret   []byte
	sendKeys chan struct{}
}

func (s *stubRejoinClient) Start(ctx context.Context, _ string) ([]byte, []byte) {
	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case <-s.sendKeys:
			return s.key, s.secret
		}
	}
}

type stubMapper struct {
	formatDiskCalled bool
	formatDiskErr    error
	mapDiskCalled    bool
	mapDiskErr       error
	unmapDiskCalled  bool
	unmapDiskErr     error
	uuid             string
}

func (s *stubMapper) DiskUUID() string {
	return s.uuid
}

func (s *stubMapper) FormatDisk(string) error {
	s.formatDiskCalled = true
	return s.formatDiskErr
}

func (s *stubMapper) MapDisk(string, string) error {
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

func (s *stubMounter) Mount(_ string, _ string, _ string, _ uintptr, _ string) error {
	s.mountCalled = true
	return s.mountErr
}

func (s *stubMounter) Unmount(_ string, _ int) error {
	s.unmountCalled = true
	return s.unmountErr
}

func (s *stubMounter) MkdirAll(_ string, _ fs.FileMode) error {
	return s.mkdirAllErr
}

type stubRecoveryDoer struct {
	passphrase  []byte
	secret      []byte
	recoveryErr error
}

func (s *stubRecoveryDoer) Do(_, _ string) (passphrase, measurementSecret []byte, err error) {
	return s.passphrase, s.secret, s.recoveryErr
}

type stubConfigurationGenerator struct {
	generateErr error
}

func (s *stubConfigurationGenerator) Generate(_, _, _, _ string) error {
	return s.generateErr
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package initserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/crypto/testvector"
	"github.com/edgelesssys/constellation/v2/internal/file"
	kmssetup "github.com/edgelesssys/constellation/v2/internal/kms/setup"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestNew(t *testing.T) {
	fh := file.NewHandler(afero.NewMemMapFs())

	testCases := map[string]struct {
		metadata stubMetadata
		wantErr  bool
	}{
		"success": {
			metadata: stubMetadata{initSecretHashVal: []byte("hash")},
		},
		"empty init secret hash": {
			metadata: stubMetadata{initSecretHashVal: nil},
			wantErr:  true,
		},
		"metadata error": {
			metadata: stubMetadata{initSecretHashErr: errors.New("error")},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			server, err := New(context.TODO(), newFakeLock(), &stubClusterInitializer{}, atls.NewFakeIssuer(variant.Dummy{}), fh, &tc.metadata, logger.NewTest(t))
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.NotNil(server)
			assert.NotEmpty(server.initSecretHash)
			assert.NotNil(server.log)
			assert.NotNil(server.nodeLock)
			assert.NotNil(server.initializer)
			assert.NotNil(server.grpcServer)
			assert.NotNil(server.fileHandler)
			assert.NotNil(server.disk)
		})
	}
}

func TestInit(t *testing.T) {
	someErr := errors.New("failed")
	lockedLock := newFakeLock()
	aqcuiredLock, lockErr := lockedLock.TryLockOnce(nil)
	require.True(t, aqcuiredLock)
	require.Nil(t, lockErr)

	initSecret := []byte("password")
	initSecretHash, err := bcrypt.GenerateFromPassword(initSecret, bcrypt.DefaultCost)
	require.NoError(t, err)

	masterSecret := uri.MasterSecret{Key: []byte("secret"), Salt: []byte("salt")}

	testCases := map[string]struct {
		nodeLock       *fakeLock
		initializer    ClusterInitializer
		disk           encryptedDisk
		fileHandler    file.Handler
		req            *initproto.InitRequest
		stream         stubStream
		logCollector   stubJournaldCollector
		initSecretHash []byte
		wantErr        bool
		wantShutdown   bool
	}{
		"successful init": {
			nodeLock:       newFakeLock(),
			initializer:    &stubClusterInitializer{},
			disk:           &stubDisk{},
			fileHandler:    file.NewHandler(afero.NewMemMapFs()),
			initSecretHash: initSecretHash,
			req:            &initproto.InitRequest{InitSecret: initSecret, KmsUri: masterSecret.EncodeToURI(), StorageUri: uri.NoStoreURI},
			stream:         stubStream{},
			logCollector:   stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte{})}},
			wantShutdown:   true,
		},
		"node locked": {
			nodeLock:       lockedLock,
			initializer:    &stubClusterInitializer{},
			disk:           &stubDisk{},
			fileHandler:    file.NewHandler(afero.NewMemMapFs()),
			req:            &initproto.InitRequest{InitSecret: initSecret, KmsUri: masterSecret.EncodeToURI(), StorageUri: uri.NoStoreURI},
			stream:         stubStream{},
			logCollector:   stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte{})}},
			initSecretHash: initSecretHash,
			wantErr:        true,
		},
		"disk open error": {
			nodeLock:       newFakeLock(),
			initializer:    &stubClusterInitializer{},
			disk:           &stubDisk{openErr: someErr},
			fileHandler:    file.NewHandler(afero.NewMemMapFs()),
			req:            &initproto.InitRequest{InitSecret: initSecret, KmsUri: masterSecret.EncodeToURI(), StorageUri: uri.NoStoreURI},
			stream:         stubStream{},
			logCollector:   stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte{})}},
			initSecretHash: initSecretHash,
			wantErr:        true,
			wantShutdown:   true,
		},
		"disk uuid error": {
			nodeLock:       newFakeLock(),
			initializer:    &stubClusterInitializer{},
			disk:           &stubDisk{uuidErr: someErr},
			fileHandler:    file.NewHandler(afero.NewMemMapFs()),
			req:            &initproto.InitRequest{InitSecret: initSecret, KmsUri: masterSecret.EncodeToURI(), StorageUri: uri.NoStoreURI},
			stream:         stubStream{},
			logCollector:   stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte{})}},
			initSecretHash: initSecretHash,
			wantErr:        true,
			wantShutdown:   true,
		},
		"disk update passphrase error": {
			nodeLock:       newFakeLock(),
			initializer:    &stubClusterInitializer{},
			disk:           &stubDisk{updatePassphraseErr: someErr},
			fileHandler:    file.NewHandler(afero.NewMemMapFs()),
			req:            &initproto.InitRequest{InitSecret: initSecret, KmsUri: masterSecret.EncodeToURI(), StorageUri: uri.NoStoreURI},
			stream:         stubStream{},
			logCollector:   stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte{})}},
			initSecretHash: initSecretHash,
			wantErr:        true,
			wantShutdown:   true,
		},
		"write state file error": {
			nodeLock:       newFakeLock(),
			initializer:    &stubClusterInitializer{},
			disk:           &stubDisk{},
			fileHandler:    file.NewHandler(afero.NewReadOnlyFs(afero.NewMemMapFs())),
			req:            &initproto.InitRequest{InitSecret: initSecret, KmsUri: masterSecret.EncodeToURI(), StorageUri: uri.NoStoreURI},
			stream:         stubStream{},
			logCollector:   stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte{})}},
			initSecretHash: initSecretHash,
			wantErr:        true,
			wantShutdown:   true,
		},
		"initialize cluster error": {
			nodeLock:       newFakeLock(),
			initializer:    &stubClusterInitializer{initClusterErr: someErr},
			disk:           &stubDisk{},
			fileHandler:    file.NewHandler(afero.NewMemMapFs()),
			req:            &initproto.InitRequest{InitSecret: initSecret, KmsUri: masterSecret.EncodeToURI(), StorageUri: uri.NoStoreURI},
			stream:         stubStream{},
			logCollector:   stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte{})}},
			initSecretHash: initSecretHash,
			wantErr:        true,
			wantShutdown:   true,
		},
		"wrong initSecret": {
			nodeLock:       newFakeLock(),
			initializer:    &stubClusterInitializer{},
			disk:           &stubDisk{},
			fileHandler:    file.NewHandler(afero.NewMemMapFs()),
			initSecretHash: initSecretHash,
			req:            &initproto.InitRequest{InitSecret: []byte("wrongpassword")},
			stream:         stubStream{},
			logCollector:   stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte{})}},
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			serveStopper := newStubServeStopper()
			server := &Server{
				nodeLock:          tc.nodeLock,
				initializer:       tc.initializer,
				disk:              tc.disk,
				fileHandler:       tc.fileHandler,
				log:               logger.NewTest(t),
				grpcServer:        serveStopper,
				cleaner:           &fakeCleaner{serveStopper: serveStopper},
				initSecretHash:    tc.initSecretHash,
				journaldCollector: &tc.logCollector,
			}

			err := server.Init(tc.req, &tc.stream)

			if tc.wantErr {
				assert.Error(err)

				if tc.wantShutdown {
					select {
					case <-serveStopper.shutdownCalled:
					case <-time.After(time.Second):
						t.Fatal("grpc server did not shut down")
					}
				}

				return
			}

			for _, res := range tc.stream.res {
				assert.NotNil(res.GetInitSuccess())
			}
			assert.NoError(err)
			assert.False(server.nodeLock.TryLockOnce(nil)) // lock should be locked
		})
	}
}

func TestSendLogs(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		logCollector   journaldCollection
		stream         stubStream
		expectedResult string
		wantErr        bool
	}{
		"success": {
			logCollector:   &stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte("asdf"))}},
			stream:         stubStream{},
			expectedResult: "asdf",
		},
		"fail collection": {
			logCollector: &stubJournaldCollector{collectErr: someError},
			wantErr:      true,
		},
		"fail to send": {
			logCollector: &stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewReader([]byte("asdf"))}},
			stream:       stubStream{sendError: someError},
			wantErr:      true,
		},
		"fail to read": {
			logCollector: &stubJournaldCollector{logPipe: &stubReadCloser{readErr: someError}},
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			serverStopper := newStubServeStopper()
			server := &Server{
				grpcServer:        serverStopper,
				journaldCollector: tc.logCollector,
			}

			err := server.sendLogs(&tc.stream)

			if tc.wantErr {
				assert.Error(err)

				return
			}

			assert.NoError(err)
			for _, res := range tc.stream.res {
				assert.Equal(tc.expectedResult, string(res.GetLog().Log))
			}
		})
	}
}

func TestSetupDisk(t *testing.T) {
	testCases := map[string]struct {
		uuid      string
		masterKey []byte
		salt      []byte
		wantKey   []byte
	}{
		"lower case uuid": {
			uuid:      strings.ToLower(testvector.HKDF0xFF.Info),
			masterKey: testvector.HKDF0xFF.Secret,
			salt:      testvector.HKDF0xFF.Salt,
			wantKey:   testvector.HKDF0xFF.Output,
		},
		"upper case uuid": {
			uuid:      strings.ToUpper(testvector.HKDF0xFF.Info),
			masterKey: testvector.HKDF0xFF.Secret,
			salt:      testvector.HKDF0xFF.Salt,
			wantKey:   testvector.HKDF0xFF.Output,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			disk := &fakeDisk{
				uuid:    tc.uuid,
				wantKey: tc.wantKey,
			}
			server := &Server{
				disk: disk,
			}

			masterSecret := uri.MasterSecret{Key: tc.masterKey, Salt: tc.salt}

			cloudKms, err := kmssetup.KMS(context.Background(), uri.NoStoreURI, masterSecret.EncodeToURI())
			require.NoError(err)
			assert.NoError(server.setupDisk(context.Background(), cloudKms))
		})
	}
}

type fakeDisk struct {
	uuid    string
	wantKey []byte
}

func (d *fakeDisk) Open() error {
	return nil
}

func (d *fakeDisk) Close() error {
	return nil
}

func (d *fakeDisk) UUID() (string, error) {
	return d.uuid, nil
}

func (d *fakeDisk) UpdatePassphrase(passphrase string) error {
	if passphrase != string(d.wantKey) {
		return errors.New("wrong passphrase")
	}
	return nil
}

type stubDisk struct {
	openErr                error
	closeErr               error
	uuid                   string
	uuidErr                error
	updatePassphraseErr    error
	updatePassphraseCalled bool
}

func (d *stubDisk) Open() error {
	return d.openErr
}

func (d *stubDisk) Close() error {
	return d.closeErr
}

func (d *stubDisk) UUID() (string, error) {
	return d.uuid, d.uuidErr
}

func (d *stubDisk) UpdatePassphrase(string) error {
	d.updatePassphraseCalled = true
	return d.updatePassphraseErr
}

type stubClusterInitializer struct {
	initClusterKubeconfig []byte
	initClusterErr        error
}

func (i *stubClusterInitializer) InitCluster(
	context.Context, string, string, string, []byte,
	[]byte, bool, components.Components, *logger.Logger,
) ([]byte, error) {
	return i.initClusterKubeconfig, i.initClusterErr
}

type stubServeStopper struct {
	shutdownCalled chan struct{}
}

func newStubServeStopper() *stubServeStopper {
	return &stubServeStopper{shutdownCalled: make(chan struct{}, 1)}
}

func (s *stubServeStopper) Serve(net.Listener) error {
	panic("should not be called in a test")
}

func (s *stubServeStopper) GracefulStop() {
	s.shutdownCalled <- struct{}{}
}

type fakeLock struct {
	state *sync.Mutex
}

func newFakeLock() *fakeLock {
	return &fakeLock{
		state: &sync.Mutex{},
	}
}

func (l *fakeLock) TryLockOnce(_ []byte) (bool, error) {
	return l.state.TryLock(), nil
}

type fakeCleaner struct {
	serveStopper
}

func (f *fakeCleaner) Clean() {
	go f.serveStopper.GracefulStop() // this is not the correct way to do this, but it's fine for testing
}

type stubMetadata struct {
	initSecretHashVal []byte
	initSecretHashErr error
}

func (m *stubMetadata) InitSecretHash(context.Context) ([]byte, error) {
	return m.initSecretHashVal, m.initSecretHashErr
}

type stubStream struct {
	res       []*initproto.InitResponse
	sendError error
	grpc.ServerStream
}

func (s *stubStream) Send(m *initproto.InitResponse) error {
	if s.sendError == nil {
		// we append here since we don't receive anything
		// if that if doesn't trigger
		s.res = append(s.res, m)
	}
	return s.sendError
}

func (s *stubStream) Context() context.Context {
	return context.Background()
}

type stubJournaldCollector struct {
	logPipe    io.ReadCloser
	collectErr error
}

func (s *stubJournaldCollector) Start() (io.ReadCloser, error) {
	return s.logPipe, s.collectErr
}

type stubReadCloser struct {
	reader   io.Reader
	readErr  error
	closeErr error
}

func (s *stubReadCloser) Read(p []byte) (n int, err error) {
	if s.readErr != nil {
		return 0, s.readErr
	}
	return s.reader.Read(p)
}

func (s *stubReadCloser) Close() error {
	return s.closeErr
}

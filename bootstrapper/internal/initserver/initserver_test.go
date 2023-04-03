/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package initserver

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
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
			wantShutdown:   true,
		},
		"node locked": {
			nodeLock:       lockedLock,
			initializer:    &stubClusterInitializer{},
			disk:           &stubDisk{},
			fileHandler:    file.NewHandler(afero.NewMemMapFs()),
			req:            &initproto.InitRequest{InitSecret: initSecret, KmsUri: masterSecret.EncodeToURI(), StorageUri: uri.NoStoreURI},
			initSecretHash: initSecretHash,
			wantErr:        true,
		},
		"disk open error": {
			nodeLock:       newFakeLock(),
			initializer:    &stubClusterInitializer{},
			disk:           &stubDisk{openErr: someErr},
			fileHandler:    file.NewHandler(afero.NewMemMapFs()),
			req:            &initproto.InitRequest{InitSecret: initSecret, KmsUri: masterSecret.EncodeToURI(), StorageUri: uri.NoStoreURI},
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
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			serveStopper := newStubServeStopper()
			server := &Server{
				nodeLock:       tc.nodeLock,
				initializer:    tc.initializer,
				disk:           tc.disk,
				fileHandler:    tc.fileHandler,
				log:            logger.NewTest(t),
				grpcServer:     serveStopper,
				cleaner:        &fakeCleaner{serveStopper: serveStopper},
				initSecretHash: tc.initSecretHash,
			}

			kubeconfig, err := server.Init(context.Background(), tc.req)

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

			assert.NoError(err)
			assert.NotNil(kubeconfig)
			assert.False(server.nodeLock.TryLockOnce(nil)) // lock should be locked
		})
	}
}

func TestGetLogs(t *testing.T) {
	initSecret := []byte("password")
	initSecretHash, err := bcrypt.GenerateFromPassword(initSecret, bcrypt.DefaultCost)
	require.NoError(t, err)

	masterSecret := uri.MasterSecret{Key: []byte("secret"), Salt: []byte("salt")}

	decryptor := func() cipher.AEAD {
		key, err := crypto.DeriveKey(masterSecret.Key, masterSecret.Salt, nil, 16) // 16 byte = 128 bit
		require.NoError(t, err)

		block, err := aes.NewCipher(key)
		require.NoError(t, err)
		decryptor, err := cipher.NewGCM(block)
		require.NoError(t, err)

		return decryptor
	}
	testCases := map[string]struct {
		initSecretHash    []byte
		masterSecret      uri.MasterSecret
		req               *initproto.LogRequest
		stream            stubLogStream
		decryptor         cipher.AEAD
		journaldCollector stubJournaldCollector
		wantErr           bool
		wantNoRes         bool
		tryDecrypt        bool // useful if sending fails
		wantDecryptable   bool
	}{
		"success": {
			initSecretHash:    initSecretHash,
			masterSecret:      masterSecret,
			req:               &initproto.LogRequest{InitSecret: initSecret},
			journaldCollector: stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewBuffer([]byte("asdf"))}},
			stream:            stubLogStream{},
			decryptor:         decryptor(),
			wantDecryptable:   true,
		},
		"collection error": {
			initSecretHash:    initSecretHash,
			req:               &initproto.LogRequest{InitSecret: initSecret},
			journaldCollector: stubJournaldCollector{collectErr: errors.New("failed")},
			wantErr:           true,
			wantNoRes:         true,
		},
		"wrong init secret": {
			initSecretHash: initSecretHash,
			req:            &initproto.LogRequest{InitSecret: []byte("asdf")},
			wantErr:        true,
			wantNoRes:      true,
		},
		"error executing command": {
			initSecretHash: initSecretHash,
			req:            &initproto.LogRequest{InitSecret: initSecret},
			wantErr:        true,
			wantNoRes:      true,
		},
		"decode master secret fail": {
			initSecretHash: initSecretHash,
			req:            &initproto.LogRequest{InitSecret: initSecret},
			masterSecret:   uri.MasterSecret{Key: nil, Salt: nil},
			wantErr:        true,
			wantNoRes:      true,
		},
		"send error": {
			initSecretHash:    initSecretHash,
			req:               &initproto.LogRequest{InitSecret: initSecret},
			stream:            stubLogStream{sendError: errors.New("failed")},
			journaldCollector: stubJournaldCollector{logPipe: &stubReadCloser{reader: bytes.NewBuffer([]byte("asdf"))}},
			masterSecret:      masterSecret,
			decryptor:         decryptor(),
			wantErr:           true,
			wantNoRes:         true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			server := &Server{
				initSecretHash:    tc.initSecretHash,
				kmsURI:            tc.masterSecret.EncodeToURI(),
				journaldCollector: &tc.journaldCollector,
				log:               logger.NewTest(t),
			}

			err := server.GetLogs(tc.req, &tc.stream)

			if tc.wantDecryptable {
				for _, res := range tc.stream.res {
					nonce := res.Nonce
					ciphertext := res.Log
					decrypted, err := tc.decryptor.Open(nil, nonce, ciphertext, nil)
					require.NoError(err)
					assert.Equal([]byte("asdf"), decrypted)
				}
			}

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			if tc.wantNoRes {
				assert.Empty(tc.stream.res)
			} else {
				assert.NotEmpty(tc.stream.res)
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

type stubLogStream struct {
	res       []*initproto.LogResponse
	sendError error
	grpc.ServerStream
}

func (s *stubLogStream) Send(m *initproto.LogResponse) error {
	if s.sendError == nil {
		// we append here since we don't receive anything
		// if that if doesn't trigger
		s.res = append(s.res, m)
	}
	return s.sendError
}

func (s *stubLogStream) Context() context.Context {
	return context.Background()
}

type journaldCollection interface {
	Start() (io.ReadCloser, error)
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

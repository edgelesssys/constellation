package initserver

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/internal/crypto/testvector"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestNew(t *testing.T) {
	assert := assert.New(t)

	fh := file.NewHandler(afero.NewMemMapFs())
	server := New(newFakeLock(), &stubClusterInitializer{}, nil, fh, logger.NewTest(t))
	assert.NotNil(server)
	assert.NotNil(server.log)
	assert.NotNil(server.nodeLock)
	assert.NotNil(server.initializer)
	assert.NotNil(server.grpcServer)
	assert.NotNil(server.fileHandler)
	assert.NotNil(server.disk)
}

func TestInit(t *testing.T) {
	someErr := errors.New("failed")
	lockedLock := newFakeLock()
	aqcuiredLock, lockErr := lockedLock.TryLockOnce(nil)
	require.True(t, aqcuiredLock)
	require.Nil(t, lockErr)

	testCases := map[string]struct {
		nodeLock     *fakeLock
		initializer  ClusterInitializer
		disk         encryptedDisk
		fileHandler  file.Handler
		req          *initproto.InitRequest
		wantErr      bool
		wantShutdown bool
	}{
		"successful init": {
			nodeLock:    newFakeLock(),
			initializer: &stubClusterInitializer{},
			disk:        &stubDisk{},
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
			req:         &initproto.InitRequest{},
		},
		"node locked": {
			nodeLock:     lockedLock,
			initializer:  &stubClusterInitializer{},
			disk:         &stubDisk{},
			fileHandler:  file.NewHandler(afero.NewMemMapFs()),
			req:          &initproto.InitRequest{},
			wantErr:      true,
			wantShutdown: true,
		},
		"disk open error": {
			nodeLock:    newFakeLock(),
			initializer: &stubClusterInitializer{},
			disk:        &stubDisk{openErr: someErr},
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
			req:         &initproto.InitRequest{},
			wantErr:     true,
		},
		"disk uuid error": {
			nodeLock:    newFakeLock(),
			initializer: &stubClusterInitializer{},
			disk:        &stubDisk{uuidErr: someErr},
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
			req:         &initproto.InitRequest{},
			wantErr:     true,
		},
		"disk update passphrase error": {
			nodeLock:    newFakeLock(),
			initializer: &stubClusterInitializer{},
			disk:        &stubDisk{updatePassphraseErr: someErr},
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
			req:         &initproto.InitRequest{},
			wantErr:     true,
		},
		"write state file error": {
			nodeLock:    newFakeLock(),
			initializer: &stubClusterInitializer{},
			disk:        &stubDisk{},
			fileHandler: file.NewHandler(afero.NewReadOnlyFs(afero.NewMemMapFs())),
			req:         &initproto.InitRequest{},
			wantErr:     true,
		},
		"initialize cluster error": {
			nodeLock:    newFakeLock(),
			initializer: &stubClusterInitializer{initClusterErr: someErr},
			disk:        &stubDisk{},
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
			req:         &initproto.InitRequest{},
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			serveStopper := newStubServeStopper()
			server := &Server{
				nodeLock:    tc.nodeLock,
				initializer: tc.initializer,
				disk:        tc.disk,
				fileHandler: tc.fileHandler,
				log:         logger.NewTest(t),
				grpcServer:  serveStopper,
				cleaner:     &fakeCleaner{serveStopper: serveStopper},
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

func TestSSHProtoKeysToMap(t *testing.T) {
	testCases := map[string]struct {
		keys []*initproto.SSHUserKey
		want map[string]string
	}{
		"empty": {
			keys: []*initproto.SSHUserKey{},
			want: map[string]string{},
		},
		"one key": {
			keys: []*initproto.SSHUserKey{
				{Username: "key1", PublicKey: "key1-key"},
			},
			want: map[string]string{
				"key1": "key1-key",
			},
		},
		"two keys": {
			keys: []*initproto.SSHUserKey{
				{Username: "key1", PublicKey: "key1-key"},
				{Username: "key2", PublicKey: "key2-key"},
			},
			want: map[string]string{
				"key1": "key1-key",
				"key2": "key2-key",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			got := sshProtoKeysToMap(tc.keys)
			assert.Equal(tc.want, got)
		})
	}
}

func TestSetupDisk(t *testing.T) {
	testCases := map[string]struct {
		uuid         string
		masterSecret []byte
		salt         []byte
		wantKey      []byte
	}{
		"lower case uuid": {
			uuid:         strings.ToLower(testvector.HKDF0xFF.Info),
			masterSecret: testvector.HKDF0xFF.Secret,
			salt:         testvector.HKDF0xFF.Salt,
			wantKey:      testvector.HKDF0xFF.Output,
		},
		"upper case uuid": {
			uuid:         strings.ToUpper(testvector.HKDF0xFF.Info),
			masterSecret: testvector.HKDF0xFF.Secret,
			salt:         testvector.HKDF0xFF.Salt,
			wantKey:      testvector.HKDF0xFF.Output,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			disk := &fakeDisk{
				uuid:    tc.uuid,
				wantKey: tc.wantKey,
			}
			server := &Server{
				disk: disk,
			}

			assert.NoError(server.setupDisk(tc.masterSecret, tc.salt))
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
	context.Context, []string, string, string, []byte, []uint32,
	resources.KMSConfig, map[string]string, []byte, *logger.Logger,
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

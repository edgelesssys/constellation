/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package joinclient

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	testclock "k8s.io/utils/clock/testing"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestClient(t *testing.T) {
	someErr := errors.New("failed")
	lockedLock := newFakeLock()
	aqcuiredLock, lockErr := lockedLock.TryLockOnce(nil)
	require.True(t, aqcuiredLock)
	require.Nil(t, lockErr)
	workerSelf := metadata.InstanceMetadata{Role: role.Worker, Name: "node-1"}
	controlSelf := metadata.InstanceMetadata{Role: role.ControlPlane, Name: "node-5"}
	peers := []metadata.InstanceMetadata{
		{Role: role.Worker, Name: "node-2", VPCIP: "192.0.2.8"},
		{Role: role.ControlPlane, Name: "node-3", VPCIP: "192.0.2.1"},
		{Role: role.ControlPlane, Name: "node-4", VPCIP: "192.0.2.2"},
		{Role: role.ControlPlane, Name: "node-5", VPCIP: "192.0.2.3"},
	}

	testCases := map[string]struct {
		role          role.Role
		clusterJoiner *stubClusterJoiner
		disk          encryptedDisk
		nodeLock      *fakeLock
		apiAnswers    []any
		wantLock      bool
		wantJoin      bool
	}{
		"on worker: metadata self: errors occur": {
			role: role.Worker,
			apiAnswers: []any{
				selfAnswer{err: someErr},
				selfAnswer{err: someErr},
				selfAnswer{err: someErr},
				selfAnswer{instance: workerSelf},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{},
			},
			clusterJoiner: &stubClusterJoiner{},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{},
			wantJoin:      true,
			wantLock:      true,
		},
		"on worker: metadata self: invalid answer": {
			role: role.Worker,
			apiAnswers: []any{
				selfAnswer{},
				selfAnswer{instance: metadata.InstanceMetadata{Role: role.Worker}},
				selfAnswer{instance: metadata.InstanceMetadata{Name: "node-1"}},
				selfAnswer{instance: workerSelf},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{},
			},
			clusterJoiner: &stubClusterJoiner{},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{},
			wantJoin:      true,
			wantLock:      true,
		},
		"on worker: metadata list: errors occur": {
			role: role.Worker,
			apiAnswers: []any{
				selfAnswer{instance: workerSelf},
				listAnswer{err: someErr},
				listAnswer{err: someErr},
				listAnswer{err: someErr},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{},
			},
			clusterJoiner: &stubClusterJoiner{},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{},
			wantJoin:      true,
			wantLock:      true,
		},
		"on worker: metadata list: no control plane nodes in answer": {
			role: role.Worker,
			apiAnswers: []any{
				selfAnswer{instance: workerSelf},
				listAnswer{},
				listAnswer{},
				listAnswer{},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{},
			},
			clusterJoiner: &stubClusterJoiner{},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{},
			wantJoin:      true,
			wantLock:      true,
		},
		"on worker: issueJoinTicket errors": {
			role: role.Worker,
			apiAnswers: []any{
				selfAnswer{instance: workerSelf},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{err: someErr},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{err: someErr},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{},
			},
			clusterJoiner: &stubClusterJoiner{},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{},
			wantJoin:      true,
			wantLock:      true,
		},
		"on control plane: issueJoinTicket errors": {
			role: role.ControlPlane,
			apiAnswers: []any{
				selfAnswer{instance: controlSelf},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{err: someErr},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{err: someErr},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{},
			},
			clusterJoiner: &stubClusterJoiner{},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{},
			wantJoin:      true,
			wantLock:      true,
		},
		"on control plane: joinCluster fails": {
			role: role.ControlPlane,
			apiAnswers: []any{
				selfAnswer{instance: controlSelf},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{},
			},
			clusterJoiner: &stubClusterJoiner{joinClusterErr: someErr},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{},
			wantJoin:      true,
			wantLock:      true,
		},
		"on control plane: node already locked": {
			role: role.ControlPlane,
			apiAnswers: []any{
				selfAnswer{instance: controlSelf},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{},
			},
			clusterJoiner: &stubClusterJoiner{},
			nodeLock:      lockedLock,
			disk:          &stubDisk{},
			wantLock:      true,
		},
		"on control plane: disk open fails": {
			role:          role.ControlPlane,
			clusterJoiner: &stubClusterJoiner{},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{openErr: someErr},
		},
		"on control plane: disk uuid fails": {
			role:          role.ControlPlane,
			clusterJoiner: &stubClusterJoiner{},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{uuidErr: someErr},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			clock := testclock.NewFakeClock(time.Now())
			metadataAPI := newStubMetadataAPI()
			fileHandler := file.NewHandler(afero.NewMemMapFs())

			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, nil, netDialer)

			client := &JoinClient{
				nodeLock:    tc.nodeLock,
				timeout:     30 * time.Second,
				interval:    time.Millisecond,
				dialer:      dialer,
				disk:        tc.disk,
				joiner:      tc.clusterJoiner,
				fileHandler: fileHandler,
				metadataAPI: metadataAPI,
				clock:       clock,
				log:         logger.NewTest(t),
			}

			serverCreds := atlscredentials.New(nil, nil)
			joinServer := grpc.NewServer(grpc.Creds(serverCreds))
			joinserviceAPI := newStubJoinServiceAPI()
			joinproto.RegisterAPIServer(joinServer, joinserviceAPI)
			port := strconv.Itoa(constants.JoinServiceNodePort)
			listener := netDialer.GetListener(net.JoinHostPort("192.0.2.3", port))
			go joinServer.Serve(listener)
			defer joinServer.GracefulStop()

			client.Start(stubCleaner{})

			for _, a := range tc.apiAnswers {
				switch a := a.(type) {
				case selfAnswer:
					metadataAPI.selfAnswerC <- a
				case listAnswer:
					metadataAPI.listAnswerC <- a
				case issueJoinTicketAnswer:
					joinserviceAPI.issueJoinTicketAnswerC <- a
				}
				clock.Step(time.Second)
			}

			client.Stop()

			if tc.wantJoin {
				assert.True(tc.clusterJoiner.joinClusterCalled)
			} else {
				assert.False(tc.clusterJoiner.joinClusterCalled)
			}
			if tc.wantLock {
				assert.False(client.nodeLock.TryLockOnce(nil)) // lock should be locked
			} else {
				assert.True(client.nodeLock.TryLockOnce(nil))
			}
		})
	}
}

func TestClientConcurrentStartStop(t *testing.T) {
	netDialer := testdialer.NewBufconnDialer()
	dialer := dialer.New(nil, nil, netDialer)
	client := &JoinClient{
		nodeLock:    newFakeLock(),
		timeout:     30 * time.Second,
		interval:    30 * time.Second,
		dialer:      dialer,
		disk:        &stubDisk{},
		joiner:      &stubClusterJoiner{},
		fileHandler: file.NewHandler(afero.NewMemMapFs()),
		metadataAPI: &stubRepeaterMetadataAPI{},
		clock:       testclock.NewFakeClock(time.Now()),
		log:         logger.NewTest(t),
	}

	wg := sync.WaitGroup{}

	start := func() {
		defer wg.Done()
		client.Start(stubCleaner{})
	}

	stop := func() {
		defer wg.Done()
		client.Stop()
	}

	wg.Add(10)
	go stop()
	go start()
	go start()
	go stop()
	go stop()
	go start()
	go start()
	go stop()
	go stop()
	go start()
	wg.Wait()

	client.Stop()
}

func TestIsUnrecoverable(t *testing.T) {
	assert := assert.New(t)

	some := errors.New("failed")
	unrec := unrecoverableError{some}
	assert.True(isUnrecoverable(unrec))
	assert.False(isUnrecoverable(some))
}

type stubRepeaterMetadataAPI struct {
	selfInstance  metadata.InstanceMetadata
	selfErr       error
	listInstances []metadata.InstanceMetadata
	listErr       error
}

func (s *stubRepeaterMetadataAPI) Self(_ context.Context) (metadata.InstanceMetadata, error) {
	return s.selfInstance, s.selfErr
}

func (s *stubRepeaterMetadataAPI) List(_ context.Context) ([]metadata.InstanceMetadata, error) {
	return s.listInstances, s.listErr
}

func (s *stubRepeaterMetadataAPI) GetLoadBalancerEndpoint(_ context.Context) (string, string, error) {
	return "", "", nil
}

type stubMetadataAPI struct {
	selfAnswerC chan selfAnswer
	listAnswerC chan listAnswer
}

func newStubMetadataAPI() *stubMetadataAPI {
	return &stubMetadataAPI{
		selfAnswerC: make(chan selfAnswer),
		listAnswerC: make(chan listAnswer),
	}
}

func (s *stubMetadataAPI) Self(_ context.Context) (metadata.InstanceMetadata, error) {
	answer := <-s.selfAnswerC
	return answer.instance, answer.err
}

func (s *stubMetadataAPI) List(_ context.Context) ([]metadata.InstanceMetadata, error) {
	answer := <-s.listAnswerC
	return answer.instances, answer.err
}

func (s *stubMetadataAPI) GetLoadBalancerEndpoint(_ context.Context) (string, string, error) {
	return "", "", nil
}

type selfAnswer struct {
	instance metadata.InstanceMetadata
	err      error
}

type listAnswer struct {
	instances []metadata.InstanceMetadata
	err       error
}

type stubJoinServiceAPI struct {
	issueJoinTicketAnswerC chan issueJoinTicketAnswer

	joinproto.UnimplementedAPIServer
}

func newStubJoinServiceAPI() *stubJoinServiceAPI {
	return &stubJoinServiceAPI{
		issueJoinTicketAnswerC: make(chan issueJoinTicketAnswer),
	}
}

func (s *stubJoinServiceAPI) IssueJoinTicket(_ context.Context, _ *joinproto.IssueJoinTicketRequest,
) (*joinproto.IssueJoinTicketResponse, error) {
	answer := <-s.issueJoinTicketAnswerC
	if answer.resp == nil {
		answer.resp = &joinproto.IssueJoinTicketResponse{}
	}
	return answer.resp, answer.err
}

type issueJoinTicketAnswer struct {
	resp *joinproto.IssueJoinTicketResponse
	err  error
}

type stubClusterJoiner struct {
	joinClusterCalled bool
	joinClusterErr    error
}

func (j *stubClusterJoiner) JoinCluster(context.Context, *kubeadm.BootstrapTokenDiscovery, role.Role, components.Components, *logger.Logger) error {
	j.joinClusterCalled = true
	return j.joinClusterErr
}

type stubDisk struct {
	openErr                error
	uuid                   string
	uuidErr                error
	updatePassphraseErr    error
	updatePassphraseCalled bool
}

func (d *stubDisk) Open() (func(), error) {
	return func() {}, d.openErr
}

func (d *stubDisk) UUID() (string, error) {
	return d.uuid, d.uuidErr
}

func (d *stubDisk) UpdatePassphrase(string) error {
	d.updatePassphraseCalled = true
	return d.updatePassphraseErr
}

type stubCleaner struct{}

func (c stubCleaner) Clean() {}

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

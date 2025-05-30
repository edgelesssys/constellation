/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package joinclient

import (
	"context"
	"crypto/ed25519"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
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
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	testclock "k8s.io/utils/clock/testing"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestClient(t *testing.T) {
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
	caDerivationKey := make([]byte, 256)
	respCaKey := &joinproto.IssueJoinTicketResponse{AuthorizedCaPublicKey: caDerivationKey}

	makeIssueJoinTicketAnswerWithValidCert := func(t *testing.T, originalAnswer issueJoinTicketAnswer, fh file.Handler) issueJoinTicketAnswer {
		require := require.New(t)

		tries := 0
		var sshKeyBytes []byte
		for {
			t.Logf("Trying to read ssh host key: %d/3", tries)
			sshKey, err := fh.Read(constants.SSHHostKeyPath)
			if err != nil {
				tries++
				if tries >= 3 {
					assert.Fail(t, "ssh key never written: ", err)
				}
				time.Sleep(time.Second * 3)
				continue
			} else {
				sshKeyBytes = sshKey
				break
			}
		}
		sshKey, err := ssh.ParsePrivateKey(sshKeyBytes)
		require.NoError(err)
		_, randomCAKey, err := ed25519.GenerateKey(nil)
		require.NoError(err)
		randomCA, err := ssh.NewSignerFromSigner(randomCAKey)
		require.NoError(err)

		cert, err := crypto.GenerateSSHHostCertificate([]string{"asdf"}, sshKey.PublicKey(), randomCA)
		require.NoError(err)

		certBytes := ssh.MarshalAuthorizedKey(cert)

		if originalAnswer.resp == nil {
			originalAnswer.resp = &joinproto.IssueJoinTicketResponse{HostCertificate: certBytes}
		} else {
			originalAnswer.resp.HostCertificate = certBytes
		}

		return originalAnswer
	}

	testCases := map[string]struct {
		role                role.Role
		clusterJoiner       *stubClusterJoiner
		disk                encryptedDisk
		nodeLock            *fakeLock
		apiAnswers          []any
		wantLock            bool
		wantJoin            bool
		wantNumJoins        int
		wantNotMatchingCert bool
	}{
		"on worker: metadata self: errors occur": {
			role: role.Worker,
			apiAnswers: []any{
				selfAnswer{err: assert.AnError},
				selfAnswer{err: assert.AnError},
				selfAnswer{err: assert.AnError},
				selfAnswer{instance: workerSelf},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{resp: respCaKey},
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
				issueJoinTicketAnswer{resp: respCaKey},
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
				listAnswer{err: assert.AnError},
				listAnswer{err: assert.AnError},
				listAnswer{err: assert.AnError},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{resp: respCaKey},
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
				issueJoinTicketAnswer{resp: respCaKey},
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
				issueJoinTicketAnswer{err: assert.AnError},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{err: assert.AnError},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{resp: respCaKey},
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
				issueJoinTicketAnswer{err: assert.AnError},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{err: assert.AnError},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{resp: respCaKey},
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
				issueJoinTicketAnswer{resp: respCaKey},
			},
			clusterJoiner: &stubClusterJoiner{numBadCalls: -1, joinClusterErr: assert.AnError},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{},
			wantJoin:      true,
			wantLock:      true,
		},
		"on control plane: joinCluster fails transiently": {
			role: role.ControlPlane,
			apiAnswers: []any{
				selfAnswer{instance: controlSelf},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{resp: respCaKey},
			},
			clusterJoiner: &stubClusterJoiner{numBadCalls: 1, joinClusterErr: assert.AnError},
			nodeLock:      newFakeLock(),
			disk:          &stubDisk{},
			wantJoin:      true,
			wantLock:      true,
			wantNumJoins:  2,
		},
		"on control plane: node already locked": {
			role: role.ControlPlane,
			apiAnswers: []any{
				selfAnswer{instance: controlSelf},
				listAnswer{instances: peers},
				issueJoinTicketAnswer{resp: respCaKey},
			},
			clusterJoiner:       &stubClusterJoiner{},
			nodeLock:            lockedLock,
			disk:                &stubDisk{},
			wantLock:            true,
			wantNotMatchingCert: true,
		},
		"on control plane: disk open fails": {
			role:                role.ControlPlane,
			clusterJoiner:       &stubClusterJoiner{},
			nodeLock:            newFakeLock(),
			disk:                &stubDisk{openErr: assert.AnError},
			wantNotMatchingCert: true,
		},
		"on control plane: disk uuid fails": {
			role:                role.ControlPlane,
			clusterJoiner:       &stubClusterJoiner{},
			nodeLock:            newFakeLock(),
			disk:                &stubDisk{uuidErr: assert.AnError},
			wantNotMatchingCert: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

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

				stopC:    make(chan struct{}, 1),
				stopDone: make(chan struct{}, 1),
			}

			serverCreds := atlscredentials.New(nil, nil)
			joinServer := grpc.NewServer(grpc.Creds(serverCreds))
			joinserviceAPI := newStubJoinServiceAPI()
			joinproto.RegisterAPIServer(joinServer, joinserviceAPI)
			port := strconv.Itoa(constants.JoinServiceNodePort)
			listener := netDialer.GetListener(net.JoinHostPort("192.0.2.3", port))
			go joinServer.Serve(listener)
			defer joinServer.GracefulStop()

			go func() { _ = client.Start(stubCleaner{}) }()

			for _, a := range tc.apiAnswers {
				switch a := a.(type) {
				case selfAnswer:
					metadataAPI.selfAnswerC <- a
				case listAnswer:
					metadataAPI.listAnswerC <- a
				case issueJoinTicketAnswer:
					joinserviceAPI.issueJoinTicketAnswerC <- makeIssueJoinTicketAnswerWithValidCert(t, a, fileHandler)
				}
				clock.Step(time.Second)
			}

			client.Stop()

			if !tc.wantNotMatchingCert {
				hostCertBytes, err := fileHandler.Read(constants.SSHHostCertificatePath)
				require.NoError(err)
				hostKeyBytes, err := fileHandler.Read(constants.SSHHostKeyPath)
				require.NoError(err)

				hostCertKey, _, _, _, err := ssh.ParseAuthorizedKey(hostCertBytes)
				require.NoError(err)
				hostCert, ok := hostCertKey.(*ssh.Certificate)
				require.True(ok)

				hostKey, err := ssh.ParsePrivateKey(hostKeyBytes)
				require.NoError(err)

				// TODO: for some reason, the private keys are different
				assert.Equal(string(ssh.MarshalAuthorizedKey(hostKey.PublicKey())), string(ssh.MarshalAuthorizedKey(hostCert.Key)))
			}

			if tc.wantJoin {
				assert.Greater(tc.clusterJoiner.joinClusterCalled, 0)
			} else {
				assert.Equal(0, tc.clusterJoiner.joinClusterCalled)
			}
			if tc.wantNumJoins > 0 {
				assert.GreaterOrEqual(tc.clusterJoiner.joinClusterCalled, tc.wantNumJoins)
			}
			if tc.wantLock {
				assert.False(client.nodeLock.TryLockOnce(nil)) // lock should be locked
			} else {
				assert.True(client.nodeLock.TryLockOnce(nil))
			}
		})
	}
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
	joinClusterCalled int
	numBadCalls       int
	joinClusterErr    error
}

func (j *stubClusterJoiner) JoinCluster(context.Context, *kubeadm.BootstrapTokenDiscovery, role.Role, components.Components) error {
	j.joinClusterCalled++
	if j.numBadCalls == 0 {
		return nil
	}
	j.numBadCalls--
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

func (d *stubDisk) MarkDiskForReset() error {
	return nil
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

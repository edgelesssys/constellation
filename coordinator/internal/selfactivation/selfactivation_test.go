package selfactivation

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/activation/activationproto"
	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	testclock "k8s.io/utils/clock/testing"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestClient(t *testing.T) {
	someErr := errors.New("failed")
	peers := []cloudtypes.Instance{
		{Role: role.Node, PrivateIPs: []string{"192.0.2.8"}},
		{Role: role.Coordinator, PrivateIPs: []string{"192.0.2.1"}},
		{Role: role.Coordinator, PrivateIPs: []string{"192.0.2.2", "192.0.2.3"}},
	}

	testCases := map[string]struct {
		role       role.Role
		apiAnswers []any
		setterAPI  *stubActiveSetter
	}{
		"on node: metadata self: errors occur": {
			role: role.Node,
			apiAnswers: []any{
				selfAnswer{err: someErr},
				selfAnswer{err: someErr},
				selfAnswer{err: someErr},
				selfAnswer{instance: cloudtypes.Instance{Role: role.Node}},
				listAnswer{instances: peers},
				activateNodeAnswer{},
			},
			setterAPI: &stubActiveSetter{},
		},
		"on node: metadata self: no role in answer": {
			role: role.Node,
			apiAnswers: []any{
				selfAnswer{},
				selfAnswer{},
				selfAnswer{},
				selfAnswer{instance: cloudtypes.Instance{Role: role.Node}},
				listAnswer{instances: peers},
				activateNodeAnswer{},
			},
			setterAPI: &stubActiveSetter{},
		},
		"on node: metadata list: errors occur": {
			role: role.Node,
			apiAnswers: []any{
				selfAnswer{instance: cloudtypes.Instance{Role: role.Node}},
				listAnswer{err: someErr},
				listAnswer{err: someErr},
				listAnswer{err: someErr},
				listAnswer{instances: peers},
				activateNodeAnswer{},
			},
			setterAPI: &stubActiveSetter{},
		},
		"on node: metadata list: no coordinators in answer": {
			role: role.Node,
			apiAnswers: []any{
				selfAnswer{instance: cloudtypes.Instance{Role: role.Node}},
				listAnswer{},
				listAnswer{},
				listAnswer{},
				listAnswer{instances: peers},
				activateNodeAnswer{},
			},
			setterAPI: &stubActiveSetter{},
		},
		"on node: aaas ActivateNode: errors": {
			role: role.Node,
			apiAnswers: []any{
				selfAnswer{instance: cloudtypes.Instance{Role: role.Node}},
				listAnswer{instances: peers},
				activateNodeAnswer{err: someErr},
				listAnswer{instances: peers},
				activateNodeAnswer{err: someErr},
				listAnswer{instances: peers},
				activateNodeAnswer{},
			},
			setterAPI: &stubActiveSetter{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			clock := testclock.NewFakeClock(time.Now())
			metadataAPI := newStubMetadataAPI()

			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, nil, netDialer)

			client := &SelfActivationClient{
				timeout:     30 * time.Second,
				interval:    time.Millisecond,
				dialer:      dialer,
				setterAPI:   tc.setterAPI,
				metadataAPI: metadataAPI,
				clock:       clock,
				log:         zaptest.NewLogger(t),
			}

			serverCreds := atlscredentials.New(nil, nil)
			activationSever := grpc.NewServer(grpc.Creds(serverCreds))
			activationAPI := newStubActivationServiceAPI()
			activationproto.RegisterAPIServer(activationSever, activationAPI)
			port := strconv.Itoa(constants.ActivationServicePort)
			listener := netDialer.GetListener(net.JoinHostPort("192.0.2.3", port))
			go activationSever.Serve(listener)
			defer activationSever.GracefulStop()

			client.Start()

			for _, a := range tc.apiAnswers {
				switch a := a.(type) {
				case selfAnswer:
					metadataAPI.selfAnswerC <- a
				case listAnswer:
					metadataAPI.listAnswerC <- a
				case activateNodeAnswer:
					activationAPI.activateNodeAnswerC <- a
				}
				clock.Step(time.Second)
			}

			client.Stop()

			if tc.role == role.Node {
				assert.Equal(1, tc.setterAPI.setNodeActiveCalled)
			} else {
				assert.Equal(1, tc.setterAPI.setCoordinatorActiveCalled)
			}
		})
	}
}

func TestClientConcurrentStartStop(t *testing.T) {
	client := &SelfActivationClient{
		setterAPI:   &stubActiveSetter{},
		metadataAPI: &stubRepeaterMetadataAPI{},
		clock:       testclock.NewFakeClock(time.Now()),
		log:         zap.NewNop(),
	}

	wg := sync.WaitGroup{}

	start := func() {
		defer wg.Done()
		client.Start()
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

type stubActiveSetter struct {
	setNodeActiveErr           error
	setNodeActiveCalled        int
	setCoordinatorActiveErr    error
	setCoordinatorActiveCalled int
}

func (s *stubActiveSetter) SetNodeActive(_, _, _ []byte, _, _, _ string) error {
	s.setNodeActiveCalled++
	return s.setNodeActiveErr
}

func (s *stubActiveSetter) SetCoordinatorActive() error {
	s.setCoordinatorActiveCalled++
	return s.setCoordinatorActiveErr
}

type stubRepeaterMetadataAPI struct {
	selfInstance  cloudtypes.Instance
	selfErr       error
	listInstances []cloudtypes.Instance
	listErr       error
}

func (s *stubRepeaterMetadataAPI) Self(_ context.Context) (cloudtypes.Instance, error) {
	return s.selfInstance, s.selfErr
}

func (s *stubRepeaterMetadataAPI) List(_ context.Context) ([]cloudtypes.Instance, error) {
	return s.listInstances, s.listErr
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

func (s *stubMetadataAPI) Self(_ context.Context) (cloudtypes.Instance, error) {
	answer := <-s.selfAnswerC
	return answer.instance, answer.err
}

func (s *stubMetadataAPI) List(_ context.Context) ([]cloudtypes.Instance, error) {
	answer := <-s.listAnswerC
	return answer.instances, answer.err
}

type selfAnswer struct {
	instance cloudtypes.Instance
	err      error
}

type listAnswer struct {
	instances []cloudtypes.Instance
	err       error
}

type stubActivationServiceAPI struct {
	activateNodeAnswerC chan activateNodeAnswer

	activationproto.UnimplementedAPIServer
}

func newStubActivationServiceAPI() *stubActivationServiceAPI {
	return &stubActivationServiceAPI{
		activateNodeAnswerC: make(chan activateNodeAnswer),
	}
}

func (s *stubActivationServiceAPI) ActivateWorkerNode(_ context.Context, _ *activationproto.ActivateWorkerNodeRequest,
) (*activationproto.ActivateWorkerNodeResponse, error) {
	answer := <-s.activateNodeAnswerC
	if answer.resp == nil {
		answer.resp = &activationproto.ActivateWorkerNodeResponse{}
	}
	return answer.resp, answer.err
}

type activateNodeAnswer struct {
	resp *activationproto.ActivateWorkerNodeResponse
	err  error
}

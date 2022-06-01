package keyservice

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/edgelesssys/constellation/joinservice/joinproto"
	"github.com/edgelesssys/constellation/state/keyservice/keyproto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	testclock "k8s.io/utils/clock/testing"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRequestKeyLoop(t *testing.T) {
	clockstep := struct{}{}
	someErr := errors.New("failed")
	defaultInstance := metadata.InstanceMetadata{
		Name:       "test-instance",
		ProviderID: "/test/provider",
		Role:       role.ControlPlane,
		PrivateIPs: []string{"192.0.2.1"},
	}

	testCases := map[string]struct {
		answers []any
	}{
		"success": {
			answers: []any{
				listAnswer{listResponse: []metadata.InstanceMetadata{defaultInstance}},
				issueRejoinTicketAnswer{stateDiskKey: []byte{0x1}, measurementSecret: []byte{0x2}},
				pushStateDiskKeyAnswer{},
			},
		},
		"recover metadata list error": {
			answers: []any{
				listAnswer{err: someErr},
				clockstep,
				listAnswer{listResponse: []metadata.InstanceMetadata{defaultInstance}},
				issueRejoinTicketAnswer{stateDiskKey: []byte{0x1}, measurementSecret: []byte{0x2}},
				pushStateDiskKeyAnswer{},
			},
		},
		"recover issue rejoin ticket error": {
			answers: []any{
				listAnswer{listResponse: []metadata.InstanceMetadata{defaultInstance}},
				issueRejoinTicketAnswer{err: someErr},
				clockstep,
				listAnswer{listResponse: []metadata.InstanceMetadata{defaultInstance}},
				issueRejoinTicketAnswer{stateDiskKey: []byte{0x1}, measurementSecret: []byte{0x2}},
				pushStateDiskKeyAnswer{},
			},
		},
		"recover push key error": {
			answers: []any{
				listAnswer{listResponse: []metadata.InstanceMetadata{defaultInstance}},
				issueRejoinTicketAnswer{stateDiskKey: []byte{0x1}, measurementSecret: []byte{0x2}},
				pushStateDiskKeyAnswer{err: someErr},
				clockstep,
				listAnswer{listResponse: []metadata.InstanceMetadata{defaultInstance}},
				issueRejoinTicketAnswer{stateDiskKey: []byte{0x1}, measurementSecret: []byte{0x2}},
				pushStateDiskKeyAnswer{},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			metadataServer := newStubMetadataServer()
			joinServer := newStubJoinAPIServer()
			keyServer := newStubKeyAPIServer()

			listener := bufconn.Listen(1024)
			defer listener.Close()
			creds := atlscredentials.New(atls.NewFakeIssuer(oid.Dummy{}), nil)
			grpcServer := grpc.NewServer(grpc.Creds(creds))
			joinproto.RegisterAPIServer(grpcServer, joinServer)
			keyproto.RegisterAPIServer(grpcServer, keyServer)
			go grpcServer.Serve(listener)
			defer grpcServer.GracefulStop()

			clock := testclock.NewFakeClock(time.Now())
			keyReceived := make(chan struct{}, 1)
			keyWaiter := &KeyAPI{
				listenAddr:  "192.0.2.1:30090",
				log:         logger.NewTest(t),
				metadata:    metadataServer,
				keyReceived: keyReceived,
				clock:       clock,
				timeout:     1 * time.Second,
				interval:    1 * time.Second,
			}
			grpcOpts := []grpc.DialOption{
				grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
					return listener.DialContext(ctx)
				}),
			}

			// Start the request loop under tests.
			done := make(chan struct{})
			go func() {
				defer close(done)
				keyWaiter.requestKeyLoop("1234", grpcOpts...)
			}()

			// Play test case answers.
			for _, answ := range tc.answers {
				switch answ := answ.(type) {
				case listAnswer:
					metadataServer.listAnswerC <- answ
				case issueRejoinTicketAnswer:
					joinServer.issueRejoinTicketAnswerC <- answ
				case pushStateDiskKeyAnswer:
					keyServer.pushStateDiskKeyAnswerC <- answ
				default:
					clock.Step(time.Second)
				}
			}

			// Stop the request loop.
			keyReceived <- struct{}{}
		})
	}
}

func TestPushStateDiskKey(t *testing.T) {
	testCases := map[string]struct {
		testAPI *KeyAPI
		request *keyproto.PushStateDiskKeyRequest
		wantErr bool
	}{
		"success": {
			testAPI: &KeyAPI{keyReceived: make(chan struct{}, 1)},
			request: &keyproto.PushStateDiskKeyRequest{StateDiskKey: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), MeasurementSecret: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")},
		},
		"key already set": {
			testAPI: &KeyAPI{
				keyReceived: make(chan struct{}, 1),
				key:         []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
			},
			request: &keyproto.PushStateDiskKeyRequest{StateDiskKey: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"), MeasurementSecret: []byte("CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC")},
			wantErr: true,
		},
		"incorrect size of pushed key": {
			testAPI: &KeyAPI{keyReceived: make(chan struct{}, 1)},
			request: &keyproto.PushStateDiskKeyRequest{StateDiskKey: []byte("AAAAAAAAAAAAAAAA"), MeasurementSecret: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")},
			wantErr: true,
		},
		"incorrect size of measurement secret": {
			testAPI: &KeyAPI{keyReceived: make(chan struct{}, 1)},
			request: &keyproto.PushStateDiskKeyRequest{StateDiskKey: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), MeasurementSecret: []byte("BBBBBBBBBBBBBBBB")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			tc.testAPI.log = logger.NewTest(t)
			_, err := tc.testAPI.PushStateDiskKey(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.request.StateDiskKey, tc.testAPI.key)
			}
		})
	}
}

func TestResetKey(t *testing.T) {
	api := New(logger.NewTest(t), nil, nil, time.Second, time.Millisecond)

	api.key = []byte{0x1, 0x2, 0x3}
	api.ResetKey()
	assert.Nil(t, api.key)
}

type stubMetadataServer struct {
	listAnswerC chan listAnswer
}

func newStubMetadataServer() *stubMetadataServer {
	return &stubMetadataServer{
		listAnswerC: make(chan listAnswer),
	}
}

func (s *stubMetadataServer) List(context.Context) ([]metadata.InstanceMetadata, error) {
	answer := <-s.listAnswerC
	return answer.listResponse, answer.err
}

type listAnswer struct {
	listResponse []metadata.InstanceMetadata
	err          error
}

type stubJoinAPIServer struct {
	issueRejoinTicketAnswerC chan issueRejoinTicketAnswer
	joinproto.UnimplementedAPIServer
}

func newStubJoinAPIServer() *stubJoinAPIServer {
	return &stubJoinAPIServer{
		issueRejoinTicketAnswerC: make(chan issueRejoinTicketAnswer),
	}
}

func (s *stubJoinAPIServer) IssueRejoinTicket(context.Context, *joinproto.IssueRejoinTicketRequest) (*joinproto.IssueRejoinTicketResponse, error) {
	answer := <-s.issueRejoinTicketAnswerC
	resp := &joinproto.IssueRejoinTicketResponse{
		StateDiskKey:      answer.stateDiskKey,
		MeasurementSecret: answer.measurementSecret,
	}
	return resp, answer.err
}

type issueRejoinTicketAnswer struct {
	stateDiskKey      []byte
	measurementSecret []byte
	err               error
}

type stubKeyAPIServer struct {
	pushStateDiskKeyAnswerC chan pushStateDiskKeyAnswer
	keyproto.UnimplementedAPIServer
}

func newStubKeyAPIServer() *stubKeyAPIServer {
	return &stubKeyAPIServer{
		pushStateDiskKeyAnswerC: make(chan pushStateDiskKeyAnswer),
	}
}

func (s *stubKeyAPIServer) PushStateDiskKey(context.Context, *keyproto.PushStateDiskKeyRequest) (*keyproto.PushStateDiskKeyResponse, error) {
	answer := <-s.pushStateDiskKeyAnswerC
	return &keyproto.PushStateDiskKeyResponse{}, answer.err
}

type pushStateDiskKeyAnswer struct {
	err error
}

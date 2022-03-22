package deploy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/debugd/coordinator"
	"github.com/edgelesssys/constellation/debugd/debugd"
	pb "github.com/edgelesssys/constellation/debugd/service"
	"github.com/edgelesssys/constellation/debugd/service/testdialer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func TestDownloadCoordinator(t *testing.T) {
	filename := "/opt/coordinator"

	testCases := map[string]struct {
		server              fakeOnlyDownloadServer
		downloadClient      stubDownloadClient
		serviceManager      stubServiceManager
		attemptedDownloads  map[string]time.Time
		expectedChunks      [][]byte
		expectDownloadErr   bool
		expectFile          bool
		expectSystemdAction bool
		expectDeployed      bool
	}{
		"download works": {
			server: fakeOnlyDownloadServer{
				chunks: [][]byte{[]byte("test")},
			},
			attemptedDownloads: map[string]time.Time{},
			expectedChunks: [][]byte{
				[]byte("test"),
			},
			expectDownloadErr:   false,
			expectFile:          true,
			expectSystemdAction: true,
			expectDeployed:      true,
		},
		"second download is not attempted twice": {
			server: fakeOnlyDownloadServer{
				chunks: [][]byte{[]byte("test")},
			},
			attemptedDownloads: map[string]time.Time{
				"192.0.2.0:4000": time.Now(),
			},
			expectDownloadErr:   false,
			expectFile:          false,
			expectSystemdAction: false,
			expectDeployed:      false,
		},
		"download rpc call error is detected": {
			server: fakeOnlyDownloadServer{
				downladErr: errors.New("download rpc error"),
			},
			attemptedDownloads: map[string]time.Time{},
			expectDownloadErr:  true,
		},
		"service restart error is detected": {
			server: fakeOnlyDownloadServer{
				chunks: [][]byte{[]byte("test")},
			},
			serviceManager: stubServiceManager{
				systemdActionErr: errors.New("systemd error"),
			},
			attemptedDownloads: map[string]time.Time{},
			expectedChunks: [][]byte{
				[]byte("test"),
			},
			expectDownloadErr:   true,
			expectFile:          true,
			expectDeployed:      true,
			expectSystemdAction: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ip := "192.0.2.0"
			writer := &fakeStreamToFileWriter{}
			dialer := testdialer.NewBufconnDialer()

			grpcServ := grpc.NewServer()
			pb.RegisterDebugdServer(grpcServ, &tc.server)
			lis := dialer.GetListener(net.JoinHostPort(ip, debugd.DebugdPort))
			go grpcServ.Serve(lis)

			download := &Download{
				dialer:             dialer,
				writer:             writer,
				serviceManager:     &tc.serviceManager,
				attemptedDownloads: tc.attemptedDownloads,
			}
			err := download.DownloadCoordinator(context.Background(), ip)
			grpcServ.GracefulStop()

			if tc.expectDownloadErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			if tc.expectFile {
				assert.Equal(tc.expectedChunks, writer.chunks)
				assert.Equal(filename, writer.filename)
			}
			if tc.expectSystemdAction {
				assert.ElementsMatch(
					[]ServiceManagerRequest{
						{Unit: debugd.CoordinatorSystemdUnitName, Action: Restart},
					},
					tc.serviceManager.requests,
				)
			}
		})
	}
}

type stubDownloadClient struct {
	requests    []*pb.DownloadCoordinatorRequest
	stream      coordinator.ReadChunkStream
	downloadErr error
}

func (s *stubDownloadClient) DownloadCoordinator(ctx context.Context, in *pb.DownloadCoordinatorRequest, opts ...grpc.CallOption) (coordinator.ReadChunkStream, error) {
	s.requests = append(s.requests, proto.Clone(in).(*pb.DownloadCoordinatorRequest))
	return s.stream, s.downloadErr
}

type stubServiceManager struct {
	requests         []ServiceManagerRequest
	systemdActionErr error
}

func (s *stubServiceManager) SystemdAction(ctx context.Context, request ServiceManagerRequest) error {
	s.requests = append(s.requests, request)
	return s.systemdActionErr
}

type fakeStreamToFileWriter struct {
	chunks   [][]byte
	filename string
}

func (f *fakeStreamToFileWriter) WriteStream(filename string, stream coordinator.ReadChunkStream, showProgress bool) error {
	f.filename = filename
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("reading stream failed: %w", err)
		}
		f.chunks = append(f.chunks, chunk.Content)
	}
}

// fakeOnlyDownloadServer implements DebugdServer; only fakes DownloadCoordinator, panics on every other rpc.
type fakeOnlyDownloadServer struct {
	chunks     [][]byte
	downladErr error
	pb.UnimplementedDebugdServer
}

func (f *fakeOnlyDownloadServer) DownloadCoordinator(request *pb.DownloadCoordinatorRequest, stream pb.Debugd_DownloadCoordinatorServer) error {
	for _, chunk := range f.chunks {
		if err := stream.Send(&pb.Chunk{Content: chunk}); err != nil {
			return fmt.Errorf("sending chunk failed: %w", err)
		}
	}
	return f.downladErr
}

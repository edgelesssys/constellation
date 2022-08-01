package deploy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/debugd/bootstrapper"
	"github.com/edgelesssys/constellation/debugd/debugd"
	pb "github.com/edgelesssys/constellation/debugd/service"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestDownloadBootstrapper(t *testing.T) {
	filename := "/opt/bootstrapper"

	testCases := map[string]struct {
		server             fakeOnlyDownloadServer
		downloadClient     stubDownloadClient
		serviceManager     stubServiceManager
		attemptedDownloads map[string]time.Time
		wantChunks         [][]byte
		wantDownloadErr    bool
		wantFile           bool
		wantSystemdAction  bool
		wantDeployed       bool
	}{
		"download works": {
			server: fakeOnlyDownloadServer{
				chunks: [][]byte{[]byte("test")},
			},
			attemptedDownloads: map[string]time.Time{},
			wantChunks: [][]byte{
				[]byte("test"),
			},
			wantDownloadErr:   false,
			wantFile:          true,
			wantSystemdAction: true,
			wantDeployed:      true,
		},
		"second download is not attempted twice": {
			server: fakeOnlyDownloadServer{
				chunks: [][]byte{[]byte("test")},
			},
			attemptedDownloads: map[string]time.Time{
				"192.0.2.0:" + strconv.Itoa(constants.DebugdPort): time.Now(),
			},
			wantDownloadErr: true,
		},
		"download rpc call error is detected": {
			server: fakeOnlyDownloadServer{
				downladErr: errors.New("download rpc error"),
			},
			attemptedDownloads: map[string]time.Time{},
			wantDownloadErr:    true,
		},
		"service restart error is detected": {
			server: fakeOnlyDownloadServer{
				chunks: [][]byte{[]byte("test")},
			},
			serviceManager: stubServiceManager{
				systemdActionErr: errors.New("systemd error"),
			},
			attemptedDownloads: map[string]time.Time{},
			wantChunks: [][]byte{
				[]byte("test"),
			},
			wantDownloadErr:   true,
			wantFile:          true,
			wantDeployed:      true,
			wantSystemdAction: false,
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
			lis := dialer.GetListener(net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort)))
			go grpcServ.Serve(lis)

			download := &Download{
				log:                logger.NewTest(t),
				dialer:             dialer,
				writer:             writer,
				serviceManager:     &tc.serviceManager,
				attemptedDownloads: tc.attemptedDownloads,
			}
			err := download.DownloadBootstrapper(context.Background(), ip)
			grpcServ.GracefulStop()

			if tc.wantDownloadErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			if tc.wantFile {
				assert.Equal(tc.wantChunks, writer.chunks)
				assert.Equal(filename, writer.filename)
			}
			if tc.wantSystemdAction {
				assert.ElementsMatch(
					[]ServiceManagerRequest{
						{Unit: debugd.BootstrapperSystemdUnitName, Action: Restart},
					},
					tc.serviceManager.requests,
				)
			}
		})
	}
}

type stubDownloadClient struct {
	requests    []*pb.DownloadBootstrapperRequest
	stream      bootstrapper.ReadChunkStream
	downloadErr error
}

func (s *stubDownloadClient) DownloadBootstrapper(ctx context.Context, in *pb.DownloadBootstrapperRequest, opts ...grpc.CallOption) (bootstrapper.ReadChunkStream, error) {
	s.requests = append(s.requests, proto.Clone(in).(*pb.DownloadBootstrapperRequest))
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

func (f *fakeStreamToFileWriter) WriteStream(filename string, stream bootstrapper.ReadChunkStream, showProgress bool) error {
	f.filename = filename
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("reading stream: %w", err)
		}
		f.chunks = append(f.chunks, chunk.Content)
	}
}

// fakeOnlyDownloadServer implements DebugdServer; only fakes DownloadBootstrapper, panics on every other rpc.
type fakeOnlyDownloadServer struct {
	chunks     [][]byte
	downladErr error
	pb.UnimplementedDebugdServer
}

func (f *fakeOnlyDownloadServer) DownloadBootstrapper(request *pb.DownloadBootstrapperRequest, stream pb.Debugd_DownloadBootstrapperServer) error {
	for _, chunk := range f.chunks {
		if err := stream.Send(&pb.Chunk{Content: chunk}); err != nil {
			return fmt.Errorf("sending chunk: %w", err)
		}
	}
	return f.downladErr
}

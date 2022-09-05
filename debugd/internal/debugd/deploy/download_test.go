/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

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

	"github.com/edgelesssys/constellation/debugd/internal/bootstrapper"
	"github.com/edgelesssys/constellation/debugd/internal/debugd"
	pb "github.com/edgelesssys/constellation/debugd/service"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestDownloadBootstrapper(t *testing.T) {
	filename := "/opt/bootstrapper"
	someErr := errors.New("failed")

	testCases := map[string]struct {
		server             fakeDownloadServer
		serviceManager     stubServiceManager
		attemptedDownloads map[string]time.Time
		wantChunks         [][]byte
		wantDownloadErr    bool
		wantFile           bool
		wantSystemdAction  bool
		wantDeployed       bool
		wantKeys           []ssh.UserKey
	}{
		"download works": {
			server: fakeDownloadServer{
				chunks: [][]byte{[]byte("test")},
				keys:   []*pb.AuthorizedKey{{Username: "name", KeyValue: "key"}},
			},
			attemptedDownloads: map[string]time.Time{},
			wantChunks:         [][]byte{[]byte("test")},
			wantDownloadErr:    false,
			wantFile:           true,
			wantSystemdAction:  true,
			wantDeployed:       true,
			wantKeys:           []ssh.UserKey{{Username: "name", PublicKey: "key"}},
		},
		"second download is not attempted twice": {
			server:             fakeDownloadServer{chunks: [][]byte{[]byte("test")}},
			attemptedDownloads: map[string]time.Time{"192.0.2.0:" + strconv.Itoa(constants.DebugdPort): time.Now()},
			wantDownloadErr:    true,
		},
		"download rpc call error is detected": {
			server:             fakeDownloadServer{downladErr: someErr},
			attemptedDownloads: map[string]time.Time{},
			wantDownloadErr:    true,
		},
		"download key error": {
			server: fakeDownloadServer{
				chunks:                    [][]byte{[]byte("test")},
				downloadAuthorizedKeysErr: someErr,
			},
			attemptedDownloads: map[string]time.Time{},
			wantDownloadErr:    true,
		},
		"service restart error is detected": {
			server:             fakeDownloadServer{chunks: [][]byte{[]byte("test")}},
			serviceManager:     stubServiceManager{systemdActionErr: someErr},
			attemptedDownloads: map[string]time.Time{},
			wantChunks:         [][]byte{[]byte("test")},
			wantDownloadErr:    true,
			wantFile:           true,
			wantDeployed:       true,
			wantSystemdAction:  false,
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
			defer grpcServ.GracefulStop()

			download := &Download{
				log:                logger.NewTest(t),
				dialer:             dialer,
				writer:             writer,
				serviceManager:     &tc.serviceManager,
				attemptedDownloads: tc.attemptedDownloads,
			}

			keys, err := download.DownloadDeployment(context.Background(), ip)

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
			assert.Equal(tc.wantKeys, keys)
		})
	}
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

// fakeDownloadServer implements DebugdServer; only fakes DownloadBootstrapper, panics on every other rpc.
type fakeDownloadServer struct {
	chunks                    [][]byte
	downladErr                error
	keys                      []*pb.AuthorizedKey
	downloadAuthorizedKeysErr error

	pb.UnimplementedDebugdServer
}

func (f *fakeDownloadServer) DownloadBootstrapper(request *pb.DownloadBootstrapperRequest, stream pb.Debugd_DownloadBootstrapperServer) error {
	for _, chunk := range f.chunks {
		if err := stream.Send(&pb.Chunk{Content: chunk}); err != nil {
			return fmt.Errorf("sending chunk: %w", err)
		}
	}
	return f.downladErr
}

func (s *fakeDownloadServer) DownloadAuthorizedKeys(context.Context, *pb.DownloadAuthorizedKeysRequest) (*pb.DownloadAuthorizedKeysResponse, error) {
	return &pb.DownloadAuthorizedKeysResponse{Keys: s.keys}, s.downloadAuthorizedKeysErr
}

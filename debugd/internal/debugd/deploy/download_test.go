/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"testing"

	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
	)
}

func TestDownloadDeployment(t *testing.T) {
	testCases := map[string]struct {
		files                  []filetransfer.FileStat
		recvFilesErr           error
		overrideServiceUnitErr error
		wantErr                bool
		wantOverrideCalls      []struct{ UnitName, ExecStart string }
	}{
		"download works": {
			files: []filetransfer.FileStat{
				{
					SourcePath:          "source/testfileA",
					TargetPath:          "target/testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "unitA",
				},
				{
					SourcePath: "source/testfileB",
					TargetPath: "target/testfileB",
					Mode:       0o644,
				},
			},
			wantOverrideCalls: []struct{ UnitName, ExecStart string }{
				{"unitA", "target/testfileA"},
			},
		},
		"recv files error is detected": {
			recvFilesErr: errors.New("some error"),
			wantErr:      true,
		},
		"recv already running": {
			recvFilesErr: filetransfer.ErrReceiveRunning,
			wantErr:      true,
		},
		"recv already finished": {
			files: []filetransfer.FileStat{
				{
					SourcePath:          "source/testfileA",
					TargetPath:          "target/testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "unitA",
				},
			},
			recvFilesErr: filetransfer.ErrReceiveFinished,
			wantErr:      false,
		},
		"service unit fail does not stop further tries": {
			files: []filetransfer.FileStat{
				{
					SourcePath:          "source/testfileA",
					TargetPath:          "target/testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "unitA",
				},
				{
					SourcePath:          "source/testfileB",
					TargetPath:          "target/testfileB",
					Mode:                0o644,
					OverrideServiceUnit: "unitB",
				},
			},
			overrideServiceUnitErr: errors.New("some error"),
			wantOverrideCalls: []struct{ UnitName, ExecStart string }{
				{"unitA", "target/testfileA"},
				{"unitB", "target/testfileB"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ip := "192.0.2.0"
			transfer := &stubTransfer{recvFilesErr: tc.recvFilesErr, files: tc.files}
			serviceMgr := &stubServiceManager{overrideServiceUnitExecStartErr: tc.overrideServiceUnitErr}
			dialer := testdialer.NewBufconnDialer()

			server := &stubDownloadServer{}
			grpcServ := grpc.NewServer()
			pb.RegisterDebugdServer(grpcServ, server)
			lis := dialer.GetListener(net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort)))
			go grpcServ.Serve(lis)
			defer grpcServ.GracefulStop()

			download := &Download{
        log:            slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				dialer:         dialer,
				transfer:       transfer,
				serviceManager: serviceMgr,
			}

			err := download.DownloadDeployment(context.Background(), ip)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			assert.Equal(tc.wantOverrideCalls, serviceMgr.overrideCalls)
		})
	}
}

func TestDownloadInfo(t *testing.T) {
	someErr := errors.New("failed")
	someInfo := []*pb.Info{
		{Key: "foo", Value: "bar"},
		{Key: "baz", Value: "qux"},
	}

	testCases := map[string]struct {
		server     stubDebugdServer
		infoSetter stubInfoSetter
		wantErr    bool
		wantInfo   []*pb.Info
	}{
		"download works": {
			server:     stubDebugdServer{info: someInfo},
			infoSetter: stubInfoSetter{},
			wantInfo:   someInfo,
		},
		"empty info ok": {
			server:     stubDebugdServer{info: []*pb.Info{}},
			infoSetter: stubInfoSetter{},
			wantInfo:   nil,
		},
		"nil info ok": {
			server:     stubDebugdServer{},
			infoSetter: stubInfoSetter{},
			wantInfo:   nil,
		},
		"getInfo fails": {
			server:     stubDebugdServer{getInfoErr: someErr},
			infoSetter: stubInfoSetter{},
			wantErr:    true,
		},
		"setInfo fails": {
			server:     stubDebugdServer{info: someInfo},
			infoSetter: stubInfoSetter{setProtoErr: someErr},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ip := "192.0.2.1"
			dialer := testdialer.NewBufconnDialer()
			grpcServer := grpc.NewServer()
			pb.RegisterDebugdServer(grpcServer, &tc.server)
			lis := dialer.GetListener(net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort)))
			go grpcServer.Serve(lis)
			defer grpcServer.GracefulStop()

			download := &Download{
        log:    slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				dialer: dialer,
				info:   &tc.infoSetter,
			}

			err := download.DownloadInfo(context.Background(), ip)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(len(tc.wantInfo), len(tc.infoSetter.info))
			}
		})
	}
}

type stubServiceManager struct {
	requests         []ServiceManagerRequest
	systemdActionErr error

	overrideCalls                   []struct{ UnitName, ExecStart string }
	overrideServiceUnitExecStartErr error
}

func (s *stubServiceManager) SystemdAction(_ context.Context, request ServiceManagerRequest) error {
	s.requests = append(s.requests, request)
	return s.systemdActionErr
}

func (s *stubServiceManager) OverrideServiceUnitExecStart(_ context.Context, unitName string, execStart string) error {
	s.overrideCalls = append(s.overrideCalls, struct {
		UnitName, ExecStart string
	}{UnitName: unitName, ExecStart: execStart})
	return s.overrideServiceUnitExecStartErr
}

type stubTransfer struct {
	recvFilesErr error
	files        []filetransfer.FileStat
}

func (t *stubTransfer) RecvFiles(_ filetransfer.RecvFilesStream) error {
	return t.recvFilesErr
}

func (t *stubTransfer) GetFiles() []filetransfer.FileStat {
	return t.files
}

// stubDownloadServer implements DebugdServer; only stubs DownloadFiles, panics on every other rpc.
type stubDownloadServer struct {
	downladErr error

	pb.UnimplementedDebugdServer
}

func (s *stubDownloadServer) DownloadFiles(_ *pb.DownloadFilesRequest, _ pb.Debugd_DownloadFilesServer) error {
	return s.downladErr
}

type stubDebugdServer struct {
	info       []*pb.Info
	getInfoErr error
	pb.UnimplementedDebugdServer
}

func (s *stubDebugdServer) GetInfo(_ context.Context, _ *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
	return &pb.GetInfoResponse{Info: s.info}, s.getInfoErr
}

type stubInfoSetter struct {
	info        []*pb.Info
	received    bool
	setProtoErr error
}

func (s *stubInfoSetter) SetProto(infos []*pb.Info) error {
	s.info = infos
	return s.setProtoErr
}

func (s *stubInfoSetter) Received() bool {
	return s.received
}

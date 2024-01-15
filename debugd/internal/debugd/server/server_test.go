/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"strconv"
	"testing"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/deploy"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/info"
	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestSetInfo(t *testing.T) {
	endpoint := "192.0.2.1:" + strconv.Itoa(constants.DebugdPort)

	testCases := map[string]struct {
		info         *info.Map
		infoReceived bool
		setInfo      []*pb.Info
		wantStatus   pb.SetInfoStatus
	}{
		"set info works": {
			setInfo:    []*pb.Info{{Key: "foo", Value: "bar"}},
			info:       info.NewMap(),
			wantStatus: pb.SetInfoStatus_SET_INFO_SUCCESS,
		},
		"set empty info works": {
			setInfo:    []*pb.Info{},
			info:       info.NewMap(),
			wantStatus: pb.SetInfoStatus_SET_INFO_SUCCESS,
		},
		"set fails when info already set": {
			info:         info.NewMap(),
			infoReceived: true,
			setInfo:      []*pb.Info{{Key: "foo", Value: "bar"}},
			wantStatus:   pb.SetInfoStatus_SET_INFO_ALREADY_SET,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serv := debugdServer{
				log:  slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				info: tc.info,
			}

			if tc.infoReceived {
				err := tc.info.SetProto(tc.setInfo)
				require.NoError(err)
			}

			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)

			setInfoStatus, err := client.SetInfo(context.Background(), &pb.SetInfoRequest{Info: tc.setInfo})
			grpcServ.GracefulStop()

			assert.NoError(err)
			assert.Equal(tc.wantStatus, setInfoStatus.Status)
			for i := range tc.setInfo {
				value, ok, err := tc.info.Get(tc.setInfo[i].Key)
				assert.NoError(err)
				assert.True(ok)
				assert.Equal(tc.setInfo[i].Value, value)
			}
		})
	}
}

func TestGetInfo(t *testing.T) {
	endpoint := "192.0.2.1:" + strconv.Itoa(constants.DebugdPort)

	testCases := map[string]struct {
		info    *info.Map
		getInfo []*pb.Info
		wantErr bool
	}{
		"get info works": {
			getInfo: []*pb.Info{{Key: "foo", Value: "bar"}},
			info:    info.NewMap(),
		},
		"get empty info works": {
			getInfo: []*pb.Info{},
			info:    info.NewMap(),
		},
		"get unset info fails": {
			getInfo: nil,
			info:    info.NewMap(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			if tc.getInfo != nil {
				err := tc.info.SetProto(tc.getInfo)
				require.NoError(err)
			}

			serv := debugdServer{
				log:  slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				info: tc.info,
			}

			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)

			resp, err := client.GetInfo(context.Background(), &pb.GetInfoRequest{})
			grpcServ.GracefulStop()

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(len(tc.getInfo), len(resp.Info))
			}
		})
	}
}

func TestUploadFiles(t *testing.T) {
	endpoint := "192.0.2.1:" + strconv.Itoa(constants.DebugdPort)

	testCases := map[string]struct {
		files              []filetransfer.FileStat
		recvFilesErr       error
		wantResponseStatus pb.UploadFilesStatus
		wantOverrideCalls  []struct{ UnitName, ExecStart string }
	}{
		"upload works": {
			files: []filetransfer.FileStat{
				{SourcePath: "source/testA", TargetPath: "target/testA", Mode: 0o644, OverrideServiceUnit: "testA"},
				{SourcePath: "source/testB", TargetPath: "target/testB", Mode: 0o644},
			},
			wantOverrideCalls: []struct{ UnitName, ExecStart string }{
				{"testA", "target/testA"},
			},
			wantResponseStatus: pb.UploadFilesStatus_UPLOAD_FILES_SUCCESS,
		},
		"recv fails": {
			recvFilesErr:       errors.New("recv error"),
			wantResponseStatus: pb.UploadFilesStatus_UPLOAD_FILES_UPLOAD_FAILED,
		},
		"upload in progress": {
			recvFilesErr:       filetransfer.ErrReceiveRunning,
			wantResponseStatus: pb.UploadFilesStatus_UPLOAD_FILES_ALREADY_STARTED,
		},
		"upload already finished": {
			recvFilesErr:       filetransfer.ErrReceiveFinished,
			wantResponseStatus: pb.UploadFilesStatus_UPLOAD_FILES_ALREADY_FINISHED,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serviceMgr := &stubServiceManager{}
			transfer := &stubTransfer{files: tc.files, recvFilesErr: tc.recvFilesErr}

			serv := debugdServer{
				log:            slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				serviceManager: serviceMgr,
				transfer:       transfer,
			}

			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)
			stream, err := client.UploadFiles(context.Background())
			require.NoError(err)
			resp, err := stream.CloseAndRecv()

			grpcServ.GracefulStop()

			require.NoError(err)
			assert.Equal(tc.wantResponseStatus, resp.Status)
			assert.Equal(tc.wantOverrideCalls, serviceMgr.overrideCalls)
		})
	}
}

func TestDownloadFiles(t *testing.T) {
	endpoint := "192.0.2.1:" + strconv.Itoa(constants.DebugdPort)

	testCases := map[string]struct {
		request           *pb.DownloadFilesRequest
		canSend           bool
		wantRecvErr       bool
		wantSendFileCalls int
	}{
		"download works": {
			request:           &pb.DownloadFilesRequest{},
			canSend:           true,
			wantSendFileCalls: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			transfer := &stubTransfer{canSend: tc.canSend}
			serv := debugdServer{
				log:      slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				transfer: transfer,
			}

			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)
			stream, err := client.DownloadFiles(context.Background(), tc.request)
			require.NoError(err)
			_, recvErr := stream.Recv()
			if tc.wantRecvErr {
				require.Error(recvErr)
			} else {
				require.ErrorIs(recvErr, io.EOF)
			}
			require.NoError(stream.CloseSend())
			grpcServ.GracefulStop()
			require.NoError(err)

			assert.Equal(tc.wantSendFileCalls, transfer.sendFilesCount)
		})
	}
}

func TestUploadSystemServiceUnits(t *testing.T) {
	endpoint := "192.0.2.1:" + strconv.Itoa(constants.DebugdPort)

	testCases := map[string]struct {
		serviceManager     stubServiceManager
		request            *pb.UploadSystemdServiceUnitsRequest
		wantErr            bool
		wantResponseStatus pb.UploadSystemdServiceUnitsStatus
		wantUnitFiles      []deploy.SystemdUnit
	}{
		"upload systemd service units": {
			request: &pb.UploadSystemdServiceUnitsRequest{
				Units: []*pb.ServiceUnit{
					{
						Name:     "test.service",
						Contents: "testcontents",
					},
				},
			},
			wantResponseStatus: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS,
			wantUnitFiles: []deploy.SystemdUnit{
				{
					Name:     "test.service",
					Contents: "testcontents",
				},
			},
		},
		"writing fails": {
			request: &pb.UploadSystemdServiceUnitsRequest{
				Units: []*pb.ServiceUnit{
					{
						Name:     "test.service",
						Contents: "testcontents",
					},
				},
			},
			serviceManager: stubServiceManager{
				writeSystemdUnitFileErr: errors.New("write error"),
			},
			wantResponseStatus: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_FAILURE,
			wantUnitFiles: []deploy.SystemdUnit{
				{
					Name:     "test.service",
					Contents: "testcontents",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serv := debugdServer{
				log:            slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				serviceManager: &tc.serviceManager,
			}
			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)
			resp, err := client.UploadSystemServiceUnits(context.Background(), tc.request)

			grpcServ.GracefulStop()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			require.NotNil(resp.Status)
			assert.Equal(tc.wantResponseStatus, resp.Status)
			assert.ElementsMatch(tc.wantUnitFiles, tc.serviceManager.unitFiles)
		})
	}
}

type stubServiceManager struct {
	requests      []deploy.ServiceManagerRequest
	unitFiles     []deploy.SystemdUnit
	overrideCalls []struct{ UnitName, ExecStart string }

	systemdActionErr                error
	writeSystemdUnitFileErr         error
	overrideServiceUnitExecStartErr error
}

func (s *stubServiceManager) SystemdAction(_ context.Context, request deploy.ServiceManagerRequest) error {
	s.requests = append(s.requests, request)
	return s.systemdActionErr
}

func (s *stubServiceManager) WriteSystemdUnitFile(_ context.Context, unit deploy.SystemdUnit) error {
	s.unitFiles = append(s.unitFiles, unit)
	return s.writeSystemdUnitFileErr
}

func (s *stubServiceManager) OverrideServiceUnitExecStart(_ context.Context, unitName string, execStart string) error {
	s.overrideCalls = append(s.overrideCalls, struct {
		UnitName, ExecStart string
	}{UnitName: unitName, ExecStart: execStart})
	return s.overrideServiceUnitExecStartErr
}

type netDialer interface {
	DialContext(_ context.Context, network, address string) (net.Conn, error)
}

func dial(ctx context.Context, dialer netDialer, target string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target,
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp", addr)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

type stubTransfer struct {
	recvFilesCount int
	sendFilesCount int
	files          []filetransfer.FileStat
	canSend        bool
	recvFilesErr   error
	sendFilesErr   error
}

func (t *stubTransfer) RecvFiles(_ filetransfer.RecvFilesStream) error {
	t.recvFilesCount++
	return t.recvFilesErr
}

func (t *stubTransfer) SendFiles(_ filetransfer.SendFilesStream) error {
	t.sendFilesCount++
	return t.sendFilesErr
}

func (t *stubTransfer) GetFiles() []filetransfer.FileStat {
	return t.files
}

func (t *stubTransfer) CanSend() bool {
	return t.canSend
}

func setupServerWithConn(endpoint string, serv *debugdServer) (*grpc.Server, *grpc.ClientConn, error) {
	dialer := testdialer.NewBufconnDialer()
	grpcServ := grpc.NewServer()
	pb.RegisterDebugdServer(grpcServ, serv)
	lis := dialer.GetListener(endpoint)
	go grpcServ.Serve(lis)

	conn, err := dial(context.Background(), dialer, endpoint)
	if err != nil {
		return nil, nil, err
	}

	return grpcServ, conn, nil
}

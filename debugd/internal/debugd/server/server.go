/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package server implements the gRPC endpoint of Constellation's debugd.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/deploy"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/info"
	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type debugdServer struct {
	log            *slog.Logger
	serviceManager serviceManager
	transfer       fileTransferer
	info           *info.Map

	pb.UnimplementedDebugdServer
}

// New creates a new debugdServer according to the gRPC spec.
func New(log *slog.Logger, serviceManager serviceManager, transfer fileTransferer, infos *info.Map) pb.DebugdServer {
	return &debugdServer{
		log:            log,
		serviceManager: serviceManager,
		transfer:       transfer,
		info:           infos,
	}
}

// SetInfo sets the info of the debugd instance.
func (s *debugdServer) SetInfo(_ context.Context, req *pb.SetInfoRequest) (*pb.SetInfoResponse, error) {
	s.log.Info("Received SetInfo request")

	if len(req.Info) == 0 {
		s.log.Info("Info is empty")
	}

	setProtoErr := s.info.SetProto(req.Info)
	if errors.Is(setProtoErr, info.ErrInfoAlreadySet) {
		s.log.Warn("Setting info failed (already set)")
		return &pb.SetInfoResponse{
			Status: pb.SetInfoStatus_SET_INFO_ALREADY_SET,
		}, nil
	}

	if setProtoErr != nil {
		s.log.With(slog.Any("error", setProtoErr)).Error("Setting info failed")
		return nil, setProtoErr
	}
	s.log.Info("Info set")

	return &pb.SetInfoResponse{
		Status: pb.SetInfoStatus_SET_INFO_SUCCESS,
	}, nil
}

// GetInfo returns the info of the debugd instance.
func (s *debugdServer) GetInfo(_ context.Context, _ *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
	s.log.Info("Received GetInfo request")

	info, err := s.info.GetProto()
	if err != nil {
		return nil, err
	}

	return &pb.GetInfoResponse{Info: info}, nil
}

// UploadFiles receives a stream of files (each consisting of a header and a stream of chunks) and writes them to the filesystem.
func (s *debugdServer) UploadFiles(stream pb.Debugd_UploadFilesServer) error {
	s.log.Info("Received UploadFiles request")
	err := s.transfer.RecvFiles(stream)
	switch {
	case err == nil:
		s.log.Info("Uploading files succeeded")
	case errors.Is(err, filetransfer.ErrReceiveRunning):
		s.log.Warn("Upload already in progress")
		return stream.SendAndClose(&pb.UploadFilesResponse{
			Status: pb.UploadFilesStatus_UPLOAD_FILES_ALREADY_STARTED,
		})
	case errors.Is(err, filetransfer.ErrReceiveFinished):
		s.log.Warn("Upload already finished")
		return stream.SendAndClose(&pb.UploadFilesResponse{
			Status: pb.UploadFilesStatus_UPLOAD_FILES_ALREADY_FINISHED,
		})
	default:
		s.log.With(slog.Any("error", err)).Error("Uploading files failed")
		return stream.SendAndClose(&pb.UploadFilesResponse{
			Status: pb.UploadFilesStatus_UPLOAD_FILES_UPLOAD_FAILED,
			Error:  err.Error(),
		})
	}

	files := s.transfer.GetFiles()
	var overrideUnitErr error
	for _, file := range files {
		if file.OverrideServiceUnit == "" {
			continue
		}
		// continue on error to allow other units to be overridden
		err = s.serviceManager.OverrideServiceUnitExecStart(stream.Context(), file.OverrideServiceUnit, file.TargetPath)
		overrideUnitErr = errors.Join(overrideUnitErr, err)
	}

	if overrideUnitErr != nil {
		s.log.With(slog.Any("error", overrideUnitErr)).Error("Overriding service units failed")
		return stream.SendAndClose(&pb.UploadFilesResponse{
			Status: pb.UploadFilesStatus_UPLOAD_FILES_START_FAILED,
			Error:  overrideUnitErr.Error(),
		})
	}
	return stream.SendAndClose(&pb.UploadFilesResponse{
		Status: pb.UploadFilesStatus_UPLOAD_FILES_SUCCESS,
	})
}

// DownloadFiles streams the previously received files to other instances.
func (s *debugdServer) DownloadFiles(_ *pb.DownloadFilesRequest, stream pb.Debugd_DownloadFilesServer) error {
	s.log.Info("Sending files to other instance")
	return s.transfer.SendFiles(stream)
}

// UploadSystemServiceUnits receives systemd service units, writes them to a service file and schedules a daemon-reload.
func (s *debugdServer) UploadSystemServiceUnits(ctx context.Context, in *pb.UploadSystemdServiceUnitsRequest) (*pb.UploadSystemdServiceUnitsResponse, error) {
	s.log.Info("Uploading systemd service units")
	for _, unit := range in.Units {
		if err := s.serviceManager.WriteSystemdUnitFile(ctx, deploy.SystemdUnit{Name: unit.Name, Contents: unit.Contents}); err != nil {
			return &pb.UploadSystemdServiceUnitsResponse{Status: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_FAILURE}, nil
		}
	}

	return &pb.UploadSystemdServiceUnitsResponse{Status: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS}, nil
}

// Start will start the gRPC server as goroutine.
func Start(log *slog.Logger, wg *sync.WaitGroup, serv pb.DebugdServer) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		grpcLog := logger.GRPCLogger(log)
		logger.ReplaceGRPCLogger(grpcLog)

		grpcServer := grpc.NewServer(
			logger.GetServerStreamInterceptor(grpcLog),
			logger.GetServerUnaryInterceptor(grpcLog),
			grpc.KeepaliveParams(keepalive.ServerParameters{Time: 15 * time.Second}),
		)
		pb.RegisterDebugdServer(grpcServer, serv)
		lis, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", strconv.Itoa(constants.DebugdPort)))
		if err != nil {
			log.With(slog.Any("error", err)).Error("Listening failed")
			os.Exit(1)
		}
		log.Info("gRPC server is waiting for connections")
		grpcServer.Serve(lis)
	}()
}

type serviceManager interface {
	SystemdAction(ctx context.Context, request deploy.ServiceManagerRequest) error
	WriteSystemdUnitFile(ctx context.Context, unit deploy.SystemdUnit) error
	OverrideServiceUnitExecStart(ctx context.Context, unitName string, execStart string) error
}

type fileTransferer interface {
	RecvFiles(stream filetransfer.RecvFilesStream) error
	SendFiles(stream filetransfer.SendFilesStream) error
	GetFiles() []filetransfer.FileStat
}

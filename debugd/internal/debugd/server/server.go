/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/debugd/internal/bootstrapper"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/deploy"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type debugdServer struct {
	log            *logger.Logger
	ssh            sshDeployer
	serviceManager serviceManager
	streamer       streamer
	pb.UnimplementedDebugdServer
}

// New creates a new debugdServer according to the gRPC spec.
func New(log *logger.Logger, ssh sshDeployer, serviceManager serviceManager, streamer streamer) pb.DebugdServer {
	return &debugdServer{
		log:            log,
		ssh:            ssh,
		serviceManager: serviceManager,
		streamer:       streamer,
	}
}

// TODO (stateless-ssh): re-enable once ssh keys can be deployed on readonly rootfs.
// UploadAuthorizedKeys receives a list of authorized keys and forwards them to a channel.
//
//	func (s *debugdServer) UploadAuthorizedKeys(ctx context.Context, in *pb.UploadAuthorizedKeysRequest) (*pb.UploadAuthorizedKeysResponse, error) {
//		s.log.Infof("Uploading authorized keys")
//		for _, key := range in.Keys {
//			if err := s.ssh.DeployAuthorizedKey(ctx, ssh.UserKey{Username: key.Username, PublicKey: key.KeyValue}); err != nil {
//				s.log.With(zap.Error(err)).Errorf("Uploading authorized keys failed")
//				return &pb.UploadAuthorizedKeysResponse{
//					Status: pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_FAILURE,
//				}, nil
//			}
//		}
//		return &pb.UploadAuthorizedKeysResponse{
//			Status: pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_SUCCESS,
//		}, nil
//	}
//
// UploadAuthorizedKeys receives a list of authorized keys and forwards them to a channel.
func (s *debugdServer) UploadAuthorizedKeys(ctx context.Context, in *pb.UploadAuthorizedKeysRequest) (*pb.UploadAuthorizedKeysResponse, error) {
	s.log.Infof("Uploading authorized keys (Disabled feature)")
	return &pb.UploadAuthorizedKeysResponse{
		Status: pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_SUCCESS,
	}, nil
}

// UploadBootstrapper receives a bootstrapper binary in a stream of chunks and writes to a file.
func (s *debugdServer) UploadBootstrapper(stream pb.Debugd_UploadBootstrapperServer) error {
	startAction := deploy.ServiceManagerRequest{
		Unit:   debugd.BootstrapperSystemdUnitName,
		Action: deploy.Start,
	}
	var responseStatus pb.UploadBootstrapperStatus
	defer func() {
		if err := s.serviceManager.SystemdAction(stream.Context(), startAction); err != nil {
			s.log.With(zap.Error(err)).Errorf("Starting uploaded bootstrapper failed")
			if responseStatus == pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_SUCCESS {
				responseStatus = pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_START_FAILED
			}
		}
		stream.SendAndClose(&pb.UploadBootstrapperResponse{
			Status: responseStatus,
		})
	}()
	s.log.Infof("Starting bootstrapper upload")
	if err := s.streamer.WriteStream(debugd.BootstrapperDeployFilename, stream, true); err != nil {
		if errors.Is(err, fs.ErrExist) {
			// bootstrapper was already uploaded
			s.log.Warnf("Bootstrapper already uploaded")
			responseStatus = pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_FILE_EXISTS
			return nil
		}
		s.log.With(zap.Error(err)).Errorf("Uploading bootstrapper failed")
		responseStatus = pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_UPLOAD_FAILED
		return fmt.Errorf("uploading bootstrapper: %w", err)
	}

	s.log.Infof("Successfully uploaded bootstrapper")
	responseStatus = pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_SUCCESS
	return nil
}

// DownloadBootstrapper streams the local bootstrapper binary to other instances.
func (s *debugdServer) DownloadBootstrapper(request *pb.DownloadBootstrapperRequest, stream pb.Debugd_DownloadBootstrapperServer) error {
	s.log.Infof("Sending bootstrapper to other instance")
	return s.streamer.ReadStream(debugd.BootstrapperDeployFilename, stream, debugd.Chunksize, true)
}

// DownloadAuthorizedKeys streams the local authorized keys to other instances.
func (s *debugdServer) DownloadAuthorizedKeys(_ context.Context, req *pb.DownloadAuthorizedKeysRequest) (*pb.DownloadAuthorizedKeysResponse, error) {
	s.log.Infof("Sending authorized keys to other instance")

	var authKeys []*pb.AuthorizedKey
	for _, key := range s.ssh.GetAuthorizedKeys() {
		authKeys = append(authKeys, &pb.AuthorizedKey{
			Username: key.Username,
			KeyValue: key.PublicKey,
		})
	}

	return &pb.DownloadAuthorizedKeysResponse{Keys: authKeys}, nil
}

// UploadSystemServiceUnits receives systemd service units, writes them to a service file and schedules a daemon-reload.
func (s *debugdServer) UploadSystemServiceUnits(ctx context.Context, in *pb.UploadSystemdServiceUnitsRequest) (*pb.UploadSystemdServiceUnitsResponse, error) {
	s.log.Infof("Uploading systemd service units")
	for _, unit := range in.Units {
		if err := s.serviceManager.WriteSystemdUnitFile(ctx, deploy.SystemdUnit{Name: unit.Name, Contents: unit.Contents}); err != nil {
			return &pb.UploadSystemdServiceUnitsResponse{Status: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_FAILURE}, nil
		}
	}

	return &pb.UploadSystemdServiceUnitsResponse{Status: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS}, nil
}

// Start will start the gRPC server and block.
func Start(log *logger.Logger, wg *sync.WaitGroup, serv pb.DebugdServer) {
	defer wg.Done()

	grpcLog := log.Named("gRPC")
	grpcLog.WithIncreasedLevel(zap.WarnLevel).ReplaceGRPCLogger()

	grpcServer := grpc.NewServer(
		grpcLog.GetServerStreamInterceptor(),
		grpcLog.GetServerUnaryInterceptor(),
		grpc.KeepaliveParams(keepalive.ServerParameters{Time: 15 * time.Second}),
	)
	pb.RegisterDebugdServer(grpcServer, serv)
	lis, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", strconv.Itoa(constants.DebugdPort)))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Listening failed")
	}
	log.Infof("gRPC server is waiting for connections")
	grpcServer.Serve(lis)
}

type sshDeployer interface {
	DeployAuthorizedKey(ctx context.Context, sshKey ssh.UserKey) error
	GetAuthorizedKeys() []ssh.UserKey
}

type serviceManager interface {
	SystemdAction(ctx context.Context, request deploy.ServiceManagerRequest) error
	WriteSystemdUnitFile(ctx context.Context, unit deploy.SystemdUnit) error
}

type streamer interface {
	WriteStream(filename string, stream bootstrapper.ReadChunkStream, showProgress bool) error
	ReadStream(filename string, stream bootstrapper.WriteChunkStream, chunksize uint, showProgress bool) error
}

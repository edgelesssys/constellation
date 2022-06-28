package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"sync"

	"github.com/edgelesssys/constellation/debugd/coordinator"
	"github.com/edgelesssys/constellation/debugd/debugd"
	"github.com/edgelesssys/constellation/debugd/debugd/deploy"
	pb "github.com/edgelesssys/constellation/debugd/service"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

// UploadAuthorizedKeys receives a list of authorized keys and forwards them to a channel.
func (s *debugdServer) UploadAuthorizedKeys(ctx context.Context, in *pb.UploadAuthorizedKeysRequest) (*pb.UploadAuthorizedKeysResponse, error) {
	s.log.Infof("Uploading authorized keys")
	for _, key := range in.Keys {
		if err := s.ssh.DeployAuthorizedKey(ctx, ssh.UserKey{Username: key.Username, PublicKey: key.KeyValue}); err != nil {
			s.log.With(zap.Error(err)).Errorf("Uploading authorized keys failed")
			return &pb.UploadAuthorizedKeysResponse{
				Status: pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_FAILURE,
			}, nil
		}
	}
	return &pb.UploadAuthorizedKeysResponse{
		Status: pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_SUCCESS,
	}, nil
}

// UploadCoordinator receives a coordinator binary in a stream of chunks and writes to a file.
func (s *debugdServer) UploadCoordinator(stream pb.Debugd_UploadCoordinatorServer) error {
	startAction := deploy.ServiceManagerRequest{
		Unit:   debugd.CoordinatorSystemdUnitName,
		Action: deploy.Start,
	}
	var responseStatus pb.UploadCoordinatorStatus
	defer func() {
		if err := s.serviceManager.SystemdAction(stream.Context(), startAction); err != nil {
			s.log.With(zap.Error(err)).Errorf("Starting uploaded coordinator failed")
			if responseStatus == pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_SUCCESS {
				responseStatus = pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_START_FAILED
			}
		}
		stream.SendAndClose(&pb.UploadCoordinatorResponse{
			Status: responseStatus,
		})
	}()
	s.log.Infof("Starting coordinator upload")
	if err := s.streamer.WriteStream(debugd.CoordinatorDeployFilename, stream, true); err != nil {
		if errors.Is(err, fs.ErrExist) {
			// coordinator was already uploaded
			s.log.Warnf("Coordinator already uploaded")
			responseStatus = pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_FILE_EXISTS
			return nil
		}
		s.log.With(zap.Error(err)).Errorf("Uploading coordinator failed")
		responseStatus = pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_UPLOAD_FAILED
		return fmt.Errorf("uploading coordinator: %w", err)
	}

	s.log.Infof("Successfully uploaded coordinator")
	responseStatus = pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_SUCCESS
	return nil
}

// DownloadCoordinator streams the local coordinator binary to other instances.
func (s *debugdServer) DownloadCoordinator(request *pb.DownloadCoordinatorRequest, stream pb.Debugd_DownloadCoordinatorServer) error {
	s.log.Infof("Sending coordinator to other instance")
	return s.streamer.ReadStream(debugd.CoordinatorDeployFilename, stream, debugd.Chunksize, true)
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

	grpcServer := grpc.NewServer(grpcLog.GetServerStreamInterceptor(), grpcLog.GetServerUnaryInterceptor())
	pb.RegisterDebugdServer(grpcServer, serv)
	lis, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", debugd.DebugdPort))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Listening failed")
	}
	log.Infof("gRPC server is waiting for connections")
	grpcServer.Serve(lis)
}

type sshDeployer interface {
	DeployAuthorizedKey(ctx context.Context, sshKey ssh.UserKey) error
}

type serviceManager interface {
	SystemdAction(ctx context.Context, request deploy.ServiceManagerRequest) error
	WriteSystemdUnitFile(ctx context.Context, unit deploy.SystemdUnit) error
}

type streamer interface {
	WriteStream(filename string, stream coordinator.ReadChunkStream, showProgress bool) error
	ReadStream(filename string, stream coordinator.WriteChunkStream, chunksize uint, showProgress bool) error
}

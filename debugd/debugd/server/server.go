package server

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/edgelesssys/constellation/debugd/coordinator"
	"github.com/edgelesssys/constellation/debugd/debugd"
	"github.com/edgelesssys/constellation/debugd/debugd/deploy"
	pb "github.com/edgelesssys/constellation/debugd/service"
	"github.com/edgelesssys/constellation/debugd/ssh"
	"google.golang.org/grpc"
)

type debugdServer struct {
	ssh            sshDeployer
	serviceManager serviceManager
	streamer       streamer
	pb.UnimplementedDebugdServer
}

// New creates a new debugdServer according to the gRPC spec.
func New(ssh sshDeployer, serviceManager serviceManager, streamer streamer) pb.DebugdServer {
	return &debugdServer{
		ssh:            ssh,
		serviceManager: serviceManager,
		streamer:       streamer,
	}
}

// UploadAuthorizedKeys receives a list of authorized keys and forwards them to a channel.
func (s *debugdServer) UploadAuthorizedKeys(ctx context.Context, in *pb.UploadAuthorizedKeysRequest) (*pb.UploadAuthorizedKeysResponse, error) {
	log.Println("Uploading authorized keys")
	for _, key := range in.Keys {
		if err := s.ssh.DeploySSHAuthorizedKey(ctx, ssh.SSHKey{Username: key.Username, KeyValue: key.KeyValue}); err != nil {
			log.Printf("Uploading authorized keys failed: %v\n", err)
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
	// try to stop the coordinator before upload but ignore failure since there might be no coordinator running yet.
	_ = s.serviceManager.SystemdAction(stream.Context(), deploy.ServiceManagerRequest{
		Unit:   debugd.CoordinatorSystemdUnitName,
		Action: deploy.Stop,
	})
	log.Println("Starting coordinator upload")
	if err := s.streamer.WriteStream(debugd.CoordinatorDeployFilename, stream, true); err != nil {
		log.Printf("Uploading coordinator failed: %v\n", err)
		return stream.SendAndClose(&pb.UploadCoordinatorResponse{
			Status: pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_UPLOAD_FAILED,
		})
	}

	log.Println("Successfully uploaded coordinator")

	// after the upload succeeds, try to restart the coordinator
	restartAction := deploy.ServiceManagerRequest{
		Unit:   debugd.CoordinatorSystemdUnitName,
		Action: deploy.Restart,
	}
	if err := s.serviceManager.SystemdAction(stream.Context(), restartAction); err != nil {
		log.Printf("Starting uploaded coordinator failed: %v\n", err)
		return stream.SendAndClose(&pb.UploadCoordinatorResponse{
			Status: pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_START_FAILED,
		})
	}

	return stream.SendAndClose(&pb.UploadCoordinatorResponse{
		Status: pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_SUCCESS,
	})
}

// DownloadCoordinator streams the local coordinator binary to other instances.
func (s *debugdServer) DownloadCoordinator(request *pb.DownloadCoordinatorRequest, stream pb.Debugd_DownloadCoordinatorServer) error {
	log.Println("Sending coordinator to other instance")
	return s.streamer.ReadStream(debugd.CoordinatorDeployFilename, stream, debugd.Chunksize, true)
}

// UploadSystemServiceUnits receives systemd service units, writes them to a service file and schedules a daemon-reload.
func (s *debugdServer) UploadSystemServiceUnits(ctx context.Context, in *pb.UploadSystemdServiceUnitsRequest) (*pb.UploadSystemdServiceUnitsResponse, error) {
	log.Println("Uploading systemd service units")
	for _, unit := range in.Units {
		if err := s.serviceManager.WriteSystemdUnitFile(ctx, deploy.SystemdUnit{Name: unit.Name, Contents: unit.Contents}); err != nil {
			return &pb.UploadSystemdServiceUnitsResponse{Status: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_FAILURE}, nil
		}
	}

	return &pb.UploadSystemdServiceUnitsResponse{Status: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS}, nil
}

// Start will start the gRPC server and block.
func Start(wg *sync.WaitGroup, serv pb.DebugdServer) {
	defer wg.Done()
	grpcServer := grpc.NewServer()
	pb.RegisterDebugdServer(grpcServer, serv)
	lis, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", debugd.DebugdPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("gRPC server is waiting for connections")
	grpcServer.Serve(lis)
}

type sshDeployer interface {
	DeploySSHAuthorizedKey(ctx context.Context, sshKey ssh.SSHKey) error
}

type serviceManager interface {
	SystemdAction(ctx context.Context, request deploy.ServiceManagerRequest) error
	WriteSystemdUnitFile(ctx context.Context, unit deploy.SystemdUnit) error
}

type streamer interface {
	WriteStream(filename string, stream coordinator.ReadChunkStream, showProgress bool) error
	ReadStream(filename string, stream coordinator.WriteChunkStream, chunksize uint, showProgress bool) error
}

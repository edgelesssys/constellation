package recoveryserver

import (
	"context"
	"net"
	"sync"

	"github.com/edgelesssys/constellation/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryServer is a gRPC server that can be used by an admin to recover a restarting node.
type RecoveryServer struct {
	mux sync.Mutex

	diskUUID          string
	stateDiskKey      []byte
	measurementSecret []byte
	grpcServer        server

	log *logger.Logger

	recoverproto.UnimplementedAPIServer
}

// New returns a new RecoveryServer.
func New(issuer atls.Issuer, log *logger.Logger) *RecoveryServer {
	server := &RecoveryServer{
		log: log,
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(atlscredentials.New(issuer, nil)),
		log.Named("gRPC").GetServerStreamInterceptor(),
	)
	recoverproto.RegisterAPIServer(grpcServer, server)

	server.grpcServer = grpcServer
	return server
}

// Serve starts the recovery server.
// It blocks until a recover request call is successful.
// The server will shut down when the call is successful and the keys are returned.
// Additionally, the server can be shutdown by canceling the context.
func (s *RecoveryServer) Serve(ctx context.Context, listener net.Listener, diskUUID string) (diskKey, measurementSecret []byte, err error) {
	s.log.Infof("Starting RecoveryServer")
	s.diskUUID = diskUUID
	recoveryDone := make(chan struct{}, 1)
	var serveErr error

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		serveErr = s.grpcServer.Serve(listener)
		recoveryDone <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("Context canceled, shutting down server")
			s.grpcServer.GracefulStop()
			return nil, nil, ctx.Err()
		case <-recoveryDone:
			if serveErr != nil {
				return nil, nil, serveErr
			}
			return s.stateDiskKey, s.measurementSecret, nil
		}
	}
}

// Recover is a bidirectional streaming RPC that is used to send recovery keys to a restarting node.
func (s *RecoveryServer) Recover(stream recoverproto.API_RecoverServer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.log.Infof("Received recover call")

	msg, err := stream.Recv()
	if err != nil {
		return status.Error(codes.Internal, "failed to receive message")
	}

	measurementSecret, ok := msg.GetRequest().(*recoverproto.RecoverMessage_MeasurementSecret)
	if !ok {
		s.log.Errorf("Received invalid first message: not a measurement secret")
		return status.Error(codes.InvalidArgument, "first message is not a measurement secret")
	}

	if err := stream.Send(&recoverproto.RecoverResponse{DiskUuid: s.diskUUID}); err != nil {
		s.log.With(zap.Error(err)).Errorf("Failed to send disk UUID")
		return status.Error(codes.Internal, "failed to send response")
	}

	msg, err = stream.Recv()
	if err != nil {
		s.log.With(zap.Error(err)).Errorf("Failed to receive disk key")
		return status.Error(codes.Internal, "failed to receive message")
	}

	stateDiskKey, ok := msg.GetRequest().(*recoverproto.RecoverMessage_StateDiskKey)
	if !ok {
		s.log.Errorf("Received invalid second message: not a state disk key")
		return status.Error(codes.InvalidArgument, "second message is not a state disk key")
	}

	s.stateDiskKey = stateDiskKey.StateDiskKey
	s.measurementSecret = measurementSecret.MeasurementSecret
	s.log.Infof("Received state disk key and measurement secret, shutting down server")

	go s.grpcServer.GracefulStop()
	return nil
}

type server interface {
	Serve(net.Listener) error
	GracefulStop()
}

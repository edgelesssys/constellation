// Package server implements an API to manage encryption keys.
package server

import (
	"context"
	"net"
	"sync"

	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/kms/kms"
	"github.com/edgelesssys/constellation/kms/kmsproto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements an encryption key management server.
// The server serves aTLS for cluster external requests
// and plain gRPC for cluster internal requests.
type Server struct {
	log    *logger.Logger
	conKMS kms.CloudKMS
	kmsproto.UnimplementedAPIServer
}

// New creates a new Server.
func New(log *logger.Logger, conKMS kms.CloudKMS) *Server {
	return &Server{
		log:    log,
		conKMS: conKMS,
	}
}

// Run starts both the plain gRPC server and the aTLS gRPC server.
// If one of the servers fails, the other server will be closed and the error will be returned.
func (s *Server) Run(atlsListener, plainListener net.Listener, credentials *atlscredentials.Credentials) error {
	var err error
	var once sync.Once
	var wg sync.WaitGroup

	atlsServer := grpc.NewServer(
		grpc.Creds(credentials),
		s.log.Named("gRPC.aTLS").GetServerUnaryInterceptor(),
	)
	kmsproto.RegisterAPIServer(atlsServer, s)

	plainServer := grpc.NewServer(s.log.Named("gRPC.cluster").GetServerUnaryInterceptor())
	kmsproto.RegisterAPIServer(plainServer, s)

	s.log.Named("gRPC").WithIncreasedLevel(zapcore.WarnLevel).ReplaceGRPCLogger()

	// start the plain gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer atlsServer.GracefulStop()

		s.log.Infof("Starting Constellation key management service on %s", plainListener.Addr().String())
		plainErr := plainServer.Serve(plainListener)
		if plainErr != nil {
			once.Do(func() { err = plainErr })
		}
	}()

	// start the aTLS server
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer plainServer.GracefulStop()

		s.log.Infof("Starting Constellation aTLS key management service on %s", atlsListener.Addr().String())
		atlsErr := atlsServer.Serve(atlsListener)
		if atlsErr != nil {
			once.Do(func() { err = atlsErr })
		}
	}()

	wg.Wait()
	return err
}

// GetDataKey returns a data key.
func (s *Server) GetDataKey(ctx context.Context, in *kmsproto.GetDataKeyRequest) (*kmsproto.GetDataKeyResponse, error) {
	log := s.log.With("peerAddress", grpclog.PeerAddrFromContext(ctx))

	// Error on 0 key length
	if in.Length == 0 {
		log.Errorf("Requested key length is zero")
		return nil, status.Error(codes.InvalidArgument, "can't derive key with length zero")
	}

	// Error on empty DataKeyId
	if in.DataKeyId == "" {
		log.Errorf("No data key ID specified")
		return nil, status.Error(codes.InvalidArgument, "no data key ID specified")
	}

	key, err := s.conKMS.GetDEK(ctx, "Constellation", "key-"+in.DataKeyId, int(in.Length))
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to get data key")
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &kmsproto.GetDataKeyResponse{DataKey: key}, nil
}

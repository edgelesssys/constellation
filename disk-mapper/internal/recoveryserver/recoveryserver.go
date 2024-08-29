/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package recoveryserver implements the gRPC endpoints for recovering a restarting node.

The endpoint is only available for control-plane nodes,
worker nodes should only rejoin the cluster using Constellation's JoinService.

This endpoint can be used by an admin in case of a complete cluster shutdown,
in which case a node is unable to rejoin the cluster automatically.
*/
package recoveryserver

import (
	"context"
	"log/slog"
	"net"
	"sync"

	"github.com/edgelesssys/constellation/v2/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type kmsFactory func(ctx context.Context, storageURI string, kmsURI string) (kms.CloudKMS, error)

// RecoveryServer is a gRPC server that can be used by an admin to recover a restarting node.
type RecoveryServer struct {
	mux sync.Mutex

	diskUUID          string
	stateDiskKey      []byte
	measurementSecret []byte
	grpcServer        server
	factory           kmsFactory

	log *slog.Logger

	recoverproto.UnimplementedAPIServer
}

// New returns a new RecoveryServer.
func New(issuer atls.Issuer, factory kmsFactory, log *slog.Logger) *RecoveryServer {
	server := &RecoveryServer{
		log:     log,
		factory: factory,
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(atlscredentials.New(issuer, nil)),
		logger.GetServerStreamInterceptor(logger.GRPCLogger(log)),
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
	s.log.Info("Starting RecoveryServer")
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
			s.log.Info("Context canceled, shutting down server")
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
func (s *RecoveryServer) Recover(ctx context.Context, req *recoverproto.RecoverMessage) (*recoverproto.RecoverResponse, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	log := s.log.With(slog.String("peer", grpclog.PeerAddrFromContext(ctx)))

	log.Info("Received recover call")

	cloudKms, err := s.factory(ctx, req.StorageUri, req.KmsUri)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating kms client: %s", err)
	}

	measurementSecret, err := cloudKms.GetDEK(ctx, crypto.DEKPrefix+crypto.MeasurementSecretKeyID, crypto.DerivedKeyLengthDefault)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "requesting measurementSecret: %s", err)
	}
	stateDiskKey, err := cloudKms.GetDEK(ctx, crypto.DEKPrefix+s.diskUUID, crypto.StateDiskKeyLength)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "requesting stateDiskKey: %s", err)
	}
	s.stateDiskKey = stateDiskKey
	s.measurementSecret = measurementSecret
	log.Info("Received state disk key and measurement secret, shutting down server")

	go s.grpcServer.GracefulStop()
	return &recoverproto.RecoverResponse{}, nil
}

// StubServer implements the RecoveryServer interface but does not actually start a server.
type StubServer struct {
	log *slog.Logger
}

// NewStub returns a new stubbed RecoveryServer.
// We use this to avoid having to start a server for worker nodes, since they don't require manual recovery.
func NewStub(log *slog.Logger) *StubServer {
	return &StubServer{log: log}
}

// Serve waits until the context is canceled and returns nil.
func (s *StubServer) Serve(ctx context.Context, _ net.Listener, _ string) ([]byte, []byte, error) {
	s.log.Info("Running as worker node, skipping recovery server")
	<-ctx.Done()
	return nil, nil, ctx.Err()
}

type server interface {
	Serve(net.Listener) error
	GracefulStop()
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package server implements an API to manage encryption keys.
package server

import (
	"context"
	"fmt"
	"net"

	"github.com/edgelesssys/constellation/internal/crypto"
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

// Run starts the gRPC server.
func (s *Server) Run(port string) error {
	// set up listener
	listener, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", port, err)
	}

	server := grpc.NewServer(s.log.Named("gRPC").GetServerUnaryInterceptor())
	kmsproto.RegisterAPIServer(server, s)
	s.log.Named("gRPC").WithIncreasedLevel(zapcore.WarnLevel).ReplaceGRPCLogger()

	// start the server
	s.log.Infof("Starting Constellation key management service on %s", listener.Addr().String())
	return server.Serve(listener)
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

	key, err := s.conKMS.GetDEK(ctx, "Constellation", crypto.HKDFInfoPrefix+in.DataKeyId, int(in.Length))
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to get data key")
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &kmsproto.GetDataKeyResponse{DataKey: key}, nil
}

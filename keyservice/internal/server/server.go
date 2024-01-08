/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package server implements an API to manage encryption keys.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/keyservice/keyserviceproto"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements an encryption key management server.
// The server serves aTLS for cluster external requests
// and plain gRPC for cluster internal requests.
type Server struct {
	log    *slog.Logger
	conKMS kms.CloudKMS
	keyserviceproto.UnimplementedAPIServer
}

// New creates a new Server.
func New(log *slog.Logger, conKMS kms.CloudKMS) *Server {
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

	server := grpc.NewServer(logger.GetServerUnaryInterceptor(s.log.WithGroup("gRPC")))
	keyserviceproto.RegisterAPIServer(server, s)
  // TODO(miampf): Find out a good way to pass an increased Level to slog.
  // A reference implementation for something like that exists
  // [here](https://pkg.go.dev/log/slog#Handler), however, this would
  // utilise structs in the logger package again which is not optimal.
	logger.ReplaceGRPCLogger(s.log.WithGroup("gRPC").WithIncreasedLevel(zapcore.WarnLevel))

	// start the server
	s.log.Info("Starting Constellation key management service on %s", listener.Addr().String())
	return server.Serve(listener)
}

// GetDataKey returns a data key.
func (s *Server) GetDataKey(ctx context.Context, in *keyserviceproto.GetDataKeyRequest) (*keyserviceproto.GetDataKeyResponse, error) {
	log := s.log.With("peerAddress", grpclog.PeerAddrFromContext(ctx))

	// Error on 0 key length
	if in.Length == 0 {
		log.Error("Requested key length is zero")
		return nil, status.Error(codes.InvalidArgument, "can't derive key with length zero")
	}

	// Error on empty DataKeyId
	if in.DataKeyId == "" {
		log.Error("No data key ID specified")
		return nil, status.Error(codes.InvalidArgument, "no data key ID specified")
	}

	key, err := s.conKMS.GetDEK(ctx, crypto.DEKPrefix+in.DataKeyId, int(in.Length))
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to get data key")
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &keyserviceproto.GetDataKeyResponse{DataKey: key}, nil
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package server implements the gRPC and REST endpoints for retrieving attestation statements.
package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/verify/verifyproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type attestation struct {
	Data []byte `json:"data"`
}

// Server implements Constellation's verify API.
// The server exposes both HTTP and gRPC endpoints
// to retrieve attestation statements.
type Server struct {
	log    *slog.Logger
	issuer AttestationIssuer
	verifyproto.UnimplementedAPIServer
}

// New initializes a new verification server.
func New(log *slog.Logger, issuer AttestationIssuer) *Server {
	return &Server{
		log:    log,
		issuer: issuer,
	}
}

// Run starts the HTTP and gRPC servers.
// If one of the servers fails, other server will be closed and the error will be returned.
func (s *Server) Run(httpListener, grpcListener net.Listener) error {
	var err error
	var wg sync.WaitGroup
	var once sync.Once

	grpcLog := logger.GRPCLogger(s.log)
	logger.ReplaceGRPCLogger(grpcLog)
	grpcServer := grpc.NewServer(
		logger.GetServerUnaryInterceptor(grpcLog),
		grpc.KeepaliveParams(keepalive.ServerParameters{Time: 15 * time.Second}),
	)
	verifyproto.RegisterAPIServer(grpcServer, s)

	httpHandler := http.NewServeMux()
	httpHandler.HandleFunc("/", s.getAttestationHTTP)
	httpServer := &http.Server{Handler: httpHandler}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer grpcServer.GracefulStop()

		s.log.Info(fmt.Sprintf("Starting HTTP server on %s", httpListener.Addr().String()))
		httpErr := httpServer.Serve(httpListener)
		if httpErr != nil && httpErr != http.ErrServerClosed {
			once.Do(func() { err = httpErr })
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { _ = httpServer.Shutdown(context.Background()) }()

		s.log.Info(fmt.Sprintf("Starting gRPC server on %s", grpcListener.Addr().String()))
		grpcErr := grpcServer.Serve(grpcListener)
		if grpcErr != nil {
			once.Do(func() { err = grpcErr })
		}
	}()

	wg.Wait()
	return err
}

// GetAttestation implements the gRPC endpoint for requesting attestation statements.
func (s *Server) GetAttestation(ctx context.Context, req *verifyproto.GetAttestationRequest) (*verifyproto.GetAttestationResponse, error) {
	peerAddr := "unknown"
	if peer, ok := peer.FromContext(ctx); ok {
		peerAddr = peer.Addr.String()
	}

	log := s.log.With(slog.String("peerAddress", peerAddr)).WithGroup("gRPC")
	s.log.Info("Received attestation request")
	if len(req.Nonce) == 0 {
		log.Error("Received attestation request with empty nonce")
		return nil, status.Error(codes.InvalidArgument, "nonce is required to issue attestation")
	}

	log.Info("Creating attestation")
	statement, err := s.issuer.Issue(ctx, []byte(constants.ConstellationVerifyServiceUserData), req.Nonce)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issuing attestation statement: %v", err)
	}

	log.Info("Attestation request successful")
	return &verifyproto.GetAttestationResponse{Attestation: statement}, nil
}

// getAttestationHTTP implements the HTTP endpoint for retrieving attestation statements.
func (s *Server) getAttestationHTTP(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(slog.String("peerAddress", r.RemoteAddr)).WithGroup("http")

	nonceB64 := r.URL.Query()["nonce"]
	if len(nonceB64) != 1 || nonceB64[0] == "" {
		log.Error("Received attestation request with empty or multiple nonce parameter")
		http.Error(w, "nonce parameter is required exactly once", http.StatusBadRequest)
		return
	}

	nonce, err := base64.URLEncoding.DecodeString(nonceB64[0])
	if err != nil {
		log.With(slog.Any("error", err)).Error("Received attestation request with invalid nonce")
		http.Error(w, fmt.Sprintf("invalid base64 encoding for nonce: %v", err), http.StatusBadRequest)
		return
	}

	log.Info("Creating attestation")
	quote, err := s.issuer.Issue(r.Context(), []byte(constants.ConstellationVerifyServiceUserData), nonce)
	if err != nil {
		http.Error(w, fmt.Sprintf("issuing attestation statement: %v", err), http.StatusInternalServerError)
		return
	}

	log.Info("Attestation request successful")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(attestation{quote}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AttestationIssuer issues an attestation document for the provided userData and nonce.
type AttestationIssuer interface {
	Issue(ctx context.Context, userData []byte, nonce []byte) (quote []byte, err error)
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
)

// Server provides measurements.
type Server struct {
	log          *logger.Logger
	server       http.Server
	measurements map[uint32][]byte
	done         chan<- struct{}
}

// New creates a new Server.
func New(log *logger.Logger, done chan<- struct{}) *Server {
	return &Server{
		log:  log,
		done: done,
	}
}

// ListenAndServe on given port.
func (s *Server) ListenAndServe(port string) error {
	mux := http.NewServeMux()
	mux.Handle("/pcrs", http.HandlerFunc(s.logPCRs))

	s.server = http.Server{
		Handler: mux,
	}

	lis, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		return err
	}
	s.log.Infof("Starting QEMU metadata API on %s", lis.Addr())
	return s.server.Serve(lis)
}

// Shutdown server.
func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

// logPCRs allows QEMU instances to export their TPM state during boot.
func (s *Server) logPCRs(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(zap.String("peer", r.RemoteAddr))
	if r.Method != http.MethodPost {
		log.With(zap.String("method", r.Method)).Errorf("Invalid method for /log")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Infof("Serving POST request for /pcrs")

	if r.Body == nil {
		log.Errorf("Request body is empty")
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	// unmarshal the request body into a map of PCRs
	var pcrs map[uint32][]byte
	if err := json.NewDecoder(r.Body).Decode(&pcrs); err != nil {
		log.With(zap.Error(err)).Errorf("Failed to read request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("PCR 4 %x", pcrs[4])
	log.Infof("PCR 8 %x", pcrs[8])
	log.Infof("PCR 9 %x", pcrs[9])

	s.measurements = pcrs

	s.done <- struct{}{}
}

// GetMeasurements returns the static measurements for QEMU environment.
func (s *Server) GetMeasurements() map[uint32][]byte {
	return s.measurements
}

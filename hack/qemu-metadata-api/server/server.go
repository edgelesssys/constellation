/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/virtwrapper"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

// Server that provides QEMU metadata.
type Server struct {
	log               *slog.Logger
	virt              virConnect
	network           string
	initSecretHashVal []byte
}

// New creates a new Server.
func New(log *slog.Logger, network, initSecretHash string, conn virConnect) *Server {
	return &Server{
		log:               log,
		virt:              conn,
		network:           network,
		initSecretHashVal: []byte(initSecretHash),
	}
}

// ListenAndServe on a given port.
func (s *Server) ListenAndServe(port string) error {
	mux := http.NewServeMux()
	mux.Handle("/self", http.HandlerFunc(s.listSelf))
	mux.Handle("/peers", http.HandlerFunc(s.listPeers))
	mux.Handle("/endpoint", http.HandlerFunc(s.getEndpoint))
	mux.Handle("/initsecrethash", http.HandlerFunc(s.initSecretHash))

	server := http.Server{
		Handler: mux,
	}

	lis, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		return err
	}

	s.log.Info("Starting QEMU metadata API on %s", lis.Addr())
	return server.Serve(lis)
}

// listSelf returns peer information about the instance issuing the request.
func (s *Server) listSelf(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(slog.String("peer", r.RemoteAddr))
	log.Info("Serving GET request for /self")

	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to parse remote address")
		http.Error(w, fmt.Sprintf("Failed to parse remote address: %s\n", err), http.StatusInternalServerError)
		return
	}

	peers, err := s.listAll()
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to list peer metadata")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, peer := range peers {
		if peer.VPCIP == remoteIP {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(peer); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Info("Request successful")
			return
		}
	}

	log.Error("Failed to find peer in active leases")
	http.Error(w, "No matching peer found", http.StatusNotFound)
}

// listPeers returns a list of all active peers.
func (s *Server) listPeers(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(slog.String("peer", r.RemoteAddr))
	log.Info("Serving GET request for /peers")

	peers, err := s.listAll()
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to list peer metadata")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(peers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("Request successful")
}

// initSecretHash returns the hash of the init secret.
func (s *Server) initSecretHash(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(slog.String("initSecretHash", r.RemoteAddr))
	if r.Method != http.MethodGet {
		log.With(slog.String("method", r.Method)).Error("Invalid method for /initSecretHash")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Info("Serving GET request for /initsecrethash")

	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write(s.initSecretHashVal)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to write init secret hash")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("Request successful")
}

// getEndpoint returns the IP address of the first control-plane instance.
// This allows us to fake a load balancer for QEMU instances.
func (s *Server) getEndpoint(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(slog.String("peer", r.RemoteAddr))
	log.Info("Serving GET request for /endpoint")

	net, err := s.virt.LookupNetworkByName(s.network)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to lookup network")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer net.Free()

	leases, err := net.GetDHCPLeases()
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to get DHCP leases")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	for _, lease := range leases {
		if strings.HasPrefix(lease.Hostname, "control-plane") &&
			strings.HasSuffix(lease.Hostname, "0") {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(lease.IPaddr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Info("Request successful")
			return
		}
	}

	log.Error("Failed to find control-plane peer in active leases")
	http.Error(w, "No matching peer found", http.StatusNotFound)
}

// listAll returns a list of all active peers.
func (s *Server) listAll() ([]metadata.InstanceMetadata, error) {
	net, err := s.virt.LookupNetworkByName(s.network)
	if err != nil {
		return nil, err
	}
	defer net.Free()

	leases, err := net.GetDHCPLeases()
	if err != nil {
		return nil, err
	}
	var peers []metadata.InstanceMetadata

	for _, lease := range leases {
		instanceRole := role.Worker
		if strings.HasPrefix(lease.Hostname, "control-plane") {
			instanceRole = role.ControlPlane
		}

		peers = append(peers, metadata.InstanceMetadata{
			Name:       lease.Hostname,
			Role:       instanceRole,
			VPCIP:      lease.IPaddr,
			ProviderID: "qemu:///hostname/" + lease.Hostname,
		})
	}

	return peers, nil
}

type virConnect interface {
	LookupNetworkByName(name string) (*virtwrapper.Network, error)
}

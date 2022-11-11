/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/virtwrapper"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"go.uber.org/zap"
)

// Server that provides QEMU metadata.
type Server struct {
	log     *logger.Logger
	virt    virConnect
	network string
}

// New creates a new Server.
func New(log *logger.Logger, network string, conn virConnect) *Server {
	return &Server{
		log:     log,
		virt:    conn,
		network: network,
	}
}

// ListenAndServe on a given port.
func (s *Server) ListenAndServe(port string) error {
	mux := http.NewServeMux()
	mux.Handle("/self", http.HandlerFunc(s.listSelf))
	mux.Handle("/peers", http.HandlerFunc(s.listPeers))
	mux.Handle("/log", http.HandlerFunc(s.postLog))
	mux.Handle("/pcrs", http.HandlerFunc(s.exportPCRs))
	mux.Handle("/endpoint", http.HandlerFunc(s.getEndpoint))

	server := http.Server{
		Handler: mux,
	}

	lis, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		return err
	}

	s.log.Infof("Starting QEMU metadata API on %s", lis.Addr())
	return server.Serve(lis)
}

// listSelf returns peer information about the instance issuing the request.
func (s *Server) listSelf(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(zap.String("peer", r.RemoteAddr))
	log.Infof("Serving GET request for /self")

	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to parse remote address")
		http.Error(w, fmt.Sprintf("Failed to parse remote address: %s\n", err), http.StatusInternalServerError)
		return
	}

	peers, err := s.listAll()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to list peer metadata")
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
			log.Infof("Request successful")
			return
		}
	}

	log.Errorf("Failed to find peer in active leases")
	http.Error(w, "No matching peer found", http.StatusNotFound)
}

// listPeers returns a list of all active peers.
func (s *Server) listPeers(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(zap.String("peer", r.RemoteAddr))
	log.Infof("Serving GET request for /peers")

	peers, err := s.listAll()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to list peer metadata")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(peers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof("Request successful")
}

// getEndpoint returns the IP address of the first control-plane instance.
// This allows us to fake a load balancer for QEMU instances.
func (s *Server) getEndpoint(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(zap.String("peer", r.RemoteAddr))
	log.Infof("Serving GET request for /endpoint")

	net, err := s.virt.LookupNetworkByName(s.network)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to lookup network")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer net.Free()

	leases, err := net.GetDHCPLeases()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to get DHCP leases")
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
			log.Infof("Request successful")
			return
		}
	}

	log.Errorf("Failed to find control-plane peer in active leases")
	http.Error(w, "No matching peer found", http.StatusNotFound)
}

// postLog writes implements cloud-logging for QEMU instances.
func (s *Server) postLog(w http.ResponseWriter, r *http.Request) {
	log := s.log.With(zap.String("peer", r.RemoteAddr))
	if r.Method != http.MethodPost {
		log.With(zap.String("method", r.Method)).Errorf("Invalid method for /log")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Infof("Serving POST request for /log")

	if r.Body == nil {
		log.Errorf("Request body is empty")
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	msg, err := io.ReadAll(r.Body)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to read request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.With(zap.String("message", string(msg))).Infof("Cloud-logging entry")
}

// exportPCRs allows QEMU instances to export their TPM state during boot.
// This can be used to check expected PCRs for GCP/Azure cloud images locally.
func (s *Server) exportPCRs(w http.ResponseWriter, r *http.Request) {
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
	var pcrs measurements.Measurements
	if err := json.NewDecoder(r.Body).Decode(&pcrs); err != nil {
		log.With(zap.Error(err)).Errorf("Failed to read request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get name of the node sending the export request
	var nodeName string
	peers, err := s.listAll()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to list peer metadata")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to parse remote address")
		http.Error(w, fmt.Sprintf("Failed to parse remote address: %s\n", err), http.StatusInternalServerError)
		return
	}
	for _, peer := range peers {
		if peer.VPCIP == remoteIP {
			nodeName = peer.Name
		}
	}

	log.With(zap.String("node", nodeName)).With(zap.Any("pcrs", pcrs)).Infof("Received PCRs from node")
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

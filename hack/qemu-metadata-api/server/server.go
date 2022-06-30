package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/hack/qemu-metadata-api/virtwrapper"
	"github.com/edgelesssys/constellation/internal/logger"
	"go.uber.org/zap"
)

type Server struct {
	log  *logger.Logger
	virt virConnect
}

func New(log *logger.Logger, conn virConnect) *Server {
	return &Server{
		log:  log,
		virt: conn,
	}
}

func (s *Server) ListenAndServe(port string) error {
	mux := http.NewServeMux()
	mux.Handle("/self", http.HandlerFunc(s.listSelf))
	mux.Handle("/peers", http.HandlerFunc(s.listPeers))

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
		for _, ip := range peer.PublicIPs {
			if ip == remoteIP {
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(peer); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				log.Infof("Request successful")
				return
			}
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

// listAll returns a list of all active peers.
func (s *Server) listAll() ([]cloudtypes.Instance, error) {
	net, err := s.virt.LookupNetworkByName("constellation")
	if err != nil {
		return nil, err
	}
	defer net.Free()

	leases, err := net.GetDHCPLeases()
	if err != nil {
		return nil, err
	}
	var peers []cloudtypes.Instance

	for _, lease := range leases {
		instanceRole := role.Node
		if strings.HasPrefix(lease.Hostname, "control-plane") {
			instanceRole = role.Coordinator
		}

		peers = append(peers, cloudtypes.Instance{
			Name:       lease.Hostname,
			Role:       instanceRole,
			PrivateIPs: []string{lease.IPaddr},
			PublicIPs:  []string{lease.IPaddr},
			ProviderID: "qemu:///hostname/" + lease.Hostname,
		})
	}

	return peers, nil
}

type virConnect interface {
	LookupNetworkByName(name string) (*virtwrapper.Network, error)
}

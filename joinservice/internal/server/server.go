package server

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"time"

	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/joinservice/joinproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// Server implements the core logic of Constellation's node join service.
type Server struct {
	log             *logger.Logger
	file            file.Handler
	joinTokenGetter joinTokenGetter
	dataKeyGetter   dataKeyGetter
	ca              certificateAuthority
	joinproto.UnimplementedAPIServer
}

// New initializes a new Server.
func New(log *logger.Logger, fileHandler file.Handler, ca certificateAuthority, joinTokenGetter joinTokenGetter, dataKeyGetter dataKeyGetter) *Server {
	return &Server{
		log:             log,
		file:            fileHandler,
		joinTokenGetter: joinTokenGetter,
		dataKeyGetter:   dataKeyGetter,
		ca:              ca,
	}
}

// Run starts the gRPC server on the given port, using the provided tlsConfig.
func (s *Server) Run(creds credentials.TransportCredentials, port string) error {
	s.log.WithIncreasedLevel(zap.WarnLevel).Named("gRPC").ReplaceGRPCLogger()
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		s.log.Named("gRPC").GetServerUnaryInterceptor(),
	)

	joinproto.RegisterAPIServer(grpcServer, s)

	lis, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	s.log.Infof("Starting join service on %s", lis.Addr().String())
	return grpcServer.Serve(lis)
}

// IssueJoinTicket handles join requests of Constellation nodes.
// A node will receive:
// - stateful disk encryption key.
// - Kubernetes join token.
// - cluster and owner ID to taint the node as initialized.
// In addition, control plane nodes receive:
// - a decryption key for CA certificates uploaded to the Kubernetes cluster.
func (s *Server) IssueJoinTicket(ctx context.Context, req *joinproto.IssueJoinTicketRequest) (resp *joinproto.IssueJoinTicketResponse, retErr error) {
	s.log.Infof("IssueJoinTicket called")
	log := s.log.With(zap.String("peerAddress", grpclog.PeerAddrFromContext(ctx)))
	log.Infof("Loading IDs")
	var id attestationtypes.ID
	if err := s.file.ReadJSON(filepath.Join(constants.ServiceBasePath, constants.IDFilename), &id); err != nil {
		log.With(zap.Error(err)).Errorf("Unable to load IDs")
		return nil, status.Errorf(codes.Internal, "unable to load IDs: %s", err)
	}

	log.Infof("Requesting disk encryption key")
	stateDiskKey, err := s.dataKeyGetter.GetDataKey(ctx, req.DiskUuid, constants.StateDiskKeyLength)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Unable to get key for stateful disk")
		return nil, status.Errorf(codes.Internal, "unable to get key for stateful disk: %s", err)
	}

	log.Infof("Creating Kubernetes join token")
	kubeArgs, err := s.joinTokenGetter.GetJoinToken(constants.KubernetesJoinTokenTTL)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Unable to generate Kubernetes join arguments")
		return nil, status.Errorf(codes.Internal, "unable to generate Kubernetes join arguments: %s", err)
	}

	log.Infof("Creating signed kubelet certificate")
	kubeletCert, kubeletKey, err := s.ca.GetCertificate(req.NodeName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to generate kubelet certificate: %s", err)
	}

	var controlPlaneFiles []*joinproto.ControlPlaneCertOrKey
	if req.IsControlPlane {
		log.Infof("Loading control plane certificates and keys")
		filesMap, err := s.joinTokenGetter.GetControlPlaneCertificatesAndKeys()
		if err != nil {
			log.With(zap.Error(err)).Errorf("Failed to load control plane certificates and keys")
			return nil, status.Errorf(codes.Internal, "ActivateControlPlane failed: %s", err)
		}

		for k, v := range filesMap {
			controlPlaneFiles = append(controlPlaneFiles, &joinproto.ControlPlaneCertOrKey{
				Name: k,
				Data: v,
			})
		}
	}

	s.log.Infof("IssueJoinTicket successful")
	return &joinproto.IssueJoinTicketResponse{
		StateDiskKey:             stateDiskKey,
		ClusterId:                id.Cluster,
		OwnerId:                  id.Owner,
		ApiServerEndpoint:        kubeArgs.APIServerEndpoint,
		Token:                    kubeArgs.Token,
		DiscoveryTokenCaCertHash: kubeArgs.CACertHashes[0],
		KubeletCert:              kubeletCert,
		KubeletKey:               kubeletKey,
		ControlPlaneFiles:        controlPlaneFiles,
	}, nil
}

// joinTokenGetter returns Kubernetes bootstrap (join) tokens.
type joinTokenGetter interface {
	// GetJoinToken returns a bootstrap (join) token.
	GetJoinToken(ttl time.Duration) (*kubeadmv1.BootstrapTokenDiscovery, error)
	GetControlPlaneCertificatesAndKeys() (map[string][]byte, error)
}

// dataKeyGetter interacts with Constellation's key management system to retrieve keys.
type dataKeyGetter interface {
	// GetDataKey returns a key derived from Constellation's KMS.
	GetDataKey(ctx context.Context, uuid string, length int) ([]byte, error)
}

type certificateAuthority interface {
	// GetCertificate returns a certificate and private key, signed by the issuer.
	GetCertificate(nodeName string) (kubeletCert []byte, kubeletKey []byte, err error)
}

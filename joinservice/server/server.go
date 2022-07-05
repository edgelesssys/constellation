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
	proto "github.com/edgelesssys/constellation/joinservice/joinproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// Server implements the core logic of Constellation's node activation service.
type Server struct {
	log             *logger.Logger
	file            file.Handler
	joinTokenGetter joinTokenGetter
	dataKeyGetter   dataKeyGetter
	ca              certificateAuthority
	proto.UnimplementedAPIServer
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

	proto.RegisterAPIServer(grpcServer, s)

	lis, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	s.log.Infof("Starting activation service on %s", lis.Addr().String())
	return grpcServer.Serve(lis)
}

// ActivateWorkerNode handles activation requests of Constellation worker nodes.
// A worker node will receive:
// - stateful disk encryption key.
// - Kubernetes join token.
// - cluster and owner ID to taint the node as initialized.
func (s *Server) ActivateWorkerNode(ctx context.Context, req *proto.ActivateWorkerNodeRequest) (*proto.ActivateWorkerNodeResponse, error) {
	s.log.Infof("ActivateWorkerNode called")
	nodeParameters, err := s.activateNode(ctx, req.DiskUuid, req.NodeName)
	if err != nil {
		return nil, fmt.Errorf("ActivateWorkerNode failed: %w", err)
	}

	s.log.Infof("ActivateWorkerNode successful")

	return &proto.ActivateWorkerNodeResponse{
		StateDiskKey:             nodeParameters.stateDiskKey,
		ClusterId:                nodeParameters.id.Cluster,
		OwnerId:                  nodeParameters.id.Owner,
		ApiServerEndpoint:        nodeParameters.kubeArgs.APIServerEndpoint,
		Token:                    nodeParameters.kubeArgs.Token,
		DiscoveryTokenCaCertHash: nodeParameters.kubeArgs.CACertHashes[0],
		KubeletCert:              nodeParameters.kubeletCert,
		KubeletKey:               nodeParameters.kubeletKey,
	}, nil
}

// ActivateControlPlaneNode handles activation requests of Constellation control-plane nodes.
// A control-plane node will receive:
// - stateful disk encryption key.
// - Kubernetes join token.
// - cluster and owner ID to taint the node as initialized.
// - a decryption key for CA certificates uploaded to the Kubernetes cluster.
func (s *Server) ActivateControlPlaneNode(ctx context.Context, req *proto.ActivateControlPlaneNodeRequest) (*proto.ActivateControlPlaneNodeResponse, error) {
	s.log.Infof("ActivateControlPlaneNode called")
	nodeParameters, err := s.activateNode(ctx, req.DiskUuid, req.NodeName)
	if err != nil {
		return nil, fmt.Errorf("ActivateControlPlaneNode failed: %w", err)
	}

	certKey, err := s.joinTokenGetter.GetControlPlaneCertificateKey()
	if err != nil {
		return nil, fmt.Errorf("ActivateControlPlane failed: %w", err)
	}

	s.log.Infof("ActivateControlPlaneNode successful")

	return &proto.ActivateControlPlaneNodeResponse{
		StateDiskKey:             nodeParameters.stateDiskKey,
		ClusterId:                nodeParameters.id.Cluster,
		OwnerId:                  nodeParameters.id.Owner,
		ApiServerEndpoint:        nodeParameters.kubeArgs.APIServerEndpoint,
		Token:                    nodeParameters.kubeArgs.Token,
		DiscoveryTokenCaCertHash: nodeParameters.kubeArgs.CACertHashes[0],
		KubeletCert:              nodeParameters.kubeletCert,
		KubeletKey:               nodeParameters.kubeletKey,
		CertificateKey:           certKey,
	}, nil
}

func (s *Server) activateNode(ctx context.Context, diskUUID, nodeName string) (nodeParameters, error) {
	log := s.log.With(zap.String("peerAddress", grpclog.PeerAddrFromContext(ctx)))
	log.Infof("Loading IDs")
	var id attestationtypes.ID
	if err := s.file.ReadJSON(filepath.Join(constants.ServiceBasePath, constants.IDFilename), &id); err != nil {
		log.With(zap.Error(err)).Errorf("Unable to load IDs")
		return nodeParameters{}, status.Errorf(codes.Internal, "unable to load IDs: %s", err)
	}

	log.Infof("Requesting disk encryption key")
	stateDiskKey, err := s.dataKeyGetter.GetDataKey(ctx, diskUUID, constants.StateDiskKeyLength)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Unable to get key for stateful disk")
		return nodeParameters{}, status.Errorf(codes.Internal, "unable to get key for stateful disk: %s", err)
	}

	log.Infof("Creating Kubernetes join token")
	kubeArgs, err := s.joinTokenGetter.GetJoinToken(constants.KubernetesJoinTokenTTL)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Unable to generate Kubernetes join arguments")
		return nodeParameters{}, status.Errorf(codes.Internal, "unable to generate Kubernetes join arguments: %s", err)
	}

	log.Infof("Creating signed kubelet certificate")
	kubeletCert, kubeletKey, err := s.ca.GetCertificate(nodeName)
	if err != nil {
		return nodeParameters{}, status.Errorf(codes.Internal, "unable to generate kubelet certificate: %s", err)
	}

	return nodeParameters{
		stateDiskKey: stateDiskKey,
		id:           id,
		kubeArgs:     kubeArgs,
		kubeletCert:  kubeletCert,
		kubeletKey:   kubeletKey,
	}, nil
}

type nodeParameters struct {
	stateDiskKey []byte
	id           attestationtypes.ID
	kubeArgs     *kubeadmv1.BootstrapTokenDiscovery
	kubeletCert  []byte
	kubeletKey   []byte
}

// joinTokenGetter returns Kubernetes bootstrap (join) tokens.
type joinTokenGetter interface {
	// GetJoinToken returns a bootstrap (join) token.
	GetJoinToken(ttl time.Duration) (*kubeadmv1.BootstrapTokenDiscovery, error)
	GetControlPlaneCertificateKey() (string, error)
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

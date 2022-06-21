package server

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"time"

	proto "github.com/edgelesssys/constellation/activation/activationproto"
	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/grpc_klog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"

	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// Server implements the core logic of Constellation's node activation service.
type Server struct {
	file            file.Handler
	joinTokenGetter joinTokenGetter
	dataKeyGetter   dataKeyGetter
	ca              certificateAuthority
	proto.UnimplementedAPIServer
}

// New initializes a new Server.
func New(fileHandler file.Handler, ca certificateAuthority, joinTokenGetter joinTokenGetter, dataKeyGetter dataKeyGetter) *Server {
	return &Server{
		file:            fileHandler,
		joinTokenGetter: joinTokenGetter,
		dataKeyGetter:   dataKeyGetter,
		ca:              ca,
	}
}

// Run starts the gRPC server on the given port, using the provided tlsConfig.
func (s *Server) Run(creds credentials.TransportCredentials, port string) error {
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(grpc_klog.LogGRPC(2)),
	)

	proto.RegisterAPIServer(grpcServer, s)

	lis, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	klog.V(2).Infof("starting activation service on %s", lis.Addr().String())
	return grpcServer.Serve(lis)
}

// ActivateWorkerNode handles activation requests of Constellation worker nodes.
// A worker node will receive:
// - stateful disk encryption key.
// - Kubernetes join token.
// - cluster and owner ID to taint the node as initialized.
func (s *Server) ActivateWorkerNode(ctx context.Context, req *proto.ActivateWorkerNodeRequest) (*proto.ActivateWorkerNodeResponse, error) {
	nodeParameters, err := s.activateNode(ctx, "ActivateWorker", req.DiskUuid, req.NodeName)
	if err != nil {
		return nil, fmt.Errorf("ActivateNode failed: %w", err)
	}

	klog.V(4).Info("ActivateNode successful")

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
	nodeParameters, err := s.activateNode(ctx, "ActivateControlPlane", req.DiskUuid, req.NodeName)
	if err != nil {
		return nil, fmt.Errorf("ActivateControlPlane failed: %w", err)
	}

	certKey, err := s.joinTokenGetter.GetControlPlaneCertificateKey()
	if err != nil {
		return nil, fmt.Errorf("ActivateControlPlane failed: %w", err)
	}

	klog.V(4).Info("ActivateControlPlane successful")

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

func (s *Server) activateNode(ctx context.Context, logPrefix, diskUUID, nodeName string) (nodeParameters, error) {
	klog.V(4).Infof("%s: loading IDs", logPrefix)
	var id attestationtypes.ID
	if err := s.file.ReadJSON(filepath.Join(constants.ActivationBasePath, constants.ActivationIDFilename), &id); err != nil {
		klog.Errorf("unable to load IDs: %s", err)
		return nodeParameters{}, status.Errorf(codes.Internal, "unable to load IDs: %s", err)
	}

	klog.V(4).Infof("%s: requesting disk encryption key", logPrefix)
	stateDiskKey, err := s.dataKeyGetter.GetDataKey(ctx, diskUUID, constants.StateDiskKeyLength)
	if err != nil {
		klog.Errorf("unable to get key for stateful disk: %s", err)
		return nodeParameters{}, status.Errorf(codes.Internal, "unable to get key for stateful disk: %s", err)
	}

	klog.V(4).Infof("%s: creating Kubernetes join token", logPrefix)
	kubeArgs, err := s.joinTokenGetter.GetJoinToken(constants.KubernetesJoinTokenTTL)
	if err != nil {
		klog.Errorf("unable to generate Kubernetes join arguments: %s", err)
		return nodeParameters{}, status.Errorf(codes.Internal, "unable to generate Kubernetes join arguments: %s", err)
	}

	klog.V(4).Infof("%s: creating signed kubelet certificate", logPrefix)
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

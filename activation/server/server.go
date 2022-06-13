package server

import (
	"context"
	"fmt"
	"net"
	"time"

	proto "github.com/edgelesssys/constellation/activation/activationproto"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
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
		grpc.UnaryInterceptor(logGRPC),
	)

	proto.RegisterAPIServer(grpcServer, s)

	lis, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	klog.V(2).Infof("starting activation service on %s", lis.Addr().String())
	return grpcServer.Serve(lis)
}

// ActivateNode handles activation requests of Constellation worker nodes.
// A worker node will receive:
// - stateful disk encryption key.
// - Kubernetes join token.
// - cluster and owner ID to taint the node as initialized.
func (s *Server) ActivateNode(ctx context.Context, req *proto.ActivateNodeRequest) (*proto.ActivateNodeResponse, error) {
	klog.V(4).Info("ActivateNode: loading IDs")
	var id id
	if err := s.file.ReadJSON(constants.ActivationIDFilename, &id); err != nil {
		klog.Errorf("unable to load IDs: %s", err)
		return nil, status.Errorf(codes.Internal, "unable to load IDs: %s", err)
	}

	klog.V(4).Info("ActivateNode: requesting disk encryption key")
	stateDiskKey, err := s.dataKeyGetter.GetDataKey(ctx, req.DiskUuid, constants.StateDiskKeyLength)
	if err != nil {
		klog.Errorf("unable to get key for stateful disk: %s", err)
		return nil, status.Errorf(codes.Internal, "unable to get key for stateful disk: %s", err)
	}

	klog.V(4).Info("ActivateNode: creating Kubernetes join token")
	kubeArgs, err := s.joinTokenGetter.GetJoinToken(constants.KubernetesJoinTokenTTL)
	if err != nil {
		klog.Errorf("unable to generate Kubernetes join arguments: %s", err)
		return nil, status.Errorf(codes.Internal, "unable to generate Kubernetes join arguments: %s", err)
	}

	klog.V(4).Info("ActivateNode: creating signed kubelet certificate")
	kubeletCert, kubeletKey, err := s.ca.GetCertificate(req.NodeName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to generate kubelet certificate: %s", err)
	}

	klog.V(4).Info("ActivateNode successful")

	return &proto.ActivateNodeResponse{
		StateDiskKey:             stateDiskKey,
		ClusterId:                id.Cluster,
		OwnerId:                  id.Owner,
		ApiServerEndpoint:        kubeArgs.APIServerEndpoint,
		Token:                    kubeArgs.Token,
		DiscoveryTokenCaCertHash: kubeArgs.CACertHashes[0],
		KubeletCert:              kubeletCert,
		KubeletKey:               kubeletKey,
	}, nil
}

// ActivateCoordinator handles activation requests of Constellation control-plane nodes.
func (s *Server) ActivateCoordinator(ctx context.Context, req *proto.ActivateCoordinatorRequest) (*proto.ActivateCoordinatorResponse, error) {
	panic("not implemented")
}

// joinTokenGetter returns Kubernetes bootstrap (join) tokens.
type joinTokenGetter interface {
	// GetJoinToken returns a bootstrap (join) token.
	GetJoinToken(ttl time.Duration) (*kubeadmv1.BootstrapTokenDiscovery, error)
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

type id struct {
	Cluster []byte `json:"cluster"`
	Owner   []byte `json:"owner"`
}

// logGRPC writes a log with the name of every gRPC call or error it receives.
func logGRPC(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// log the requests method name
	klog.V(2).Infof("GRPC call: %s", info.FullMethod)

	// log errors, if any
	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("GRPC error: %v", err)
	}
	return resp, err
}

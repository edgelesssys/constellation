package initserver

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/edgelesssys/constellation/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/bootstrapper/internal/diskencryption"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes"
	"github.com/edgelesssys/constellation/bootstrapper/internal/nodelock"
	"github.com/edgelesssys/constellation/bootstrapper/nodestate"
	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/bootstrapper/util"
	"github.com/edgelesssys/constellation/internal/atls"
	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the initialization server, which is started on each node.
// The server handles initialization calls from the CLI and initializes the
// Kubernetes cluster.
type Server struct {
	nodeLock    *nodelock.Lock
	initializer ClusterInitializer
	disk        encryptedDisk
	fileHandler file.Handler
	grpcServer  serveStopper

	logger *zap.Logger

	initproto.UnimplementedAPIServer
}

// New creates a new initialization server.
func New(lock *nodelock.Lock, kube ClusterInitializer, issuer atls.Issuer, fh file.Handler, logger *zap.Logger) *Server {
	logger = logger.Named("initServer")
	server := &Server{
		nodeLock:    lock,
		disk:        diskencryption.New(),
		initializer: kube,
		fileHandler: fh,
		logger:      logger,
	}

	creds := atlscredentials.New(issuer, nil)
	grpcLogger := logger.Named("gRPC")
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(grpcLogger),
		)),
	)
	initproto.RegisterAPIServer(grpcServer, server)

	server.grpcServer = grpcServer

	return server
}

func (s *Server) Serve(ip, port string) error {
	lis, err := net.Listen("tcp", net.JoinHostPort(ip, port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	return s.grpcServer.Serve(lis)
}

// Init initializes the cluster.
func (s *Server) Init(ctx context.Context, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	s.logger.Info("Init called")

	if ok := s.nodeLock.TryLockOnce(); !ok {
		// The join client seems to already have a connection to an
		// existing join service. At this point, any further call to
		// init does not make sense, so we just stop.
		//
		// The server stops itself after the current call is done.
		go s.grpcServer.GracefulStop()
		s.logger.Info("node is already in a join process")
		return nil, status.Error(codes.FailedPrecondition, "node is already being activated")
	}

	id, err := s.deriveAttestationID(req.MasterSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if err := s.setupDisk(req.MasterSecret); err != nil {
		return nil, status.Errorf(codes.Internal, "setting up disk: %s", err)
	}

	state := nodestate.NodeState{
		Role:      role.ControlPlane,
		OwnerID:   id.Owner,
		ClusterID: id.Cluster,
	}
	if err := state.ToFile(s.fileHandler); err != nil {
		return nil, status.Errorf(codes.Internal, "persisting node state: %s", err)
	}

	kubeconfig, err := s.initializer.InitCluster(ctx,
		req.AutoscalingNodeGroups,
		req.CloudServiceAccountUri,
		req.KubernetesVersion,
		id,
		kubernetes.KMSConfig{
			MasterSecret:       req.MasterSecret,
			KMSURI:             req.KmsUri,
			StorageURI:         req.StorageUri,
			KeyEncryptionKeyID: req.KeyEncryptionKeyId,
			UseExistingKEK:     req.UseExistingKek,
		},
		sshProtoKeysToMap(req.SshUserKeys),
		s.logger,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "initializing cluster: %s", err)
	}

	s.logger.Info("Init succeeded")
	return &initproto.InitResponse{
		Kubeconfig: kubeconfig,
		OwnerId:    id.Owner,
		ClusterId:  id.Cluster,
	}, nil
}

func (s *Server) setupDisk(masterSecret []byte) error {
	if err := s.disk.Open(); err != nil {
		return fmt.Errorf("opening encrypted disk: %w", err)
	}
	defer s.disk.Close()

	uuid, err := s.disk.UUID()
	if err != nil {
		return fmt.Errorf("retrieving uuid of disk: %w", err)
	}
	uuid = strings.ToLower(uuid)

	// TODO: Choose a way to salt the key derivation
	diskKey, err := util.DeriveKey(masterSecret, []byte("Constellation"), []byte("key"+uuid), 32)
	if err != nil {
		return err
	}

	return s.disk.UpdatePassphrase(string(diskKey))
}

func (s *Server) deriveAttestationID(masterSecret []byte) (attestationtypes.ID, error) {
	clusterID, err := util.GenerateRandomBytes(constants.RNGLengthDefault)
	if err != nil {
		return attestationtypes.ID{}, err
	}

	// TODO: Choose a way to salt the key derivation
	ownerID, err := util.DeriveKey(masterSecret, []byte("Constellation"), []byte("id"), constants.RNGLengthDefault)
	if err != nil {
		return attestationtypes.ID{}, err
	}

	return attestationtypes.ID{Owner: ownerID, Cluster: clusterID}, nil
}

func sshProtoKeysToMap(keys []*initproto.SSHUserKey) map[string]string {
	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[key.Username] = key.PublicKey
	}
	return keyMap
}

// ClusterInitializer has the ability to initialize a cluster.
type ClusterInitializer interface {
	// InitCluster initializes a new Kubernetes cluster.
	InitCluster(
		ctx context.Context,
		autoscalingNodeGroups []string,
		cloudServiceAccountURI string,
		k8sVersion string,
		id attestationtypes.ID,
		kmsConfig kubernetes.KMSConfig,
		sshUserKeys map[string]string,
		logger *zap.Logger,
	) ([]byte, error)
}

type encryptedDisk interface {
	// Open prepares the underlying device for disk operations.
	Open() error
	// Close closes the underlying device.
	Close() error
	// UUID gets the device's UUID.
	UUID() (string, error)
	// UpdatePassphrase switches the initial random passphrase of the encrypted disk to a permanent passphrase.
	UpdatePassphrase(passphrase string) error
}

type serveStopper interface {
	// Serve starts the server.
	Serve(lis net.Listener) error
	// GracefulStop stops the server and blocks until all requests are done.
	GracefulStop()
}

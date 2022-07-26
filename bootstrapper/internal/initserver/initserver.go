package initserver

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/edgelesssys/constellation/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/bootstrapper/internal/diskencryption"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes"
	"github.com/edgelesssys/constellation/bootstrapper/nodestate"
	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the initialization server, which is started on each node.
// The server handles initialization calls from the CLI and initializes the
// Kubernetes cluster.
type Server struct {
	nodeLock    locker
	initializer ClusterInitializer
	disk        encryptedDisk
	fileHandler file.Handler
	grpcServer  serveStopper
	cleaner     cleaner

	log *logger.Logger

	initproto.UnimplementedAPIServer
}

// New creates a new initialization server.
func New(lock locker, kube ClusterInitializer, issuer atls.Issuer, fh file.Handler, log *logger.Logger) *Server {
	log = log.Named("initServer")
	server := &Server{
		nodeLock:    lock,
		disk:        diskencryption.New(),
		initializer: kube,
		fileHandler: fh,
		log:         log,
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(atlscredentials.New(issuer, nil)),
		log.Named("gRPC").GetServerUnaryInterceptor(),
	)
	initproto.RegisterAPIServer(grpcServer, server)

	server.grpcServer = grpcServer

	return server
}

// Serve starts the initialization server.
func (s *Server) Serve(ip, port string, cleaner cleaner) error {
	s.cleaner = cleaner
	lis, err := net.Listen("tcp", net.JoinHostPort(ip, port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	return s.grpcServer.Serve(lis)
}

// Init initializes the cluster.
func (s *Server) Init(ctx context.Context, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	defer s.cleaner.Clean()
	log := s.log.With(zap.String("peer", grpclog.PeerAddrFromContext(ctx)))
	log.Infof("Init called")

	// generate values for cluster attestation
	measurementSalt, clusterID, err := deriveMeasurementValues(req.MasterSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deriving measurement values: %s", err)
	}

	nodeLockAcquired, err := s.nodeLock.TryLockOnce(clusterID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "locking node: %s", err)
	}
	if !nodeLockAcquired {
		// The join client seems to already have a connection to an
		// existing join service. At this point, any further call to
		// init does not make sense, so we just stop.
		//
		// The server stops itself after the current call is done.
		log.Warnf("Node is already in a join process")
		return nil, status.Error(codes.FailedPrecondition, "node is already being activated")
	}

	if err := s.setupDisk(req.MasterSecret); err != nil {
		return nil, status.Errorf(codes.Internal, "setting up disk: %s", err)
	}

	state := nodestate.NodeState{
		Role:            role.ControlPlane,
		MeasurementSalt: measurementSalt,
	}
	if err := state.ToFile(s.fileHandler); err != nil {
		return nil, status.Errorf(codes.Internal, "persisting node state: %s", err)
	}

	kubeconfig, err := s.initializer.InitCluster(ctx,
		req.AutoscalingNodeGroups,
		req.CloudServiceAccountUri,
		req.KubernetesVersion,
		measurementSalt,
		kubernetes.KMSConfig{
			MasterSecret:       req.MasterSecret,
			KMSURI:             req.KmsUri,
			StorageURI:         req.StorageUri,
			KeyEncryptionKeyID: req.KeyEncryptionKeyId,
			UseExistingKEK:     req.UseExistingKek,
		},
		sshProtoKeysToMap(req.SshUserKeys),
		s.log,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "initializing cluster: %s", err)
	}

	log.Infof("Init succeeded")
	return &initproto.InitResponse{
		Kubeconfig: kubeconfig,
		ClusterId:  clusterID,
	}, nil
}

// Stop stops the initialization server gracefully.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
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
	diskKey, err := crypto.DeriveKey(masterSecret, []byte("Constellation"), []byte(crypto.HKDFInfoPrefix+uuid), 32)
	if err != nil {
		return err
	}

	return s.disk.UpdatePassphrase(string(diskKey))
}

func sshProtoKeysToMap(keys []*initproto.SSHUserKey) map[string]string {
	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[key.Username] = key.PublicKey
	}
	return keyMap
}

func deriveMeasurementValues(masterSecret []byte) (salt, clusterID []byte, err error) {
	salt, err = crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return nil, nil, err
	}
	secret, err := attestation.DeriveMeasurementSecret(masterSecret)
	if err != nil {
		return nil, nil, err
	}
	clusterID, err = attestation.DeriveClusterID(salt, secret)
	if err != nil {
		return nil, nil, err
	}

	return salt, clusterID, nil
}

// ClusterInitializer has the ability to initialize a cluster.
type ClusterInitializer interface {
	// InitCluster initializes a new Kubernetes cluster.
	InitCluster(
		ctx context.Context,
		autoscalingNodeGroups []string,
		cloudServiceAccountURI string,
		k8sVersion string,
		measurementSalt []byte,
		kmsConfig kubernetes.KMSConfig,
		sshUserKeys map[string]string,
		log *logger.Logger,
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

type locker interface {
	// TryLockOnce tries to lock the node. If the node is already locked, it
	// returns false. If the node is unlocked, it locks it and returns true.
	TryLockOnce(clusterID []byte) (bool, error)
}

type cleaner interface {
	Clean()
}

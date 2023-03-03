/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# InitServer

The InitServer is one of the two main components of the bootstrapper.
It is responsible for the initial setup of a node, and the initialization of the Kubernetes cluster.

The InitServer is started on each node, and waits for either a call from the CLI,
or for the JoinClient to connect to an existing cluster.

If a call from the CLI is received, the InitServer bootstraps the Kubernetes cluster, and stops the JoinClient.
*/
package initserver

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/diskencryption"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms"
	kmssetup "github.com/edgelesssys/constellation/v2/internal/kms/setup"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/nodestate"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
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
	issuer      atls.Issuer

	initSecretHash []byte

	log *logger.Logger

	initproto.UnimplementedAPIServer
}

// New creates a new initialization server.
func New(ctx context.Context, lock locker, kube ClusterInitializer, issuer atls.Issuer, fh file.Handler, metadata MetadataAPI, log *logger.Logger) (*Server, error) {
	log = log.Named("initServer")

	initSecretHash, err := metadata.InitSecretHash(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving init secret hash: %w", err)
	}
	if len(initSecretHash) == 0 {
		return nil, fmt.Errorf("init secret hash is empty")
	}

	server := &Server{
		nodeLock:       lock,
		disk:           diskencryption.New(),
		initializer:    kube,
		fileHandler:    fh,
		issuer:         issuer,
		log:            log,
		initSecretHash: initSecretHash,
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(atlscredentials.New(issuer, nil)),
		grpc.KeepaliveParams(keepalive.ServerParameters{Time: 15 * time.Second}),
		log.Named("gRPC").GetServerUnaryInterceptor(),
	)
	initproto.RegisterAPIServer(grpcServer, server)

	server.grpcServer = grpcServer
	return server, nil
}

// Serve starts the initialization server.
func (s *Server) Serve(ip, port string, cleaner cleaner) error {
	s.cleaner = cleaner
	lis, err := net.Listen("tcp", net.JoinHostPort(ip, port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.log.Infof("Starting")
	return s.grpcServer.Serve(lis)
}

// Init initializes the cluster.
func (s *Server) Init(ctx context.Context, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	log := s.log.With(zap.String("peer", grpclog.PeerAddrFromContext(ctx)))
	log.Infof("Init called")

	if err := bcrypt.CompareHashAndPassword(s.initSecretHash, req.InitSecret); err != nil {
		return nil, status.Errorf(codes.Internal, "invalid init secret %s", err)
	}

	cloudKms, err := kmssetup.KMS(ctx, req.StorageUri, req.KmsUri)
	if err != nil {
		return nil, fmt.Errorf("creating kms client: %w", err)
	}

	// generate values for cluster attestation
	measurementSalt, clusterID, err := deriveMeasurementValues(ctx, cloudKms)
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

	// Stop the join client -> We no longer expect to join an existing cluster,
	// since we are bootstrapping a new one.
	// Any errors following this call will result in a failed node that may not join any cluster.
	s.cleaner.Clean()

	if err := s.setupDisk(ctx, cloudKms); err != nil {
		return nil, status.Errorf(codes.Internal, "setting up disk: %s", err)
	}

	state := nodestate.NodeState{
		Role:            role.ControlPlane,
		MeasurementSalt: measurementSalt,
	}
	if err := state.ToFile(s.fileHandler); err != nil {
		return nil, status.Errorf(codes.Internal, "persisting node state: %s", err)
	}

	// Check if we are running on a CVM
	_, isCVM := s.issuer.(*snp.Issuer)

	clusterName := req.ClusterName
	if clusterName == "" {
		clusterName = "constellation"
	}

	kubeconfig, err := s.initializer.InitCluster(ctx,
		req.CloudServiceAccountUri,
		req.KubernetesVersion,
		clusterName,
		measurementSalt,
		req.EnforcedPcrs,
		req.EnforceIdkeydigest,
		isCVM,
		req.HelmDeployments,
		req.ConformanceMode,
		components.NewComponentsFromInitProto(req.KubernetesComponents),
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
	s.log.Infof("Stopping")

	s.grpcServer.GracefulStop()

	s.log.Infof("Stopped")
}

func (s *Server) setupDisk(ctx context.Context, cloudKms kms.CloudKMS) error {
	if err := s.disk.Open(); err != nil {
		return fmt.Errorf("opening encrypted disk: %w", err)
	}
	defer s.disk.Close()

	uuid, err := s.disk.UUID()
	if err != nil {
		return fmt.Errorf("retrieving uuid of disk: %w", err)
	}
	uuid = strings.ToLower(uuid)

	diskKey, err := cloudKms.GetDEK(ctx, crypto.DEKPrefix+uuid, crypto.StateDiskKeyLength)
	if err != nil {
		return err
	}

	return s.disk.UpdatePassphrase(string(diskKey))
}

func deriveMeasurementValues(ctx context.Context, cloudKms kms.CloudKMS) (salt, clusterID []byte, err error) {
	salt, err = crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return nil, nil, err
	}
	secret, err := cloudKms.GetDEK(ctx, crypto.DEKPrefix+crypto.MeasurementSecretKeyID, crypto.DerivedKeyLengthDefault)
	if err != nil {
		return nil, nil, err
	}
	clusterID, err = attestation.DeriveClusterID(secret, salt)
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
		cloudServiceAccountURI string,
		k8sVersion string,
		clusterName string,
		measurementSalt []byte,
		enforcedPcrs []uint32,
		enforceIDKeyDigest bool,
		azureCVM bool,
		helmDeployments []byte,
		conformanceMode bool,
		kubernetesComponents components.Components,
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

// MetadataAPI provides information about the instances.
type MetadataAPI interface {
	// InitSecretHash returns the initSecretHash of the instance.
	InitSecretHash(ctx context.Context) ([]byte, error)
}

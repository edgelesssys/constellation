/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
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
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/cloud/vmtype"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/nodestate"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// Server is the initialization server, which is started on each node.
// The server handles initialization calls from the CLI and initializes the
// Kubernetes cluster.
type Server struct {
	nodeLock      locker
	initializer   ClusterInitializer
	disk          encryptedDisk
	fileHandler   file.Handler
	grpcServer    serveStopper
	cleaner       cleaner
	issuerWrapper IssuerWrapper

	log *logger.Logger

	initproto.UnimplementedAPIServer
}

// New creates a new initialization server.
func New(lock locker, kube ClusterInitializer, issuerWrapper IssuerWrapper, fh file.Handler, log *logger.Logger) *Server {
	log = log.Named("initServer")

	server := &Server{
		nodeLock:      lock,
		disk:          diskencryption.New(),
		initializer:   kube,
		fileHandler:   fh,
		issuerWrapper: issuerWrapper,
		log:           log,
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(atlscredentials.New(issuerWrapper, nil)),
		grpc.KeepaliveParams(keepalive.ServerParameters{Time: 15 * time.Second}),
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

	s.log.Infof("Starting")
	return s.grpcServer.Serve(lis)
}

// Init initializes the cluster.
func (s *Server) Init(ctx context.Context, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	defer s.cleaner.Clean()
	log := s.log.With(zap.String("peer", grpclog.PeerAddrFromContext(ctx)))
	log.Infof("Init called")

	// generate values for cluster attestation
	measurementSalt, clusterID, err := deriveMeasurementValues(req.MasterSecret, req.Salt)
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

	if err := s.setupDisk(req.MasterSecret, req.Salt); err != nil {
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
		req.CloudServiceAccountUri,
		req.KubernetesVersion,
		measurementSalt,
		req.EnforcedPcrs,
		req.EnforceIdkeydigest,
		s.issuerWrapper.IDKeyDigest(),
		s.issuerWrapper.VMType() == vmtype.AzureCVM,
		resources.KMSConfig{
			MasterSecret:       req.MasterSecret,
			Salt:               req.Salt,
			KMSURI:             req.KmsUri,
			StorageURI:         req.StorageUri,
			KeyEncryptionKeyID: req.KeyEncryptionKeyId,
			UseExistingKEK:     req.UseExistingKek,
		},
		sshProtoKeysToMap(req.SshUserKeys),
		req.HelmDeployments,
		req.ConformanceMode,
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

func (s *Server) setupDisk(masterSecret, salt []byte) error {
	if err := s.disk.Open(); err != nil {
		return fmt.Errorf("opening encrypted disk: %w", err)
	}
	defer s.disk.Close()

	uuid, err := s.disk.UUID()
	if err != nil {
		return fmt.Errorf("retrieving uuid of disk: %w", err)
	}
	uuid = strings.ToLower(uuid)

	diskKey, err := crypto.DeriveKey(masterSecret, salt, []byte(crypto.HKDFInfoPrefix+uuid), crypto.DerivedKeyLengthDefault)
	if err != nil {
		return err
	}

	return s.disk.UpdatePassphrase(string(diskKey))
}

type IssuerWrapper struct {
	atls.Issuer
	vmType      vmtype.VMType
	idkeydigest []byte
}

func NewIssuerWrapper(issuer atls.Issuer, vmType vmtype.VMType, idkeydigest []byte) IssuerWrapper {
	return IssuerWrapper{
		Issuer:      issuer,
		vmType:      vmType,
		idkeydigest: idkeydigest,
	}
}

func (i *IssuerWrapper) VMType() vmtype.VMType {
	return i.vmType
}

func (i *IssuerWrapper) IDKeyDigest() []byte {
	return i.idkeydigest
}

func sshProtoKeysToMap(keys []*initproto.SSHUserKey) map[string]string {
	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[key.Username] = key.PublicKey
	}
	return keyMap
}

func deriveMeasurementValues(masterSecret, hkdfSalt []byte) (salt, clusterID []byte, err error) {
	salt, err = crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return nil, nil, err
	}
	secret, err := attestation.DeriveMeasurementSecret(masterSecret, hkdfSalt)
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
		measurementSalt []byte,
		enforcedPcrs []uint32,
		enforceIDKeyDigest bool,
		idKeyDigest []byte,
		azureCVM bool,
		kmsConfig resources.KMSConfig,
		sshUserKeys map[string]string,
		helmDeployments []byte,
		conformanceMode bool,
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

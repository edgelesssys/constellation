/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
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
	"bufio"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/addresses"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/journald"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/constants"
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
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// Server is the initialization server, which is started on each node.
// The server handles initialization calls from the CLI and initializes the
// Kubernetes cluster.
type Server struct {
	nodeLock     locker
	initializer  ClusterInitializer
	disk         encryptedDisk
	fileHandler  file.Handler
	grpcServer   serveStopper
	cleaner      cleaner
	issuer       atls.Issuer
	shutdownLock sync.RWMutex

	initSecretHash []byte
	initFailure    error

	kmsURI string

	log *slog.Logger

	journaldCollector journaldCollection

	initproto.UnimplementedAPIServer
}

// New creates a new initialization server.
func New(
	ctx context.Context, lock locker, kube ClusterInitializer, issuer atls.Issuer,
	disk encryptedDisk, fh file.Handler, metadata MetadataAPI, log *slog.Logger,
) (*Server, error) {
	log = log.WithGroup("initServer")

	initSecretHash, err := metadata.InitSecretHash(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving init secret hash: %w", err)
	}
	if len(initSecretHash) == 0 {
		return nil, fmt.Errorf("init secret hash is empty")
	}

	jctlCollector, err := journald.NewCollector(ctx)
	if err != nil {
		return nil, err
	}

	server := &Server{
		nodeLock:          lock,
		disk:              disk,
		initializer:       kube,
		fileHandler:       fh,
		issuer:            issuer,
		log:               log,
		initSecretHash:    initSecretHash,
		journaldCollector: jctlCollector,
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(atlscredentials.New(issuer, nil)),
		grpc.KeepaliveParams(keepalive.ServerParameters{Time: 15 * time.Second}),
		logger.GetServerUnaryInterceptor(logger.GRPCLogger(log)),
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

	s.log.Info("Starting")
	err = s.grpcServer.Serve(lis)

	// If Init failed, we mark the disk for reset, so the node can restart the process
	// In this case we don't care about any potential errors from the grpc server
	if s.initFailure != nil {
		s.log.Error("Fatal error during Init request", "error", s.initFailure)
		return err
	}

	return err
}

// Init initializes the cluster.
func (s *Server) Init(req *initproto.InitRequest, stream initproto.API_InitServer) (retErr error) {
	// Acquire lock to prevent shutdown while Init is still running
	s.shutdownLock.RLock()
	defer s.shutdownLock.RUnlock()

	log := s.log.With(slog.String("peer", grpclog.PeerAddrFromContext(stream.Context())))
	log.Info("Init called")

	s.kmsURI = req.KmsUri

	if err := bcrypt.CompareHashAndPassword(s.initSecretHash, req.InitSecret); err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "invalid init secret %s", err)))
	}

	cloudKms, err := kmssetup.KMS(stream.Context(), req.StorageUri, req.KmsUri)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "creating kms client: %s", err)))
	}

	// generate values for cluster attestation
	clusterID, err := deriveMeasurementValues(stream.Context(), req.MeasurementSalt, cloudKms)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "deriving measurement values: %s", err)))
	}

	nodeLockAcquired, err := s.nodeLock.TryLockOnce(clusterID)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "locking node: %s", err)))
	}
	if !nodeLockAcquired {
		// The join client seems to already have a connection to an
		// existing join service. At this point, any further call to
		// init does not make sense, so we just stop.
		//
		// The server stops itself after the current call is done.
		log.Warn("Node is already in a join process")

		err = status.Error(codes.FailedPrecondition, "node is already being activated")

		if e := s.sendLogsWithMessage(stream, err); e != nil {
			err = errors.Join(err, e)
		}
		return err
	}

	// Stop the join client -> We no longer expect to join an existing cluster,
	// since we are bootstrapping a new one.
	// Any errors following this call will result in a failed node that may not join any cluster.
	s.cleaner.Clean()
	defer func() {
		s.initFailure = retErr
	}()

	if err := s.setupDisk(stream.Context(), cloudKms); err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "setting up disk: %s", err)))
	}

	state := nodestate.NodeState{
		Role:            role.ControlPlane,
		MeasurementSalt: req.MeasurementSalt,
	}
	if err := state.ToFile(s.fileHandler); err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "persisting node state: %s", err)))
	}

	// Derive the emergency ssh CA key
	key, err := cloudKms.GetDEK(stream.Context(), crypto.DEKPrefix+constants.SSHCAKeySuffix, ed25519.SeedSize)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "retrieving DEK for key derivation: %s", err)))
	}
	ca, err := crypto.GenerateEmergencySSHCAKey(key)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "generating emergency SSH CA key: %s", err)))
	}
	if err := s.fileHandler.Write(constants.SSHCAKeyPath, ssh.MarshalAuthorizedKey(ca.PublicKey()), file.OptMkdirAll); err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "writing ssh CA pubkey: %s", err)))
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "getting network interfaces: %s", err)))
	}
	// Needed since go doesn't implicitly convert slices of structs to slices of interfaces
	interfacesForFunc := make([]addresses.NetInterface, len(interfaces))
	for i := range interfaces {
		interfacesForFunc[i] = &interfaces[i]
	}

	principalList, err := addresses.GetMachineNetworkAddresses(interfacesForFunc)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "failed to get network addresses: %s", err)))
	}
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "failed to get hostname: %s", err)))
	}

	principalList = append(principalList, hostname)
	principalList = append(principalList, req.ApiserverCertSans...)

	hostKeyContent, err := s.fileHandler.Read(constants.SSHHostKeyPath)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "failed to read host SSH key: %s", err)))
	}

	hostPrivateKey, err := ssh.ParsePrivateKey(hostKeyContent)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "failed to parse host SSH key: %s", err)))
	}

	hostKeyPubSSH := hostPrivateKey.PublicKey()

	hostCertificate, err := crypto.GenerateSSHHostCertificate(principalList, hostKeyPubSSH, ca)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "generating SSH host certificate: %s", err)))
	}

	if err := s.fileHandler.Write(constants.SSHAdditionalPrincipalsPath, []byte(strings.Join(req.ApiserverCertSans, ",")), file.OptMkdirAll); err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "writing list of public ssh principals: %s", err)))
	}

	if err := s.fileHandler.Write(constants.SSHHostCertificatePath, ssh.MarshalAuthorizedKey(hostCertificate), file.OptMkdirAll); err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "writing ssh host certificate: %s", err)))
	}

	clusterName := req.ClusterName
	if clusterName == "" {
		clusterName = "constellation"
	}

	kubeconfig, err := s.initializer.InitCluster(stream.Context(),
		req.KubernetesVersion,
		clusterName,
		req.ConformanceMode,
		req.KubernetesComponents,
		req.ApiserverCertSans,
		req.ServiceCidr,
	)
	if err != nil {
		return errors.Join(err, s.sendLogsWithMessage(stream, status.Errorf(codes.Internal, "initializing cluster: %s", err)))
	}

	log.Info("Init succeeded")

	successMessage := &initproto.InitResponse_InitSuccess{
		InitSuccess: &initproto.InitSuccessResponse{
			Kubeconfig: kubeconfig,
			ClusterId:  clusterID,
		},
	}

	return stream.Send(&initproto.InitResponse{Kind: successMessage})
}

func (s *Server) sendLogsWithMessage(stream initproto.API_InitServer, message error) error {
	// send back the error message
	if err := stream.Send(&initproto.InitResponse{
		Kind: &initproto.InitResponse_InitFailure{
			InitFailure: &initproto.InitFailureResponse{Error: message.Error()},
		},
	}); err != nil {
		return err
	}

	logPipe, err := s.journaldCollector.Start()
	if err != nil {
		return status.Errorf(codes.Internal, "failed starting the log collector: %s", err)
	}

	reader := bufio.NewReader(logPipe)
	buffer := make([]byte, 1024)

	for {
		n, err := io.ReadFull(reader, buffer)
		buffer = buffer[:n] // cap the buffer so that we don't have a bunch of nullbytes at the end
		if err != nil {
			if err == io.EOF {
				break
			}
			if err != io.ErrUnexpectedEOF {
				return status.Errorf(codes.Internal, "failed to read from pipe: %s", err)
			}
		}

		err = stream.Send(&initproto.InitResponse{
			Kind: &initproto.InitResponse_Log{
				Log: &initproto.LogResponseType{
					Log: buffer,
				},
			},
		})
		if err != nil {
			return status.Errorf(codes.Internal, "failed to send chunk: %s", err)
		}
	}

	return nil
}

// Stop stops the initialization server gracefully.
func (s *Server) Stop() {
	s.log.Info("Stopping")

	// Make sure to only stop the server if no Init calls are running
	s.shutdownLock.Lock()
	defer s.shutdownLock.Unlock()
	s.grpcServer.GracefulStop()

	s.log.Info("Stopped")
}

func (s *Server) setupDisk(ctx context.Context, cloudKms kms.CloudKMS) error {
	free, err := s.disk.Open()
	if err != nil {
		return fmt.Errorf("opening encrypted disk: %w", err)
	}
	defer free()

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

func deriveMeasurementValues(ctx context.Context, measurementSalt []byte, cloudKms kms.CloudKMS) (clusterID []byte, err error) {
	secret, err := cloudKms.GetDEK(ctx, crypto.DEKPrefix+crypto.MeasurementSecretKeyID, crypto.DerivedKeyLengthDefault)
	if err != nil {
		return nil, err
	}
	clusterID, err = attestation.DeriveClusterID(secret, measurementSalt)
	if err != nil {
		return nil, err
	}

	return clusterID, nil
}

// ClusterInitializer has the ability to initialize a cluster.
type ClusterInitializer interface {
	// InitCluster initializes a new Kubernetes cluster.
	InitCluster(
		ctx context.Context,
		k8sVersion string,
		clusterName string,
		conformanceMode bool,
		kubernetesComponents components.Components,
		apiServerCertSANs []string,
		serviceCIDR string,
	) ([]byte, error)
}

type encryptedDisk interface {
	// Open prepares the underlying device for disk operations.
	Open() (free func(), err error)
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

// journaldCollection is an interface for collecting journald logs.
type journaldCollection interface {
	// Start starts the journald collector and returns a pipe from which the system logs can be read.
	Start() (io.ReadCloser, error)
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// Server implements the core logic of Constellation's node join service.
type Server struct {
	measurementSalt []byte

	log             *logger.Logger
	file            file.Handler
	joinTokenGetter joinTokenGetter
	dataKeyGetter   dataKeyGetter
	ca              certificateAuthority
	kubeClient      kubeClient
	joinproto.UnimplementedAPIServer
}

// New initializes a new Server.
func New(
	measurementSalt []byte, fileHandler file.Handler, ca certificateAuthority,
	joinTokenGetter joinTokenGetter, dataKeyGetter dataKeyGetter, log *logger.Logger,
) (*Server, error) {
	kubeClient, err := kubernetes.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	return &Server{
		measurementSalt: measurementSalt,
		log:             log,
		file:            fileHandler,
		joinTokenGetter: joinTokenGetter,
		dataKeyGetter:   dataKeyGetter,
		ca:              ca,
		kubeClient:      kubeClient,
	}, nil
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
// - measurement salt and secret, to mark the node as initialized.
// In addition, control plane nodes receive:
// - a decryption key for CA certificates uploaded to the Kubernetes cluster.
func (s *Server) IssueJoinTicket(ctx context.Context, req *joinproto.IssueJoinTicketRequest) (*joinproto.IssueJoinTicketResponse, error) {
	log := s.log.With(zap.String("peerAddress", grpclog.PeerAddrFromContext(ctx)))
	log.Infof("IssueJoinTicket called")

	log.Infof("Requesting measurement secret")
	measurementSecret, err := s.dataKeyGetter.GetDataKey(ctx, attestation.MeasurementSecretContext, crypto.DerivedKeyLengthDefault)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Unable to get measurement secret")
		return nil, status.Errorf(codes.Internal, "unable to get measurement secret: %s", err)
	}

	log.Infof("Requesting disk encryption key")
	stateDiskKey, err := s.dataKeyGetter.GetDataKey(ctx, req.DiskUuid, crypto.StateDiskKeyLength)
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

	log.Infof("Querying K8sVersion ConfigMap for Kubernetes version")
	k8sVersion, err := s.getK8sVersion()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get k8s version: %s", err)
	}

	log.Infof("Querying K8sVersion ConfigMap for components ConfigMap name")
	componentsConfigMapName, err := s.getK8sComponentsConfigMapName()
	if errors.Is(err, fs.ErrNotExist) {
		// If the file does not exist, the Constellation was initialized with a version before 2.3.0
		// As a migration step, the join service will create the ConfigMap with the K8s components which
		// match the K8s minor version of the cluster.
		log.Warnf("Reference to K8sVersion ConfigMap does not exist, creating fallback Components ConfigMap and referencing it in K8sVersion ConfigMap")
		log.Warnf("This is expected if the Constellation was initialized with a CLI before version 2.3.0")
		log.Warnf("DEPRECATION WARNING: This is a migration step and will be removed in a future release")
		componentsConfigMapName, err = s.createFallbackComponentsConfigMap(ctx, k8sVersion)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "unable to create fallback k8s components configmap: %s", err)
		}
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get components ConfigMap name: %s", err)
	}

	log.Infof("Querying %s ConfigMap for components", componentsConfigMapName)
	components, err := s.kubeClient.GetComponents(ctx, componentsConfigMapName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get components: %s", err)
	}

	log.Infof("Creating signed kubelet certificate")
	kubeletCert, err := s.ca.GetCertificate(req.CertificateRequest)
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

	nodeName, err := s.ca.GetNodeNameFromCSR(req.CertificateRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get node name from CSR: %s", err)
	}

	if err := s.kubeClient.AddNodeToJoiningNodes(ctx, nodeName, components.GetHash(), req.IsControlPlane); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to add node to joining nodes: %s", err)
	}

	log.Infof("IssueJoinTicket successful")
	return &joinproto.IssueJoinTicketResponse{
		StateDiskKey:             stateDiskKey,
		MeasurementSalt:          s.measurementSalt,
		MeasurementSecret:        measurementSecret,
		ApiServerEndpoint:        kubeArgs.APIServerEndpoint,
		Token:                    kubeArgs.Token,
		DiscoveryTokenCaCertHash: kubeArgs.CACertHashes[0],
		KubeletCert:              kubeletCert,
		ControlPlaneFiles:        controlPlaneFiles,
		KubernetesVersion:        k8sVersion,
		KubernetesComponents:     components.ToJoinProto(),
	}, nil
}

// IssueRejoinTicket issues a ticket for nodes to rejoin cluster.
func (s *Server) IssueRejoinTicket(ctx context.Context, req *joinproto.IssueRejoinTicketRequest) (*joinproto.IssueRejoinTicketResponse, error) {
	log := s.log.With(zap.String("peerAddress", grpclog.PeerAddrFromContext(ctx)))
	log.Infof("IssueRejoinTicket called")

	log.Infof("Requesting measurement secret")
	measurementSecret, err := s.dataKeyGetter.GetDataKey(ctx, attestation.MeasurementSecretContext, crypto.DerivedKeyLengthDefault)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Unable to get measurement secret")
		return nil, status.Errorf(codes.Internal, "unable to get measurement secret: %s", err)
	}

	log.Infof("Requesting disk encryption key")
	stateDiskKey, err := s.dataKeyGetter.GetDataKey(ctx, req.DiskUuid, crypto.StateDiskKeyLength)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Unable to get key for stateful disk")
		return nil, status.Errorf(codes.Internal, "unable to get key for stateful disk: %s", err)
	}

	return &joinproto.IssueRejoinTicketResponse{
		StateDiskKey:      stateDiskKey,
		MeasurementSecret: measurementSecret,
	}, nil
}

// getK8sVersion reads the k8s version from a VolumeMount that is backed by the k8s-version ConfigMap.
func (s *Server) getK8sVersion() (string, error) {
	fileContent, err := s.file.Read(filepath.Join(constants.ServiceBasePath, constants.K8sVersionConfigMapName))
	if err != nil {
		return "", fmt.Errorf("could not read k8s version file: %w", err)
	}
	k8sVersion := string(fileContent)

	return k8sVersion, nil
}

// getK8sComponentsConfigMapName reads the k8s components config map name from a VolumeMount that is backed by the k8s-version ConfigMap.
func (s *Server) getK8sComponentsConfigMapName() (string, error) {
	fileContent, err := s.file.Read(filepath.Join(constants.ServiceBasePath, constants.K8sComponentsFieldName))
	if err != nil {
		return "", fmt.Errorf("could not read k8s version file: %w", err)
	}
	componentsConfigMapName := string(fileContent)

	return componentsConfigMapName, nil
}

// This function mimics the creation of the components ConfigMap which is now done in the bootstrapper
// during the first initialization of the Constellation .
// For more information see setupK8sVersionConfigMap() in bootstrapper/internal/kubernetes/kubernetes.go.
// This is a migration step and will be removed in a future release.
func (s *Server) createFallbackComponentsConfigMap(ctx context.Context, k8sVersion string) (string, error) {
	validK8sVersion, err := versions.NewValidK8sVersion(k8sVersion)
	if err != nil {
		return "", fmt.Errorf("could not create fallback components config map: %w", err)
	}
	components := versions.VersionConfigs[validK8sVersion].KubernetesComponents
	componentsMarshalled, err := json.Marshal(components)
	if err != nil {
		return "", fmt.Errorf("marshalling component versions: %w", err)
	}
	componentsHash := components.GetHash()
	componentConfigMapName := fmt.Sprintf("k8s-component-%s", strings.ReplaceAll(componentsHash, ":", "-"))

	componentsConfig := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		Immutable: to.Ptr(true),
		ObjectMeta: metav1.ObjectMeta{
			Name:      componentConfigMapName,
			Namespace: "kube-system",
		},
		Data: map[string]string{
			constants.K8sComponentsFieldName: string(componentsMarshalled),
		},
	}

	if err := s.kubeClient.CreateConfigMap(ctx, componentsConfig); err != nil {
		return "", fmt.Errorf("creating fallback components config map: %w", err)
	}

	if err := s.kubeClient.AddReferenceToK8sVersionConfigMap(ctx, "k8s-version", componentConfigMapName); err != nil {
		return "", fmt.Errorf("adding reference to fallback components config map: %w", err)
	}

	return componentConfigMapName, nil
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
	GetCertificate(certificateRequest []byte) (kubeletCert []byte, err error)
	// GetNodeNameFromCSR returns the node name from the CSR.
	GetNodeNameFromCSR(csr []byte) (string, error)
}

type kubeClient interface {
	GetComponents(ctx context.Context, configMapName string) (versions.ComponentVersions, error)
	CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error
	AddNodeToJoiningNodes(ctx context.Context, nodeName string, componentsHash string, isControlPlane bool) error
	AddReferenceToK8sVersionConfigMap(ctx context.Context, k8sVersionsConfigMapName string, componentsConfigMapName string) error
}

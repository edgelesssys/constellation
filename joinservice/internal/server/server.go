/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package server implements the gRPC endpoint of Constellation's node join service.
package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// Server implements the core logic of Constellation's node join service.
type Server struct {
	measurementSalt []byte

	log             *logger.Logger
	joinTokenGetter joinTokenGetter
	dataKeyGetter   dataKeyGetter
	ca              certificateAuthority
	kubeClient      kubeClient
	joinproto.UnimplementedAPIServer
}

// New initializes a new Server.
func New(
	measurementSalt []byte, ca certificateAuthority,
	joinTokenGetter joinTokenGetter, dataKeyGetter dataKeyGetter, kubeClient kubeClient, log *logger.Logger,
) (*Server, error) {
	return &Server{
		measurementSalt: measurementSalt,
		log:             log,
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
		log.With(zap.Error(err)).Errorf("Failed to get measurement secret")
		return nil, status.Errorf(codes.Internal, "getting measurement secret: %s", err)
	}

	log.Infof("Requesting disk encryption key")
	stateDiskKey, err := s.dataKeyGetter.GetDataKey(ctx, req.DiskUuid, crypto.StateDiskKeyLength)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to get key for stateful disk")
		return nil, status.Errorf(codes.Internal, "getting key for stateful disk: %s", err)
	}

	log.Infof("Creating Kubernetes join token")
	kubeArgs, err := s.joinTokenGetter.GetJoinToken(constants.KubernetesJoinTokenTTL)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to generate Kubernetes join arguments")
		return nil, status.Errorf(codes.Internal, "generating Kubernetes join arguments: %s", err)
	}

	log.Infof("Querying NodeVersion custom resource for components ConfigMap name")
	componentsConfigMapName, err := s.getK8sComponentsConfigMapName(ctx)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed getting components ConfigMap name")
		return nil, status.Errorf(codes.Internal, "getting components ConfigMap name: %s", err)
	}

	log.Infof("Querying %s ConfigMap for components", componentsConfigMapName)
	components, err := s.kubeClient.GetComponents(ctx, componentsConfigMapName)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed getting components from ConfigMap")
		return nil, status.Errorf(codes.Internal, "getting components: %s", err)
	}

	log.Infof("Creating signed kubelet certificate")
	kubeletCert, err := s.ca.GetCertificate(req.CertificateRequest)
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed generating kubelet certificate")
		return nil, status.Errorf(codes.Internal, "Generating kubelet certificate: %s", err)
	}

	var controlPlaneFiles []*joinproto.ControlPlaneCertOrKey
	if req.IsControlPlane {
		log.Infof("Loading control plane certificates and keys")
		filesMap, err := s.joinTokenGetter.GetControlPlaneCertificatesAndKeys()
		if err != nil {
			log.With(zap.Error(err)).Errorf("Failed to load control plane certificates and keys")
			return nil, status.Errorf(codes.Internal, "loading control-plane certificates and keys: %s", err)
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
		log.With(zap.Error(err)).Errorf("Failed getting node name from CSR")
		return nil, status.Errorf(codes.Internal, "getting node name from CSR: %s", err)
	}

	if err := s.kubeClient.AddNodeToJoiningNodes(ctx, nodeName, componentsConfigMapName, req.IsControlPlane); err != nil {
		log.With(zap.Error(err)).Errorf("Failed adding node to joining nodes")
		return nil, status.Errorf(codes.Internal, "adding node to joining nodes: %s", err)
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

	log.Infof("IssueRejoinTicket successful")
	return &joinproto.IssueRejoinTicketResponse{
		StateDiskKey:      stateDiskKey,
		MeasurementSecret: measurementSecret,
	}, nil
}

// getK8sComponentsConfigMapName reads the k8s components config map name from a VolumeMount that is backed by the k8s-version ConfigMap.
func (s *Server) getK8sComponentsConfigMapName(ctx context.Context) (string, error) {
	k8sComponentsRef, err := s.kubeClient.GetK8sComponentsRefFromNodeVersionCRD(ctx, "constellation-version")
	if err != nil {
		return "", fmt.Errorf("could not get k8s components config map name: %w", err)
	}
	return k8sComponentsRef, nil
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
	GetK8sComponentsRefFromNodeVersionCRD(ctx context.Context, nodeName string) (string, error)
	GetComponents(ctx context.Context, configMapName string) (components.Components, error)
	AddNodeToJoiningNodes(ctx context.Context, nodeName string, componentsHash string, isControlPlane bool) error
}

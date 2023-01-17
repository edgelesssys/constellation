/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package server implements the gRPC server for the upgrade agent.

The server is responsible for using kubeadm to upgrade the Kubernetes
release of a Constellation node.
*/
package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/installer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/upgrade-agent/upgradeproto"
	"golang.org/x/mod/semver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errInvalidKubernetesVersion = errors.New("invalid kubernetes version")

// Server is the upgrade-agent server.
type Server struct {
	file       file.Handler
	grpcServer serveStopper
	log        *logger.Logger
	upgradeproto.UnimplementedUpdateServer
}

// New creates a new upgrade-agent server.
func New(log *logger.Logger, fileHandler file.Handler) (*Server, error) {
	log = log.Named("upgradeServer")

	server := &Server{
		log:  log,
		file: fileHandler,
	}

	grpcServer := grpc.NewServer(
		log.Named("gRPC").GetServerUnaryInterceptor(),
	)
	upgradeproto.RegisterUpdateServer(grpcServer, server)

	server.grpcServer = grpcServer
	return server, nil
}

// Run starts the upgrade-agent server on the given port, using the provided protocol and socket address.
func (s *Server) Run(protocol string, sockAddr string) error {
	grpcServer := grpc.NewServer()

	upgradeproto.RegisterUpdateServer(grpcServer, s)

	cleanup := func() error {
		err := os.RemoveAll(sockAddr)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}
		return nil
	}
	if err := cleanup(); err != nil {
		return fmt.Errorf("failed to clean socket file: %s", err)
	}

	lis, err := net.Listen(protocol, sockAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}

	s.log.Infof("Starting")
	return grpcServer.Serve(lis)
}

// Stop stops the upgrade-agent server gracefully.
func (s *Server) Stop() {
	s.log.Infof("Stopping")

	s.grpcServer.GracefulStop()

	s.log.Infof("Stopped")
}

// ExecuteUpdate installs & verifies the provided kubeadm, then executes `kubeadm upgrade plan` & `kubeadm upgrade apply {wanted_Kubernetes_Version}` to upgrade to the specified version.
func (s *Server) ExecuteUpdate(ctx context.Context, updateRequest *upgradeproto.ExecuteUpdateRequest) (*upgradeproto.ExecuteUpdateResponse, error) {
	s.log.Infof("Upgrade to Kubernetes version started: %s", updateRequest.WantedKubernetesVersion)

	installer := installer.NewOSInstaller()
	err := prepareUpdate(ctx, installer, updateRequest)
	if errors.Is(err, errInvalidKubernetesVersion) {
		return nil, status.Errorf(codes.Internal, "unable to verify the Kubernetes version %s: %s", updateRequest.WantedKubernetesVersion, err)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to install the kubeadm binary: %s", err)
	}

	upgradeCmd := exec.CommandContext(ctx, "kubeadm", "upgrade", "plan")
	if out, err := upgradeCmd.CombinedOutput(); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to execute kubeadm upgrade plan: %s: %s", err, string(out))
	}

	applyCmd := exec.CommandContext(ctx, "kubeadm", "upgrade", "apply", "--yes", updateRequest.WantedKubernetesVersion)
	if out, err := applyCmd.CombinedOutput(); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to execute kubeadm upgrade apply: %s: %s", err, string(out))
	}

	s.log.Infof("Upgrade to Kubernetes version succeeded: %s", updateRequest.WantedKubernetesVersion)
	return &upgradeproto.ExecuteUpdateResponse{}, nil
}

// prepareUpdate downloads & installs the specified kubeadm version and verifies the desired Kubernetes version.
func prepareUpdate(ctx context.Context, installer osInstaller, updateRequest *upgradeproto.ExecuteUpdateRequest) error {
	// verify Kubernetes version
	err := verifyVersion(updateRequest.WantedKubernetesVersion)
	if err != nil {
		return err
	}

	// download & install the kubeadm binary
	return installer.Install(ctx, components.Component{
		URL:         updateRequest.KubeadmUrl,
		Hash:        updateRequest.KubeadmHash,
		InstallPath: constants.KubeadmPath,
		Extract:     false,
	})
}

// verifyVersion verifies the provided Kubernetes version.
func verifyVersion(version string) error {
	if !semver.IsValid(version) {
		return errInvalidKubernetesVersion
	}
	return nil
}

type osInstaller interface {
	// Install downloads, installs and verifies the kubernetes component.
	Install(ctx context.Context, kubernetesComponent components.Component) error
}

type serveStopper interface {
	// Serve starts the server.
	Serve(lis net.Listener) error
	// GracefulStop stops the server and blocks until all requests are done.
	GracefulStop()
}

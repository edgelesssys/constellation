package main

import (
	"net"
	"sync"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/internal/initserver"
	"github.com/edgelesssys/constellation/coordinator/internal/joinclient"
	"github.com/edgelesssys/constellation/coordinator/logging"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

var version = "0.0.0"

func run(issuer core.QuoteIssuer, tpm vtpm.TPMOpenFunc, fileHandler file.Handler,
	kube ClusterInitJoiner, metadata core.ProviderMetadata,
	bindIP, bindPort string, logger *zap.Logger,
	cloudLogger logging.CloudLogger, fs afero.Fs,
) {
	defer logger.Sync()
	logger.Info("starting coordinator", zap.String("version", version))

	defer cloudLogger.Close()
	cloudLogger.Disclose("Coordinator started running...")

	nodeActivated, err := vtpm.IsNodeInitialized(tpm)
	if err != nil {
		logger.Fatal("failed to check for previous activation using vTPM", zap.Error(err))
	}

	if nodeActivated {
		if err := kube.StartKubelet(); err != nil {
			logger.Fatal("failed to restart kubelet", zap.Error(err))
		}
		return
	}

	nodeLock := &sync.Mutex{}
	initServer := initserver.New(nodeLock, kube, logger)

	dialer := dialer.New(issuer, nil, &net.Dialer{})
	joinClient := joinclient.New(nodeLock, dialer, kube, metadata, logger)

	joinClient.Start()
	defer joinClient.Stop()

	if err := initServer.Serve(bindIP, bindPort); err != nil {
		logger.Error("Failed to serve init server", zap.Error(err))
	}
}

type ClusterInitJoiner interface {
	joinclient.ClusterJoiner
	initserver.ClusterInitializer
	StartKubelet() error
}

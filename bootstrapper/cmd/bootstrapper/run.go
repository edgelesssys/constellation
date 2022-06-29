package main

import (
	"net"

	"github.com/edgelesssys/constellation/bootstrapper/internal/initserver"
	"github.com/edgelesssys/constellation/bootstrapper/internal/joinclient"
	"github.com/edgelesssys/constellation/bootstrapper/internal/logging"
	"github.com/edgelesssys/constellation/bootstrapper/internal/nodelock"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

var version = "0.0.0"

func run(issuer quoteIssuer, tpm vtpm.TPMOpenFunc, fileHandler file.Handler,
	kube clusterInitJoiner, metadata joinclient.MetadataAPI,
	bindIP, bindPort string, logger *zap.Logger,
	cloudLogger logging.CloudLogger, fs afero.Fs,
) {
	defer logger.Sync()
	logger.Info("starting bootstrapper", zap.String("version", version))

	defer cloudLogger.Close()
	cloudLogger.Disclose("bootstrapper started running...")

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

	nodeLock := nodelock.New()
	initServer := initserver.New(nodeLock, kube, logger)

	dialer := dialer.New(issuer, nil, &net.Dialer{})
	joinClient := joinclient.New(nodeLock, dialer, kube, metadata, logger)

	joinClient.Start()
	defer joinClient.Stop()

	if err := initServer.Serve(bindIP, bindPort); err != nil {
		logger.Error("Failed to serve init server", zap.Error(err))
	}
}

type clusterInitJoiner interface {
	joinclient.ClusterJoiner
	initserver.ClusterInitializer
	StartKubelet() error
}

type quoteIssuer interface {
	oid.Getter
	// Issue issues a quote for remote attestation for a given message
	Issue(userData []byte, nonce []byte) (quote []byte, err error)
}

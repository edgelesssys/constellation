package main

import (
	"net"

	"github.com/edgelesssys/constellation/bootstrapper/internal/exit"
	"github.com/edgelesssys/constellation/bootstrapper/internal/initserver"
	"github.com/edgelesssys/constellation/bootstrapper/internal/joinclient"
	"github.com/edgelesssys/constellation/bootstrapper/internal/logging"
	"github.com/edgelesssys/constellation/bootstrapper/internal/nodelock"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/oid"
	"go.uber.org/zap"
)

var version = "0.0.0"

func run(issuer quoteIssuer, tpm vtpm.TPMOpenFunc, fileHandler file.Handler,
	kube clusterInitJoiner, metadata joinclient.MetadataAPI,
	bindIP, bindPort string, logger *zap.Logger,
	cloudLogger logging.CloudLogger,
) {
	defer logger.Sync()
	defer cloudLogger.Close()

	logger.Info("starting bootstrapper", zap.String("version", version))
	cloudLogger.Disclose("bootstrapper started running...")

	nodeBootstrapped, err := vtpm.IsNodeBootstrapped(tpm)
	if err != nil {
		logger.Fatal("failed to check for previous bootstrapping using vTPM", zap.Error(err))
	}

	if nodeBootstrapped {
		if err := kube.StartKubelet(); err != nil {
			logger.Fatal("failed to restart kubelet", zap.Error(err))
		}
		return
	}

	nodeLock := nodelock.New(tpm)
	initServer := initserver.New(nodeLock, kube, issuer, fileHandler, logger)

	dialer := dialer.New(issuer, nil, &net.Dialer{})
	joinClient := joinclient.New(nodeLock, dialer, kube, metadata, logger)

	cleaner := exit.New().
		With(initServer).
		With(joinClient)
	joinClient.Start(cleaner)

	if err := initServer.Serve(bindIP, bindPort, cleaner); err != nil {
		logger.Error("Failed to serve init server", zap.Error(err))
	}

	// wait for join client and server to exit cleanly
	cleaner.Clean()

	// if node lock was never acquired, then we didn't bootstrap successfully.
	if !nodeLock.Locked() {
		cloudLogger.Disclose("bootstrapper failed")
		logger.Fatal("bootstrapper failed")
	}

	logger.Info("bootstrapper done")
	cloudLogger.Disclose("bootstrapper done")
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

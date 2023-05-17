/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"net"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/clean"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/diskencryption"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/initserver"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/joinclient"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/logging"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/nodelock"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/initialize"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
)

func run(issuer atls.Issuer, openDevice vtpm.TPMOpenFunc, fileHandler file.Handler,
	kube clusterInitJoiner, metadata metadataAPI,
	bindIP, bindPort string, log *logger.Logger,
	cloudLogger logging.CloudLogger,
) {
	defer cloudLogger.Close()

	log.With(zap.String("version", constants.VersionInfo())).Infof("Starting bootstrapper")
	cloudLogger.Disclose("bootstrapper started running...")

	uuid, err := getDiskUUID()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to get disk UUID")
		cloudLogger.Disclose("Failed to get disk UUID")
	} else {
		log.Infof("Disk UUID: %s", uuid)
		cloudLogger.Disclose("Disk UUID: " + uuid)
	}

	nodeBootstrapped, err := initialize.IsNodeBootstrapped(openDevice)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to check if node was previously bootstrapped")
	}

	if nodeBootstrapped {
		if err := kube.StartKubelet(log); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to restart kubelet")
		}
		return
	}

	nodeLock := nodelock.New(openDevice)
	initServer, err := initserver.New(context.Background(), nodeLock, kube, issuer, fileHandler, metadata, log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create init server")
	}

	dialer := dialer.New(issuer, nil, &net.Dialer{})
	joinClient := joinclient.New(nodeLock, dialer, kube, metadata, log)

	cleaner := clean.New().With(initServer).With(joinClient)
	go cleaner.Start()
	defer cleaner.Done()

	joinClient.Start(cleaner)

	if err := initServer.Serve(bindIP, bindPort, cleaner); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to serve init server")
	}

	log.Infof("bootstrapper done")
	cloudLogger.Disclose("bootstrapper done")
}

func getDiskUUID() (string, error) {
	disk := diskencryption.New()
	if err := disk.Open(); err != nil {
		return "", err
	}
	defer disk.Close()
	return disk.UUID()
}

type clusterInitJoiner interface {
	joinclient.ClusterJoiner
	initserver.ClusterInitializer
	StartKubelet(*logger.Logger) error
}

type metadataAPI interface {
	joinclient.MetadataAPI
	initserver.MetadataAPI
	GetLoadBalancerEndpoint(ctx context.Context) (string, error)
}

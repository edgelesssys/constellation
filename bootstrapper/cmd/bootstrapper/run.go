/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
  "context"
  "log/slog"
  "net"
  "os"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/clean"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/diskencryption"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/initserver"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/joinclient"
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
) {
	log.With(zap.String("version", constants.BinaryVersion().String())).Infof("Starting bootstrapper")

	uuid, err := getDiskUUID()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to get disk UUID")
	} else {
		log.Infof("Disk UUID: %s", uuid)
	}

  nodeBootstrapped, err := initialize.IsNodeBootstrapped(openDevice)
  if err != nil {
    log.With(slog.Any("error", err)).Error("Failed to check if node was previously bootstrapped")
    os.Exit(1)
  }

  if nodeBootstrapped {
    if err := kube.StartKubelet(); err != nil {
      log.With(slog.Any("error", err)).Error("Failed to restart kubelet")
      os.Exit(1)
    }
    return
  }

  nodeLock := nodelock.New(openDevice)
  initServer, err := initserver.New(context.Background(), nodeLock, kube, issuer, fileHandler, metadata, log)
  if err != nil {
    log.With(slog.Any("error", err)).Error("Failed to create init server")
    os.Exit(1)
  }

  dialer := dialer.New(issuer, nil, &net.Dialer{})
  joinClient := joinclient.New(nodeLock, dialer, kube, metadata, log)

  cleaner := clean.New().With(initServer).With(joinClient)
  go cleaner.Start()
  defer cleaner.Done()

  joinClient.Start(cleaner)

  if err := initServer.Serve(bindIP, bindPort, cleaner); err != nil {
    log.With(slog.Any("error", err)).Error("Failed to serve init server")
    os.Exit(1)
  }

	log.Infof("bootstrapper done")
}

func getDiskUUID() (string, error) {
  disk := diskencryption.New()
  free, err := disk.Open()
  if err != nil {
    return "", err
  }
  defer free()
  return disk.UUID()
}

type clusterInitJoiner interface {
  joinclient.ClusterJoiner
  initserver.ClusterInitializer
  StartKubelet() error
}

type metadataAPI interface {
  joinclient.MetadataAPI
  initserver.MetadataAPI
}

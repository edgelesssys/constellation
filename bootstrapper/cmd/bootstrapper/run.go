/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"syscall"

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
)

func run(issuer atls.Issuer, openDevice vtpm.TPMOpenFunc, fileHandler file.Handler,
	kube clusterInitJoiner, metadata metadataAPI,
	bindIP, bindPort string, log *slog.Logger,
) {
	log.With(slog.String("version", constants.BinaryVersion().String())).Info("Starting bootstrapper")

	uuid, err := getDiskUUID()
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to get disk UUID")
	} else {
		log.Info(fmt.Sprintf("Disk UUID: %s", uuid))
	}

	nodeBootstrapped, err := initialize.IsNodeBootstrapped(openDevice)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to check if node was previously bootstrapped")
		reboot()
	}

	if nodeBootstrapped {
		if err := kube.StartKubelet(); err != nil {
			log.With(slog.Any("error", err)).Error("Failed to restart kubelet")
			reboot()
		}
		return
	}

	nodeLock := nodelock.New(openDevice)
	initServer, err := initserver.New(context.Background(), nodeLock, kube, issuer, fileHandler, metadata, log)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to create init server")
		reboot()
	}

	dialer := dialer.New(issuer, nil, &net.Dialer{})
	joinClient := joinclient.New(nodeLock, dialer, kube, metadata, log)

	cleaner := clean.New().With(initServer).With(joinClient)
	go cleaner.Start()
	defer cleaner.Done()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := joinClient.Start(cleaner); err != nil {
			log.With(slog.Any("error", err)).Error("Failed to join cluster")
			reboot()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := initServer.Serve(bindIP, bindPort, cleaner); err != nil {
			log.With(slog.Any("error", err)).Error("Failed to serve init server")
			reboot()
		}
	}()
	wg.Wait()

	log.Info("bootstrapper done")
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

// reboot the system.
// We call this instead of os.Exit() since failures in the bootstrapper usually require a node reset.
func reboot() {
	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
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

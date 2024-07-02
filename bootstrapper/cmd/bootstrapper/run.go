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

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/clean"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/diskencryption"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/initserver"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/joinclient"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/nodelock"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/reboot"
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

	disk := diskencryption.New()
	uuid, err := getDiskUUID(disk)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to get disk UUID")
	} else {
		log.Info(fmt.Sprintf("Disk UUID: %s", uuid))
	}

	nodeBootstrapped, err := initialize.IsNodeBootstrapped(openDevice)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to check if node was previously bootstrapped")
		reboot.Reboot(fmt.Errorf("checking if node was previously bootstrapped: %w", err))
	}

	if nodeBootstrapped {
		if err := kube.StartKubelet(); err != nil {
			log.With(slog.Any("error", err)).Error("Failed to restart kubelet")
			reboot.Reboot(fmt.Errorf("restarting kubelet: %w", err))
		}
		return
	}

	nodeLock := nodelock.New(openDevice)
	initServer, err := initserver.New(context.Background(), nodeLock, kube, issuer, disk, fileHandler, metadata, log)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to create init server")
		reboot.Reboot(fmt.Errorf("creating init server: %w", err))
	}

	dialer := dialer.New(issuer, nil, &net.Dialer{})
	joinClient := joinclient.New(nodeLock, dialer, kube, metadata, disk, log)

	cleaner := clean.New().With(initServer).With(joinClient)
	go cleaner.Start()
	defer cleaner.Done()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := joinClient.Start(cleaner); err != nil {
			log.With(slog.Any("error", err)).Error("Failed to join cluster")
			markDiskForReset(disk)
			reboot.Reboot(fmt.Errorf("joining cluster: %w", err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := initServer.Serve(bindIP, bindPort, cleaner); err != nil {
			log.With(slog.Any("error", err)).Error("Failed to serve init server")
			markDiskForReset(disk)
			reboot.Reboot(fmt.Errorf("serving init server: %w", err))
		}
	}()
	wg.Wait()

	log.Info("bootstrapper done")
}

func getDiskUUID(disk *diskencryption.DiskEncryption) (string, error) {
	free, err := disk.Open()
	if err != nil {
		return "", err
	}
	defer free()
	return disk.UUID()
}

// markDiskForReset sets a token in the cryptsetup header of the disk to indicate the disk should be reset on next boot.
// This is used to reset all state of a node in case the bootstrapper encountered a non recoverable error
// after the node successfully retrieved a join ticket from the JoinService.
// As setting this token is safe as long as we are certain we don't need the data on the disk anymore, we call this
// unconditionally when either the JoinClient or the InitServer encounter an error.
// We don't call it before that, as the node may be restarting after a previous, successful bootstrapping,
// and now encountered a transient error on rejoining the cluster. Wiping the disk now would delete existing data.
func markDiskForReset(disk *diskencryption.DiskEncryption) {
	free, err := disk.Open()
	if err != nil {
		return
	}
	defer free()
	_ = disk.MarkDiskForReset()
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

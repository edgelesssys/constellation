/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/diskencryption"
	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/recoveryserver"
	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/rejoinclient"
	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/setup"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/tdx"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	awscloud "github.com/edgelesssys/constellation/v2/internal/cloud/aws"
	azurecloud "github.com/edgelesssys/constellation/v2/internal/cloud/azure"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	gcpcloud "github.com/edgelesssys/constellation/v2/internal/cloud/gcp"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	qemucloud "github.com/edgelesssys/constellation/v2/internal/cloud/qemu"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	kmssetup "github.com/edgelesssys/constellation/v2/internal/kms/setup"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/spf13/afero"
)

const (
	gcpStateDiskPath       = "/dev/disk/by-id/google-state-disk"
	azureStateDiskPath     = "/dev/disk/azure/scsi1/lun0"
	awsStateDiskPath       = "/dev/sdb"
	qemuStateDiskPath      = "/dev/vdb"
	openstackStateDiskPath = "/dev/vdb"
)

func main() {
	csp := flag.String("csp", "", "Cloud Service Provider the image is running on")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)

	flag.Parse()
	log := logger.NewJSONLogger(logger.VerbosityFromInt(*verbosity))
	log.With(slog.String("version", constants.BinaryVersion().String()), slog.String("cloudProvider", *csp)).
		Info("Starting disk-mapper")

	// set up quote issuer for aTLS connections
	attestVariant, err := variant.FromString(os.Getenv(constants.AttestationVariant))
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to parse attestation variant")
		os.Exit(1)
	}
	issuer, err := choose.Issuer(attestVariant, log)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to select issuer")
		os.Exit(1)
	}

	// set up metadata API
	var diskPath string
	var metadataClient setup.MetadataAPI
	switch cloudprovider.FromString(*csp) {
	case cloudprovider.AWS:
		// on AWS Nitro platform, disks are attached over NVMe
		// using udev rules, a symlink for our disk is created at /dev/sdb
		diskPath, err = filepath.EvalSymlinks(awsStateDiskPath)
		if err != nil {
			log.With(slog.Any("error", err)).Error("Unable to resolve Azure state disk path")
			os.Exit(1)
		}
		metadataClient, err = awscloud.New(context.Background())
		if err != nil {
			log.With(slog.Any("error", err)).Error("Failed to set up AWS metadata client")
			os.Exit(1)
		}

	case cloudprovider.Azure:
		diskPath, err = filepath.EvalSymlinks(azureStateDiskPath)
		if err != nil {
			log.With(slog.Any("error", err)).Error("Unable to resolve Azure state disk path")
			os.Exit(1)
		}
		metadataClient, err = azurecloud.New(context.Background())
		if err != nil {
			log.With(slog.Any("error", err)).Error("Failed to set up Azure metadata client")
			os.Exit(1)
		}

	case cloudprovider.GCP:
		diskPath, err = filepath.EvalSymlinks(gcpStateDiskPath)
		if err != nil {
			log.With(slog.Any("error", err)).Error("Unable to resolve GCP state disk path")
			os.Exit(1)
		}
		gcpMeta, err := gcpcloud.New(context.Background())
		if err != nil {
			log.With(slog.Any("error", err)).Error(("Failed to create GCP metadata client"))
			os.Exit(1)
		}
		defer gcpMeta.Close()
		metadataClient = gcpMeta

	case cloudprovider.OpenStack:
		diskPath = openstackStateDiskPath
		metadataClient, err = openstack.New(context.Background())
		if err != nil {
			log.With(slog.Any("error", err)).Error(("Failed to create OpenStack metadata client"))
			os.Exit(1)
		}

	case cloudprovider.QEMU:
		diskPath = qemuStateDiskPath
		metadataClient = qemucloud.New()

	default:
		log.Error(fmt.Sprintf("CSP %s is not supported by Constellation", *csp))
		os.Exit(1)
	}

	// initialize device mapper
	mapper, free, err := diskencryption.New(diskPath, log)
	if err != nil {
		log.With(slog.Any("error", err)).Error(("Failed to initialize device mapper"))
		os.Exit(1)
	}
	defer free()

	// Use TDX if available
	openDevice := vtpm.OpenVTPM
	if attestVariant.OID().Equal(variant.QEMUTDX{}.OID()) {
		openDevice = func() (io.ReadWriteCloser, error) {
			return tdx.Open()
		}
	}
	setupManger := setup.New(
		log.WithGroup("setupManager"),
		*csp,
		diskPath,
		afero.Afero{Fs: afero.NewOsFs()},
		mapper,
		setup.DiskMounter{},
		openDevice,
	)

	if err := setupManger.LogDevices(); err != nil {
		log.With(slog.Any("error", err)).Error(("Failed to log devices"))
		os.Exit(1)
	}

	// prepare the state disk
	if mapper.IsInitialized() {
		// set up rejoin client
		var self metadata.InstanceMetadata
		self, err = metadataClient.Self(context.Background())
		if err != nil {
			log.With(slog.Any("error", err)).Error(("Failed to get self metadata"))
			os.Exit(1)
		}
		rejoinClient := rejoinclient.New(
			dialer.New(issuer, nil, &net.Dialer{}),
			self,
			metadataClient,
			log.WithGroup("rejoinClient"),
		)

		// set up recovery server if control-plane node
		var recoveryServer setup.RecoveryServer
		if self.Role == role.ControlPlane {
			recoveryServer = recoveryserver.New(issuer, kmssetup.KMS, log.WithGroup("recoveryServer"))
		} else {
			recoveryServer = recoveryserver.NewStub(log.WithGroup("recoveryServer"))
		}

		err = setupManger.PrepareExistingDisk(setup.NewNodeRecoverer(recoveryServer, rejoinClient))
	} else {
		err = setupManger.PrepareNewDisk()
	}
	if err != nil {
		log.With(slog.Any("error", err)).Error(("Failed to prepare state disk"))
		os.Exit(1)
	}
}

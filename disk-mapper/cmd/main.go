/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"flag"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/mapper"
	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/recoveryserver"
	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/rejoinclient"
	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/setup"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/tdx"
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
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/spf13/afero"
	"go.uber.org/zap"
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
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))
	log.With(zap.String("version", constants.VersionInfo()), zap.String("cloudProvider", *csp)).
		Infof("Starting disk-mapper")

	// set up quote issuer for aTLS connections
	attestVariant, err := variant.FromString(os.Getenv(constants.AttestationVariant))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to parse attestation variant")
	}
	issuer, err := choose.Issuer(attestVariant, log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to select issuer")
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
			log.With(zap.Error(err)).Fatalf("Unable to resolve Azure state disk path")
		}
		metadataClient, err = awscloud.New(context.Background())
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to set up AWS metadata client")
		}

	case cloudprovider.Azure:
		diskPath, err = filepath.EvalSymlinks(azureStateDiskPath)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Unable to resolve Azure state disk path")
		}
		metadataClient, err = azurecloud.New(context.Background())
		if err != nil {
			log.With(zap.Error).Fatalf("Failed to set up Azure metadata client")
		}

	case cloudprovider.GCP:
		diskPath, err = filepath.EvalSymlinks(gcpStateDiskPath)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Unable to resolve GCP state disk path")
		}
		gcpMeta, err := gcpcloud.New(context.Background())
		if err != nil {
			log.With(zap.Error).Fatalf("Failed to create GCP metadata client")
		}
		defer gcpMeta.Close()
		metadataClient = gcpMeta

	case cloudprovider.OpenStack:
		diskPath = openstackStateDiskPath
		metadataClient, err = openstack.New(context.Background())
		if err != nil {
			log.With(zap.Error).Fatalf("Failed to create OpenStack metadata client")
		}

	case cloudprovider.QEMU:
		diskPath = qemuStateDiskPath
		metadataClient = qemucloud.New()

	default:
		log.Fatalf("CSP %s is not supported by Constellation", *csp)
	}

	// initialize device mapper
	mapper, err := mapper.New(diskPath, log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to initialize device mapper")
	}
	defer mapper.Close()

	// Use TDX if available
	openDevice := vtpm.OpenVTPM
	if attestVariant.OID().Equal(variant.QEMUTDX{}.OID()) {
		openDevice = func() (io.ReadWriteCloser, error) {
			return tdx.Open()
		}
	}
	setupManger := setup.New(
		log.Named("setupManager"),
		*csp,
		diskPath,
		afero.Afero{Fs: afero.NewOsFs()},
		mapper,
		setup.DiskMounter{},
		openDevice,
	)

	if err := setupManger.LogDevices(); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to log devices")
	}

	// prepare the state disk
	if mapper.IsLUKSDevice() {
		// set up rejoin client
		var self metadata.InstanceMetadata
		self, err = metadataClient.Self(context.Background())
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get self metadata")
		}
		rejoinClient := rejoinclient.New(
			dialer.New(issuer, nil, &net.Dialer{}),
			self,
			metadataClient,
			log.Named("rejoinClient"),
		)

		// set up recovery server if control-plane node
		var recoveryServer setup.RecoveryServer
		if self.Role == role.ControlPlane {
			recoveryServer = recoveryserver.New(issuer, kmssetup.KMS, log.Named("recoveryServer"))
		} else {
			recoveryServer = recoveryserver.NewStub(log.Named("recoveryServer"))
		}

		err = setupManger.PrepareExistingDisk(setup.NewNodeRecoverer(recoveryServer, rejoinClient))
	} else {
		err = setupManger.PrepareNewDisk()
	}
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to prepare state disk")
	}
}

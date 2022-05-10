package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/coordinator/attestation/azure"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/coordinator/attestation/qemu"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	azurecloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/gcp"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/state/keyservice"
	"github.com/edgelesssys/constellation/state/mapper"
	"github.com/edgelesssys/constellation/state/setup"
	"github.com/spf13/afero"
)

const (
	gcpStateDiskPath   = "/dev/disk/by-id/google-state-disk"
	azureStateDiskPath = "/dev/disk/azure/scsi1/lun0"
	qemuStateDiskPath  = "/dev/vda"
)

var csp = flag.String("csp", "", "Cloud Service Provider the image is running on")

func main() {
	flag.Parse()

	log.Printf("Starting disk-mapper for csp %q\n", *csp)

	// set up metadata API and quote issuer for aTLS connections
	var err error
	var diskPathErr error
	var diskPath string
	var issuer core.QuoteIssuer
	var metadata core.ProviderMetadata
	switch strings.ToLower(*csp) {
	case "azure":
		diskPath, diskPathErr = filepath.EvalSymlinks(azureStateDiskPath)
		metadata, err = azurecloud.NewMetadata(context.Background())
		if err != nil {
			exit(err)
		}
		issuer = azure.NewIssuer()

	case "gcp":
		diskPath, diskPathErr = filepath.EvalSymlinks(gcpStateDiskPath)
		issuer = gcp.NewIssuer()
		gcpClient, err := gcpcloud.NewClient(context.Background())
		if err != nil {
			exit(err)
		}
		metadata = gcpcloud.New(gcpClient)

	case "qemu":
		diskPath = qemuStateDiskPath
		issuer = qemu.NewIssuer()
		fmt.Fprintf(os.Stderr, "warning: cloud services are not supported for csp %q\n", *csp)
		metadata = &core.ProviderMetadataFake{}

	default:
		diskPathErr = fmt.Errorf("csp %q is not supported by Constellation", *csp)
	}
	if diskPathErr != nil {
		exit(fmt.Errorf("unable to determine state disk path: %w", diskPathErr))
	}

	// initialize device mapper
	mapper, err := mapper.New(diskPath)
	if err != nil {
		exit(err)
	}
	defer mapper.Close()

	setupManger := setup.New(
		*csp,
		afero.Afero{Fs: afero.NewOsFs()},
		keyservice.New(issuer, metadata, 20*time.Second), // try to request a key every 20 seconds
		mapper,
		setup.DiskMounter{},
		vtpm.OpenVTPM,
	)

	// prepare the state disk
	if mapper.IsLUKSDevice() {
		err = setupManger.PrepareExistingDisk()
	} else {
		err = setupManger.PrepareNewDisk()
	}
	exit(err)
}

func exit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

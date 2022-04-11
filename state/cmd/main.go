package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/coordinator/attestation/azure"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	azurecloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/gcp"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/internal/utils"
	"github.com/edgelesssys/constellation/state/keyservice"
	"github.com/edgelesssys/constellation/state/mapper"
	"github.com/edgelesssys/constellation/state/setup"
	"github.com/spf13/afero"
)

const (
	gcpStateDiskPath   = "/dev/disk/by-id/google-state-disk"
	azureStateDiskPath = "/dev/disk/azure/scsi1/lun0"
	fallBackPath       = "/dev/disk/by-id/state-disk"
)

var csp = flag.String("csp", "", "Cloud Service Provider the image is running on")

func main() {
	flag.Parse()

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
			utils.KernelPanic(err)
		}
		issuer = azure.NewIssuer()

	case "gcp":
		diskPath, diskPathErr = filepath.EvalSymlinks(gcpStateDiskPath)
		issuer = gcp.NewIssuer()
		gcpClient, err := gcpcloud.NewClient(context.Background())
		if err != nil {
			utils.KernelPanic(err)
		}
		metadata = gcpcloud.New(gcpClient)

	default:
		diskPath, err = filepath.EvalSymlinks(fallBackPath)
		if err != nil {
			utils.KernelPanic(err)
		}
		issuer = core.NewMockIssuer()
		fmt.Fprintf(os.Stderr, "warning: csp %q is not supported, unable to automatically request decryption keys on reboot\n", *csp)
		metadata = &core.ProviderMetadataFake{}
	}
	if diskPathErr != nil {
		fmt.Fprintf(os.Stderr, "warning: no attached disk detected, trying to use boot-disk state partition as fallback")
		diskPath, err = filepath.EvalSymlinks(fallBackPath)
		if err != nil {
			utils.KernelPanic(err)
		}
	}

	// initialize device mapper
	mapper, err := mapper.New(diskPath)
	if err != nil {
		utils.KernelPanic(err)
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

	if err != nil {
		utils.KernelPanic(err)
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	azurecloud "github.com/edgelesssys/constellation/bootstrapper/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/bootstrapper/cloudprovider/gcp"
	qemucloud "github.com/edgelesssys/constellation/bootstrapper/cloudprovider/qemu"
	"github.com/edgelesssys/constellation/internal/attestation/azure"
	"github.com/edgelesssys/constellation/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/state/internal/keyservice"
	"github.com/edgelesssys/constellation/state/internal/mapper"
	"github.com/edgelesssys/constellation/state/internal/setup"
	tpmClient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/tpm2"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

const (
	gcpStateDiskPath   = "/dev/disk/by-id/google-state-disk"
	azureStateDiskPath = "/dev/disk/azure/scsi1/lun0"
	qemuStateDiskPath  = "/dev/vda"
)

func main() {
	csp := flag.String("csp", "", "Cloud Service Provider the image is running on")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)

	flag.Parse()
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))
	log.With(zap.String("version", constants.VersionInfo), zap.String("cloudProvider", *csp)).
		Infof("Starting disk-mapper")

	// set up metadata API and quote issuer for aTLS connections
	var err error
	var diskPath string
	var issuer keyservice.QuoteIssuer
	var metadata metadata.InstanceLister
	switch strings.ToLower(*csp) {
	case "azure":
		diskPath, err = filepath.EvalSymlinks(azureStateDiskPath)
		if err != nil {
			_ = exportPCRs()
			log.With(zap.Error(err)).Fatalf("Unable to resolve Azure state disk path")
		}
		metadata, err = azurecloud.NewMetadata(context.Background())
		if err != nil {
			log.With(zap.Error).Fatalf("Failed to create Azure metadata API")
		}
		issuer = azure.NewIssuer()

	case "gcp":
		diskPath, err = filepath.EvalSymlinks(gcpStateDiskPath)
		if err != nil {
			_ = exportPCRs()
			log.With(zap.Error(err)).Fatalf("Unable to resolve GCP state disk path")
		}
		issuer = gcp.NewIssuer()
		gcpClient, err := gcpcloud.NewClient(context.Background())
		if err != nil {
			log.With(zap.Error).Fatalf("Failed to create GCP client")
		}
		metadata = gcpcloud.New(gcpClient)

	case "qemu":
		diskPath = qemuStateDiskPath
		issuer = qemu.NewIssuer()
		metadata = &qemucloud.Metadata{}

	default:
		log.Fatalf("CSP %s is not supported by Constellation", *csp)
	}

	// initialize device mapper
	mapper, err := mapper.New(diskPath, log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to initialize device mapper")
	}
	defer mapper.Close()

	setupManger := setup.New(
		log.Named("setupManager"),
		*csp,
		afero.Afero{Fs: afero.NewOsFs()},
		keyservice.New(log.Named("keyService"), issuer, metadata, 20*time.Second, 20*time.Second), // try to request a key every 20 seconds
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
		log.With(zap.Error(err)).Fatalf("Failed to prepare state disk")
	}
}

// exportPCRs tries to export the node's PCRs to QEMU's metadata API.
// This function is called when an Azure or GCP image boots, but is unable to find a state disk.
// This happens when we boot such an image in QEMU.
// We can use this to calculate the PCRs of the image locally.
func exportPCRs() error {
	// get TPM state
	pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, tpmClient.FullPcrSel(tpm2.AlgSHA256))
	if err != nil {
		return err
	}
	pcrsPretty, err := json.Marshal(pcrs)
	if err != nil {
		return err
	}

	// send PCRs to metadata API
	url := &url.URL{
		Scheme: "http",
		Host:   "10.42.0.1:8080", // QEMU metadata endpoint
		Path:   "/pcrs",
	}
	_, err = http.Post(url.String(), "application/json", bytes.NewBuffer(pcrsPretty))
	return err
}

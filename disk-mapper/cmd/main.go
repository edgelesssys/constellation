/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/disk-mapper/internal/mapper"
	"github.com/edgelesssys/constellation/disk-mapper/internal/recoveryserver"
	"github.com/edgelesssys/constellation/disk-mapper/internal/rejoinclient"
	"github.com/edgelesssys/constellation/disk-mapper/internal/setup"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/azure"
	"github.com/edgelesssys/constellation/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	azurecloud "github.com/edgelesssys/constellation/internal/cloud/azure"
	gcpcloud "github.com/edgelesssys/constellation/internal/cloud/gcp"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	qemucloud "github.com/edgelesssys/constellation/internal/cloud/qemu"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/logger"
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
	var issuer atls.Issuer
	var metadataAPI setup.MetadataAPI
	switch strings.ToLower(*csp) {
	case "azure":
		diskPath, err = filepath.EvalSymlinks(azureStateDiskPath)
		if err != nil {
			_ = exportPCRs()
			log.With(zap.Error(err)).Fatalf("Unable to resolve Azure state disk path")
		}
		metadataAPI, err = azurecloud.NewMetadata(context.Background())
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
		metadataAPI = gcpcloud.New(gcpClient)

	case "qemu":
		diskPath = qemuStateDiskPath
		issuer = qemu.NewIssuer()
		metadataAPI = &qemucloud.Metadata{}

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
		diskPath,
		afero.Afero{Fs: afero.NewOsFs()},
		mapper,
		setup.DiskMounter{},
		vtpm.OpenVTPM,
	)

	// prepare the state disk
	if mapper.IsLUKSDevice() {
		// set up rejoin client
		var self metadata.InstanceMetadata
		self, err = metadataAPI.Self(context.Background())
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get self metadata")
		}
		rejoinClient := rejoinclient.New(
			dialer.New(issuer, nil, &net.Dialer{}),
			self,
			metadataAPI,
			log.Named("rejoinClient"),
		)

		// set up recovery server
		recoveryServer := recoveryserver.New(issuer, log.Named("recoveryServer"))

		err = setupManger.PrepareExistingDisk(setup.NewNodeRecoverer(recoveryServer, rejoinClient))
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
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url.String(), bytes.NewBuffer(pcrsPretty))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

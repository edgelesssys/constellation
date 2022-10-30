/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"path/filepath"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	azurecloud "github.com/edgelesssys/constellation/v2/internal/cloud/azure"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	gcpcloud "github.com/edgelesssys/constellation/v2/internal/cloud/gcp"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	qemucloud "github.com/edgelesssys/constellation/v2/internal/cloud/qemu"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/watcher"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/kms"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/kubeadm"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/kubernetesca"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/server"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// vpcIPTimeout is the maximum amount of time to wait for retrieval of the VPC ip.
const vpcIPTimeout = 30 * time.Second

func main() {
	provider := flag.String("cloud-provider", "", "cloud service provider this binary is running on")
	kmsEndpoint := flag.String("kms-endpoint", "", "endpoint of Constellations key management service")
	// FIXME: define flag again once its definition no longer collides with glog.
	// This should happen as soon as https://github.com/google/go-sev-guest/issues/23 is merged and consumed by us.
	verbosity := 0 // flag.Int("v", 0, logger.CmdLineVerbosityDescription)
	flag.Parse()

	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(verbosity))
	log.With(zap.String("version", constants.VersionInfo), zap.String("cloudProvider", *provider)).
		Infof("Constellation Node Join Service")

	handler := file.NewHandler(afero.NewOsFs())

	cvmRaw, err := handler.Read(filepath.Join(constants.ServiceBasePath, constants.AzureCVM))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to get azureCVM from config map")
	}
	azureCVM, err := strconv.ParseBool(string(cvmRaw))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to parse content of AzureCVM: %s", cvmRaw)
	}

	validator, err := watcher.NewValidator(log.Named("validator"), *provider, handler, azureCVM)
	if err != nil {
		flag.Usage()
		log.With(zap.Error(err)).Fatalf("Failed to create validator")
	}

	creds := atlscredentials.New(nil, []atls.Validator{validator})

	ctx, cancel := context.WithTimeout(context.Background(), vpcIPTimeout)
	defer cancel()
	vpcIP, err := getVPCIP(ctx, *provider)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to get IP in VPC")
	}
	apiServerEndpoint := net.JoinHostPort(vpcIP, strconv.Itoa(constants.KubernetesPort))
	kubeadm, err := kubeadm.New(apiServerEndpoint, log.Named("kubeadm"))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create kubeadm")
	}
	kms := kms.New(log.Named("kms"), *kmsEndpoint)

	measurementSalt, err := handler.Read(filepath.Join(constants.ServiceBasePath, constants.MeasurementSaltFilename))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to read measurement salt")
	}

	server := server.New(
		measurementSalt,
		handler,
		kubernetesca.New(log.Named("certificateAuthority"), handler),
		kubeadm,
		kms,
		log.Named("server"),
	)

	watcher, err := watcher.New(log.Named("fileWatcher"), validator)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create watcher for measurements updates")
	}
	defer watcher.Close()

	go func() {
		log.Infof("starting file watcher for measurements file %s", filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename))
		if err := watcher.Watch(filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename)); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to watch measurements file")
		}
	}()

	if err := server.Run(creds, strconv.Itoa(constants.JoinServicePort)); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to run server")
	}
}

func getVPCIP(ctx context.Context, provider string) (string, error) {
	var metadata metadataAPI
	var err error

	switch cloudprovider.FromString(provider) {
	case cloudprovider.Azure:
		metadata, err = azurecloud.NewMetadata(ctx)
		if err != nil {
			return "", err
		}
	case cloudprovider.GCP:
		gcpClient, err := gcpcloud.NewClient(ctx)
		if err != nil {
			return "", err
		}
		metadata = gcpcloud.New(gcpClient)
	case cloudprovider.QEMU:
		metadata = &qemucloud.Metadata{}
	default:
		return "", errors.New("unsupported cloud provider")
	}

	self, err := metadata.Self(ctx)
	if err != nil {
		return "", err
	}
	return self.VPCIP, nil
}

type metadataAPI interface {
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
}

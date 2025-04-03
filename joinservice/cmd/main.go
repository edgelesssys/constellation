/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	awscloud "github.com/edgelesssys/constellation/v2/internal/cloud/aws"
	azurecloud "github.com/edgelesssys/constellation/v2/internal/cloud/azure"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	gcpcloud "github.com/edgelesssys/constellation/v2/internal/cloud/gcp"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	qemucloud "github.com/edgelesssys/constellation/v2/internal/cloud/qemu"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/certcache"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/kms"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/kubeadm"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/kubernetesca"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/server"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/watcher"
	"github.com/spf13/afero"
)

// vpcIPTimeout is the maximum amount of time to wait for retrieval of the VPC ip.
const vpcIPTimeout = 30 * time.Second

func main() {
	provider := flag.String("cloud-provider", "", "cloud service provider this binary is running on")
	keyServiceEndpoint := flag.String("key-service-endpoint", "", "endpoint of Constellations key management service")
	attestationVariant := flag.String("attestation-variant", "", "attestation variant to use for aTLS connections")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)
	flag.Parse()

	log := logger.NewJSONLogger(logger.VerbosityFromInt(*verbosity))
	log.With(
		slog.String("version", constants.BinaryVersion().String()),
		slog.String("cloudProvider", *provider),
		slog.String("attestationVariant", *attestationVariant),
	).Info("Constellation Node Join Service")

	handler := file.NewHandler(afero.NewOsFs())

	kubeClient, err := kubernetes.New()
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to create Kubernetes client")
		os.Exit(1)
	}

	attVariant, err := variant.FromString(*attestationVariant)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to parse attestation variant")
		os.Exit(1)
	}

	certCacheClient := certcache.NewClient(log.WithGroup("certcache"), kubeClient, attVariant)
	cachedCerts, err := certCacheClient.CreateCertChainCache(context.Background())
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to create certificate chain cache")
		os.Exit(1)
	}

	validator, err := watcher.NewValidator(log.WithGroup("validator"), attVariant, handler, cachedCerts)
	if err != nil {
		flag.Usage()
		log.With(slog.Any("error", err)).Error("Failed to create validator")
		os.Exit(1)
	}

	creds := atlscredentials.New(nil, []atls.Validator{validator})

	vpcCtx, cancel := context.WithTimeout(context.Background(), vpcIPTimeout)
	defer cancel()

	vpcIP, err := getVPCIP(vpcCtx, *provider)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to get IP in VPC")
		os.Exit(1)
	}
	apiServerEndpoint := net.JoinHostPort(vpcIP, strconv.Itoa(constants.KubernetesPort))
	kubeadm, err := kubeadm.New(apiServerEndpoint, log.WithGroup("kubeadm"))
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to create kubeadm")
	}
	keyServiceClient := kms.New(log.WithGroup("keyServiceClient"), *keyServiceEndpoint)

	measurementSalt, err := handler.Read(filepath.Join(constants.ServiceBasePath, constants.MeasurementSaltFilename))
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to read measurement salt")
		os.Exit(1)
	}

	server, err := server.New(
		measurementSalt,
		kubernetesca.New(log.WithGroup("certificateAuthority"), handler),
		kubeadm,
		keyServiceClient,
		kubeClient,
		log.WithGroup("server"),
		file.NewHandler(afero.NewOsFs()),
	)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to create server")
		os.Exit(1)
	}

	watcher, err := watcher.New(log.WithGroup("fileWatcher"), validator)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to create watcher for measurements updates")
		os.Exit(1)
	}
	defer watcher.Close()

	go func() {
		log.Info(fmt.Sprintf("starting file watcher for measurements file %s", filepath.Join(constants.ServiceBasePath, constants.AttestationConfigFilename)))
		if err := watcher.Watch(filepath.Join(constants.ServiceBasePath, constants.AttestationConfigFilename)); err != nil {
			log.With(slog.Any("error", err)).Error("Failed to watch measurements file")
			os.Exit(1)
		}
	}()

	if err := server.Run(creds, strconv.Itoa(constants.JoinServicePort)); err != nil {
		log.With(slog.Any("error", err)).Error("Failed to run server")
		os.Exit(1)
	}
}

func getVPCIP(ctx context.Context, provider string) (string, error) {
	var metadataClient metadataAPI
	var err error

	switch cloudprovider.FromString(provider) {
	case cloudprovider.AWS:
		metadataClient, err = awscloud.New(ctx)
		if err != nil {
			return "", err
		}
	case cloudprovider.Azure:
		metadataClient, err = azurecloud.New(ctx)
		if err != nil {
			return "", err
		}
	case cloudprovider.GCP:
		gcpMeta, err := gcpcloud.New(ctx)
		if err != nil {
			return "", err
		}
		defer gcpMeta.Close()
		metadataClient = gcpMeta
	case cloudprovider.OpenStack:
		metadataClient, err = openstack.New(ctx)
		if err != nil {
			return "", err
		}
	case cloudprovider.QEMU:
		metadataClient = qemucloud.New()
	default:
		return "", errors.New("unsupported cloud provider")
	}

	self, err := metadataClient.Self(ctx)
	if err != nil {
		return "", err
	}
	return self.VPCIP, nil
}

type metadataAPI interface {
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
}

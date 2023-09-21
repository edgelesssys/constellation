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
	"go.uber.org/zap"
)

// vpcIPTimeout is the maximum amount of time to wait for retrieval of the VPC ip.
const vpcIPTimeout = 30 * time.Second

func main() {
	provider := flag.String("cloud-provider", "", "cloud service provider this binary is running on")
	keyServiceEndpoint := flag.String("key-service-endpoint", "", "endpoint of Constellations key management service")
	attestationVariant := flag.String("attestation-variant", "", "attestation variant to use for aTLS connections")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), vpcIPTimeout)
	defer cancel()

	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))
	log.With(
		zap.String("version", constants.BinaryVersion().String()),
		zap.String("cloudProvider", *provider),
		zap.String("attestationVariant", *attestationVariant),
	).Infof("Constellation Node Join Service")

	handler := file.NewHandler(afero.NewOsFs())

	kubeClient, err := kubernetes.New()
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create Kubernetes client")
	}

	attVariant, err := variant.FromString(*attestationVariant)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to parse attestation variant")
	}

	certCacheClient := certcache.NewClient(log.Named("certcache"), kubeClient, &attVariant)
	cachedCerts, err := certCacheClient.CreateCertChainCache(ctx)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create certificate chain cache")
	}

	validator := watcher.
		NewValidator(log.Named("validator"), attVariant, handler).
		WithCachedCerts(cachedCerts)
	if err := validator.Update(); err != nil {
		flag.Usage()
		log.With(zap.Error(err)).Fatalf("Failed to create validator")
	}

	creds := atlscredentials.New(nil, []atls.Validator{validator})

	vpcIP, err := getVPCIP(ctx, *provider)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to get IP in VPC")
	}
	apiServerEndpoint := net.JoinHostPort(vpcIP, strconv.Itoa(constants.KubernetesPort))
	kubeadm, err := kubeadm.New(apiServerEndpoint, log.Named("kubeadm"))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create kubeadm")
	}
	keyServiceClient := kms.New(log.Named("keyServiceClient"), *keyServiceEndpoint)

	measurementSalt, err := handler.Read(filepath.Join(constants.ServiceBasePath, constants.MeasurementSaltFilename))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to read measurement salt")
	}

	server, err := server.New(
		measurementSalt,
		kubernetesca.New(log.Named("certificateAuthority"), handler),
		kubeadm,
		keyServiceClient,
		kubeClient,
		log.Named("server"),
	)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create server")
	}

	watcher, err := watcher.New(log.Named("fileWatcher"), validator)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create watcher for measurements updates")
	}
	defer watcher.Close()

	go func() {
		log.Infof("starting file watcher for measurements file %s", filepath.Join(constants.ServiceBasePath, constants.AttestationConfigFilename))
		if err := watcher.Watch(filepath.Join(constants.ServiceBasePath, constants.AttestationConfigFilename)); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to watch measurements file")
		}
	}()

	if err := server.Run(creds, strconv.Itoa(constants.JoinServicePort)); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to run server")
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

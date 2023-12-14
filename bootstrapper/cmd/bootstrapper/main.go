/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"flag"
	"io"
	"os"
	"strconv"

	"github.com/spf13/afero"
	"go.uber.org/zap"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubewaiter"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/logging"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/attestation/tdx"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	awscloud "github.com/edgelesssys/constellation/v2/internal/cloud/aws"
	azurecloud "github.com/edgelesssys/constellation/v2/internal/cloud/azure"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	gcpcloud "github.com/edgelesssys/constellation/v2/internal/cloud/gcp"
	openstackcloud "github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	qemucloud "github.com/edgelesssys/constellation/v2/internal/cloud/qemu"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

const (
	// constellationCSP is the environment variable stating which Cloud Service Provider Constellation is running on.
	constellationCSP = "CONSTEL_CSP"
)

func main() {
	gRPCDebug := flag.Bool("debug", false, "Enable gRPC debug logging")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)
	flag.Parse()
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity)).Named("bootstrapper")
	defer log.Sync()

	if *gRPCDebug {
		log.Named("gRPC").ReplaceGRPCLogger()
	} else {
		log.Named("gRPC").WithIncreasedLevel(zap.WarnLevel).ReplaceGRPCLogger()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bindIP := "0.0.0.0"
	bindPort := strconv.Itoa(constants.BootstrapperPort)
	var clusterInitJoiner clusterInitJoiner
	var metadataAPI metadataAPI
	var cloudLogger logging.CloudLogger
	var openDevice vtpm.TPMOpenFunc
	var fs afero.Fs

	attestVariant, err := variant.FromString(os.Getenv(constants.AttestationVariant))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to parse attestation variant")
	}
	issuer, err := choose.Issuer(attestVariant, log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to select issuer")
	}

	switch cloudprovider.FromString(os.Getenv(constellationCSP)) {
	case cloudprovider.AWS:
		metadata, err := awscloud.New(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to set up AWS metadata API")
		}
		metadataAPI = metadata

		cloudLogger, err = awscloud.NewLogger(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to set up cloud logger")
		}

		clusterInitJoiner = kubernetes.New(
			"aws", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{},
		)
		openDevice = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.GCP:
		metadata, err := gcpcloud.New(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to create GCP metadata client")
		}
		defer metadata.Close()

		cloudLogger, err = gcpcloud.NewLogger(ctx, "constellation-boot-log")
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to set up cloud logger")
		}

		metadataAPI = metadata
		clusterInitJoiner = kubernetes.New(
			"gcp", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{},
		)
		openDevice = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.Azure:
		metadata, err := azurecloud.New(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to create Azure metadata client")
		}
		cloudLogger, err = azurecloud.NewLogger(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to set up cloud logger")
		}
		if err := metadata.PrepareControlPlaneNode(ctx, log); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to prepare Azure control plane node")
		}

		metadataAPI = metadata
		clusterInitJoiner = kubernetes.New(
			"azure", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{},
		)

		openDevice = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.QEMU:
		cloudLogger = qemucloud.NewLogger()
		metadata := qemucloud.New()
		clusterInitJoiner = kubernetes.New(
			"qemu", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{},
		)
		metadataAPI = metadata

		switch attestVariant {
		case variant.QEMUVTPM{}:
			openDevice = vtpm.OpenVTPM
		case variant.QEMUTDX{}:
			openDevice = func() (io.ReadWriteCloser, error) {
				return tdx.Open()
			}
		default:
			log.Fatalf("Unsupported attestation variant: %s", attestVariant)
		}
		fs = afero.NewOsFs()
	case cloudprovider.OpenStack:
		cloudLogger = &logging.NopLogger{}
		metadata, err := openstackcloud.New(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to create OpenStack metadata client")
		}
		clusterInitJoiner = kubernetes.New(
			"openstack", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{},
		)
		metadataAPI = metadata
		openDevice = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	default:
		clusterInitJoiner = &clusterFake{}
		metadataAPI = &providerMetadataFake{}
		cloudLogger = &logging.NopLogger{}
		var simulatedTPMCloser io.Closer
		openDevice, simulatedTPMCloser = simulator.NewSimulatedTPMOpenFunc()
		defer simulatedTPMCloser.Close()
		fs = afero.NewMemMapFs()
	}

	fileHandler := file.NewHandler(fs)

	run(issuer, openDevice, fileHandler, clusterInitJoiner, metadataAPI, bindIP, bindPort, log, cloudLogger)
}

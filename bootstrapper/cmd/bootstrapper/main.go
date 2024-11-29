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
	"os"
	"strconv"

	"github.com/spf13/afero"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubewaiter"
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
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)
	flag.Parse()
	log := logger.NewJSONLogger(logger.VerbosityFromInt(*verbosity)).WithGroup("bootstrapper")
	logger.ReplaceGRPCLogger(logger.GRPCLogger(log))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bindIP := "0.0.0.0"
	bindPort := strconv.Itoa(constants.BootstrapperPort)
	var clusterInitJoiner clusterInitJoiner
	var metadataAPI metadataAPI
	var openDevice vtpm.TPMOpenFunc
	var fs afero.Fs

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

	switch cloudprovider.FromString(os.Getenv(constellationCSP)) {
	case cloudprovider.AWS:
		metadata, err := awscloud.New(ctx)
		if err != nil {
			log.With(slog.Any("error", err)).Error("Failed to set up AWS metadata API")
			os.Exit(1)
		}
		metadataAPI = metadata

		clusterInitJoiner = kubernetes.New(
			"aws", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{}, log,
		)
		openDevice = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.GCP:
		metadata, err := gcpcloud.New(ctx)
		if err != nil {
			log.With(slog.Any("error", err)).Error("Failed to create GCP metadata client")
			os.Exit(1)
		}
		defer metadata.Close()

		metadataAPI = metadata
		clusterInitJoiner = kubernetes.New(
			"gcp", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{}, log,
		)
		openDevice = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.Azure:
		metadata, err := azurecloud.New(ctx)
		if err != nil {
			log.With(slog.Any("error", err)).Error("Failed to create Azure metadata client")
			os.Exit(1)
		}

		if err := metadata.PrepareControlPlaneNode(ctx, log); err != nil {
			log.With(slog.Any("error", err)).Error("Failed to prepare Azure control plane node")
			os.Exit(1)
		}

		metadataAPI = metadata
		clusterInitJoiner = kubernetes.New(
			"azure", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{}, log,
		)

		openDevice = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.QEMU:
		metadata := qemucloud.New()
		clusterInitJoiner = kubernetes.New(
			"qemu", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{}, log,
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
			log.Error(fmt.Sprintf("Unsupported attestation variant: %s", attestVariant))
		}
		fs = afero.NewOsFs()
	case cloudprovider.OpenStack:
		metadata, err := openstackcloud.New(ctx)
		if err != nil {
			log.With(slog.Any("error", err)).Error("Failed to create OpenStack metadata client")
			os.Exit(1)
		}
		clusterInitJoiner = kubernetes.New(
			"openstack", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.NewUninitialized(),
			metadata, &kubewaiter.CloudKubeAPIWaiter{}, log,
		)
		metadataAPI = metadata
		openDevice = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	default:
		clusterInitJoiner = &clusterFake{}
		metadataAPI = &providerMetadataFake{}
		var simulatedTPMCloser io.Closer
		openDevice, simulatedTPMCloser = simulator.NewSimulatedTPMOpenFunc()
		defer simulatedTPMCloser.Close()
		fs = afero.NewMemMapFs()
	}

	fileHandler := file.NewHandler(fs)

	run(issuer, openDevice, fileHandler, clusterInitJoiner, metadataAPI, bindIP, bindPort, log)
}

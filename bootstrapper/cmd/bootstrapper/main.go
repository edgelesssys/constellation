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

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/helm"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/initserver"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	kubewaiter "github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubeWaiter"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/logging"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/aws"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	awscloud "github.com/edgelesssys/constellation/v2/internal/cloud/aws"
	azurecloud "github.com/edgelesssys/constellation/v2/internal/cloud/azure"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	gcpcloud "github.com/edgelesssys/constellation/v2/internal/cloud/gcp"
	qemucloud "github.com/edgelesssys/constellation/v2/internal/cloud/qemu"
	"github.com/edgelesssys/constellation/v2/internal/cloud/vmtype"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

const (
	// ConstellationCSP is the environment variable stating which Cloud Service Provider Constellation is running on.
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
	var issuer initserver.IssuerWrapper
	var openTPM vtpm.TPMOpenFunc
	var fs afero.Fs

	helmClient, err := helm.New(log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Helm client could not be initialized")
	}

	switch cloudprovider.FromString(os.Getenv(constellationCSP)) {
	case cloudprovider.AWS:
		measurements, err := vtpm.GetSelectedMeasurements(vtpm.OpenVTPM, vtpm.AWSPCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		issuer = initserver.NewIssuerWrapper(aws.NewIssuer(), vmtype.Unknown, nil)

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
			"aws", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.New(),
			metadata, measurements, helmClient, &kubewaiter.CloudKubeAPIWaiter{},
		)
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.GCP:
		measurements, err := vtpm.GetSelectedMeasurements(vtpm.OpenVTPM, vtpm.GCPPCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		issuer = initserver.NewIssuerWrapper(gcp.NewIssuer(), vmtype.Unknown, nil)

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
			"gcp", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.New(),
			metadata, measurements, helmClient, &kubewaiter.CloudKubeAPIWaiter{},
		)
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
		log.Infof("Added load balancer IP to routing table")

	case cloudprovider.Azure:
		measurements, err := vtpm.GetSelectedMeasurements(vtpm.OpenVTPM, vtpm.AzurePCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		if idkeydigest, err := snp.GetIDKeyDigest(vtpm.OpenVTPM); err == nil {
			issuer = initserver.NewIssuerWrapper(snp.NewIssuer(), vmtype.AzureCVM, idkeydigest)
		} else {
			// assume we are running in a trusted-launch VM
			issuer = initserver.NewIssuerWrapper(trustedlaunch.NewIssuer(), vmtype.AzureTrustedLaunch, idkeydigest)
		}

		metadata, err := azurecloud.New(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to create Azure metadata client")
		}
		cloudLogger, err = azurecloud.NewLogger(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to set up cloud logger")
		}
		metadataAPI = metadata
		clusterInitJoiner = kubernetes.New(
			"azure", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.New(),
			metadata, measurements, helmClient, &kubewaiter.CloudKubeAPIWaiter{},
		)

		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.QEMU:
		measurements, err := vtpm.GetSelectedMeasurements(vtpm.OpenVTPM, vtpm.QEMUPCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		issuer = initserver.NewIssuerWrapper(qemu.NewIssuer(), vmtype.Unknown, nil)

		cloudLogger = qemucloud.NewLogger()
		metadata := qemucloud.New()
		clusterInitJoiner = kubernetes.New(
			"qemu", k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.New(),
			metadata, measurements, helmClient, &kubewaiter.CloudKubeAPIWaiter{},
		)
		metadataAPI = metadata

		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	default:
		issuer = initserver.NewIssuerWrapper(atls.NewFakeIssuer(oid.Dummy{}), vmtype.Unknown, nil)
		clusterInitJoiner = &clusterFake{}
		metadataAPI = &providerMetadataFake{}
		cloudLogger = &logging.NopLogger{}
		var simulatedTPMCloser io.Closer
		openTPM, simulatedTPMCloser = simulator.NewSimulatedTPMOpenFunc()
		defer simulatedTPMCloser.Close()
		fs = afero.NewMemMapFs()
	}

	fileHandler := file.NewHandler(fs)

	run(issuer, openTPM, fileHandler, clusterInitJoiner, metadataAPI, bindIP, bindPort, log, cloudLogger)
}

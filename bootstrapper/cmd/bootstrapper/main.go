/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"strconv"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/helm"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi/kubectl"
	kubewaiter "github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubeWaiter"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/logging"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/aws"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	awscloud "github.com/edgelesssys/constellation/v2/internal/cloud/aws"
	azurecloud "github.com/edgelesssys/constellation/v2/internal/cloud/azure"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	gcpcloud "github.com/edgelesssys/constellation/v2/internal/cloud/gcp"
	qemucloud "github.com/edgelesssys/constellation/v2/internal/cloud/qemu"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
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
	var issuer atls.Issuer
	var openTPM vtpm.TPMOpenFunc
	var fs afero.Fs

	helmClient, err := helm.New(log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Helm client could not be initialized")
	}

	switch cloudprovider.FromString(os.Getenv(constellationCSP)) {
	case cloudprovider.AWS:
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.AWSPCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to marshal PCRs")
		}

		issuer = aws.NewIssuer()

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
			cloudprovider.AWS, k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.New(),
			metadata, pcrsJSON, nil, helmClient, &kubewaiter.CloudKubeAPIWaiter{},
		)
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.GCP:
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.GCPPCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		issuer = gcp.NewIssuer()

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
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to marshal PCRs")
		}
		clusterInitJoiner = kubernetes.New(
			cloudprovider.GCP, k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.New(),
			metadata, pcrsJSON, nil, helmClient, &kubewaiter.CloudKubeAPIWaiter{},
		)
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
		log.Infof("Added load balancer IP to routing table")

	case cloudprovider.Azure:
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.AzurePCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		issuer = azure.NewIssuer()
		idKeyDigest, err := azure.GetIDKeyDigest(vtpm.OpenVTPM)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to determine Azure idKeyDigest client")
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
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to marshal PCRs")
		}
		clusterInitJoiner = kubernetes.New(
			cloudprovider.Azure, k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.New(),
			metadata, pcrsJSON, idKeyDigest, helmClient, &kubewaiter.CloudKubeAPIWaiter{},
		)

		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()

	case cloudprovider.QEMU:
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.QEMUPCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		issuer = qemu.NewIssuer()

		cloudLogger = qemucloud.NewLogger()
		metadata := qemucloud.New()
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to marshal PCRs")
		}
		clusterInitJoiner = kubernetes.New(
			cloudprovider.QEMU, k8sapi.NewKubernetesUtil(), &k8sapi.KubdeadmConfiguration{}, kubectl.New(),
			metadata, pcrsJSON, nil, helmClient, &kubewaiter.CloudKubeAPIWaiter{},
		)
		metadataAPI = metadata

		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	default:
		issuer = atls.NewFakeIssuer(oid.Dummy{})
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

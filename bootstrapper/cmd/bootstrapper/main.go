package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	azurecloud "github.com/edgelesssys/constellation/bootstrapper/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/bootstrapper/cloudprovider/gcp"
	qemucloud "github.com/edgelesssys/constellation/bootstrapper/cloudprovider/qemu"
	"github.com/edgelesssys/constellation/bootstrapper/internal/joinclient"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/kubectl"
	"github.com/edgelesssys/constellation/bootstrapper/internal/logging"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/azure"
	"github.com/edgelesssys/constellation/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/oid"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

const (
	defaultIP   = "0.0.0.0"
	defaultPort = "9000"
	// ConstellationCSP is the Cloud Service Provider Constellation is running on.
	constellationCSP = "CONSTEL_CSP"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var bindIP, bindPort string
	var clusterInitJoiner clusterInitJoiner
	var metadataAPI joinclient.MetadataAPI
	var cloudLogger logging.CloudLogger
	cfg := zap.NewDevelopmentConfig()

	logLevelUser := flag.Bool("debug", false, "enables gRPC debug output")
	flag.Parse()
	cfg.Level.SetLevel(zap.DebugLevel)

	logger, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}
	if *logLevelUser {
		grpc_zap.ReplaceGrpcLoggerV2(logger.Named("gRPC"))
	} else {
		grpc_zap.ReplaceGrpcLoggerV2(logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel)).Named("gRPC"))
	}
	logger = logger.Named("bootstrapper")

	var issuer atls.Issuer
	var openTPM vtpm.TPMOpenFunc
	var fs afero.Fs

	switch strings.ToLower(os.Getenv(constellationCSP)) {
	case "aws":
		panic("AWS cloud provider currently unsupported")
	case "gcp":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.GCPPCRSelection)
		if err != nil {
			// TODO: Is there a reason we use log. instead of zapLogger?
			log.Fatal(err)
		}

		issuer = gcp.NewIssuer()

		gcpClient, err := gcpcloud.NewClient(ctx)
		if err != nil {
			log.Fatalf("failed to create GCP client: %v\n", err)
		}
		metadata := gcpcloud.New(gcpClient)
		descr, err := metadata.Self(ctx)
		if err != nil {
			log.Fatal(err)
		}
		cloudLogger, err = gcpcloud.NewLogger(ctx, descr.ProviderID, "constellation-boot-log")
		if err != nil {
			log.Fatal(err)
		}
		metadataAPI = metadata
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.Fatal(err)
		}
		clusterInitJoiner = kubernetes.New(
			"gcp", k8sapi.NewKubernetesUtil(), &k8sapi.CoreOSConfiguration{}, kubectl.New(), &gcpcloud.CloudControllerManager{},
			&gcpcloud.CloudNodeManager{}, &gcpcloud.Autoscaler{}, metadata, pcrsJSON,
		)
		bindIP = defaultIP
		bindPort = defaultPort
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	case "azure":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.AzurePCRSelection)
		if err != nil {
			log.Fatal(err)
		}

		issuer = azure.NewIssuer()

		metadata, err := azurecloud.NewMetadata(ctx)
		if err != nil {
			log.Fatal(err)
		}
		cloudLogger, err = azurecloud.NewLogger(ctx, metadata)
		if err != nil {
			log.Fatal(err)
		}
		metadataAPI = metadata
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.Fatal(err)
		}
		clusterInitJoiner = kubernetes.New(
			"azure", k8sapi.NewKubernetesUtil(), &k8sapi.CoreOSConfiguration{}, kubectl.New(), azurecloud.NewCloudControllerManager(metadata),
			&azurecloud.CloudNodeManager{}, &azurecloud.Autoscaler{}, metadata, pcrsJSON,
		)

		bindIP = defaultIP
		bindPort = defaultPort
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	case "qemu":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.QEMUPCRSelection)
		if err != nil {
			log.Fatal(err)
		}

		issuer = qemu.NewIssuer()

		cloudLogger = qemucloud.NewLogger()
		metadata := &qemucloud.Metadata{}
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.Fatal(err)
		}
		clusterInitJoiner = kubernetes.New(
			"qemu", k8sapi.NewKubernetesUtil(), &k8sapi.CoreOSConfiguration{}, kubectl.New(), &qemucloud.CloudControllerManager{},
			&qemucloud.CloudNodeManager{}, &qemucloud.Autoscaler{}, metadata, pcrsJSON,
		)
		metadataAPI = metadata

		bindIP = defaultIP
		bindPort = defaultPort
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	default:
		issuer = atls.NewFakeIssuer(oid.Dummy{})
		clusterInitJoiner = &clusterFake{}
		metadataAPI = &providerMetadataFake{}
		cloudLogger = &logging.NopLogger{}
		bindIP = defaultIP
		bindPort = defaultPort
		var simulatedTPMCloser io.Closer
		openTPM, simulatedTPMCloser = simulator.NewSimulatedTPMOpenFunc()
		defer simulatedTPMCloser.Close()
		fs = afero.NewMemMapFs()
	}

	fileHandler := file.NewHandler(fs)

	run(issuer, openTPM, fileHandler, clusterInitJoiner, metadataAPI, bindIP, bindPort, logger, cloudLogger)
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/bootstrapper/internal/logging"
	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/azure"
	"github.com/edgelesssys/constellation/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	azurecloud "github.com/edgelesssys/constellation/internal/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/internal/cloudprovider/gcp"
	qemucloud "github.com/edgelesssys/constellation/internal/cloudprovider/qemu"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/iproute"
	"github.com/edgelesssys/constellation/internal/kubernetes"
	"github.com/edgelesssys/constellation/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/internal/kubernetes/k8sapi/kubectl"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/oid"
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

	switch strings.ToLower(os.Getenv(constellationCSP)) {
	case "aws":
		panic("AWS cloud provider currently unsupported")
	case "gcp":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.GCPPCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		issuer = gcp.NewIssuer()

		gcpClient, err := gcpcloud.NewClient(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to create GCP metadata client")
		}
		metadata := gcpcloud.New(gcpClient)
		descr, err := metadata.Self(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get instance metadata")
		}
		cloudLogger, err = gcpcloud.NewLogger(ctx, descr.ProviderID, "constellation-boot-log")
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to set up cloud logger")
		}
		metadataAPI = metadata
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to marshal PCRs")
		}
		clusterInitJoiner = kubernetes.New(
			"gcp", k8sapi.NewKubernetesUtil(), &k8sapi.CoreOSConfiguration{}, kubectl.New(), &gcpcloud.CloudControllerManager{},
			&gcpcloud.CloudNodeManager{}, &gcpcloud.Autoscaler{}, metadata, pcrsJSON,
		)
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
		if err := setLoadbalancerRoute(ctx, metadata); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to set loadbalancer route")
		}
		log.Infof("Added load balancer IP to routing table")
	case "azure":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.AzurePCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		issuer = azure.NewIssuer()

		metadata, err := azurecloud.NewMetadata(ctx)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to create Azure metadata client")
		}
		cloudLogger, err = azurecloud.NewLogger(ctx, metadata)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to set up cloud logger")
		}
		metadataAPI = metadata
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to marshal PCRs")
		}
		clusterInitJoiner = kubernetes.New(
			"azure", k8sapi.NewKubernetesUtil(), &k8sapi.CoreOSConfiguration{}, kubectl.New(), azurecloud.NewCloudControllerManager(metadata),
			&azurecloud.CloudNodeManager{}, &azurecloud.Autoscaler{}, metadata, pcrsJSON,
		)

		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	case "qemu":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.QEMUPCRSelection)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get selected PCRs")
		}

		issuer = qemu.NewIssuer()

		cloudLogger = qemucloud.NewLogger()
		metadata := &qemucloud.Metadata{}
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to marshal PCRs")
		}
		clusterInitJoiner = kubernetes.New(
			"qemu", k8sapi.NewKubernetesUtil(), &k8sapi.CoreOSConfiguration{}, kubectl.New(), &qemucloud.CloudControllerManager{},
			&qemucloud.CloudNodeManager{}, &qemucloud.Autoscaler{}, metadata, pcrsJSON,
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

func setLoadbalancerRoute(ctx context.Context, meta metadataAPI) error {
	self, err := meta.Self(ctx)
	if err != nil {
		return err
	}
	if self.Role != role.ControlPlane {
		return nil
	}
	endpoint, err := meta.GetLoadBalancerEndpoint(ctx)
	if err != nil {
		return err
	}
	ip, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return err
	}
	return iproute.AddToLocalRoutingTable(ctx, ip)
}

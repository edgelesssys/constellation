package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"strings"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/attestation/aws"
	"github.com/edgelesssys/constellation/coordinator/attestation/azure"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	awscloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/aws"
	azurecloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/gcp"
	"github.com/edgelesssys/constellation/coordinator/config"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/kubernetes"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/kubectl"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/coordinator/wireguard"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

const (
	gvisorIP            = "192.168.127.2"
	defaultIP           = "0.0.0.0"
	defaultPort         = "9000"
	defaultEtcdEndpoint = "127.0.0.1:2379"
)

func main() {
	var bindIP, bindPort, etcdEndpoint string
	var enforceEtcdTls bool
	var kube core.Cluster
	var metadata core.ProviderMetadata
	var cloudControllerManager core.CloudControllerManager
	var cloudNodeManager core.CloudNodeManager
	var autoscaler core.ClusterAutoscaler
	cfg := zap.NewDevelopmentConfig()

	logLevelUser := flag.Bool("debug", false, "enables gRPC debug output")
	flag.Parse()
	cfg.Level.SetLevel(zap.DebugLevel)

	zapLogger, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}
	if *logLevelUser {
		grpc_zap.ReplaceGrpcLoggerV2(zapLogger.Named("gRPC"))
	} else {
		grpc_zap.ReplaceGrpcLoggerV2(zapLogger.WithOptions(zap.IncreaseLevel(zap.WarnLevel)).Named("gRPC"))
	}
	zapLoggerCore := zapLogger.Named("core")

	wg, err := wireguard.New()
	if err != nil {
		zapLogger.Panic("error opening wgctrl client")
	}
	defer wg.Close()

	var issuer core.QuoteIssuer
	var validator core.QuoteValidator
	var openTPM vtpm.TPMOpenFunc
	var fs afero.Fs

	switch strings.ToLower(os.Getenv(config.ConstellationCSP)) {
	case "aws":
		issuer = aws.NewIssuer()
		validator = aws.NewValidator(aws.NaAdGetVerifiedPayloadAsJson)
		kube = kubernetes.New(&k8sapi.KubernetesUtil{}, &k8sapi.AWSConfiguration{}, kubectl.New())
		metadata = awscloud.Metadata{}
		cloudControllerManager = awscloud.CloudControllerManager{}
		cloudNodeManager = &awscloud.CloudNodeManager{}
		autoscaler = awscloud.Autoscaler{}
		bindIP = gvisorIP
		bindPort = defaultPort
		etcdEndpoint = defaultEtcdEndpoint
		enforceEtcdTls = true
		openTPM = vtpm.OpenNOPTPM
		fs = afero.NewOsFs()
	case "gcp":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.GCPPCRSelection)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: Remove once we no longer use non cvms
		isCVM, err := gcp.IsCVM()
		if err != nil {
			log.Fatal(err)
		}
		if isCVM {
			issuer = gcp.NewIssuer()
			validator = gcp.NewValidator(pcrs)
		} else {
			issuer = gcp.NewNonCVMIssuer()
			validator = gcp.NewNonCVMValidator(pcrs)
		}
		kube = kubernetes.New(&k8sapi.KubernetesUtil{}, &k8sapi.CoreOSConfiguration{}, kubectl.New())
		gcpClient, err := gcpcloud.NewClient(context.Background())
		if err != nil {
			log.Fatalf("creating GCP client failed: %v\n", err)
		}
		metadata = gcpcloud.New(gcpClient)
		cloudControllerManager = &gcpcloud.CloudControllerManager{}
		cloudNodeManager = &gcpcloud.CloudNodeManager{}
		autoscaler = &gcpcloud.Autoscaler{}
		bindIP = defaultIP
		bindPort = defaultPort
		etcdEndpoint = defaultEtcdEndpoint
		enforceEtcdTls = true
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	case "azure":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.AzurePCRSelection)
		if err != nil {
			log.Fatal(err)
		}

		issuer = azure.NewIssuer()
		validator = azure.NewValidator(pcrs)

		kube = kubernetes.New(&k8sapi.KubernetesUtil{}, &k8sapi.CoreOSConfiguration{}, kubectl.New())
		metadata, err = azurecloud.NewMetadata(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		cloudControllerManager = &azurecloud.CloudControllerManager{}
		cloudNodeManager = &azurecloud.CloudNodeManager{}
		autoscaler = &azurecloud.Autoscaler{}
		bindIP = defaultIP
		bindPort = defaultPort
		etcdEndpoint = defaultEtcdEndpoint
		enforceEtcdTls = true
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	default:
		issuer = core.NewMockIssuer()
		validator = core.NewMockValidator()
		kube = &core.ClusterFake{}
		metadata = &core.ProviderMetadataFake{}
		cloudControllerManager = &core.CloudControllerManagerFake{}
		cloudNodeManager = &core.CloudNodeManagerFake{}
		autoscaler = &core.ClusterAutoscalerFake{}
		bindIP = defaultIP
		bindPort = defaultPort
		etcdEndpoint = "etcd-storage:2379"
		enforceEtcdTls = false
		var simulatedTPMCloser io.Closer
		openTPM, simulatedTPMCloser = vtpm.NewSimulatedTPMOpenFunc()
		defer simulatedTPMCloser.Close()
		fs = afero.NewMemMapFs()
	}

	fileHandler := file.NewHandler(fs)
	dialer := &net.Dialer{}
	run(validator, issuer, wg, openTPM, util.GetIPAddr, dialer, fileHandler, kube,
		metadata, cloudControllerManager, cloudNodeManager, autoscaler, etcdEndpoint, enforceEtcdTls, bindIP, bindPort, zapLoggerCore)
}

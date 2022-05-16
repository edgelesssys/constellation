package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/edgelesssys/constellation/coordinator/attestation/azure"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/coordinator/attestation/qemu"
	"github.com/edgelesssys/constellation/coordinator/attestation/simulator"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	azurecloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/gcp"
	qemucloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/qemu"
	"github.com/edgelesssys/constellation/coordinator/config"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/diskencryption"
	"github.com/edgelesssys/constellation/coordinator/kubernetes"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/kubectl"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/coordinator/util/grpcutil"
	"github.com/edgelesssys/constellation/coordinator/wireguard"
	"github.com/edgelesssys/constellation/internal/file"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

const (
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
	var encryptedDisk core.EncryptedDisk
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
		panic("AWS cloud provider currently unsupported")
	case "gcp":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.GCPPCRSelection)
		if err != nil {
			log.Fatal(err)
		}

		issuer = gcp.NewIssuer()
		validator = gcp.NewValidator(pcrs)

		kube = kubernetes.New(&k8sapi.KubernetesUtil{}, &k8sapi.CoreOSConfiguration{}, kubectl.New())
		gcpClient, err := gcpcloud.NewClient(context.Background())
		if err != nil {
			log.Fatalf("creating GCP client failed: %v\n", err)
		}
		metadata = gcpcloud.New(gcpClient)
		cloudControllerManager = &gcpcloud.CloudControllerManager{}
		cloudNodeManager = &gcpcloud.CloudNodeManager{}
		autoscaler = &gcpcloud.Autoscaler{}
		encryptedDisk = diskencryption.New()
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
		encryptedDisk = diskencryption.New()
		bindIP = defaultIP
		bindPort = defaultPort
		etcdEndpoint = defaultEtcdEndpoint
		enforceEtcdTls = true
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	case "qemu":
		pcrs, err := vtpm.GetSelectedPCRs(vtpm.OpenVTPM, vtpm.QEMUPCRSelection)
		if err != nil {
			log.Fatal(err)
		}

		issuer = qemu.NewIssuer()
		validator = qemu.NewValidator(pcrs)

		kube = kubernetes.New(&k8sapi.KubernetesUtil{}, &k8sapi.CoreOSConfiguration{}, kubectl.New())

		// no support for cloud services in qemu
		metadata = &qemucloud.Metadata{}
		cloudControllerManager = &qemucloud.CloudControllerManager{}
		cloudNodeManager = &qemucloud.CloudNodeManager{}
		autoscaler = &qemucloud.Autoscaler{}

		encryptedDisk = diskencryption.New()
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
		encryptedDisk = &core.EncryptedDiskFake{}
		bindIP = defaultIP
		bindPort = defaultPort
		etcdEndpoint = "etcd-storage:2379"
		enforceEtcdTls = false
		var simulatedTPMCloser io.Closer
		openTPM, simulatedTPMCloser = simulator.NewSimulatedTPMOpenFunc()
		defer simulatedTPMCloser.Close()
		fs = afero.NewMemMapFs()
	}

	fileHandler := file.NewHandler(fs)
	netDialer := &net.Dialer{}
	dialer := grpcutil.NewDialer(validator, netDialer)
	run(issuer, wg, openTPM, util.GetIPAddr, dialer, fileHandler, kube,
		metadata, cloudControllerManager, cloudNodeManager, autoscaler, encryptedDisk, etcdEndpoint, enforceEtcdTls, bindIP, bindPort, zapLoggerCore, fs)
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strings"

	azurecloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/gcp"
	qemucloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/qemu"
	"github.com/edgelesssys/constellation/coordinator/config"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/diskencryption"
	"github.com/edgelesssys/constellation/coordinator/kubernetes"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/kubectl"
	"github.com/edgelesssys/constellation/coordinator/logging"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/coordinator/wireguard"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/azure"
	"github.com/edgelesssys/constellation/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/oid"
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
	var coreMetadata core.ProviderMetadata
	var encryptedDisk core.EncryptedDisk
	var cloudLogger logging.CloudLogger
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
			// TODO: Is there a reason we use log. instead of zapLogger?
			log.Fatal(err)
		}

		issuer = gcp.NewIssuer()
		validator = gcp.NewValidator(pcrs)

		gcpClient, err := gcpcloud.NewClient(context.Background())
		if err != nil {
			log.Fatalf("failed to create GCP client: %v\n", err)
		}
		metadata := gcpcloud.New(gcpClient)
		descr, err := metadata.Self(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		cloudLogger, err = gcpcloud.NewLogger(context.Background(), descr.ProviderID, "constellation-boot-log")
		if err != nil {
			log.Fatal(err)
		}
		coreMetadata = metadata
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.Fatal(err)
		}
		kube = kubernetes.New(
			"gcp", k8sapi.NewKubernetesUtil(), &k8sapi.CoreOSConfiguration{}, kubectl.New(), &gcpcloud.CloudControllerManager{},
			&gcpcloud.CloudNodeManager{}, &gcpcloud.Autoscaler{}, metadata, pcrsJSON,
		)
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

		metadata, err := azurecloud.NewMetadata(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		cloudLogger, err = azurecloud.NewLogger(context.Background(), metadata)
		if err != nil {
			log.Fatal(err)
		}
		coreMetadata = metadata
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.Fatal(err)
		}
		kube = kubernetes.New(
			"azure", k8sapi.NewKubernetesUtil(), &k8sapi.CoreOSConfiguration{}, kubectl.New(), azurecloud.NewCloudControllerManager(metadata),
			&azurecloud.CloudNodeManager{}, &azurecloud.Autoscaler{}, metadata, pcrsJSON,
		)

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

		// no support for cloud services in qemu
		metadata := &qemucloud.Metadata{}
		cloudLogger = &logging.NopLogger{}
		pcrsJSON, err := json.Marshal(pcrs)
		if err != nil {
			log.Fatal(err)
		}
		kube = kubernetes.New(
			"qemu", k8sapi.NewKubernetesUtil(), &k8sapi.CoreOSConfiguration{}, kubectl.New(), &qemucloud.CloudControllerManager{},
			&qemucloud.CloudNodeManager{}, &qemucloud.Autoscaler{}, metadata, pcrsJSON,
		)
		coreMetadata = metadata

		encryptedDisk = diskencryption.New()
		bindIP = defaultIP
		bindPort = defaultPort
		etcdEndpoint = defaultEtcdEndpoint
		enforceEtcdTls = true
		openTPM = vtpm.OpenVTPM
		fs = afero.NewOsFs()
	default:
		issuer = atls.NewFakeIssuer(oid.Dummy{})
		validator = atls.NewFakeValidator(oid.Dummy{})
		kube = &core.ClusterFake{}
		coreMetadata = &core.ProviderMetadataFake{}
		cloudLogger = &logging.NopLogger{}
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
	dialer := dialer.New(nil, validator, netDialer)
	run(issuer, wg, openTPM, util.GetIPAddr, dialer, fileHandler, kube,
		coreMetadata, encryptedDisk, etcdEndpoint, enforceEtcdTls, bindIP,
		bindPort, zapLoggerCore, cloudLogger, fs)
}

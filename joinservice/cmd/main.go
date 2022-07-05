package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"path/filepath"
	"strconv"

	azurecloud "github.com/edgelesssys/constellation/bootstrapper/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/bootstrapper/cloudprovider/gcp"
	qemucloud "github.com/edgelesssys/constellation/bootstrapper/cloudprovider/qemu"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/watcher"
	"github.com/edgelesssys/constellation/joinservice/internal/kms"
	"github.com/edgelesssys/constellation/joinservice/internal/kubeadm"
	"github.com/edgelesssys/constellation/joinservice/internal/kubernetesca"
	"github.com/edgelesssys/constellation/joinservice/internal/server"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func main() {
	provider := flag.String("cloud-provider", "", "cloud service provider this binary is running on")
	kmsEndpoint := flag.String("kms-endpoint", "", "endpoint of Constellations key management service")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))

	log.With(zap.String("version", constants.VersionInfo), zap.String("cloudProvider", *provider)).
		Infof("Constellation Node Join Service")

	handler := file.NewHandler(afero.NewOsFs())

	validator, err := watcher.NewValidator(log.Named("validator"), *provider, handler)
	if err != nil {
		flag.Usage()
		log.With(zap.Error(err)).Fatalf("Failed to create validator")
	}

	creds := atlscredentials.New(nil, []atls.Validator{validator})

	vpcIP, err := getIPinVPC(ctx, *provider)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to get IP in VPC")
	}
	apiServerEndpoint := net.JoinHostPort(vpcIP, strconv.Itoa(constants.KubernetesPort))
	kubeadm, err := kubeadm.New(apiServerEndpoint, log.Named("kubeadm"))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create kubeadm")
	}
	kms := kms.New(log.Named("kms"), *kmsEndpoint)

	server := server.New(
		log.Named("server"),
		handler,
		kubernetesca.New(log.Named("certificateAuthority"), handler),
		kubeadm,
		kms,
	)

	watcher, err := watcher.New(log.Named("fileWatcher"), validator)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create watcher for measurements updates")
	}
	defer watcher.Close()

	go func() {
		log.Infof("starting file watcher for measurements file %s", filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename))
		if err := watcher.Watch(filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename)); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to watch measurements file")
		}
	}()

	if err := server.Run(creds, strconv.Itoa(constants.JoinServicePort)); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to run server")
	}
}

func getIPinVPC(ctx context.Context, provider string) (string, error) {
	switch cloudprovider.FromString(provider) {
	case cloudprovider.Azure:
		metadata, err := azurecloud.NewMetadata(ctx)
		if err != nil {
			return "", err
		}
		self, err := metadata.Self(ctx)
		if err != nil {
			return "", err
		}
		return self.PrivateIPs[0], nil
	case cloudprovider.GCP:
		gcpClient, err := gcpcloud.NewClient(ctx)
		if err != nil {
			return "", err
		}
		metadata := gcpcloud.New(gcpClient)
		if err != nil {
			return "", err
		}
		self, err := metadata.Self(ctx)
		if err != nil {
			return "", err
		}
		return self.PrivateIPs[0], nil
	case cloudprovider.QEMU:
		metadata := &qemucloud.Metadata{}
		self, err := metadata.Self(ctx)
		if err != nil {
			return "", err
		}
		return self.PrivateIPs[0], nil
	default:
		return "", errors.New("unsupported cloud provider")
	}
}

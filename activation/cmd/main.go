package main

import (
	"flag"
	"path/filepath"
	"strconv"

	"github.com/edgelesssys/constellation/activation/kms"
	"github.com/edgelesssys/constellation/activation/kubeadm"
	"github.com/edgelesssys/constellation/activation/kubernetesca"
	"github.com/edgelesssys/constellation/activation/server"
	"github.com/edgelesssys/constellation/activation/validator"
	"github.com/edgelesssys/constellation/activation/watcher"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/spf13/afero"
	"k8s.io/klog/v2"
)

func main() {
	provider := flag.String("cloud-provider", "", "cloud service provider this binary is running on")
	kmsEndpoint := flag.String("kms-endpoint", "", "endpoint of Constellations key management service")

	klog.InitFlags(nil)
	flag.Parse()
	defer klog.Flush()

	klog.V(2).Infof("\nConstellation Node Activation Service\nVersion: %s\nRunning on: %s", constants.VersionInfo, *provider)

	handler := file.NewHandler(afero.NewOsFs())

	validator, err := validator.New(*provider, handler)
	if err != nil {
		flag.Usage()
		klog.Exitf("failed to create validator: %s", err)
	}

	creds := atlscredentials.New(nil, []atls.Validator{validator})

	kubeadm, err := kubeadm.New()
	if err != nil {
		klog.Exitf("failed to create kubeadm: %s", err)
	}
	kms := kms.New(*kmsEndpoint)

	server := server.New(handler, kubernetesca.New(handler), kubeadm, kms)

	watcher, err := watcher.New(validator)
	if err != nil {
		klog.Exitf("failed to create watcher for measurements updates: %s", err)
	}
	defer watcher.Close()

	go func() {
		klog.V(4).Infof("starting file watcher for measurements file %s", filepath.Join(constants.ActivationBasePath, constants.ActivationMeasurementsFilename))
		if err := watcher.Watch(filepath.Join(constants.ActivationBasePath, constants.ActivationMeasurementsFilename)); err != nil {
			klog.Exitf("failed to watch measurements file: %s", err)
		}
	}()

	if err := server.Run(creds, strconv.Itoa(constants.ActivationServicePort)); err != nil {
		klog.Exitf("failed to run server: %s", err)
	}
}

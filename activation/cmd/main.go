package main

import (
	"flag"

	"github.com/edgelesssys/constellation/activation/kms"
	"github.com/edgelesssys/constellation/activation/kubeadm"
	"github.com/edgelesssys/constellation/activation/server"
	"github.com/edgelesssys/constellation/activation/validator"
	"github.com/edgelesssys/constellation/activation/watcher"
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"k8s.io/klog/v2"
)

const (
	bindPort = "9090"
)

func main() {
	provider := flag.String("cloud-provider", "", "cloud service provider this binary is running on")
	kmsEndpoint := flag.String("kms-endpoint", "", "endpoint of Constellations key management service")

	klog.InitFlags(nil)
	flag.Parse()
	klog.V(2).Infof("\nConstellation Node Activation Service\nVersion: %s\nRunning on: %s", constants.VersionInfo, *provider)

	handler := file.NewHandler(afero.NewOsFs())

	validator, err := validator.New(*provider, handler)
	if err != nil {
		flag.Usage()
		klog.Exitf("failed to create validator: %s", err)
	}

	tlsConfig, err := atls.CreateAttestationServerTLSConfig(nil, []atls.Validator{validator})
	if err != nil {
		klog.Exitf("unable to create server config: %s", err)
	}

	kubeadm, err := kubeadm.New()
	if err != nil {
		klog.Exitf("failed to create kubeadm: %s", err)
	}
	kms := kms.New(*kmsEndpoint)

	server := server.New(handler, kubeadm, kms)

	watcher, err := watcher.New(validator)
	if err != nil {
		klog.Exitf("failed to create watcher for measurements updates: %s", err)
	}
	defer watcher.Close()

	go func() {
		klog.V(4).Infof("starting file watcher for measurements file %s", constants.ActivationMeasurementsFilename)
		if err := watcher.Watch(constants.ActivationMeasurementsFilename); err != nil {
			klog.Exitf("failed to watch measurements file: %s", err)
		}
	}()

	if err := server.Run(tlsConfig, bindPort); err != nil {
		klog.Exitf("failed to run server: %s", err)
	}
}

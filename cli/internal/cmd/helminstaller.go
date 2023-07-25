/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	helminstaller "github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
)

type helmInstaller interface {
	Install(ctx context.Context, provider cloudprovider.Provider, masterSecret uri.MasterSecret,
		idFile clusterid.File,
		serviceAccURI string, releases *helminstaller.Releases,
	) error
}

type helmInstallationClient struct{}

func (h helmInstallationClient) Install(ctx context.Context, provider cloudprovider.Provider, masterSecret uri.MasterSecret,
	idFile clusterid.File,
	serviceAccURI string, releases *helminstaller.Releases,
) error {
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(0)).Named("init") // TODO: use the same logger as the rest of the CLI
	defer log.Sync()
	installer, err := helminstaller.NewInstaller(log, constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("creating Helm installer: %w", err)
	}

	serviceVals, err := helm.SetupMicroserviceVals(ctx, provider, masterSecret.Salt, idFile.UID, serviceAccURI)
	if err != nil {
		return fmt.Errorf("setting up microservice values: %w", err)
	}
	fmt.Println("Installing microservices", serviceVals)

	log.Infof("Installing cert-manager")
	if err = installer.InstallChart(ctx, releases.CertManager); err != nil {
		return fmt.Errorf("installing cert-manager: %w", err)
	}

	if releases.CSI != nil {
		var csiVals map[string]any
		if provider == cloudprovider.OpenStack {
			creds, err := openstack.AccountKeyFromURI(serviceAccURI)
			if err != nil {
				return err
			}
			cinderIni := creds.CloudINI().CinderCSIConfiguration()
			csiVals = map[string]any{
				"cinder-config": map[string]any{
					"secretData": cinderIni,
				},
			}
		}

		log.Infof("Installing CSI deployments")
		if err := installer.InstallChartWithValues(ctx, *releases.CSI, csiVals); err != nil {
			return fmt.Errorf("installing CSI snapshot CRDs: %w", err)
		}
	}

	if releases.AWSLoadBalancerController != nil {
		log.Infof("Installing AWS Load Balancer Controller")
		if err = installer.InstallChart(ctx, *releases.AWSLoadBalancerController); err != nil {
			return fmt.Errorf("installing AWS Load Balancer Controller: %w", err)
		}
	}

	log.Infof("Installing constellation operators")
	operatorVals, err := helm.SetupOperatorVals(ctx, idFile.UID)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to set up operator values")
	}
	err = installer.InstallChartWithValues(ctx, releases.ConstellationOperators, operatorVals)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to install constellation operators")
	}

	// TODO(elchead): do cilium later, after upgrade
	//k8sClient := kubectl.New()
	//kubeconfig, err := os.ReadFile(constants.AdminConfFilename)
	//if err != nil {
	//	return fmt.Errorf("reading kubeconfig: %w", err)
	//}
	//err = k8sClient.Initialize(kubeconfig)
	//if err != nil {
	//	return fmt.Errorf("initializing k8s client: %w", err)
	//}
	// load cilium
	// releases.Cilium
	//

	// get instance id from tf? (create cmd) or just get metadata from cloud provider
	//switch provider {
	//case cloudprovider.AWS:
	//	// metadataAPI = metadata
	//}
	//metadata, err := helminstaller.GetMetadaClient(cmd.Context(), provider) // awscloud.New(cmd.Context())
	//if err != nil {
	//	log.With(zap.Error(err)).Fatalf("Failed to set up AWS metadata API")
	//}

	//input, err := helminstaller.GetSetupPodNetwork(cmd.Context(), metadata, provider)
	//if err != nil {
	//	return fmt.Errorf("getting setup pod network: %w", err)
	//}
	//cmd.Println("Installing Cilium")

	//if err = helminstaller.InstallCilium(cmd.Context(), installer, k8sClient, releases.Cilium, *input); err != nil {
	//	return fmt.Errorf("installing Cilium: %w", err)
	//}
	//cmd.Println("Installed Cilium")

	// installer.InstallChart()
	return nil
}

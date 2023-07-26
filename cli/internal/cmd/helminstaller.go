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
)

type helmInstaller interface {
	Install(ctx context.Context, provider cloudprovider.Provider, masterSecret uri.MasterSecret,
		idFile clusterid.File,
		serviceAccURI string, releases *helminstaller.Releases,
	) error
}

type helmInstallationClient struct {
	log debugLog
}

func (h helmInstallationClient) Install(ctx context.Context, provider cloudprovider.Provider, masterSecret uri.MasterSecret,
	idFile clusterid.File,
	serviceAccURI string, releases *helminstaller.Releases,
) error {
	installer, err := helminstaller.NewInstaller(constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("creating Helm installer: %w", err)
	}

	serviceVals, err := helm.SetupMicroserviceVals(ctx, provider, masterSecret.Salt, idFile.UID, serviceAccURI)
	if err != nil {
		return fmt.Errorf("setting up microservice values: %w", err)
	}
	h.log.Debugf("Installing microservices", serviceVals)
	if err := installer.InstallChartWithValues(ctx, releases.ConstellationServices, serviceVals); err != nil {
		return fmt.Errorf("installing microservices: %w", err)
	}

	h.log.Debugf("Installing cert-manager")
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

		h.log.Debugf("Installing CSI deployments")
		if err := installer.InstallChartWithValues(ctx, *releases.CSI, csiVals); err != nil {
			return fmt.Errorf("installing CSI snapshot CRDs: %w", err)
		}
	}

	if releases.AWSLoadBalancerController != nil {
		h.log.Debugf("Installing AWS Load Balancer Controller")
		if err = installer.InstallChart(ctx, *releases.AWSLoadBalancerController); err != nil {
			return fmt.Errorf("installing AWS Load Balancer Controller: %w", err)
		}
	}

	h.log.Debugf("Installing constellation operators")
	operatorVals, err := helm.SetupOperatorVals(ctx, idFile.UID)
	if err != nil {
		return fmt.Errorf("setting up operator values: %w", err)
	}
	err = installer.InstallChartWithValues(ctx, releases.ConstellationOperators, operatorVals)
	if err != nil {
		return fmt.Errorf("installing constellation operators: %w", err)
	}

	// TODO(elchead): AB394 do cilium after version upgrade
	return nil
}

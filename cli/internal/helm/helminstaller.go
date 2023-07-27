/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	helminstaller "github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
)

// SuiteInstaller installs all Helm charts required for a constellation cluster.
type SuiteInstaller interface {
	Install(ctx context.Context, provider cloudprovider.Provider, masterSecret uri.MasterSecret,
		idFile clusterid.File,
		serviceAccURI string, releases *helminstaller.Releases,
	) error
}

type helmInstallationClient struct {
	log       debugLog
	installer helmInstaller
}

// NewInstallationClient creates a new Helm installation client to install all Helm charts required for a constellation cluster.
func NewInstallationClient(log debugLog) (SuiteInstaller, error) {
	installer, err := helminstaller.NewInstaller(constants.AdminConfFilename)
	if err != nil {
		return nil, fmt.Errorf("creating Helm installer: %w", err)
	}
	return &helmInstallationClient{log: log, installer: installer}, nil
}

func (h helmInstallationClient) Install(ctx context.Context, provider cloudprovider.Provider, masterSecret uri.MasterSecret,
	idFile clusterid.File,
	serviceAccURI string, releases *helminstaller.Releases,
) error {
	serviceVals, err := setupMicroserviceVals(ctx, provider, masterSecret.Salt, idFile.UID, serviceAccURI)
	if err != nil {
		return fmt.Errorf("setting up microservice values: %w", err)
	}
	if err := h.installer.InstallChartWithValues(ctx, releases.ConstellationServices, serviceVals); err != nil {
		return fmt.Errorf("installing microservices: %w", err)
	}

	h.log.Debugf("Installing cert-manager")
	if err := h.installer.InstallChart(ctx, releases.CertManager); err != nil {
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
		if err := h.installer.InstallChartWithValues(ctx, *releases.CSI, csiVals); err != nil {
			return fmt.Errorf("installing CSI snapshot CRDs: %w", err)
		}
	}

	if releases.AWSLoadBalancerController != nil {
		h.log.Debugf("Installing AWS Load Balancer Controller")
		if err := h.installer.InstallChart(ctx, *releases.AWSLoadBalancerController); err != nil {
			return fmt.Errorf("installing AWS Load Balancer Controller: %w", err)
		}
	}

	h.log.Debugf("Installing constellation operators")
	operatorVals := setupOperatorVals(ctx, idFile.UID)
	if err := h.installer.InstallChartWithValues(ctx, releases.ConstellationOperators, operatorVals); err != nil {
		return fmt.Errorf("installing constellation operators: %w", err)
	}

	// TODO(elchead): AB#3301 do cilium after version upgrade
	return nil
}

type helmInstaller interface {
	InstallChart(context.Context, helminstaller.Release) error
	InstallChartWithValues(ctx context.Context, release helminstaller.Release, extraValues map[string]any) error
}

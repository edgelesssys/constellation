/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// Client handles interaction with helm.
type Client struct {
	config *action.Configuration
}

// NewClient returns a new initializes client for the namespace Client.
func NewClient(kubeConfigPath, helmNamespace string) (*Client, error) {
	settings := cli.New()
	settings.KubeConfig = kubeConfigPath // constants.AdminConfFilename

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), helmNamespace, "secret", nil); err != nil {
		return nil, fmt.Errorf("initializing config: %w", err)
	}

	return &Client{config: actionConfig}, nil
}

// CurrentVersion returns the version of the currently installed helm release.
func (c *Client) CurrentVersion(release string) (string, error) {
	client := action.NewList(c.config)
	client.Filter = release
	rel, err := client.Run()
	if err != nil {
		return "", err
	}

	if len(rel) == 0 {
		return "", fmt.Errorf("release %s not found", release)
	}
	if len(rel) > 1 {
		return "", fmt.Errorf("multiple releases found for %s", release)
	}

	if rel[0] == nil || rel[0].Chart == nil || rel[0].Chart.Metadata == nil {
		return "", fmt.Errorf("received invalid release %s", release)
	}

	return rel[0].Chart.Metadata.Version, nil
}

// Upgrade runs a helm-upgrade on all deployments that are managed via Helm.
// TODO: Verify that CRDs are upgraded.
// TODO: check helm cli how it handles ctrl+c.
func (c *Client) Upgrade(ctx context.Context, config *config.Config, conformanceMode bool, masterSecret, salt []byte) error {
	action := action.NewUpgrade(c.config)
	action.Atomic = true
	action.Namespace = constants.HelmNamespace
	action.ReuseValues = true
	action.DependencyUpdate = true

	loader := NewLoader(config.GetProvider(), versions.ValidK8sVersion(config.KubernetesVersion))
	ciliumChart, ciliumVals, err := loader.loadCiliumHelper(config.GetProvider(), conformanceMode)
	if err != nil {
		return fmt.Errorf("loading cilium: %w", err)
	}
	certManagerChart, certManagerVals, err := loader.loadCertManagerHelper()
	if err != nil {
		return fmt.Errorf("loading cilium: %w", err)
	}
	operatorChart, operatorVals, err := loader.loadOperatorsHelper(config.GetProvider())
	if err != nil {
		return fmt.Errorf("loading operators: %w", err)
	}
	conServicesChart, conServicesVals, err := loader.loadConstellationServicesHelper(config, masterSecret, salt)
	if err != nil {
		return fmt.Errorf("loading constellation-services: %w", err)
	}

	// Prevent half-finished upgrades
	action.Lock.Lock()
	defer action.Lock.Unlock()

	if _, err := action.RunWithContext(ctx, ciliumReleaseName, ciliumChart, ciliumVals); err != nil {
		return fmt.Errorf("upgrading cilium: %w", err)
	}
	if _, err := action.RunWithContext(ctx, ciliumReleaseName, certManagerChart, certManagerVals); err != nil {
		return fmt.Errorf("upgrading cert-manager: %w", err)
	}
	if _, err := action.RunWithContext(ctx, ciliumReleaseName, operatorChart, operatorVals); err != nil {
		return fmt.Errorf("upgrading services: %w", err)
	}
	if _, err := action.RunWithContext(ctx, ciliumReleaseName, conServicesChart, conServicesVals); err != nil {
		return fmt.Errorf("upgrading operators: %w", err)
	}

	return nil
}

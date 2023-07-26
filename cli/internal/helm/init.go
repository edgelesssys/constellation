/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// InitializationClient installs all Helm charts required for a Constellation cluster.
type InitializationClient struct {
	log       debugLog
	installer installer
}

// NewInitializer creates a new client to install all Helm charts required for a constellation cluster.
func NewInitializer(log debugLog, adminConfPath string) (*InitializationClient, error) {
	installer, err := NewInstaller(adminConfPath, log)
	if err != nil {
		return nil, fmt.Errorf("creating Helm installer: %w", err)
	}
	return &InitializationClient{log: log, installer: installer}, nil
}

// Install installs all Helm charts required for a constellation cluster.
func (i InitializationClient) Install(ctx context.Context, releases *Releases) error {
	if err := i.installer.InstallChart(ctx, releases.Cilium); err != nil {
		return fmt.Errorf("installing Cilium: %w", err)
	}
	i.log.Debugf("Waiting for Cilium to become ready")
	helper, err := newK8sCiliumHelper(constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}
	timeToStartWaiting := time.Now()
	// TODO(3u13r): Reduce the timeout when we switched the package repository - this is only this high because we once
	// saw polling times of ~16 minutes when hitting a slow PoP from Fastly (GitHub's / ghcr.io CDN).
	if err := helper.WaitForDS(ctx, "kube-system", "cilium", i.log); err != nil {
		return fmt.Errorf("waiting for Cilium to become healthy: %w", err)
	}
	timeUntilFinishedWaiting := time.Since(timeToStartWaiting)
	i.log.Debugf("Cilium became healthy after %s", timeUntilFinishedWaiting.String())

	i.log.Debugf("Fix Cilium through restart")
	if err := helper.RestartDS("kube-system", "cilium"); err != nil {
		return fmt.Errorf("restarting Cilium: %w", err)
	}

	i.log.Debugf("Installing microservices")
	if err := i.installer.InstallChart(ctx, releases.ConstellationServices); err != nil {
		return fmt.Errorf("installing microservices: %w", err)
	}

	i.log.Debugf("Installing cert-manager")
	if err := i.installer.InstallChart(ctx, releases.CertManager); err != nil {
		return fmt.Errorf("installing cert-manager: %w", err)
	}

	if releases.CSI != nil {
		i.log.Debugf("Installing CSI deployments")
		if err := i.installer.InstallChart(ctx, *releases.CSI); err != nil {
			return fmt.Errorf("installing CSI snapshot CRDs: %w", err)
		}
	}

	if releases.AWSLoadBalancerController != nil {
		i.log.Debugf("Installing AWS Load Balancer Controller")
		if err := i.installer.InstallChart(ctx, *releases.AWSLoadBalancerController); err != nil {
			return fmt.Errorf("installing AWS Load Balancer Controller: %w", err)
		}
	}

	i.log.Debugf("Installing constellation operators")
	if err := i.installer.InstallChart(ctx, releases.ConstellationOperators); err != nil {
		return fmt.Errorf("installing constellation operators: %w", err)
	}
	return nil
}

// installer is the interface for installing a single Helm chart.
type installer interface {
	InstallChart(context.Context, Release) error
	InstallChartWithValues(ctx context.Context, release Release, extraValues map[string]any) error
}

type cloudConfig struct {
	Cloud                       string `json:"cloud,omitempty"`
	TenantID                    string `json:"tenantId,omitempty"`
	SubscriptionID              string `json:"subscriptionId,omitempty"`
	ResourceGroup               string `json:"resourceGroup,omitempty"`
	Location                    string `json:"location,omitempty"`
	SubnetName                  string `json:"subnetName,omitempty"`
	SecurityGroupName           string `json:"securityGroupName,omitempty"`
	SecurityGroupResourceGroup  string `json:"securityGroupResourceGroup,omitempty"`
	LoadBalancerName            string `json:"loadBalancerName,omitempty"`
	LoadBalancerSku             string `json:"loadBalancerSku,omitempty"`
	VNetName                    string `json:"vnetName,omitempty"`
	VNetResourceGroup           string `json:"vnetResourceGroup,omitempty"`
	CloudProviderBackoff        bool   `json:"cloudProviderBackoff,omitempty"`
	UseInstanceMetadata         bool   `json:"useInstanceMetadata,omitempty"`
	VMType                      string `json:"vmType,omitempty"`
	UseManagedIdentityExtension bool   `json:"useManagedIdentityExtension,omitempty"`
	UserAssignedIdentityID      string `json:"userAssignedIdentityID,omitempty"`
}

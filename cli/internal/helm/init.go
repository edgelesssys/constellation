/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// Initializer installs all Helm charts required for a constellation cluster.
type Initializer interface {
	Install(ctx context.Context, releases *Releases) error
}

type initializationClient struct {
	log       debugLog
	installer installer
}

// NewInitializer creates a new client to install all Helm charts required for a constellation cluster.
func NewInitializer(log debugLog) (Initializer, error) {
	installer, err := NewInstaller(constants.AdminConfFilename, log)
	if err != nil {
		return nil, fmt.Errorf("creating Helm installer: %w", err)
	}
	return &initializationClient{log: log, installer: installer}, nil
}

// Install installs all Helm charts required for a constellation cluster.
func (h initializationClient) Install(ctx context.Context, releases *Releases,
) error {
	if err := h.installer.InstallChart(ctx, releases.Cilium); err != nil {
		return fmt.Errorf("installing Cilium: %w", err)
	}
	h.log.Debugf("Waiting for Cilium to become ready")
	helper, err := newK8sCiliumHelper(constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}
	timeToStartWaiting := time.Now()
	// TODO(3u13r): Reduce the timeout when we switched the package repository - this is only this high because we once
	// saw polling times of ~16 minutes when hitting a slow PoP from Fastly (GitHub's / ghcr.io CDN).
	if err := helper.WaitForDS(ctx, "kube-system", "cilium", h.log); err != nil {
		return fmt.Errorf("waiting for Cilium to become healthy: %w", err)
	}
	timeUntilFinishedWaiting := time.Since(timeToStartWaiting)
	h.log.Debugf("Cilium became healthy after %s", timeUntilFinishedWaiting.String())

	h.log.Debugf("Fix Cilium through restart")
	if err := helper.RestartDS("kube-system", "cilium"); err != nil {
		return fmt.Errorf("restarting Cilium: %w", err)
	}

	h.log.Debugf("Installing microservices")
	if err := h.installer.InstallChart(ctx, releases.ConstellationServices); err != nil {
		return fmt.Errorf("installing microservices: %w", err)
	}

	h.log.Debugf("Installing cert-manager")
	if err := h.installer.InstallChart(ctx, releases.CertManager); err != nil {
		return fmt.Errorf("installing cert-manager: %w", err)
	}

	if releases.CSI != nil {
		h.log.Debugf("Installing CSI deployments")
		if err := h.installer.InstallChart(ctx, *releases.CSI); err != nil {
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
	if err := h.installer.InstallChart(ctx, releases.ConstellationOperators); err != nil {
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

// GetCCMConfig returns the configuration needed for the Kubernetes Cloud Controller Manager on Azure.
func getCCMConfig(tfOutput terraform.AzureApplyOutput, serviceAccURI string) ([]byte, error) {
	creds, err := azureshared.ApplicationCredentialsFromURI(serviceAccURI)
	if err != nil {
		return nil, fmt.Errorf("getting service account key: %w", err)
	}
	useManagedIdentityExtension := creds.PreferredAuthMethod == azureshared.AuthMethodUserAssignedIdentity
	config := cloudConfig{
		Cloud:                       "AzurePublicCloud",
		TenantID:                    creds.TenantID,
		SubscriptionID:              tfOutput.SubscriptionID,
		ResourceGroup:               tfOutput.ResourceGroup,
		LoadBalancerSku:             "standard",
		SecurityGroupName:           tfOutput.NetworkSecurityGroupName,
		LoadBalancerName:            tfOutput.LoadBalancerName,
		UseInstanceMetadata:         true,
		VMType:                      "vmss",
		Location:                    creds.Location,
		UseManagedIdentityExtension: useManagedIdentityExtension,
		UserAssignedIdentityID:      tfOutput.UserAssignedIdentity,
	}

	return json.Marshal(config)
}

/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
)

// SuiteInstaller installs all Helm charts required for a constellation cluster.
type SuiteInstaller interface {
	Install(ctx context.Context, provider cloudprovider.Provider, masterSecret uri.MasterSecret,
		idFile clusterid.File,
		serviceAccURI string, releases *Releases,
	) error
}

type helmInstallationClient struct {
	log       debugLog
	installer helmInstaller
}

// NewInstallationClient creates a new Helm installation client to install all Helm charts required for a constellation cluster.
func NewInstallationClient(log debugLog) (SuiteInstaller, error) {
	installer, err := NewInstaller(constants.AdminConfFilename, log)
	if err != nil {
		return nil, fmt.Errorf("creating Helm installer: %w", err)
	}
	return &helmInstallationClient{log: log, installer: installer}, nil
}

func (h helmInstallationClient) Install(ctx context.Context, provider cloudprovider.Provider, masterSecret uri.MasterSecret,
	idFile clusterid.File,
	serviceAccURI string, releases *Releases,
) error {
	tfClient, err := terraform.New(ctx, constants.TerraformWorkingDir)
	if err != nil {
		return fmt.Errorf("creating Terraform client: %w", err)
	}
	output, err := tfClient.ShowCluster(ctx, provider)
	if err != nil {
		return fmt.Errorf("getting Terraform output: %w", err)
	}

	helper, err := newK8sCiliumHelper(constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}
	ciliumVals, err := setupCiliumVals(ctx, provider, helper, output)
	if err != nil {
		return fmt.Errorf("setting up Cilium values: %w", err)
	}
	if err := h.installer.InstallChartWithValues(ctx, releases.Cilium, ciliumVals); err != nil {
		return fmt.Errorf("installing Cilium: %w", err)
	}
	h.log.Debugf("Waiting for Cilium to become healthy")
	timeToStartWaiting := time.Now()
	// TODO(3u13r): Reduce the timeout when we switched the package repository - this is only this high because we once
	// saw polling times of ~16 minutes when hitting a slow PoP from Fastly (GitHub's / ghcr.io CDN).
	if err := helper.WaitForDS(ctx, "kube-system", "cilium", h.log); err != nil {
		return fmt.Errorf("waiting for Cilium to become healthy: %w", err)
	}
	timeUntilFinishedWaiting := time.Since(timeToStartWaiting)
	h.log.Debugf("Cilium became healthy after %s", timeUntilFinishedWaiting.String())

	h.log.Debugf("Restarting Cilium")
	if err := helper.RestartDS("kube-system", "cilium"); err != nil {
		return fmt.Errorf("restarting Cilium: %w", err)
	}

	h.log.Debugf("Installing microservices")
	serviceVals, err := setupMicroserviceVals(provider, masterSecret.Salt, idFile.UID, serviceAccURI, output)
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
	return nil
}

type helmInstaller interface {
	InstallChart(context.Context, Release) error
	InstallChartWithValues(ctx context.Context, release Release, extraValues map[string]any) error
}

// lb_port: everywhere hardcoded in TF 6443, for host use ip
// TODO(malt3): switch over to DNS name on AWS and Azure
// soon as every apiserver certificate of every control-plane node
// has the dns endpoint in its SAN list.
func setupCiliumVals(ctx context.Context, provider cloudprovider.Provider, dsClient *k8sDsClient, output terraform.ApplyOutput) (map[string]any, error) {
	vals := map[string]any{
		"k8sServiceHost": output.IP,
		"k8sServicePort": 6443, // TODO take from tf?
	}

	// GCP requires extra configuration for Cilium
	if provider == cloudprovider.GCP {
		// TODO(elchead): remove?
		if err := dsClient.PatchNode(ctx, output.GCP.IPCidrNode); err != nil {
			return nil, fmt.Errorf("patching GCP node: %w", err)
		}
		vals["ipv4NativeRoutingCIDR"] = output.GCP.IPCidrPod
		vals["strictModeCIDR"] = output.GCP.IPCidrPod
	}
	return vals, nil
}

// setupMicroserviceVals returns the values for the microservice chart.
func setupMicroserviceVals(provider cloudprovider.Provider, measurementSalt []byte, uid, serviceAccURI string, output terraform.ApplyOutput) (map[string]any, error) {
	extraVals := map[string]any{
		"join-service": map[string]any{
			"measurementSalt": base64.StdEncoding.EncodeToString(measurementSalt),
		},
		"verification-service": map[string]any{
			"loadBalancerIP": output.IP,
		},
		"konnectivity": map[string]any{
			"loadBalancerIP": output.IP,
		},
	}
	switch provider {
	case cloudprovider.GCP:
		serviceAccountKey, err := gcpshared.ServiceAccountKeyFromURI(serviceAccURI)
		if err != nil {
			return nil, fmt.Errorf("getting service account key: %w", err)
		}
		rawKey, err := json.Marshal(serviceAccountKey)
		if err != nil {
			return nil, fmt.Errorf("marshaling service account key: %w", err)
		}
		if output.GCP == nil {
			return nil, fmt.Errorf("no GCP output from Terraform")
		}
		extraVals["ccm"] = map[string]any{
			"GCP": map[string]any{
				"projectID":         output.GCP.ProjectID,
				"uid":               uid,
				"secretData":        string(rawKey),
				"subnetworkPodCIDR": output.GCP.IPCidrPod,
			},
		}
	case cloudprovider.Azure:
		if output.Azure == nil {
			return nil, fmt.Errorf("no Azure output from Terraform")
		}
		ccmConfig, err := getCCMConfig(*output.Azure, serviceAccURI)
		if err != nil {
			return nil, fmt.Errorf("getting Azure CCM config: %w", err)
		}
		extraVals["ccm"] = map[string]any{
			"Azure": map[string]any{
				"azureConfig": string(ccmConfig),
			},
		}
	}

	return extraVals, nil
}

// setupOperatorVals returns the values for the constellation-operator chart.
func setupOperatorVals(_ context.Context, uid string) map[string]any {
	return map[string]any{
		"constellation-operator": map[string]any{
			"constellationUID": uid,
		},
	}
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

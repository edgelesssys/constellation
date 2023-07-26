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

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// SetupMicroserviceVals returns the values for the microservice chart.
func SetupMicroserviceVals(ctx context.Context, provider cloudprovider.Provider, measurementSalt []byte, uid, serviceAccURI string) (map[string]any, error) {
	tfClient, err := terraform.New(ctx, constants.TerraformWorkingDir)
	if err != nil {
		return nil, fmt.Errorf("creating Terraform client: %w", err)
	}
	output, err := tfClient.ShowCluster(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting Terraform output: %w", err)
	}
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

// SetupOperatorVals returns the values for the constellation-operator chart.
func SetupOperatorVals(_ context.Context, uid string) (map[string]any, error) {
	return map[string]any{
		"constellation-operator": map[string]any{
			"constellationUID": uid,
		},
	}, nil
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

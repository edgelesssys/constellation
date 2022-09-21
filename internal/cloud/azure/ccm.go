/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"encoding/json"

	"github.com/edgelesssys/constellation/v2/internal/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	k8s "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ccmMetadata interface {
	GetNetworkSecurityGroupName(ctx context.Context) (string, error)
	GetLoadBalancerName(ctx context.Context) (string, error)
}

// CloudControllerManager holds the Azure cloud-controller-manager configuration.
type CloudControllerManager struct {
	metadata ccmMetadata
}

func NewCloudControllerManager(metadata ccmMetadata) *CloudControllerManager {
	return &CloudControllerManager{
		metadata: metadata,
	}
}

// Image returns the container image used to provide cloud-controller-manager for the cloud-provider.
func (c *CloudControllerManager) Image(k8sVersion versions.ValidK8sVersion) (string, error) {
	return versions.VersionConfigs[k8sVersion].CloudControllerManagerImageAzure, nil
}

// Path returns the path used by cloud-controller-manager executable within the container image.
func (c *CloudControllerManager) Path() string {
	return "cloud-controller-manager"
}

// Name returns the cloud-provider name as used by k8s cloud-controller-manager (k8s.gcr.io/cloud-controller-manager).
func (c *CloudControllerManager) Name() string {
	return "azure"
}

// ExtraArgs returns a list of arguments to append to the cloud-controller-manager command.
func (c *CloudControllerManager) ExtraArgs() []string {
	return []string{
		"--controllers=*,-cloud-node",
		"--cloud-config=/etc/azure/azure.json",
		"--allocate-node-cidrs=false",
		"--configure-cloud-routes=true",
	}
}

// ConfigMaps returns a list of ConfigMaps to deploy together with the k8s cloud-controller-manager
// Reference: https://kubernetes.io/docs/concepts/configuration/configmap/ .
func (c *CloudControllerManager) ConfigMaps(instance metadata.InstanceMetadata) (kubernetes.ConfigMaps, error) {
	return kubernetes.ConfigMaps{}, nil
}

// Secrets returns a list of secrets to deploy together with the k8s cloud-controller-manager.
// Reference: https://kubernetes.io/docs/concepts/configuration/secret/ .
func (c *CloudControllerManager) Secrets(ctx context.Context, providerID string, cloudServiceAccountURI string) (kubernetes.Secrets, error) {
	// Azure CCM expects cloud provider config to contain cluster configuration and service principal client secrets
	// reference: https://kubernetes-sigs.github.io/cloud-provider-azure/install/configs/

	subscriptionID, resourceGroup, err := azureshared.BasicsFromProviderID(providerID)
	if err != nil {
		return kubernetes.Secrets{}, err
	}
	creds, err := azureshared.ApplicationCredentialsFromURI(cloudServiceAccountURI)
	if err != nil {
		return kubernetes.Secrets{}, err
	}

	vmType := "standard"
	if _, _, _, _, err := azureshared.ScaleSetInformationFromProviderID(providerID); err == nil {
		vmType = "vmss"
	}

	securityGroupName, err := c.metadata.GetNetworkSecurityGroupName(ctx)
	if err != nil {
		return kubernetes.Secrets{}, err
	}

	loadBalancerName, err := c.metadata.GetLoadBalancerName(ctx)
	if err != nil {
		return kubernetes.Secrets{}, err
	}

	config := cloudConfig{
		Cloud:               "AzurePublicCloud",
		TenantID:            creds.TenantID,
		SubscriptionID:      subscriptionID,
		ResourceGroup:       resourceGroup,
		LoadBalancerSku:     "standard",
		SecurityGroupName:   securityGroupName,
		LoadBalancerName:    loadBalancerName,
		UseInstanceMetadata: true,
		VMType:              vmType,
		Location:            creds.Location,
		AADClientID:         creds.AppClientID,
		AADClientSecret:     creds.ClientSecretValue,
	}

	rawConfig, err := json.Marshal(config)
	if err != nil {
		return kubernetes.Secrets{}, err
	}

	return kubernetes.Secrets{
		&k8s.Secret{
			TypeMeta: meta.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "azureconfig",
				Namespace: "kube-system",
			},
			Data: map[string][]byte{
				"azure.json": rawConfig,
			},
		},
	}, nil
}

// Volumes returns a list of volumes to deploy together with the k8s cloud-controller-manager.
// Reference: https://kubernetes.io/docs/concepts/storage/volumes/ .
func (c *CloudControllerManager) Volumes() []k8s.Volume {
	return []k8s.Volume{
		{
			Name: "azureconfig",
			VolumeSource: k8s.VolumeSource{
				Secret: &k8s.SecretVolumeSource{
					SecretName: "azureconfig",
				},
			},
		},
	}
}

// VolumeMounts a list of of volume mounts to deploy together with the k8s cloud-controller-manager.
func (c *CloudControllerManager) VolumeMounts() []k8s.VolumeMount {
	return []k8s.VolumeMount{
		{
			Name:      "azureconfig",
			ReadOnly:  true,
			MountPath: "/etc/azure",
		},
	}
}

// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cloud-controller-manager.
func (c *CloudControllerManager) Env() []k8s.EnvVar {
	return []k8s.EnvVar{}
}

// Supported is used to determine if cloud controller manager is implemented for this cloud provider.
func (c *CloudControllerManager) Supported() bool {
	return true
}

type cloudConfig struct {
	Cloud                      string `json:"cloud,omitempty"`
	TenantID                   string `json:"tenantId,omitempty"`
	SubscriptionID             string `json:"subscriptionId,omitempty"`
	ResourceGroup              string `json:"resourceGroup,omitempty"`
	Location                   string `json:"location,omitempty"`
	SubnetName                 string `json:"subnetName,omitempty"`
	SecurityGroupName          string `json:"securityGroupName,omitempty"`
	SecurityGroupResourceGroup string `json:"securityGroupResourceGroup,omitempty"`
	LoadBalancerName           string `json:"loadBalancerName,omitempty"`
	LoadBalancerSku            string `json:"loadBalancerSku,omitempty"`
	VNetName                   string `json:"vnetName,omitempty"`
	VNetResourceGroup          string `json:"vnetResourceGroup,omitempty"`
	CloudProviderBackoff       bool   `json:"cloudProviderBackoff,omitempty"`
	UseInstanceMetadata        bool   `json:"useInstanceMetadata,omitempty"`
	VMType                     string `json:"vmType,omitempty"`
	AADClientID                string `json:"aadClientId,omitempty"`
	AADClientSecret            string `json:"aadClientSecret,omitempty"`
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
)

var (
	publicIPAddressRegexp = regexp.MustCompile(`/subscriptions/[^/]+/resourceGroups/[^/]+/providers/Microsoft.Network/publicIPAddresses/(?P<IPname>[^/]+)`)
	keyPathRegexp         = regexp.MustCompile(`^\/home\/([^\/]+)\/\.ssh\/authorized_keys$`)
)

// Metadata implements azure metadata APIs.
type Metadata struct {
	imdsAPI
	virtualNetworksAPI
	securityGroupsAPI
	networkInterfacesAPI
	publicIPAddressesAPI
	scaleSetsAPI
	loadBalancerAPI
	virtualMachineScaleSetVMsAPI
	tagsAPI
	applicationInsightsAPI
}

// NewMetadata creates a new Metadata.
func NewMetadata(ctx context.Context) (*Metadata, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	// The default http client may use a system-wide proxy and it is recommended to disable the proxy explicitly:
	// https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service?tabs=linux#proxies
	// See also: https://github.com/microsoft/azureimds/blob/master/imdssample.go#L10
	imdsAPI := imdsClient{
		client: &http.Client{Transport: &http.Transport{Proxy: nil}},
	}
	subscriptionID, err := imdsAPI.SubscriptionID(ctx)
	if err != nil {
		return nil, err
	}
	virtualNetworksAPI, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	networkInterfacesAPI, err := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	publicIPAddressesAPI, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	securityGroupsAPI, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	scaleSetsAPI, err := armcomputev2.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	loadBalancerAPI, err := armnetwork.NewLoadBalancersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	virtualMachineScaleSetVMsAPI, err := armcomputev2.NewVirtualMachineScaleSetVMsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	tagsAPI, err := armresources.NewTagsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	applicationInsightsAPI, err := armapplicationinsights.NewComponentsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	return &Metadata{
		imdsAPI:                      &imdsAPI,
		virtualNetworksAPI:           virtualNetworksAPI,
		networkInterfacesAPI:         networkInterfacesAPI,
		securityGroupsAPI:            securityGroupsAPI,
		publicIPAddressesAPI:         publicIPAddressesAPI,
		loadBalancerAPI:              loadBalancerAPI,
		scaleSetsAPI:                 scaleSetsAPI,
		virtualMachineScaleSetVMsAPI: virtualMachineScaleSetVMsAPI,
		tagsAPI:                      tagsAPI,
		applicationInsightsAPI:       applicationInsightsAPI,
	}, nil
}

// List retrieves all instances belonging to the current constellation.
func (m *Metadata) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	resourceGroup, err := m.imdsAPI.ResourceGroup(ctx)
	if err != nil {
		return nil, err
	}
	scaleSetInstances, err := m.listScaleSetVMs(ctx, resourceGroup)
	if err != nil {
		return nil, err
	}
	return scaleSetInstances, nil
}

// Self retrieves the current instance.
func (m *Metadata) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	providerID, err := m.providerID(ctx)
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}
	return m.GetInstance(ctx, providerID)
}

// GetInstance retrieves an instance using its providerID.
func (m *Metadata) GetInstance(ctx context.Context, providerID string) (metadata.InstanceMetadata, error) {
	instance, scaleSetErr := m.getScaleSetVM(ctx, providerID)
	if scaleSetErr == nil {
		return instance, nil
	}
	return metadata.InstanceMetadata{}, fmt.Errorf("retrieving instance given providerID %v: %w", providerID, scaleSetErr)
}

// GetNetworkSecurityGroupName returns the security group name of the resource group.
func (m *Metadata) GetNetworkSecurityGroupName(ctx context.Context) (string, error) {
	resourceGroup, err := m.imdsAPI.ResourceGroup(ctx)
	if err != nil {
		return "", err
	}

	nsg, err := m.getNetworkSecurityGroup(ctx, resourceGroup)
	if err != nil {
		return "", err
	}
	if nsg == nil || nsg.Name == nil {
		return "", fmt.Errorf("could not dereference network security group name")
	}
	return *nsg.Name, nil
}

// GetSubnetworkCIDR retrieves the subnetwork CIDR from cloud provider metadata.
func (m *Metadata) GetSubnetworkCIDR(ctx context.Context) (string, error) {
	resourceGroup, err := m.imdsAPI.ResourceGroup(ctx)
	if err != nil {
		return "", err
	}
	virtualNetwork, err := m.getVirtualNetwork(ctx, resourceGroup)
	if err != nil {
		return "", err
	}
	if virtualNetwork == nil || virtualNetwork.Properties == nil || len(virtualNetwork.Properties.Subnets) == 0 ||
		virtualNetwork.Properties.Subnets[0].Properties == nil || virtualNetwork.Properties.Subnets[0].Properties.AddressPrefix == nil {
		return "", fmt.Errorf("could not retrieve subnetwork CIDR from virtual network %v", virtualNetwork)
	}

	return *virtualNetwork.Properties.Subnets[0].Properties.AddressPrefix, nil
}

// UID retrieves the UID of the constellation.
func (m *Metadata) UID(ctx context.Context) (string, error) {
	return m.imdsAPI.UID(ctx)
}

// getLoadBalancer retrieves the load balancer from cloud provider metadata.
func (m *Metadata) getLoadBalancer(ctx context.Context) (*armnetwork.LoadBalancer, error) {
	resourceGroup, err := m.imdsAPI.ResourceGroup(ctx)
	if err != nil {
		return nil, err
	}

	pager := m.loadBalancerAPI.NewListPager(resourceGroup, nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving loadbalancer page: %w", err)
		}
		for _, lb := range page.Value {
			if lb != nil && lb.Properties != nil {
				return lb, nil
			}
		}
	}
	return nil, fmt.Errorf("could not get any load balancer")
}

// SupportsLoadBalancer returns true if the cloud provider supports load balancers.
func (m *Metadata) SupportsLoadBalancer() bool {
	return true
}

// GetLoadBalancerName returns the load balancer name of the resource group.
func (m *Metadata) GetLoadBalancerName(ctx context.Context) (string, error) {
	lb, err := m.getLoadBalancer(ctx)
	if err != nil {
		return "", err
	}
	if lb == nil || lb.Name == nil {
		return "", fmt.Errorf("could not dereference load balancer name")
	}
	return *lb.Name, nil
}

// GetLoadBalancerEndpoint retrieves the first load balancer IP from cloud provider metadata.
//
// The returned string is an IP address without a port, but the method name needs to satisfy the
// metadata interface.
func (m *Metadata) GetLoadBalancerEndpoint(ctx context.Context) (string, error) {
	lb, err := m.getLoadBalancer(ctx)
	if err != nil {
		return "", err
	}
	if lb == nil || lb.Properties == nil {
		return "", fmt.Errorf("could not dereference load balancer IP configuration")
	}

	var pubIPID string
	for _, fipConf := range lb.Properties.FrontendIPConfigurations {
		if fipConf == nil || fipConf.Properties == nil || fipConf.Properties.PublicIPAddress == nil || fipConf.Properties.PublicIPAddress.ID == nil {
			continue
		}
		pubIPID = *fipConf.Properties.PublicIPAddress.ID
		break
	}

	if pubIPID == "" {
		return "", fmt.Errorf("could not find public IP address reference in load balancer")
	}

	matches := publicIPAddressRegexp.FindStringSubmatch(pubIPID)
	if len(matches) != 2 {
		return "", fmt.Errorf("could not find public IP address name in load balancer: %v", pubIPID)
	}
	pubIPName := matches[1]

	resourceGroup, err := m.imdsAPI.ResourceGroup(ctx)
	if err != nil {
		return "", err
	}

	resp, err := m.publicIPAddressesAPI.Get(ctx, resourceGroup, pubIPName, nil)
	if err != nil {
		return "", fmt.Errorf("could not retrieve public IP address: %w", err)
	}
	if resp.Properties == nil || resp.Properties.IPAddress == nil {
		return "", fmt.Errorf("could not resolve public IP address reference for load balancer")
	}
	return *resp.Properties.IPAddress, nil
}

// Supported is used to determine if metadata API is implemented for this cloud provider.
func (m *Metadata) Supported() bool {
	return true
}

// providerID retrieves the current instances providerID.
func (m *Metadata) providerID(ctx context.Context) (string, error) {
	providerID, err := m.imdsAPI.ProviderID(ctx)
	if err != nil {
		return "", err
	}
	return "azure://" + providerID, nil
}

func (m *Metadata) getAppInsights(ctx context.Context) (*armapplicationinsights.Component, error) {
	resourceGroup, err := m.imdsAPI.ResourceGroup(ctx)
	if err != nil {
		return nil, err
	}

	uid, err := m.UID(ctx)
	if err != nil {
		return nil, err
	}

	pager := m.applicationInsightsAPI.NewListByResourceGroupPager(resourceGroup, nil)

	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving application insights page: %w", err)
		}
		for _, component := range nextResult.Value {
			if component == nil || component.Tags == nil {
				continue
			}

			tag, ok := component.Tags[cloud.TagUID]
			if !ok || tag == nil {
				continue
			}

			if *tag == uid {
				return component, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find correctly tagged application insights")
}

// extractInstanceTags converts azure tags into metadata key-value pairs.
func extractInstanceTags(tags map[string]*string) map[string]string {
	metadataMap := map[string]string{}
	for key, value := range tags {
		if value == nil {
			continue
		}
		metadataMap[key] = *value
	}
	return metadataMap
}

// extractSSHKeys extracts SSH public keys from azure instance OS Profile.
func extractSSHKeys(sshConfig armcomputev2.SSHConfiguration) map[string][]string {
	sshKeys := map[string][]string{}
	for _, key := range sshConfig.PublicKeys {
		if key == nil || key.Path == nil || key.KeyData == nil {
			continue
		}
		matches := keyPathRegexp.FindStringSubmatch(*key.Path)
		if len(matches) != 2 {
			continue
		}
		sshKeys[matches[1]] = append(sshKeys[matches[1]], *key.KeyData)
	}
	return sshKeys
}

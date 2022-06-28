package azure

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/edgelesssys/constellation/internal/azureshared"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
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
	instanceMetadata, err := imdsAPI.Retrieve(ctx)
	if err != nil {
		return nil, err
	}
	subscriptionID, _, err := azureshared.BasicsFromProviderID("azure://" + instanceMetadata.Compute.ResourceID)
	if err != nil {
		return nil, err
	}
	virtualNetworksAPI := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)
	networkInterfacesAPI := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
	publicIPAddressesAPI := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	securityGroupsAPI := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	scaleSetsAPI := armcompute.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)
	loadBalancerAPI := armnetwork.NewLoadBalancersClient(subscriptionID, cred, nil)
	virtualMachineScaleSetVMsAPI := armcompute.NewVirtualMachineScaleSetVMsClient(subscriptionID, cred, nil)
	tagsAPI := armresources.NewTagsClient(subscriptionID, cred, nil)
	applicationInsightsAPI := armapplicationinsights.NewComponentsClient(subscriptionID, cred, nil)

	return &Metadata{
		imdsAPI:                      &imdsAPI,
		virtualNetworksAPI:           &virtualNetworksClient{virtualNetworksAPI},
		networkInterfacesAPI:         &networkInterfacesClient{networkInterfacesAPI},
		securityGroupsAPI:            &securityGroupsClient{securityGroupsAPI},
		publicIPAddressesAPI:         &publicIPAddressesClient{publicIPAddressesAPI},
		loadBalancerAPI:              &loadBalancersClient{loadBalancerAPI},
		scaleSetsAPI:                 &scaleSetsClient{scaleSetsAPI},
		virtualMachineScaleSetVMsAPI: &virtualMachineScaleSetVMsClient{virtualMachineScaleSetVMsAPI},
		tagsAPI:                      &tagsClient{tagsAPI},
		applicationInsightsAPI:       &applicationInsightsClient{applicationInsightsAPI},
	}, nil
}

// List retrieves all instances belonging to the current constellation.
func (m *Metadata) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	providerID, err := m.providerID(ctx)
	if err != nil {
		return nil, err
	}
	_, resourceGroup, err := azureshared.BasicsFromProviderID(providerID)
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
	providerID, err := m.providerID(ctx)
	if err != nil {
		return "", err
	}
	_, resourceGroup, err := azureshared.BasicsFromProviderID(providerID)
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
	providerID, err := m.providerID(ctx)
	if err != nil {
		return "", err
	}
	_, resourceGroup, err := azureshared.BasicsFromProviderID(providerID)
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

// getLoadBalancer retrieves the load balancer from cloud provider metadata.
func (m *Metadata) getLoadBalancer(ctx context.Context) (*armnetwork.LoadBalancer, error) {
	providerID, err := m.providerID(ctx)
	if err != nil {
		return nil, err
	}
	_, resourceGroup, err := azureshared.BasicsFromProviderID(providerID)
	if err != nil {
		return nil, err
	}
	pager := m.loadBalancerAPI.List(resourceGroup, nil)

	for pager.NextPage(ctx) {
		for _, lb := range pager.PageResponse().Value {
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

// GetLoadBalancerIP retrieves the first load balancer IP from cloud provider metadata.
func (m *Metadata) GetLoadBalancerIP(ctx context.Context) (string, error) {
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

	providerID, err := m.providerID(ctx)
	if err != nil {
		return "", err
	}
	_, resourceGroup, err := azureshared.BasicsFromProviderID(providerID)
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

// SetVPNIP stores the internally used VPN IP in cloud provider metadata (not required on azure).
func (m *Metadata) SetVPNIP(ctx context.Context, vpnIP string) error {
	return nil
}

// Supported is used to determine if metadata API is implemented for this cloud provider.
func (m *Metadata) Supported() bool {
	return true
}

// providerID retrieves the current instances providerID.
func (m *Metadata) providerID(ctx context.Context) (string, error) {
	instanceMetadata, err := m.imdsAPI.Retrieve(ctx)
	if err != nil {
		return "", err
	}
	return "azure://" + instanceMetadata.Compute.ResourceID, nil
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
func extractSSHKeys(sshConfig armcompute.SSHConfiguration) map[string][]string {
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

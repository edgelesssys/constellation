/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Implements interaction with the Azure API.

Instance metadata is retrieved from the [Azure IMDS API].

Retrieving metadata of other instances is done by using the Azure API, and requires Azure credentials.

[Azure IMDS API]: https://docs.microsoft.com/en-us/azure/virtual-machines/linux/instance-metadata-service
*/
package azure

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

// Cloud provides Azure metadata and API access.
type Cloud struct {
	imds            imdsAPI
	virtNetAPI      virtualNetworksAPI
	secGroupAPI     securityGroupsAPI
	netIfacAPI      networkInterfacesAPI
	pubIPAPI        publicIPAddressesAPI
	scaleSetsAPI    scaleSetsAPI
	loadBalancerAPI loadBalancerAPI
	scaleSetsVMAPI  virtualMachineScaleSetVMsAPI
}

// New initializes Cloud with the needed API clients.
// Default credentials are used for authentication.
func New(ctx context.Context) (*Cloud, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}

	imdsAPI := NewIMDSClient()
	subscriptionID, err := imdsAPI.subscriptionID(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving subscription ID: %w", err)
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
	scaleSetsAPI, err := armcompute.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	loadBalancerAPI, err := armnetwork.NewLoadBalancersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	virtualMachineScaleSetVMsAPI, err := armcompute.NewVirtualMachineScaleSetVMsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	return &Cloud{
		imds:            imdsAPI,
		netIfacAPI:      networkInterfacesAPI,
		virtNetAPI:      virtualNetworksAPI,
		secGroupAPI:     securityGroupsAPI,
		pubIPAPI:        publicIPAddressesAPI,
		loadBalancerAPI: loadBalancerAPI,
		scaleSetsAPI:    scaleSetsAPI,
		scaleSetsVMAPI:  virtualMachineScaleSetVMsAPI,
	}, nil
}

// GetLoadBalancerEndpoint retrieves the first load balancer IP from cloud provider metadata.
//
// The returned string is an IP address without a port, but the method name needs to satisfy the
// metadata interface.
func (c *Cloud) GetLoadBalancerEndpoint(ctx context.Context) (host, port string, retErr error) {
	var multiErr error

	// Try to retrieve the public IP first
	hostname, err := c.getLoadBalancerPublicIP(ctx)
	if err == nil {
		return hostname, strconv.FormatInt(constants.KubernetesPort, 10), nil
	}
	multiErr = fmt.Errorf("retrieving load balancer public IP: %w", err)

	// If that fails, try to retrieve the private IP
	hostname, err = c.getLoadBalancerPrivateIP(ctx)
	if err == nil {
		return hostname, strconv.FormatInt(constants.KubernetesPort, 10), nil
	}
	multiErr = errors.Join(multiErr, fmt.Errorf("retrieving load balancer private IP: %w", err))

	return "", "", multiErr
}

// List retrieves all instances belonging to the current constellation.
func (c *Cloud) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	resourceGroup, err := c.imds.resourceGroup(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving resource group: %w", err)
	}

	uid, err := c.imds.uid(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instance UID: %w", err)
	}

	instances := []metadata.InstanceMetadata{}
	pager := c.scaleSetsAPI.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving scale sets: %w", err)
		}

		for _, scaleSet := range page.Value {
			if scaleSet == nil || scaleSet.Name == nil || scaleSet.Tags == nil ||
				scaleSet.Tags[cloud.TagUID] == nil || *scaleSet.Tags[cloud.TagUID] != uid {
				continue
			}
			vmPager := c.scaleSetsVMAPI.NewListPager(resourceGroup, *scaleSet.Name, nil)
			for vmPager.More() {
				vmPage, err := vmPager.NextPage(ctx)
				if err != nil {
					return nil, fmt.Errorf("retrieving vms: %w", err)
				}

				for _, vm := range vmPage.Value {
					if vm == nil || vm.InstanceID == nil {
						continue
					}
					interfaces, err := c.getVMInterfaces(ctx, *vm, resourceGroup, *scaleSet.Name, *vm.InstanceID)
					if err != nil {
						return nil, fmt.Errorf("retrieving VM network interfaces: %w", err)
					}
					instance, err := convertToInstanceMetadata(*vm, interfaces)
					if err != nil {
						return nil, fmt.Errorf("converting VM to instance metadata: %w", err)
					}
					instances = append(instances, instance)
				}
			}
		}
	}
	return instances, nil
}

// Self retrieves the current instance.
func (c *Cloud) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	providerID, err := c.imds.providerID(ctx)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving provider ID: %w", err)
	}
	return c.getInstance(ctx, "azure://"+providerID)
}

// UID retrieves the UID of the constellation.
func (c *Cloud) UID(ctx context.Context) (string, error) {
	uid, err := c.imds.uid(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving instance UID: %w", err)
	}
	return uid, nil
}

// InitSecretHash retrieves the InitSecretHash of the current instance.
func (c *Cloud) InitSecretHash(ctx context.Context) ([]byte, error) {
	initSecretHash, err := c.imds.initSecretHash(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving init secret hash: %w", err)
	}
	return []byte(initSecretHash), nil
}

// getLoadBalancer retrieves a load balancer from cloud provider metadata.
func (c *Cloud) getLoadBalancer(ctx context.Context, resourceGroup, uid string) (*armnetwork.LoadBalancer, error) {
	pager := c.loadBalancerAPI.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving available load balancers: %w", err)
		}
		for _, lb := range page.Value {
			if lb == nil || lb.Tags == nil ||
				lb.Tags[cloud.TagUID] == nil || *lb.Tags[cloud.TagUID] != uid {
				continue
			}
			return lb, nil
		}
	}
	return nil, fmt.Errorf("load balancer with UID %s not found", uid)
}

// getInstance returns an Azure instance given a providerID.
func (c *Cloud) getInstance(ctx context.Context, providerID string) (metadata.InstanceMetadata, error) {
	_, resourceGroup, scaleSet, instanceID, err := azureshared.ScaleSetInformationFromProviderID(providerID)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("invalid provider ID: %w", err)
	}
	vmResp, err := c.scaleSetsVMAPI.Get(ctx, resourceGroup, scaleSet, instanceID, nil)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving instance: %w", err)
	}
	networkInterfaces, err := c.getVMInterfaces(ctx, vmResp.VirtualMachineScaleSetVM, resourceGroup, scaleSet, instanceID)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving VM network interfaces: %w", err)
	}
	instance, err := convertToInstanceMetadata(vmResp.VirtualMachineScaleSetVM, networkInterfaces)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("converting VM to instance metadata: %w", err)
	}

	return instance, nil
}

// getNetworkSecurityGroupName returns the security group name of the resource group.
func (c *Cloud) getNetworkSecurityGroupName(ctx context.Context, resourceGroup, uid string) (string, error) {
	pager := c.secGroupAPI.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("retrieving security groups: %w", err)
		}
		for _, secGroup := range page.Value {
			if secGroup == nil || secGroup.Name == nil || secGroup.Tags == nil ||
				secGroup.Tags[cloud.TagUID] == nil || *secGroup.Tags[cloud.TagUID] != uid {
				continue
			}
			return *secGroup.Name, nil
		}
	}
	return "", fmt.Errorf("network security group with UID %s not found in resource group %s", uid, resourceGroup)
}

// getSubnetworkCIDR retrieves the subnetwork CIDR from cloud provider metadata.
func (c *Cloud) getSubnetworkCIDR(ctx context.Context) (string, error) {
	resourceGroup, err := c.imds.resourceGroup(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving resource group: %w", err)
	}

	uid, err := c.imds.uid(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving instance UID: %w", err)
	}

	pager := c.virtNetAPI.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("retrieving virtual networks: %w", err)
		}

		for _, network := range page.Value {
			if network == nil || network.Properties == nil || len(network.Properties.Subnets) == 0 ||
				network.Properties.Subnets[0] == nil || network.Properties.Subnets[0].Properties == nil ||
				network.Properties.Subnets[0].Properties.AddressPrefix == nil ||
				network.Tags == nil || network.Tags[cloud.TagUID] == nil || *network.Tags[cloud.TagUID] != uid {
				continue
			}

			return *network.Properties.Subnets[0].Properties.AddressPrefix, nil
		}
	}

	return "", fmt.Errorf("no virtual network found matching UID %s in resource group %s", uid, resourceGroup)
}

// getVMInterfaces retrieves all network interfaces referenced by a scale set virtual machine.
func (c *Cloud) getVMInterfaces(ctx context.Context, vm armcompute.VirtualMachineScaleSetVM, resourceGroup, scaleSet, instanceID string) ([]armnetwork.Interface, error) {
	if vm.Properties == nil || vm.Properties.NetworkProfile == nil {
		return []armnetwork.Interface{}, errors.New("no network profile found")
	}

	var interfaceNames []string
	for _, iface := range vm.Properties.NetworkProfile.NetworkInterfaces {
		if iface == nil || iface.ID == nil {
			continue
		}
		interfaceNames = append(interfaceNames, path.Base(*iface.ID))
	}

	networkInterfaces := []armnetwork.Interface{}
	for _, interfaceName := range interfaceNames {
		networkInterfacesResp, err := c.netIfacAPI.GetVirtualMachineScaleSetNetworkInterface(ctx, resourceGroup, scaleSet, instanceID, interfaceName, nil)
		if err != nil {
			return nil, fmt.Errorf("retrieving network interface %v: %w", interfaceName, err)
		}
		networkInterfaces = append(networkInterfaces, networkInterfacesResp.Interface)
	}
	return networkInterfaces, nil
}

// getLoadBalancerPrivateIP retrieves the first load balancer IP from cloud provider metadata.
func (c *Cloud) getLoadBalancerPrivateIP(ctx context.Context) (string, error) {
	resourceGroup, err := c.imds.resourceGroup(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving resource group: %w", err)
	}
	uid, err := c.imds.uid(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving instance UID: %w", err)
	}

	lb, err := c.getLoadBalancer(ctx, resourceGroup, uid)
	if err != nil {
		return "", fmt.Errorf("retrieving load balancer: %w", err)
	}
	if lb == nil || lb.Properties == nil {
		return "", errors.New("could not dereference load balancer IP configuration")
	}

	var privIP string
	for _, fipConf := range lb.Properties.FrontendIPConfigurations {
		if fipConf != nil && fipConf.Properties != nil && fipConf.Properties.PrivateIPAddress != nil {
			privIP = *fipConf.Properties.PrivateIPAddress
			break
		}
	}
	if privIP == "" {
		return "", errors.New("could not resolve private IP address for load balancer")
	}

	return privIP, nil
}

// getLoadBalancerPublicIP retrieves the first load balancer IP from cloud provider metadata.
func (c *Cloud) getLoadBalancerPublicIP(ctx context.Context) (string, error) {
	resourceGroup, err := c.imds.resourceGroup(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving resource group: %w", err)
	}
	uid, err := c.imds.uid(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving instance UID: %w", err)
	}

	lb, err := c.getLoadBalancer(ctx, resourceGroup, uid)
	if err != nil {
		return "", fmt.Errorf("retrieving load balancer: %w", err)
	}
	if lb == nil || lb.Properties == nil {
		return "", errors.New("could not dereference load balancer IP configuration")
	}

	var pubIP string
	for _, fipConf := range lb.Properties.FrontendIPConfigurations {
		if fipConf == nil || fipConf.Properties == nil || fipConf.Properties.PublicIPAddress == nil || fipConf.Properties.PublicIPAddress.ID == nil {
			continue
		}
		pubIP = path.Base(*fipConf.Properties.PublicIPAddress.ID)
		break
	}

	resp, err := c.pubIPAPI.Get(ctx, resourceGroup, pubIP, nil)
	if err != nil {
		return "", fmt.Errorf("retrieving load balancer public IP address: %w", err)
	}
	if resp.Properties == nil || resp.Properties.IPAddress == nil {
		return "", fmt.Errorf("could not resolve public IP address reference for load balancer")
	}
	return *resp.Properties.IPAddress, nil
}

/*
// TODO(malt3): uncomment and use as soon as we switch the primary endpoint to DNS.
// Addition from 3u13r: We have to think about how to handle DNS for internal load balancers
// that only have a private IP address and therefore no DNS name by default.
//
// getLoadBalancerDNSName retrieves the dns name of the load balancer.
// On Azure, the DNS name is the DNS name of the public IP address of the load balancer.
func (c *Cloud) getLoadBalancerDNSName(ctx context.Context) (string, error) {
	resourceGroup, err := c.imds.resourceGroup(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving resource group: %w", err)
	}
	uid, err := c.imds.uid(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving instance UID: %w", err)
	}

	lb, err := c.getLoadBalancer(ctx, resourceGroup, uid)
	if err != nil {
		return "", fmt.Errorf("retrieving load balancer: %w", err)
	}
	if lb == nil || lb.Properties == nil {
		return "", errors.New("could not dereference load balancer IP configuration")
	}

	var pubIP string
	for _, fipConf := range lb.Properties.FrontendIPConfigurations {
		if fipConf == nil || fipConf.Properties == nil || fipConf.Properties.PublicIPAddress == nil || fipConf.Properties.PublicIPAddress.ID == nil {
			continue
		}
		pubIP = path.Base(*fipConf.Properties.PublicIPAddress.ID)
		break
	}

	resp, err := c.pubIPAPI.Get(ctx, resourceGroup, pubIP, nil)
	if err != nil {
		return "", fmt.Errorf("retrieving load balancer public IP address: %w", err)
	}
	if resp.Properties == nil || resp.Properties.DNSSettings == nil || resp.Properties.DNSSettings.Fqdn == nil {
		return "", fmt.Errorf("could not resolve public IP address fqdn for load balancer")
	}
	return *resp.Properties.DNSSettings.Fqdn, nil
}
*/

// convertToInstanceMetadata converts a armcomputev2.VirtualMachineScaleSetVM to a metadata.InstanceMetadata.
func convertToInstanceMetadata(vm armcompute.VirtualMachineScaleSetVM, networkInterfaces []armnetwork.Interface,
) (metadata.InstanceMetadata, error) {
	if vm.ID == nil {
		return metadata.InstanceMetadata{}, errors.New("missing instance ID")
	}
	if vm.Properties == nil || vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil {
		return metadata.InstanceMetadata{}, errors.New("missing computer name")
	}
	var instanceRole string
	if vm.Tags != nil || vm.Tags[cloud.TagRole] != nil {
		instanceRole = *vm.Tags[cloud.TagRole]
	}

	var privateIP string
	for _, networkInterface := range networkInterfaces {
		if networkInterface.Properties == nil {
			continue
		}
		for _, config := range networkInterface.Properties.IPConfigurations {
			if config == nil || config.Properties == nil || config.Properties.PrivateIPAddress == nil || config.Properties.Primary == nil {
				continue
			}
			if *config.Properties.Primary {
				privateIP = *config.Properties.PrivateIPAddress
			}
		}
	}

	return metadata.InstanceMetadata{
		Name:       *vm.Properties.OSProfile.ComputerName,
		ProviderID: "azure://" + *vm.ID,
		Role:       role.FromString(instanceRole),
		VPCIP:      privateIP,
	}, nil
}

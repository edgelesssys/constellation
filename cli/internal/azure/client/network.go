package client

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/cli/internal/azure"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
)

type createNetworkInput struct {
	name             string
	location         string
	addressSpace     string
	nodeAddressSpace string
	podAddressSpace  string
}

const (
	nodeNetworkName     = "nodeNetwork"
	podNetworkName      = "podNetwork"
	networkAddressSpace = "10.0.0.0/8"
	nodeAddressSpace    = "10.9.0.0/16"
	podAddressSpace     = "10.10.0.0/16"
)

// CreateVirtualNetwork creates a virtual network.
func (c *Client) CreateVirtualNetwork(ctx context.Context) error {
	createNetworkInput := createNetworkInput{
		name:             "constellation-" + c.uid,
		location:         c.location,
		addressSpace:     networkAddressSpace,
		nodeAddressSpace: nodeAddressSpace,
		podAddressSpace:  podAddressSpace,
	}

	poller, err := c.networksAPI.BeginCreateOrUpdate(
		ctx, c.resourceGroup, createNetworkInput.name,
		armnetwork.VirtualNetwork{
			Name:     to.StringPtr(createNetworkInput.name), // this is supposed to be read-only
			Location: to.StringPtr(createNetworkInput.location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{
						to.StringPtr(createNetworkInput.addressSpace),
					},
				},
				Subnets: []*armnetwork.Subnet{
					{
						Name: to.StringPtr(nodeNetworkName),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr(createNetworkInput.nodeAddressSpace),
						},
					},
					{
						Name: to.StringPtr(podNetworkName),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr(createNetworkInput.podAddressSpace),
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return err
	}
	resp, err := poller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return err
	}
	c.subnetID = *resp.VirtualNetworksClientCreateOrUpdateResult.VirtualNetwork.Properties.Subnets[0].ID
	return nil
}

type createNetworkSecurityGroupInput struct {
	name     string
	location string
	rules    []*armnetwork.SecurityRule
}

// CreateSecurityGroup creates a security group containing firewall rules.
func (c *Client) CreateSecurityGroup(ctx context.Context, input NetworkSecurityGroupInput) error {
	rules, err := input.Ingress.Azure()
	if err != nil {
		return err
	}

	createNetworkSecurityGroupInput := createNetworkSecurityGroupInput{
		name:     "constellation-security-group-" + c.uid,
		location: c.location,
		rules:    rules,
	}

	poller, err := c.networkSecurityGroupsAPI.BeginCreateOrUpdate(
		ctx, c.resourceGroup, createNetworkSecurityGroupInput.name,
		armnetwork.SecurityGroup{
			Name:     to.StringPtr(createNetworkSecurityGroupInput.name),
			Location: to.StringPtr(createNetworkSecurityGroupInput.location),
			Properties: &armnetwork.SecurityGroupPropertiesFormat{
				SecurityRules: createNetworkSecurityGroupInput.rules,
			},
		},
		nil,
	)
	if err != nil {
		return err
	}
	pollerResp, err := poller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return err
	}
	c.networkSecurityGroup = *pollerResp.SecurityGroupsClientCreateOrUpdateResult.SecurityGroup.ID
	return nil
}

// createNIC creates a network interface that references a public IP address.
// TODO: deprecate as soon as scale sets are available.
func (c *Client) createNIC(ctx context.Context, name, publicIPAddressID string) (ip string, id string, err error) {
	poller, err := c.networkInterfacesAPI.BeginCreateOrUpdate(
		ctx, c.resourceGroup, name,
		armnetwork.Interface{
			Location: to.StringPtr(c.location),
			Properties: &armnetwork.InterfacePropertiesFormat{
				NetworkSecurityGroup: &armnetwork.SecurityGroup{
					ID: to.StringPtr(c.networkSecurityGroup),
				},
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.StringPtr(name),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							Subnet: &armnetwork.Subnet{
								ID: to.StringPtr(c.subnetID),
							},
							PublicIPAddress: &armnetwork.PublicIPAddress{
								ID: to.StringPtr(publicIPAddressID),
							},
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return "", "", err
	}
	pollerResp, err := poller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return "", "", err
	}

	netInterface := pollerResp.InterfacesClientCreateOrUpdateResult.Interface

	return *netInterface.Properties.IPConfigurations[0].Properties.PrivateIPAddress,
		*netInterface.ID,
		nil
}

func (c *Client) createPublicIPAddress(ctx context.Context, name string) (*armnetwork.PublicIPAddress, error) {
	poller, err := c.publicIPAddressesAPI.BeginCreateOrUpdate(
		ctx, c.resourceGroup, name,
		armnetwork.PublicIPAddress{
			Location: to.StringPtr(c.location),
			SKU: &armnetwork.PublicIPAddressSKU{
				Name: armnetwork.PublicIPAddressSKUNameStandard.ToPtr(),
			},
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: armnetwork.IPAllocationMethodStatic.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	pollerResp, err := poller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return nil, err
	}

	return &pollerResp.PublicIPAddressesClientCreateOrUpdateResult.PublicIPAddress, nil
}

// NetworkSecurityGroupInput defines firewall rules to be set.
type NetworkSecurityGroupInput struct {
	Ingress cloudtypes.Firewall
	Egress  cloudtypes.Firewall
}

// CreateExternalLoadBalancer creates an external load balancer.
func (c *Client) CreateExternalLoadBalancer(ctx context.Context) error {
	// First, create a public IP address for the load balancer.
	publicIPAddress, err := c.createPublicIPAddress(ctx, "loadbalancer-public-ip-"+c.uid)
	if err != nil {
		return err
	}

	// Then, create the load balancer.
	loadBalancerName := "constellation-load-balancer-" + c.uid
	loadBalancer := azure.LoadBalancer{
		Name:          loadBalancerName,
		Location:      c.location,
		ResourceGroup: c.resourceGroup,
		Subscription:  c.subscriptionID,
		PublicIPID:    *publicIPAddress.ID,
		UID:           c.uid,
	}
	azureLoadBalancer := loadBalancer.Azure()

	poller, err := c.loadBalancersAPI.BeginCreateOrUpdate(
		ctx, c.resourceGroup, loadBalancerName,
		azureLoadBalancer,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return err
	}
	c.loadBalancerName = loadBalancerName

	c.loadBalancerPubIP = *publicIPAddress.Properties.IPAddress
	return nil
}

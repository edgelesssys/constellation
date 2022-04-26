package client

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
)

type createNetworkInput struct {
	name         string
	location     string
	addressSpace string
}

// CreateVirtualNetwork creates a virtual network.
func (c *Client) CreateVirtualNetwork(ctx context.Context) error {
	createNetworkInput := createNetworkInput{
		name:         "constellation-" + c.uid,
		location:     c.location,
		addressSpace: "172.20.0.0/16",
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
						Name: to.StringPtr("default"),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr(createNetworkInput.addressSpace),
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

// createPublicIPAddress creates a public IP address.
// TODO: deprecate as soon as scale sets are available.
func (c *Client) createPublicIPAddress(ctx context.Context, name string) (string, error) {
	poller, err := c.publicIPAddressesAPI.BeginCreateOrUpdate(
		ctx, c.resourceGroup, name,
		armnetwork.PublicIPAddress{
			Location: to.StringPtr(c.location),
		},
		nil,
	)
	if err != nil {
		return "", err
	}
	pollerResp, err := poller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return "", err
	}

	return *pollerResp.PublicIPAddressesClientCreateOrUpdateResult.PublicIPAddress.ID, nil
}

// NetworkSecurityGroupInput defines firewall rules to be set.
type NetworkSecurityGroupInput struct {
	Ingress cloudtypes.Firewall
	Egress  cloudtypes.Firewall
}

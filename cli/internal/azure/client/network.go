/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
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
			Name:     to.Ptr(createNetworkInput.name), // this is supposed to be read-only
			Tags:     map[string]*string{"uid": to.Ptr(c.uid)},
			Location: to.Ptr(createNetworkInput.location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{
						to.Ptr(createNetworkInput.addressSpace),
					},
				},
				Subnets: []*armnetwork.Subnet{
					{
						Name: to.Ptr(nodeNetworkName),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.Ptr(createNetworkInput.nodeAddressSpace),
						},
					},
					{
						Name: to.Ptr(podNetworkName),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.Ptr(createNetworkInput.podAddressSpace),
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
	resp, err := poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: c.pollFrequency,
	})
	if err != nil {
		return err
	}
	c.subnetID = *resp.VirtualNetwork.Properties.Subnets[0].ID
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
			Name:     to.Ptr(createNetworkSecurityGroupInput.name),
			Tags:     map[string]*string{"uid": to.Ptr(c.uid)},
			Location: to.Ptr(createNetworkSecurityGroupInput.location),
			Properties: &armnetwork.SecurityGroupPropertiesFormat{
				SecurityRules: createNetworkSecurityGroupInput.rules,
			},
		},
		nil,
	)
	if err != nil {
		return err
	}
	pollerResp, err := poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: c.pollFrequency,
	})
	if err != nil {
		return err
	}
	c.networkSecurityGroup = *pollerResp.SecurityGroup.ID
	return nil
}

func (c *Client) createPublicIPAddress(ctx context.Context, name string) (*armnetwork.PublicIPAddress, error) {
	poller, err := c.publicIPAddressesAPI.BeginCreateOrUpdate(
		ctx, c.resourceGroup, name,
		armnetwork.PublicIPAddress{
			Tags:     map[string]*string{"uid": to.Ptr(c.uid)},
			Location: to.Ptr(c.location),
			SKU: &armnetwork.PublicIPAddressSKU{
				Name: to.Ptr(armnetwork.PublicIPAddressSKUNameStandard),
			},
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	pollerResp, err := poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: c.pollFrequency,
	})
	if err != nil {
		return nil, err
	}

	return &pollerResp.PublicIPAddress, nil
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

	_, err = poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: c.pollFrequency,
	})
	if err != nil {
		return err
	}
	c.loadBalancerName = loadBalancerName

	c.loadBalancerPubIP = *publicIPAddress.Properties.IPAddress
	return nil
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/utils/v2/openstack/clientconfig"
)

const (
	roleTagFormat = "constellation-role-%s"
	microversion  = "2.42"
)

// MetadataClient is the metadata client for OpenStack.
type MetadataClient struct {
	api  serversAPI
	imds imdsAPI
}

// New creates a new OpenStack metadata client.
func New(ctx context.Context) (*MetadataClient, error) {
	imds := &imdsClient{client: &http.Client{}}

	authURL, err := imds.authURL(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting auth URL: %w", err)
	}
	username, err := imds.username(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting token name: %w", err)
	}
	password, err := imds.password(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting token password: %w", err)
	}
	userDomainName, err := imds.userDomainName(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user domain name: %w", err)
	}
	regionName, err := imds.regionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting region name: %w", err)
	}

	clientOpts := &clientconfig.ClientOpts{
		AuthType: clientconfig.AuthV3Password,
		AuthInfo: &clientconfig.AuthInfo{
			AuthURL:        authURL,
			UserDomainName: userDomainName,
			Username:       username,
			Password:       password,
		},
		RegionName: regionName,
	}

	serversClient, err := clientconfig.NewServiceClient(ctx, "compute", clientOpts)
	if err != nil {
		return nil, fmt.Errorf("creating compute client: %w", err)
	}
	serversClient.Microversion = microversion

	networksClient, err := clientconfig.NewServiceClient(ctx, "network", clientOpts)
	if err != nil {
		return nil, fmt.Errorf("creating network client: %w", err)
	}
	networksClient.Microversion = microversion

	return &MetadataClient{
		imds: imds,
		api: &apiClient{
			servers:  serversClient,
			networks: networksClient,
		},
	}, nil
}

// Self returns the metadata of the current instance.
func (c *MetadataClient) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	name, err := c.imds.name(ctx)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("getting name: %w", err)
	}
	providerID, err := c.imds.providerID(ctx)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("getting provider id: %w", err)
	}
	role, err := c.imds.role(ctx)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("getting role: %w", err)
	}
	vpcIP, err := c.imds.vpcIP(ctx)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("getting vpc ip: %w", err)
	}

	return metadata.InstanceMetadata{
		Name:       name,
		ProviderID: providerID,
		Role:       role,
		VPCIP:      vpcIP,
	}, nil
}

// List returns the metadata of all instances belonging to the same Constellation cluster.
func (c *MetadataClient) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	uid, err := c.imds.uid(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting uid: %w", err)
	}

	uidTag := fmt.Sprintf("constellation-uid-%s", uid)

	subnet, err := c.getSubnetCIDR(ctx, uidTag)
	if err != nil {
		return nil, err
	}

	srvs, err := c.getServers(ctx, uidTag)
	if err != nil {
		return nil, err
	}

	var result []metadata.InstanceMetadata
	for _, s := range srvs {
		if s.Name == "" {
			continue
		}
		if s.ID == "" {
			continue
		}
		if s.Tags == nil {
			continue
		}

		var serverRole role.Role
		for _, t := range *s.Tags {
			if strings.HasPrefix(t, "constellation-role-") {
				serverRole = role.FromString(strings.TrimPrefix(t, "constellation-role-"))
				break
			}
		}
		if serverRole == role.Unknown {
			continue
		}

		subnetAddrs, err := parseSeverAddresses(s.Addresses)
		if err != nil {
			return nil, fmt.Errorf("parsing server %q addresses: %w", s.Name, err)
		}

		var vpcIP string
		// In a best effort approach, we take the first fixed IPv4 address that is in the subnet
		// belonging to our cluster.
		for _, serverSubnet := range subnetAddrs {
			for _, addr := range serverSubnet.Addresses {
				if addr.Type != fixedIP {
					continue
				}

				if addr.IPVersion != ipV4 {
					continue
				}

				if addr.IP == "" {
					continue
				}

				parsedAddr, err := netip.ParseAddr(addr.IP)
				if err != nil {
					continue
				}

				if !subnet.Contains(parsedAddr) {
					continue
				}

				vpcIP = addr.IP
				break
			}
		}
		if vpcIP == "" {
			continue
		}

		im := metadata.InstanceMetadata{
			Name:       s.Name,
			ProviderID: s.ID,
			Role:       serverRole,
			VPCIP:      vpcIP,
		}
		result = append(result, im)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no instances belonging to this cluster found")
	}

	return result, nil
}

// UID retrieves the UID of the constellation.
func (c *MetadataClient) UID(ctx context.Context) (string, error) {
	uid, err := c.imds.uid(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving instance UID: %w", err)
	}
	return uid, nil
}

// InitSecretHash retrieves the InitSecretHash of the current instance.
func (c *MetadataClient) InitSecretHash(ctx context.Context) ([]byte, error) {
	initSecretHash, err := c.imds.initSecretHash(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving init secret hash: %w", err)
	}
	return []byte(initSecretHash), nil
}

// GetLoadBalancerEndpoint returns the endpoint of the load balancer.
// For OpenStack, the load balancer is a floating ip attached to
// a control plane node.
// TODO(malt3): Rewrite to use real load balancer once it is available.
func (c *MetadataClient) GetLoadBalancerEndpoint(ctx context.Context) (host, port string, err error) {
	host, err = c.imds.loadBalancerEndpoint(ctx)
	if err != nil {
		return "", "", fmt.Errorf("getting load balancer endpoint: %w", err)
	}
	return host, strconv.FormatInt(constants.KubernetesPort, 10), nil
}

func (c *MetadataClient) getSubnetCIDR(ctx context.Context, uidTag string) (netip.Prefix, error) {
	listNetworksOpts := networks.ListOpts{Tags: uidTag}
	networksPage, err := c.api.ListNetworks(listNetworksOpts).AllPages(ctx)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("listing networks: %w", err)
	}
	nets, err := networks.ExtractNetworks(networksPage)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("extracting networks: %w", err)
	}
	if len(nets) != 1 {
		return netip.Prefix{}, fmt.Errorf("expected exactly one network, got %d", len(nets))
	}

	listSubnetsOpts := subnets.ListOpts{Tags: uidTag}
	subnetsPage, err := c.api.ListSubnets(listSubnetsOpts).AllPages(ctx)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("listing subnets: %w", err)
	}

	snets, err := subnets.ExtractSubnets(subnetsPage)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("extracting subnets: %w", err)
	}

	if len(snets) < 1 {
		return netip.Prefix{}, fmt.Errorf("expected at least one subnet, got %d", len(snets))
	}

	var rawCIDR string
	for _, n := range snets {
		if n.Name == nets[0].Name {
			rawCIDR = n.CIDR
			break
		}
	}

	cidr, err := netip.ParsePrefix(rawCIDR)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("parsing subnet CIDR: %w", err)
	}

	return cidr, nil
}

func (c *MetadataClient) getServers(ctx context.Context, uidTag string) ([]servers.Server, error) {
	listServersOpts := servers.ListOpts{Tags: uidTag}
	serversPage, err := c.api.ListServers(listServersOpts).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}
	servers, err := servers.ExtractServers(serversPage)
	if err != nil {
		return nil, fmt.Errorf("extracting servers: %w", err)
	}

	return servers, nil
}

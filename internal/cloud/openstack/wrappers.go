/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
)

type apiClient struct {
	servers  *gophercloud.ServiceClient
	networks *gophercloud.ServiceClient
}

func (c *apiClient) ListServers(opts servers.ListOptsBuilder) pagerAPI {
	return servers.List(c.servers, opts)
}

func (c *apiClient) ListNetworks(opts networks.ListOptsBuilder) pagerAPI {
	return networks.List(c.networks, opts)
}

func (c *apiClient) ListSubnets(opts subnets.ListOpts) pagerAPI {
	return subnets.List(c.networks, opts)
}

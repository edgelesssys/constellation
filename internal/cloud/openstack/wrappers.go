/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

type apiClient struct {
	servers *gophercloud.ServiceClient
	subnets *gophercloud.ServiceClient
}

func (c *apiClient) ListServers(opts servers.ListOptsBuilder) pagerAPI {
	return servers.List(c.servers, opts)
}

func (c *apiClient) ListSubnets(opts subnets.ListOpts) pagerAPI {
	return subnets.List(c.subnets, opts)
}

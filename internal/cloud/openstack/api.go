/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/pagination"
)

type imdsAPI interface {
	providerID(ctx context.Context) (string, error)
	name(ctx context.Context) (string, error)
	projectID(ctx context.Context) (string, error)
	uid(ctx context.Context) (string, error)
	initSecretHash(ctx context.Context) (string, error)
	role(ctx context.Context) (role.Role, error)
	vpcIP(ctx context.Context) (string, error)
	networkIDs(ctx context.Context) ([]string, error)
}

type serversAPI interface {
	ListServers(opts servers.ListOptsBuilder) pagerAPI
	ListSubnets(opts subnets.ListOpts) pagerAPI
}

type pagerAPI interface {
	AllPages() (pagination.Page, error)
}

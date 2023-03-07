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

type stubIMDSClient struct {
	providerIDResult     string
	providerIDErr        error
	nameResult           string
	nameErr              error
	projectIDResult      string
	projectIDErr         error
	uidResult            string
	uidErr               error
	initSecretHashResult string
	initSecretHashErr    error
	roleResult           role.Role
	roleErr              error
	vpcIPResult          string
	vpcIPErr             error
}

func (c *stubIMDSClient) providerID(ctx context.Context) (string, error) {
	return c.providerIDResult, c.providerIDErr
}

func (c *stubIMDSClient) name(ctx context.Context) (string, error) {
	return c.nameResult, c.nameErr
}

func (c *stubIMDSClient) projectID(ctx context.Context) (string, error) {
	return c.projectIDResult, c.projectIDErr
}

func (c *stubIMDSClient) uid(ctx context.Context) (string, error) {
	return c.uidResult, c.uidErr
}

func (c *stubIMDSClient) initSecretHash(ctx context.Context) (string, error) {
	return c.initSecretHashResult, c.initSecretHashErr
}

func (c *stubIMDSClient) role(ctx context.Context) (role.Role, error) {
	return c.roleResult, c.roleErr
}

func (c *stubIMDSClient) vpcIP(ctx context.Context) (string, error) {
	return c.vpcIPResult, c.vpcIPErr
}

type stubServersClient struct {
	serversPager stubPager
	subnetsPager stubPager
}

func (c *stubServersClient) ListServers(opts servers.ListOptsBuilder) pagerAPI {
	return &c.serversPager
}

func (c *stubServersClient) ListSubnets(opts subnets.ListOpts) pagerAPI {
	return &c.subnetsPager
}

type stubPager struct {
	page        pagination.Page
	allPagesErr error
}

func (p *stubPager) AllPages() (pagination.Page, error) {
	return p.page, p.allPagesErr
}

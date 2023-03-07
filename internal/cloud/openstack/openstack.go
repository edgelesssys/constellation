/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
)

// Cloud is the metadata client for OpenStack.
type Cloud struct {
	imds imdsAPI
}

// New creates a new OpenStack metadata client.
func New(ctx context.Context) (*Cloud, error) {
	imds := &imdsClient{client: &http.Client{}}
	return &Cloud{imds: imds}, nil
}

// Self returns the metadata of the current instance.
func (c *Cloud) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
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

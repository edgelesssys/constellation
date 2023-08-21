/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package cloudprovider implements a metadata service for cloud providers.
package cloudprovider

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

type providerMetadata interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]metadata.InstanceMetadata, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
	// GetLoadBalancerEndpoint returns the endpoint of the load balancer.
	GetLoadBalancerEndpoint(ctx context.Context) (host, port string, err error)
	// UID returns the UID of the current instance.
	UID(ctx context.Context) (string, error)
}

// Fetcher checks the metadata service to search for instances that were set up for debugging.
type Fetcher struct {
	metaAPI providerMetadata
}

// New creates a new Fetcher.
func New(cloud providerMetadata) *Fetcher {
	return &Fetcher{
		metaAPI: cloud,
	}
}

// Role returns node role via meta data API.
func (f *Fetcher) Role(ctx context.Context) (role.Role, error) {
	self, err := f.metaAPI.Self(ctx)
	if err != nil {
		return role.Unknown, fmt.Errorf("retrieving role from cloud provider metadata: %w", err)
	}

	return self.Role, nil
}

// UID returns node UID via meta data API.
func (f *Fetcher) UID(ctx context.Context) (string, error) {
	uid, err := f.metaAPI.UID(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving UID from cloud provider metadata: %w", err)
	}
	return uid, nil
}

// Self returns the current instance via meta data API.
func (f *Fetcher) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	return f.metaAPI.Self(ctx)
}

// DiscoverDebugdIPs will query the metadata of all instances and return any ips of instances already set up for debugging.
func (f *Fetcher) DiscoverDebugdIPs(ctx context.Context) ([]string, error) {
	self, err := f.metaAPI.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving own instance: %w", err)
	}
	instances, err := f.metaAPI.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances: %w", err)
	}
	// filter own instance from instance list
	for i, instance := range instances {
		if instance.ProviderID == self.ProviderID {
			instances = append(instances[:i], instances[i+1:]...)
			break
		}
	}
	var ips []string
	for _, instance := range instances {
		if instance.VPCIP != "" {
			ips = append(ips, instance.VPCIP)
		}
	}
	return ips, nil
}

// DiscoverLoadBalancerIP gets load balancer IP from metadata API.
func (f *Fetcher) DiscoverLoadBalancerIP(ctx context.Context) (string, error) {
	lbHost, _, err := f.metaAPI.GetLoadBalancerEndpoint(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving load balancer endpoint: %w", err)
	}

	return lbHost, nil
}

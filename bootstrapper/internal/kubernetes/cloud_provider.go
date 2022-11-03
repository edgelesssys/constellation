/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
)

// ProviderMetadata implementers read/write cloud provider metadata.
type ProviderMetadata interface {
	// UID returns the unique identifier for the constellation.
	UID(ctx context.Context) (string, error)
	// List retrieves all instances belonging to the current Constellation.
	List(ctx context.Context) ([]metadata.InstanceMetadata, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
	// GetLoadBalancerEndpoint retrieves the load balancer endpoint.
	GetLoadBalancerEndpoint(ctx context.Context) (string, error)
	// GetInstance retrieves an instance using its providerID.
	GetInstance(ctx context.Context, providerID string) (metadata.InstanceMetadata, error)
}

type stubProviderMetadata struct {
	GetLoadBalancerEndpointErr  error
	GetLoadBalancerEndpointResp string

	ListErr  error
	ListResp []metadata.InstanceMetadata

	SelfErr  error
	SelfResp metadata.InstanceMetadata

	GetInstanceErr  error
	GetInstanceResp metadata.InstanceMetadata

	UIDErr  error
	UIDResp string
}

func (m *stubProviderMetadata) GetLoadBalancerEndpoint(ctx context.Context) (string, error) {
	return m.GetLoadBalancerEndpointResp, m.GetLoadBalancerEndpointErr
}

func (m *stubProviderMetadata) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	return m.ListResp, m.ListErr
}

func (m *stubProviderMetadata) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	return m.SelfResp, m.SelfErr
}

func (m *stubProviderMetadata) GetInstance(ctx context.Context, providerID string) (metadata.InstanceMetadata, error) {
	return m.GetInstanceResp, m.GetInstanceErr
}

func (m *stubProviderMetadata) UID(ctx context.Context) (string, error) {
	return m.UIDResp, m.UIDErr
}

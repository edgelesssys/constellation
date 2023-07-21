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
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
	// GetLoadBalancerEndpoint retrieves the load balancer endpoint.
	GetLoadBalancerEndpoint(ctx context.Context) (host, port string, err error)
}

type stubProviderMetadata struct {
	getLoadBalancerEndpointErr                       error
	getLoadBalancerHostResp, getLoadBalancerPortResp string

	selfErr  error
	selfResp metadata.InstanceMetadata

	uidErr  error
	uidResp string
}

func (m *stubProviderMetadata) GetLoadBalancerEndpoint(_ context.Context) (string, string, error) {
	return m.getLoadBalancerHostResp, m.getLoadBalancerPortResp, m.getLoadBalancerEndpointErr
}

func (m *stubProviderMetadata) Self(_ context.Context) (metadata.InstanceMetadata, error) {
	return m.selfResp, m.selfErr
}

func (m *stubProviderMetadata) UID(_ context.Context) (string, error) {
	return m.uidResp, m.uidErr
}

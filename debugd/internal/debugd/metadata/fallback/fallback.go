/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package fallback implements a fake metadata backend.
package fallback

import (
	"context"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
)

// NewFallbackFetcher returns a cloudprovider.Fetcher with a fake metadata backend.
func NewFallbackFetcher() *cloudprovider.Fetcher {
	return cloudprovider.New(&fallbackMetadata{})
}

type fallbackMetadata struct{}

// List retrieves all instances belonging to the current constellation.
func (fallbackMetadata) List(context.Context) ([]metadata.InstanceMetadata, error) {
	return nil, nil
}

// Self retrieves the current instance.
func (fallbackMetadata) Self(context.Context) (metadata.InstanceMetadata, error) {
	return metadata.InstanceMetadata{}, nil
}

// GetLoadBalancerEndpoint returns the endpoint of the load balancer.
func (fallbackMetadata) GetLoadBalancerEndpoint(context.Context) (string, string, error) {
	return "", "", nil
}

// UID returns the UID of the current instance.
func (fallbackMetadata) UID(context.Context) (string, error) {
	return "", nil
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fallback

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/role"
)

// Fetcher implements metadata.Fetcher interface but does not actually fetch cloud provider metadata.
type Fetcher struct{}

func (f Fetcher) Role(_ context.Context) (role.Role, error) {
	// Fallback fetcher does not try to fetch role
	return role.Unknown, nil
}

func (f Fetcher) DiscoverDebugdIPs(ctx context.Context) ([]string, error) {
	// Fallback fetcher does not try to discover debugd IPs
	return nil, nil
}

func (f Fetcher) DiscoverLoadbalancerIP(ctx context.Context) (string, error) {
	// Fallback fetcher does not try to discover loadbalancer IP
	return "", nil
}

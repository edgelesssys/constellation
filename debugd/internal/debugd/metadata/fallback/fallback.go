/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fallback

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

// Fetcher implements metadata.Fetcher interface but does not actually fetch cloud provider metadata.
type Fetcher struct{}

// Role for fallback fetcher does not try to fetch role.
func (f Fetcher) Role(_ context.Context) (role.Role, error) {
	return role.Unknown, nil
}

// DiscoverDebugdIPs for fallback fetcher does not try to discover debugd IPs.
func (f Fetcher) DiscoverDebugdIPs(ctx context.Context) ([]string, error) {
	return nil, nil
}

// DiscoverLoadbalancerIP for fallback fetcher does not try to discover loadbalancer IP.
func (f Fetcher) DiscoverLoadbalancerIP(ctx context.Context) (string, error) {
	return "", nil
}

// FetchSSHKeys for fallback fetcher does not try to fetch ssh keys.
func (f Fetcher) FetchSSHKeys(ctx context.Context) ([]ssh.UserKey, error) {
	return nil, nil
}

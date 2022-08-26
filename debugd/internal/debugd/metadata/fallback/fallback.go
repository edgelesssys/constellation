package fallback

import (
	"context"

	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
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

func (f Fetcher) FetchSSHKeys(ctx context.Context) ([]ssh.UserKey, error) {
	// Fallback fetcher does not try to fetch ssh keys
	return nil, nil
}

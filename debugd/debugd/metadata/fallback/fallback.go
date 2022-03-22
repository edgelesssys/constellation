package fallback

import (
	"context"

	"github.com/edgelesssys/constellation/debugd/ssh"
)

// Fetcher implements metadata.Fetcher interface but does not actually fetch cloud provider metadata.
type Fetcher struct{}

func (f Fetcher) DiscoverDebugdIPs(ctx context.Context) ([]string, error) {
	// Fallback fetcher does not try to discover debugd IPs
	return nil, nil
}

func (f Fetcher) FetchSSHKeys(ctx context.Context) ([]ssh.SSHKey, error) {
	// Fallback fetcher does not try to fetch ssh keys
	return nil, nil
}

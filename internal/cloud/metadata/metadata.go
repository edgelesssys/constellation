package metadata

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/constants"
)

// Instance describes metadata of a peer.
type InstanceMetadata struct {
	Name          string
	ProviderID    string
	Role          role.Role
	PrivateIPs    []string
	PublicIPs     []string
	AliasIPRanges []string
	// SSHKeys maps usernames to ssh public keys.
	SSHKeys map[string][]string
}

type metadataAPI interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]InstanceMetadata, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (InstanceMetadata, error)
	// SignalRole signals the constellation role via cloud provider metadata (if supported by the CSP and deployment type, otherwise does nothing).
	SignalRole(ctx context.Context, role role.Role) error
	// SetVPNIP stores the internally used VPN IP in cloud provider metadata (if supported and required for autoscaling by the CSP, otherwise does nothing).
	SetVPNIP(ctx context.Context, vpnIP string) error
	// Supported is used to determine if metadata API is implemented for this cloud provider.
	Supported() bool
}

// TODO(katexochen): Rename to InitEndpoints
func CoordinatorEndpoints(ctx context.Context, api metadataAPI) ([]string, error) {
	if !api.Supported() {
		return nil, errors.New("retrieving instances list from cloud provider is not yet supported")
	}
	instances, err := api.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances list from cloud provider: %w", err)
	}
	coordinatorEndpoints := []string{}
	for _, instance := range instances {
		// check if role of instance is "Coordinator"
		if instance.Role == role.Coordinator {
			for _, ip := range instance.PrivateIPs {
				coordinatorEndpoints = append(coordinatorEndpoints, net.JoinHostPort(ip, strconv.Itoa(constants.CoordinatorPort)))
			}
		}
	}

	return coordinatorEndpoints, nil
}

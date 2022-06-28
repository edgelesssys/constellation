package metadata

import (
	"context"
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
	// Supported is used to determine if metadata API is implemented for this cloud provider.
	Supported() bool
}

type InstanceSelfer interface {
	// Self retrieves the current instance.
	Self(ctx context.Context) (InstanceMetadata, error)
}

type InstanceLister interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]InstanceMetadata, error)
}

func InitServerEndpoints(ctx context.Context, lister InstanceLister) ([]string, error) {
	instances, err := lister.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances list from cloud provider: %w", err)
	}
	initServerEndpoints := []string{}
	for _, instance := range instances {
		// check if role of instance is "Coordinator"
		if instance.Role == role.Coordinator {
			for _, ip := range instance.PrivateIPs {
				initServerEndpoints = append(initServerEndpoints, net.JoinHostPort(ip, strconv.Itoa(constants.CoordinatorPort)))
			}
		}
	}

	return initServerEndpoints, nil
}

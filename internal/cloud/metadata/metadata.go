package metadata

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/constants"
)

// InstanceMetadata describes metadata of a peer.
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

type InstanceSelfer interface {
	// Self retrieves the current instance.
	Self(ctx context.Context) (InstanceMetadata, error)
}

type InstanceLister interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]InstanceMetadata, error)
}

// InitServerEndpoints returns the list of endpoints for the init server, which are running on the control plane nodes.
func InitServerEndpoints(ctx context.Context, lister InstanceLister) ([]string, error) {
	instances, err := lister.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances list from cloud provider: %w", err)
	}
	initServerEndpoints := []string{}
	for _, instance := range instances {
		if instance.Role == role.ControlPlane {
			for _, ip := range instance.PrivateIPs {
				initServerEndpoints = append(initServerEndpoints, net.JoinHostPort(ip, strconv.Itoa(constants.BootstrapperPort)))
			}
		}
	}

	return initServerEndpoints, nil
}

// JoinServiceEndpoints returns the list of endpoints for the join service, which are running on the control plane nodes.
func JoinServiceEndpoints(ctx context.Context, lister InstanceLister) ([]string, error) {
	instances, err := lister.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances list from cloud provider: %w", err)
	}
	joinEndpoints := []string{}
	for _, instance := range instances {
		if instance.Role == role.ControlPlane {
			for _, ip := range instance.PrivateIPs {
				joinEndpoints = append(joinEndpoints, net.JoinHostPort(ip, strconv.Itoa(constants.JoinServiceNodePort)))
			}
		}
	}

	return joinEndpoints, nil
}

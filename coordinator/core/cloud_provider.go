package core

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/role"
)

var ErrUnimplemented = errors.New("unimplemented")

const (
	ConstellationUIDMetadataKey = "constellation-uid"
	coordinatorPort             = "9000"
	RoleMetadataKey             = "constellation-role"
	VPNIPMetadataKey            = "constellation-vpn-ip"
)

// ProviderMetadata implementers read/write cloud provider metadata.
type ProviderMetadata interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]cloudtypes.Instance, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (cloudtypes.Instance, error)
	// SignalRole signals the constellation role via cloud provider metadata (if supported by the CSP and deployment type, otherwise does nothing).
	SignalRole(ctx context.Context, role role.Role) error
	// SetVPNIP stores the internally used VPN IP in cloud provider metadata (if supported and required for autoscaling by the CSP, otherwise does nothing).
	SetVPNIP(ctx context.Context, vpnIP string) error
	// Supported is used to determine if metadata API is implemented for this cloud provider.
	Supported() bool
}

type ProviderMetadataFake struct{}

func (f *ProviderMetadataFake) List(ctx context.Context) ([]cloudtypes.Instance, error) {
	self, err := f.Self(ctx)
	return []cloudtypes.Instance{self}, err
}

func (f *ProviderMetadataFake) Self(ctx context.Context) (cloudtypes.Instance, error) {
	return cloudtypes.Instance{
		Name:       "instanceName",
		ProviderID: "fake://instance-id",
		Role:       role.Unknown,
		PrivateIPs: []string{"192.0.2.1"},
	}, nil
}

func (f *ProviderMetadataFake) SignalRole(ctx context.Context, role role.Role) error {
	return nil
}

func (f *ProviderMetadataFake) SetVPNIP(ctx context.Context, vpnIP string) error {
	return nil
}

func (f *ProviderMetadataFake) Supported() bool {
	return true
}

// CoordinatorEndpoints retrieves a list of constellation coordinator endpoint candidates from the cloud provider API.
func CoordinatorEndpoints(ctx context.Context, metadata ProviderMetadata) ([]string, error) {
	if !metadata.Supported() {
		return nil, errors.New("retrieving instances list from cloud provider is not yet supported")
	}
	instances, err := metadata.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances list from cloud provider: %w", err)
	}
	coordinatorEndpoints := []string{}
	for _, instance := range instances {
		// check if role of instance is "Coordinator"
		if instance.Role == role.Coordinator {
			for _, ip := range instance.PrivateIPs {
				coordinatorEndpoints = append(coordinatorEndpoints, net.JoinHostPort(ip, coordinatorPort))
			}
		}
	}

	return coordinatorEndpoints, nil
}

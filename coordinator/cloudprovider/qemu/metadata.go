package qemu

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/role"
)

// Metadata implements core.ProviderMetadata interface for QEMU (currently not supported).
type Metadata struct{}

// Supported is used to determine if metadata API is implemented for this cloud provider.
func (m *Metadata) Supported() bool {
	return false
}

// List retrieves all instances belonging to the current constellation.
func (m *Metadata) List(ctx context.Context) ([]core.Instance, error) {
	panic("function *Metadata.List not implemented")
}

// Self retrieves the current instance.
func (m *Metadata) Self(ctx context.Context) (core.Instance, error) {
	panic("function *Metdata.Self not implemented")
}

// GetInstance retrieves an instance using its providerID.
func (m Metadata) GetInstance(ctx context.Context, providerID string) (core.Instance, error) {
	panic("function *Metadata.GetInstance not implemented")
}

// SignalRole signals the constellation role via cloud provider metadata (if supported by the CSP and deployment type, otherwise does nothing).
func (m Metadata) SignalRole(ctx context.Context, role role.Role) error {
	panic("function *Metadata.SignalRole not implemented")
}

// SetVPNIP stores the internally used VPN IP in cloud provider metadata (if supported and required for autoscaling by the CSP, otherwise does nothing).
func (m Metadata) SetVPNIP(ctx context.Context, vpnIP string) error {
	panic("function *Metadata.SetVPNIP not implemented")
}

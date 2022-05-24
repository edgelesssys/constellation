package qemu

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/role"
)

// Metadata implements core.ProviderMetadata interface for QEMU (currently not supported).
type Metadata struct{}

// Supported is used to determine if metadata API is implemented for this cloud provider.
func (m *Metadata) Supported() bool {
	return false
}

// List retrieves all instances belonging to the current constellation.
func (m *Metadata) List(ctx context.Context) ([]cloudtypes.Instance, error) {
	panic("function *Metadata.List not implemented")
}

// Self retrieves the current instance.
func (m *Metadata) Self(ctx context.Context) (cloudtypes.Instance, error) {
	panic("function *Metdata.Self not implemented")
}

// GetInstance retrieves an instance using its providerID.
func (m Metadata) GetInstance(ctx context.Context, providerID string) (cloudtypes.Instance, error) {
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

// SupportsLoadBalancer returns true if the cloud provider supports load balancers.
func (m Metadata) SupportsLoadBalancer() bool {
	return false
}

// GetLoadBalancerIP returns the IP of the load balancer.
func (m Metadata) GetLoadBalancerIP(ctx context.Context) (string, error) {
	panic("function *Metadata.GetLoadBalancerIP not implemented")
}

// GetSubnetworkCIDR retrieves the subnetwork CIDR from cloud provider metadata.
func (m Metadata) GetSubnetworkCIDR(ctx context.Context) (string, error) {
	panic("function *Metadata.GetSubnetworkCIDR not implemented")
}

package gcp

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/gcpshared"
)

// API handles all GCP API requests.
type API interface {
	// RetrieveInstances retrieves a list of all accessible GCP instances with their metadata.
	RetrieveInstances(ctx context.Context, project, zone string) ([]cloudtypes.Instance, error)
	// RetrieveInstances retrieves a single GCP instances with its metadata.
	RetrieveInstance(ctx context.Context, project, zone, instanceName string) (cloudtypes.Instance, error)
	// RetrieveInstanceMetadata retrieves the GCP instance metadata of the current instance.
	RetrieveInstanceMetadata(attr string) (string, error)
	// RetrieveProjectID retrieves the GCP  projectID containing the current instance.
	RetrieveProjectID() (string, error)
	// RetrieveZone retrieves the GCP zone containing the current instance.
	RetrieveZone() (string, error)
	// RetrieveInstanceName retrieves the instance name of the current instance.
	RetrieveInstanceName() (string, error)
	// RetrieveSubnetworkAliasCIDR retrieves the subnetwork CIDR of the current instance.
	RetrieveSubnetworkAliasCIDR(ctx context.Context, project, zone, instanceName string) (string, error)
	// RetrieveLoadBalancerIP retrieves the load balancer IP of the current instance.
	RetrieveLoadBalancerIP(ctx context.Context, project, zone string) (string, error)
	// SetInstanceMetadata sets metadata key: value of the instance specified by project, zone and instanceName.
	SetInstanceMetadata(ctx context.Context, project, zone, instanceName, key, value string) error
	// UnsetInstanceMetadata removes a metadata key-value pair of the instance specified by project, zone and instanceName.
	UnsetInstanceMetadata(ctx context.Context, project, zone, instanceName, key string) error
}

// Metadata implements core.ProviderMetadata interface.
type Metadata struct {
	api API
}

// New creates a new Provider with real API and FS.
func New(api API) *Metadata {
	return &Metadata{
		api: api,
	}
}

// List retrieves all instances belonging to the current constellation.
func (m *Metadata) List(ctx context.Context) ([]cloudtypes.Instance, error) {
	project, err := m.api.RetrieveProjectID()
	if err != nil {
		return nil, err
	}
	zone, err := m.api.RetrieveZone()
	if err != nil {
		return nil, err
	}
	instances, err := m.api.RetrieveInstances(ctx, project, zone)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances list from GCP api: %w", err)
	}
	return instances, nil
}

// Self retrieves the current instance.
func (m *Metadata) Self(ctx context.Context) (cloudtypes.Instance, error) {
	project, err := m.api.RetrieveProjectID()
	if err != nil {
		return cloudtypes.Instance{}, err
	}
	zone, err := m.api.RetrieveZone()
	if err != nil {
		return cloudtypes.Instance{}, err
	}
	instanceName, err := m.api.RetrieveInstanceName()
	if err != nil {
		return cloudtypes.Instance{}, err
	}
	return m.api.RetrieveInstance(ctx, project, zone, instanceName)
}

// GetInstance retrieves an instance using its providerID.
func (m *Metadata) GetInstance(ctx context.Context, providerID string) (cloudtypes.Instance, error) {
	project, zone, instanceName, err := gcpshared.SplitProviderID(providerID)
	if err != nil {
		return cloudtypes.Instance{}, fmt.Errorf("invalid providerID: %w", err)
	}
	return m.api.RetrieveInstance(ctx, project, zone, instanceName)
}

// SignalRole signals the constellation role via cloud provider metadata.
func (m *Metadata) SignalRole(ctx context.Context, role role.Role) error {
	project, err := m.api.RetrieveProjectID()
	if err != nil {
		return err
	}
	zone, err := m.api.RetrieveZone()
	if err != nil {
		return err
	}
	instanceName, err := m.api.RetrieveInstanceName()
	if err != nil {
		return err
	}
	return m.api.SetInstanceMetadata(ctx, project, zone, instanceName, core.RoleMetadataKey, role.String())
}

// SetVPNIP stores the internally used VPN IP in cloud provider metadata.
func (m *Metadata) SetVPNIP(ctx context.Context, vpnIP string) error {
	project, err := m.api.RetrieveProjectID()
	if err != nil {
		return err
	}
	zone, err := m.api.RetrieveZone()
	if err != nil {
		return err
	}
	instanceName, err := m.api.RetrieveInstanceName()
	if err != nil {
		return err
	}
	return m.api.SetInstanceMetadata(ctx, project, zone, instanceName, core.VPNIPMetadataKey, vpnIP)
}

// GetSubnetworkCIDR returns the subnetwork CIDR of the current instance.
func (m *Metadata) GetSubnetworkCIDR(ctx context.Context) (string, error) {
	project, err := m.api.RetrieveProjectID()
	if err != nil {
		return "", err
	}
	zone, err := m.api.RetrieveZone()
	if err != nil {
		return "", err
	}
	instanceName, err := m.api.RetrieveInstanceName()
	if err != nil {
		return "", err
	}
	return m.api.RetrieveSubnetworkAliasCIDR(ctx, project, zone, instanceName)
}

// SupportsLoadBalancer returns true if the cloud provider supports load balancers.
func (m *Metadata) SupportsLoadBalancer() bool {
	return true
}

// GetLoadBalancerIP returns the IP of the load balancer.
func (m *Metadata) GetLoadBalancerIP(ctx context.Context) (string, error) {
	project, err := m.api.RetrieveProjectID()
	if err != nil {
		return "", err
	}
	zone, err := m.api.RetrieveZone()
	if err != nil {
		return "", err
	}
	return m.api.RetrieveLoadBalancerIP(ctx, project, zone)
}

// Supported is used to determine if metadata API is implemented for this cloud provider.
func (m *Metadata) Supported() bool {
	return true
}

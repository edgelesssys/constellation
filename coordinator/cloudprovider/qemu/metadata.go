package qemu

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/role"
)

const qemuMetadataEndpoint = "10.42.0.1:8080"

// Metadata implements core.ProviderMetadata interface for QEMU.
type Metadata struct{}

// Supported is used to determine if metadata API is implemented for this cloud provider.
func (m *Metadata) Supported() bool {
	return true
}

// List retrieves all instances belonging to the current constellation.
func (m *Metadata) List(ctx context.Context) ([]cloudtypes.Instance, error) {
	instancesRaw, err := m.retrieveMetadata(ctx, "/peers")
	if err != nil {
		return nil, err
	}

	var instances []cloudtypes.Instance
	err = json.Unmarshal(instancesRaw, &instances)
	return instances, err
}

// Self retrieves the current instance.
func (m *Metadata) Self(ctx context.Context) (cloudtypes.Instance, error) {
	instanceRaw, err := m.retrieveMetadata(ctx, "/self")
	if err != nil {
		return cloudtypes.Instance{}, err
	}

	var instance cloudtypes.Instance
	err = json.Unmarshal(instanceRaw, &instance)
	return instance, err
}

// GetInstance retrieves an instance using its providerID.
func (m Metadata) GetInstance(ctx context.Context, providerID string) (cloudtypes.Instance, error) {
	instances, err := m.List(ctx)
	if err != nil {
		return cloudtypes.Instance{}, err
	}

	for _, instance := range instances {
		if instance.ProviderID == providerID {
			return instance, nil
		}
	}
	return cloudtypes.Instance{}, errors.New("instance not found")
}

// SignalRole signals the constellation role via cloud provider metadata (if supported by the CSP and deployment type, otherwise does nothing).
func (m Metadata) SignalRole(ctx context.Context, role role.Role) error {
	return nil
}

// SetVPNIP stores the internally used VPN IP in cloud provider metadata (if supported and required for autoscaling by the CSP, otherwise does nothing).
func (m Metadata) SetVPNIP(ctx context.Context, vpnIP string) error {
	return nil
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
	return "10.244.0.0/16", nil
}

func (m Metadata) retrieveMetadata(ctx context.Context, uri string) ([]byte, error) {
	url := &url.URL{
		Scheme: "http",
		Host:   qemuMetadataEndpoint,
		Path:   uri,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

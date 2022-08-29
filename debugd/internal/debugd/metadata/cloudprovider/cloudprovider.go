package cloudprovider

import (
	"context"
	"fmt"
	"net"

	"github.com/edgelesssys/constellation/bootstrapper/role"
	azurecloud "github.com/edgelesssys/constellation/internal/cloud/azure"
	gcpcloud "github.com/edgelesssys/constellation/internal/cloud/gcp"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	qemucloud "github.com/edgelesssys/constellation/internal/cloud/qemu"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
)

type providerMetadata interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]metadata.InstanceMetadata, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
	// GetLoadBalancerEndpoint returns the endpoint of the load balancer.
	GetLoadBalancerEndpoint(ctx context.Context) (string, error)
}

// Fetcher checks the metadata service to search for instances that were set up for debugging and cloud provider specific SSH keys.
type Fetcher struct {
	metaAPI providerMetadata
}

// NewGCP creates a new GCP fetcher.
func NewGCP(ctx context.Context) (*Fetcher, error) {
	gcpClient, err := gcpcloud.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	metaAPI := gcpcloud.New(gcpClient)

	return &Fetcher{
		metaAPI: metaAPI,
	}, nil
}

// NewAzure creates a new Azure fetcher.
func NewAzure(ctx context.Context) (*Fetcher, error) {
	metaAPI, err := azurecloud.NewMetadata(ctx)
	if err != nil {
		return nil, err
	}

	return &Fetcher{
		metaAPI: metaAPI,
	}, nil
}

func NewQEMU() *Fetcher {
	return &Fetcher{
		metaAPI: &qemucloud.Metadata{},
	}
}

func (f *Fetcher) Role(ctx context.Context) (role.Role, error) {
	self, err := f.metaAPI.Self(ctx)
	if err != nil {
		return role.Unknown, fmt.Errorf("retrieving role from cloud provider metadata: %w", err)
	}

	return self.Role, nil
}

// DiscoverDebugdIPs will query the metadata of all instances and return any ips of instances already set up for debugging.
func (f *Fetcher) DiscoverDebugdIPs(ctx context.Context) ([]string, error) {
	self, err := f.metaAPI.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving own instance: %w", err)
	}
	instances, err := f.metaAPI.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances: %w", err)
	}
	// filter own instance from instance list
	for i, instance := range instances {
		if instance.ProviderID == self.ProviderID {
			instances = append(instances[:i], instances[i+1:]...)
			break
		}
	}
	var ips []string
	for _, instance := range instances {
		if instance.VPCIP != "" {
			ips = append(ips, instance.VPCIP)
		}
	}
	return ips, nil
}

func (f *Fetcher) DiscoverLoadbalancerIP(ctx context.Context) (string, error) {
	lbEndpoint, err := f.metaAPI.GetLoadBalancerEndpoint(ctx)
	if err != nil {
		return "", fmt.Errorf("retrieving load balancer endpoint: %w", err)
	}

	// The port of the endpoint is not the port we need. We need to strip it off.
	//
	// TODO: Tag the specific load balancer we are looking for with a distinct tag.
	// Change the GetLoadBalancerEndpoint method to return the endpoint of a load
	// balancer with a given tag.
	lbIP, _, err := net.SplitHostPort(lbEndpoint)
	if err != nil {
		return "", fmt.Errorf("parsing load balancer endpoint: %w", err)
	}

	return lbIP, nil
}

// FetchSSHKeys will query the metadata of the current instance and deploys any SSH keys found.
func (f *Fetcher) FetchSSHKeys(ctx context.Context) ([]ssh.UserKey, error) {
	self, err := f.metaAPI.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving ssh keys from cloud provider metadata: %w", err)
	}

	keys := []ssh.UserKey{}
	for username, userKeys := range self.SSHKeys {
		for _, keyValue := range userKeys {
			keys = append(keys, ssh.UserKey{Username: username, PublicKey: keyValue})
		}
	}

	return keys, nil
}

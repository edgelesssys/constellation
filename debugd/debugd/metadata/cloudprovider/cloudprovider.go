package cloudprovider

import (
	"context"
	"fmt"

	azurecloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/gcp"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
)

type providerMetadata interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]core.Instance, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (core.Instance, error)
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

// DiscoverDebugdIPs will query the metadata of all instances and return any ips of instances already set up for debugging.
func (f *Fetcher) DiscoverDebugdIPs(ctx context.Context) ([]string, error) {
	self, err := f.metaAPI.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving own instance failed: %w", err)
	}
	instances, err := f.metaAPI.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances failed: %w", err)
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
		ips = append(ips, instance.IPs...)
	}
	return ips, nil
}

// FetchSSHKeys will query the metadata of the current instance and deploys any SSH keys found.
func (f *Fetcher) FetchSSHKeys(ctx context.Context) ([]ssh.UserKey, error) {
	self, err := f.metaAPI.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving ssh keys from cloud provider metadata failed: %w", err)
	}

	keys := []ssh.UserKey{}
	for username, userKeys := range self.SSHKeys {
		for _, keyValue := range userKeys {
			keys = append(keys, ssh.UserKey{Username: username, PublicKey: keyValue})
		}
	}

	return keys, nil
}

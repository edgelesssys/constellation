/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package qemu

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
)

const qemuMetadataEndpoint = "10.42.0.1:8080"

// Metadata implements core.ProviderMetadata interface for QEMU.
type Metadata struct{}

// List retrieves all instances belonging to the current constellation.
func (m *Metadata) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	instancesRaw, err := m.retrieveMetadata(ctx, "/peers")
	if err != nil {
		return nil, err
	}

	var instances []metadata.InstanceMetadata
	err = json.Unmarshal(instancesRaw, &instances)
	return instances, err
}

// Self retrieves the current instance.
func (m *Metadata) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	instanceRaw, err := m.retrieveMetadata(ctx, "/self")
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}

	var instance metadata.InstanceMetadata
	err = json.Unmarshal(instanceRaw, &instance)
	return instance, err
}

// GetInstance retrieves an instance using its providerID.
func (m *Metadata) GetInstance(ctx context.Context, providerID string) (metadata.InstanceMetadata, error) {
	instances, err := m.List(ctx)
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}

	for _, instance := range instances {
		if instance.ProviderID == providerID {
			return instance, nil
		}
	}
	return metadata.InstanceMetadata{}, errors.New("instance not found")
}

// GetLoadBalancerEndpoint returns the endpoint of the load balancer.
// For QEMU, the load balancer is the first control plane node returned by the metadata API.
func (m *Metadata) GetLoadBalancerEndpoint(ctx context.Context) (string, error) {
	endpointRaw, err := m.retrieveMetadata(ctx, "/endpoint")
	if err != nil {
		return "", err
	}
	var endpoint string
	err = json.Unmarshal(endpointRaw, &endpoint)
	return endpoint, err
}

// UID returns the UID of the constellation.
func (m *Metadata) UID(ctx context.Context) (string, error) {
	// We expect only one constellation to be deployed in the same QEMU / libvirt environment.
	// the UID can be an empty string.
	return "", nil
}

func (m *Metadata) retrieveMetadata(ctx context.Context, uri string) ([]byte, error) {
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

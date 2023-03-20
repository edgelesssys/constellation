/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
This package provides an interface to fake a CSP API for QEMU instances.
*/
package qemu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
)

const qemuMetadataEndpoint = "10.42.0.1:8080"

// Cloud provides an interface to fake a CSP API for QEMU instances.
type Cloud struct{}

// New returns a new Cloud instance.
func New() *Cloud {
	return &Cloud{}
}

// List retrieves all instances belonging to the current constellation.
func (c *Cloud) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	instancesRaw, err := c.retrieveMetadata(ctx, "/peers")
	if err != nil {
		return nil, err
	}

	var instances []metadata.InstanceMetadata
	err = json.Unmarshal(instancesRaw, &instances)
	return instances, err
}

// Self retrieves the current instance.
func (c *Cloud) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	instanceRaw, err := c.retrieveMetadata(ctx, "/self")
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}

	var instance metadata.InstanceMetadata
	err = json.Unmarshal(instanceRaw, &instance)
	return instance, err
}

// GetLoadBalancerEndpoint returns the endpoint of the load balancer.
// For QEMU, the load balancer is the first control plane node returned by the metadata API.
func (c *Cloud) GetLoadBalancerEndpoint(ctx context.Context) (string, error) {
	endpointRaw, err := c.retrieveMetadata(ctx, "/endpoint")
	if err != nil {
		return "", err
	}
	var endpoint string
	err = json.Unmarshal(endpointRaw, &endpoint)
	return endpoint, err
}

// InitSecretHash returns the hash of the init secret.
func (c *Cloud) InitSecretHash(ctx context.Context) ([]byte, error) {
	initSecretHash, err := c.retrieveMetadata(ctx, "/initsecrethash")
	if err != nil {
		return nil, fmt.Errorf("could not retrieve init secret hash: %w", err)
	}
	return initSecretHash, nil
}

// UID returns the UID of the constellation.
func (c *Cloud) UID(_ context.Context) (string, error) {
	// We expect only one constellation to be deployed in the same QEMU / libvirt environment.
	// the UID can be an empty string.
	return "", nil
}

func (c *Cloud) retrieveMetadata(ctx context.Context, uri string) ([]byte, error) {
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

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

// subset of azure imds API: https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service?tabs=linux
// this is not yet available through the azure sdk (see https://github.com/Azure/azure-rest-api-specs/issues/4408)

const (
	imdsURL        = "http://169.254.169.254/metadata/instance"
	imdsAPIVersion = "2021-02-01"
	maxCacheAge    = 12 * time.Hour
)

// IMDSClient is a client for the Azure Instance Metadata Service.
type IMDSClient struct {
	client *http.Client

	cache     metadataResponse
	cacheTime time.Time
}

// NewIMDSClient creates a new IMDSClient.
func NewIMDSClient() IMDSClient {
	// The default http client may use a system-wide proxy and it is recommended to disable the proxy explicitly:
	// https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service?tabs=linux#proxies
	// See also: https://github.com/microsoft/azureimds/blob/master/imdssample.go#L10
	return IMDSClient{
		client: &http.Client{Transport: &http.Transport{Proxy: nil}},
	}
}

// Tags returns the tags of the instance the function is called from.
func (c *IMDSClient) Tags(ctx context.Context) (map[string]string, error) {
	if c.timeForUpdate() || len(c.cache.Compute.Tags) == 0 {
		if err := c.update(ctx); err != nil {
			return nil, err
		}
	}

	tags := make(map[string]string, len(c.cache.Compute.Tags))
	for _, tag := range c.cache.Compute.Tags {
		tags[tag.Name] = tag.Value
	}

	return tags, nil
}

// providerID returns the provider ID of the instance the function is called from.
func (c *IMDSClient) providerID(ctx context.Context) (string, error) {
	if c.timeForUpdate() || c.cache.Compute.ResourceID == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Compute.ResourceID == "" {
		return "", errors.New("unable to get provider id")
	}

	return c.cache.Compute.ResourceID, nil
}

func (c *IMDSClient) name(ctx context.Context) (string, error) {
	if c.timeForUpdate() || c.cache.Compute.OSProfile.ComputerName == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Compute.OSProfile.ComputerName == "" {
		return "", errors.New("unable to get name")
	}

	return c.cache.Compute.OSProfile.ComputerName, nil
}

// subscriptionID returns the subscription ID of the instance the function
// is called from.
func (c *IMDSClient) subscriptionID(ctx context.Context) (string, error) {
	if c.timeForUpdate() || c.cache.Compute.SubscriptionID == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Compute.SubscriptionID == "" {
		return "", errors.New("unable to get subscription id")
	}

	return c.cache.Compute.SubscriptionID, nil
}

// resourceGroup returns the resource group of the instance the function
// is called from.
func (c *IMDSClient) resourceGroup(ctx context.Context) (string, error) {
	if c.timeForUpdate() || c.cache.Compute.ResourceGroup == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Compute.ResourceGroup == "" {
		return "", errors.New("unable to get resource group")
	}

	return c.cache.Compute.ResourceGroup, nil
}

// uid returns the UID of the cluster, based on the tags on the instance
// the function is called from, which are inherited from the scale set.
func (c *IMDSClient) uid(ctx context.Context) (string, error) {
	if c.timeForUpdate() || len(c.cache.Compute.Tags) == 0 {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	for _, tag := range c.cache.Compute.Tags {
		if tag.Name == cloud.TagUID {
			return tag.Value, nil
		}
	}

	return "", fmt.Errorf("unable to get uid from metadata tags %v", c.cache.Compute.Tags)
}

// initSecretHash returns the hash of the init secret of the cluster, based on the tags on the instance
// the function is called from, which are inherited from the scale set.
func (c *IMDSClient) initSecretHash(ctx context.Context) (string, error) {
	if c.timeForUpdate() || len(c.cache.Compute.Tags) == 0 {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	for _, tag := range c.cache.Compute.Tags {
		if tag.Name == cloud.TagInitSecretHash {
			return tag.Value, nil
		}
	}

	return "", fmt.Errorf("unable to get tag %s from metadata tags %v", cloud.TagInitSecretHash, c.cache.Compute.Tags)
}

// role returns the role of the instance the function is called from.
func (c *IMDSClient) role(ctx context.Context) (role.Role, error) {
	if c.timeForUpdate() || len(c.cache.Compute.Tags) == 0 {
		if err := c.update(ctx); err != nil {
			return role.Unknown, err
		}
	}

	for _, tag := range c.cache.Compute.Tags {
		if tag.Name == cloud.TagRole {
			return role.FromString(tag.Value), nil
		}
	}

	return role.Unknown, fmt.Errorf("unable to get role from metadata tags %v", c.cache.Compute.Tags)
}

// timeForUpdate checks whether an update is needed due to cache age.
func (c *IMDSClient) timeForUpdate() bool {
	return time.Since(c.cacheTime) > maxCacheAge
}

// update updates instance metadata from the azure imds API.
func (c *IMDSClient) update(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imdsURL, http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Add("Metadata", "True")
	query := req.URL.Query()
	query.Add("format", "json")
	query.Add("api-version", imdsAPIVersion)
	req.URL.RawQuery = query.Encode()
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var res metadataResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}

	c.cache = res
	c.cacheTime = time.Now()
	return nil
}

// metadataResponse contains metadataResponse with only the required values.
type metadataResponse struct {
	Compute metadataResponseCompute `json:"compute,omitempty"`
}

type metadataResponseCompute struct {
	ResourceID     string        `json:"resourceId,omitempty"`
	SubscriptionID string        `json:"subscriptionId,omitempty"`
	ResourceGroup  string        `json:"resourceGroupName,omitempty"`
	Tags           []metadataTag `json:"tagsList,omitempty"`
	OSProfile      struct {
		ComputerName string `json:"computerName,omitempty"`
	} `json:"osProfile,omitempty"`
}

type metadataTag struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

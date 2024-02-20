/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

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

// documentation of OpenStack Metadata Service: https://docs.openstack.org/nova/rocky/user/metadata-service.html

const (
	imdsMetaDataURL = "http://169.254.169.254/openstack/2018-08-27/meta_data.json"
	ec2ImdsBaseURL  = "http://169.254.169.254/1.0/meta-data"
	maxCacheAge     = 12 * time.Hour
)

type imdsClient struct {
	client httpClient

	vpcIPCache     string
	vpcIPCacheTime time.Time
	cache          metadataResponse
	cacheTime      time.Time
}

// providerID returns the provider ID of the instance the function is called from.
func (c *imdsClient) providerID(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.cache.UUID == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.UUID == "" {
		return "", errors.New("unable to get provider id")
	}

	return c.cache.UUID, nil
}

func (c *imdsClient) name(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.cache.Name == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Name == "" {
		return "", errors.New("unable to get name")
	}

	return c.cache.Name, nil
}

// projectID returns the project ID of the instance the function
// is called from.
func (c *imdsClient) projectID(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.cache.ProjectID == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.ProjectID == "" {
		return "", errors.New("unable to get project id")
	}

	return c.cache.ProjectID, nil
}

// uid returns the UID of the cluster, based on the tags on the instance
// the function is called from, which are inherited from the scale set.
func (c *imdsClient) uid(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || len(c.cache.Tags.UID) == 0 {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if len(c.cache.Tags.UID) == 0 {
		return "", fmt.Errorf("unable to get uid from metadata tags %v", c.cache.Tags)
	}

	return c.cache.Tags.UID, nil
}

// initSecretHash returns the hash of the init secret of the cluster, based on the tags on the instance
// the function is called from, which are inherited from the scale set.
func (c *imdsClient) initSecretHash(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || len(c.cache.Tags.InitSecretHash) == 0 {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if len(c.cache.Tags.InitSecretHash) == 0 {
		return "", fmt.Errorf("unable to get tag %s from metadata tags %v", cloud.TagInitSecretHash, c.cache.Tags)
	}

	return c.cache.Tags.InitSecretHash, nil
}

// role returns the role of the instance the function is called from.
func (c *imdsClient) role(ctx context.Context) (role.Role, error) {
	if c.timeForUpdate(c.cacheTime) || len(c.cache.Tags.Role) == 0 {
		if err := c.update(ctx); err != nil {
			return role.Unknown, err
		}
	}

	if len(c.cache.Tags.Role) == 0 {
		return role.Unknown, fmt.Errorf("unable to get role from metadata tags %v", c.cache.Tags)
	}

	return role.FromString(c.cache.Tags.Role), nil
}

func (c *imdsClient) loadBalancerEndpoint(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.cache.Tags.LoadBalancerEndpoint == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Tags.LoadBalancerEndpoint == "" {
		return "", errors.New("unable to get load balancer endpoint")
	}

	return c.cache.Tags.LoadBalancerEndpoint, nil
}

func (c *imdsClient) authURL(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.cache.Tags.AuthURL == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Tags.AuthURL == "" {
		return "", errors.New("unable to get auth url")
	}

	return c.cache.Tags.AuthURL, nil
}

func (c *imdsClient) userDomainName(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.cache.Tags.UserDomainName == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Tags.UserDomainName == "" {
		return "", errors.New("unable to get user domain name")
	}

	return c.cache.Tags.UserDomainName, nil
}

func (c *imdsClient) username(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.cache.Tags.Username == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Tags.Username == "" {
		return "", errors.New("unable to get token name")
	}

	return c.cache.Tags.Username, nil
}

func (c *imdsClient) password(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.cache.Tags.Password == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.cache.Tags.Password == "" {
		return "", errors.New("unable to get token password")
	}

	return c.cache.Tags.Password, nil
}

// timeForUpdate checks whether an update is needed due to cache age.
func (c *imdsClient) timeForUpdate(t time.Time) bool {
	return time.Since(t) > maxCacheAge
}

// update updates instance metadata from the azure imds API.
func (c *imdsClient) update(ctx context.Context) error {
	resp, err := httpGet(ctx, c.client, imdsMetaDataURL)
	if err != nil {
		return err
	}
	var metadataResp metadataResponse
	if err := json.Unmarshal(resp, &metadataResp); err != nil {
		return err
	}
	c.cache = metadataResp
	c.cacheTime = time.Now()
	return nil
}

func (c *imdsClient) vpcIP(ctx context.Context) (string, error) {
	const path = "local-ipv4"

	if c.timeForUpdate(c.vpcIPCacheTime) || c.vpcIPCache == "" {
		resp, err := httpGet(ctx, c.client, ec2ImdsBaseURL+"/"+path)
		if err != nil {
			return "", err
		}
		c.vpcIPCache = string(resp)
		c.vpcIPCacheTime = time.Now()
	}

	if c.vpcIPCache == "" {
		return "", errors.New("unable to get vpc ip")
	}

	return c.vpcIPCache, nil
}

func httpGet(ctx context.Context, c httpClient, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// metadataResponse contains metadataResponse with only the required values.
type metadataResponse struct {
	UUID      string       `json:"uuid,omitempty"`
	ProjectID string       `json:"project_id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Tags      metadataTags `json:"meta,omitempty"`
}

type metadataTags struct {
	InitSecretHash       string `json:"constellation-init-secret-hash,omitempty"`
	Role                 string `json:"constellation-role,omitempty"`
	UID                  string `json:"constellation-uid,omitempty"`
	AuthURL              string `json:"openstack-auth-url,omitempty"`
	UserDomainName       string `json:"openstack-user-domain-name,omitempty"`
	Username             string `json:"openstack-username,omitempty"`
	Password             string `json:"openstack-password,omitempty"`
	LoadBalancerEndpoint string `json:"openstack-load-balancer-endpoint,omitempty"`
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

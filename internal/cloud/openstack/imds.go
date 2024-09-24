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
	imdsUserDataURL = "http://169.254.169.254/openstack/2018-08-27/user_data"
	ec2ImdsBaseURL  = "http://169.254.169.254/1.0/meta-data"
	maxCacheAge     = 12 * time.Hour
)

type imdsClient struct {
	client httpClient

	vpcIPCache     string
	vpcIPCacheTime time.Time
	cache          metadataResponse
	userDataCache  userDataResponse
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
	if c.timeForUpdate(c.cacheTime) || c.userDataCache.LoadBalancerEndpoint == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.userDataCache.LoadBalancerEndpoint == "" {
		return "", errors.New("unable to get load balancer endpoint")
	}

	return c.userDataCache.LoadBalancerEndpoint, nil
}

func (c *imdsClient) authURL(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.userDataCache.AuthURL == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.userDataCache.AuthURL == "" {
		return "", errors.New("unable to get auth url")
	}

	return c.userDataCache.AuthURL, nil
}

func (c *imdsClient) userDomainName(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.userDataCache.UserDomainName == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.userDataCache.UserDomainName == "" {
		return "", errors.New("unable to get user domain name")
	}

	return c.userDataCache.UserDomainName, nil
}

func (c *imdsClient) regionName(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.userDataCache.RegionName == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.userDataCache.RegionName == "" {
		return "", errors.New("unable to get user domain name")
	}

	return c.userDataCache.RegionName, nil
}

func (c *imdsClient) username(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.userDataCache.Username == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.userDataCache.Username == "" {
		return "", errors.New("unable to get token name")
	}

	return c.userDataCache.Username, nil
}

func (c *imdsClient) password(ctx context.Context) (string, error) {
	if c.timeForUpdate(c.cacheTime) || c.userDataCache.Password == "" {
		if err := c.update(ctx); err != nil {
			return "", err
		}
	}

	if c.userDataCache.Password == "" {
		return "", errors.New("unable to get token password")
	}

	return c.userDataCache.Password, nil
}

// timeForUpdate checks whether an update is needed due to cache age.
func (c *imdsClient) timeForUpdate(t time.Time) bool {
	return time.Since(t) > maxCacheAge
}

func (c *imdsClient) update(ctx context.Context) error {
	if err := c.updateInstanceMetadata(ctx); err != nil {
		return fmt.Errorf("updating instance metadata: %w", err)
	}
	if err := c.updateUserData(ctx); err != nil {
		return fmt.Errorf("updating user data: %w", err)
	}
	c.cacheTime = time.Now()
	return nil
}

// update updates instance metadata from the azure imds API.
func (c *imdsClient) updateInstanceMetadata(ctx context.Context) error {
	resp, err := httpGet(ctx, c.client, imdsMetaDataURL)
	if err != nil {
		return err
	}
	var metadataResp metadataResponse
	if err := json.Unmarshal(resp, &metadataResp); err != nil {
		return fmt.Errorf("unmarshalling IMDS metadata response %q: %w", resp, err)
	}
	c.cache = metadataResp
	return nil
}

func (c *imdsClient) updateUserData(ctx context.Context) error {
	resp, err := httpGet(ctx, c.client, imdsUserDataURL)
	if err != nil {
		return err
	}
	var userdataResp userDataResponse
	if err := json.Unmarshal(resp, &userdataResp); err != nil {
		return fmt.Errorf("unmarshalling IMDS user_data response %q: %w", resp, err)
	}
	c.userDataCache = userdataResp
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
		return nil, fmt.Errorf("querying the OpenStack IMDS api failed for %q: %w", url, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("IMDS api might be broken for this server. Recreate the cluster if this issue persists. Querying the OpenStack IMDS api failed for %q with error code %d", url, resp.StatusCode)
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
	InitSecretHash string `json:"constellation-init-secret-hash,omitempty"`
	Role           string `json:"constellation-role,omitempty"`
	UID            string `json:"constellation-uid,omitempty"`
}

type userDataResponse struct {
	AuthURL              string `json:"openstack-auth-url,omitempty"`
	UserDomainName       string `json:"openstack-user-domain-name,omitempty"`
	RegionName           string `json:"openstack-region-name,omitempty"`
	Username             string `json:"openstack-username,omitempty"`
	Password             string `json:"openstack-password,omitempty"`
	LoadBalancerEndpoint string `json:"openstack-load-balancer-endpoint,omitempty"`
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

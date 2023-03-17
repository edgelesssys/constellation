/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"fmt"
	"net/url"
	"regexp"
)

// AccountKey is a OpenStack account key.
type AccountKey struct {
	AuthURL           string
	Username          string
	Password          string
	ProjectID         string
	ProjectName       string
	UserDomainName    string
	ProjectDomainName string
	RegionName        string
}

// AccountKeyFromURI parses ServiceAccountKey from URI.
func AccountKeyFromURI(serviceAccountURI string) (AccountKey, error) {
	uri, err := url.Parse(serviceAccountURI)
	if err != nil {
		return AccountKey{}, err
	}
	if uri.Scheme != "serviceaccount" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: invalid scheme: %s", uri.Scheme)
	}
	if uri.Host != "openstack" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: invalid host: %s", uri.Host)
	}
	query := uri.Query()
	if query.Get("auth_url") == "" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"auth_url\": %s", uri)
	}
	if query.Get("username") == "" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"username\": %s", uri)
	}
	if query.Get("password") == "" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"password\": %s", uri)
	}
	if query.Get("project_id") == "" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"project_id\": %s", uri)
	}
	if query.Get("project_name") == "" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"project_name\": %s", uri)
	}
	if query.Get("user_domain_name") == "" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"user_domain_name\": %s", uri)
	}
	if query.Get("project_domain_name") == "" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"project_domain_name\": %s", uri)
	}
	if query.Get("region_name") == "" {
		return AccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"region_name\": %s", uri)
	}
	return AccountKey{
		AuthURL:           query.Get("auth_url"),
		Username:          query.Get("username"),
		Password:          query.Get("password"),
		ProjectID:         query.Get("project_id"),
		ProjectName:       query.Get("project_name"),
		UserDomainName:    query.Get("user_domain_name"),
		ProjectDomainName: query.Get("project_domain_name"),
		RegionName:        query.Get("region_name"),
	}, nil
}

// ToCloudServiceAccountURI converts the AccountKey into a cloud service account URI.
func (k AccountKey) ToCloudServiceAccountURI() string {
	query := url.Values{}
	query.Add("auth_url", k.AuthURL)
	query.Add("username", k.Username)
	query.Add("password", k.Password)
	query.Add("project_id", k.ProjectID)
	query.Add("project_name", k.ProjectName)
	query.Add("user_domain_name", k.UserDomainName)
	query.Add("project_domain_name", k.ProjectDomainName)
	query.Add("region_name", k.RegionName)
	uri := url.URL{
		Scheme:   "serviceaccount",
		Host:     "openstack",
		RawQuery: query.Encode(),
	}
	return uri.String()
}

// CloudINI converts the AccountKey into a CloudINI.
func (k AccountKey) CloudINI() CloudINI {
	return CloudINI{
		AuthURL:          k.AuthURL,
		Username:         k.Username,
		Password:         k.Password,
		TenantID:         k.ProjectID,
		TenantName:       k.ProjectName,
		UserDomainName:   k.UserDomainName,
		TenantDomainName: k.ProjectDomainName,
		Region:           k.RegionName,
	}
}

// CloudINI is a struct that represents the cloud.ini file used by OpenStack k8s deployments.
type CloudINI struct {
	AuthURL          string `gcfg:"auth-url" mapstructure:"auth-url" name:"os-authURL" dependsOn:"os-password|os-trustID|os-applicationCredentialSecret|os-clientCertPath"`
	Username         string `name:"os-userName" value:"optional" dependsOn:"os-password"`
	Password         string `name:"os-password" value:"optional" dependsOn:"os-domainID|os-domainName,os-projectID|os-projectName,os-userID|os-userName"`
	TenantID         string `gcfg:"tenant-id" mapstructure:"project-id" name:"os-projectID" value:"optional" dependsOn:"os-password|os-clientCertPath"`
	TenantName       string `gcfg:"tenant-name" mapstructure:"project-name" name:"os-projectName" value:"optional" dependsOn:"os-password|os-clientCertPath"`
	UserDomainName   string `gcfg:"user-domain-name" mapstructure:"user-domain-name" name:"os-userDomainName" value:"optional"`
	TenantDomainName string `gcfg:"tenant-domain-name" mapstructure:"project-domain-name" name:"os-projectDomainName" value:"optional"`
	Region           string `name:"os-region"`
}

// String returns the string representation of the CloudINI.
func (i CloudINI) String() string {
	// sanitize parameters to not include newlines
	authURL := newlineRegexp.ReplaceAllString(i.AuthURL, "")
	username := newlineRegexp.ReplaceAllString(i.Username, "")
	password := newlineRegexp.ReplaceAllString(i.Password, "")
	tenantID := newlineRegexp.ReplaceAllString(i.TenantID, "")
	tenantName := newlineRegexp.ReplaceAllString(i.TenantName, "")
	userDomainName := newlineRegexp.ReplaceAllString(i.UserDomainName, "")
	tenantDomainName := newlineRegexp.ReplaceAllString(i.TenantDomainName, "")
	region := newlineRegexp.ReplaceAllString(i.Region, "")

	return fmt.Sprintf(`[Global]
auth-url = %s
username = %s
password = %s
tenant-id = %s
tenant-name = %s
user-domain-name = %s
tenant-domain-name = %s
region = %s
`, authURL, username, password, tenantID, tenantName, userDomainName, tenantDomainName, region)
}

var newlineRegexp = regexp.MustCompile(`[\r\n]+`)

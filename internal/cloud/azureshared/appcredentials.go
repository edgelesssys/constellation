/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azureshared

import (
	"fmt"
	"net/url"
	"strings"
)

// ApplicationCredentials is a set of Azure API credentials.
// It can contain a client secret and carries the preferred authentication method.
// It is the equivalent of a service account key in other cloud providers.
type ApplicationCredentials struct {
	TenantID            string
	AppClientID         string
	ClientSecretValue   string
	Location            string
	UamiResourceID      string
	PreferredAuthMethod AuthMethod
}

// ApplicationCredentialsFromURI converts a cloudServiceAccountURI into Azure ApplicationCredentials.
func ApplicationCredentialsFromURI(cloudServiceAccountURI string) (ApplicationCredentials, error) {
	uri, err := url.Parse(cloudServiceAccountURI)
	if err != nil {
		return ApplicationCredentials{}, err
	}
	if uri.Scheme != "serviceaccount" {
		return ApplicationCredentials{}, fmt.Errorf("invalid service account URI: invalid scheme: %s", uri.Scheme)
	}
	if uri.Host != "azure" {
		return ApplicationCredentials{}, fmt.Errorf("invalid service account URI: invalid host: %s", uri.Host)
	}
	query := uri.Query()
	preferredAuthMethod := FromString(query.Get("preferred_auth_method"))
	return ApplicationCredentials{
		TenantID:            query.Get("tenant_id"),
		AppClientID:         query.Get("client_id"),
		ClientSecretValue:   query.Get("client_secret"),
		Location:            query.Get("location"),
		UamiResourceID:      query.Get("uami_resource_id"),
		PreferredAuthMethod: preferredAuthMethod,
	}, nil
}

// ToCloudServiceAccountURI converts the ApplicationCredentials into a cloud service account URI.
func (c ApplicationCredentials) ToCloudServiceAccountURI() string {
	query := url.Values{}
	query.Add("tenant_id", c.TenantID)
	query.Add("location", c.Location)
	if c.AppClientID != "" {
		query.Add("client_id", c.AppClientID)
	}
	if c.ClientSecretValue != "" {
		query.Add("client_secret", c.ClientSecretValue)
	}
	if c.UamiResourceID != "" {
		query.Add("uami_resource_id", c.UamiResourceID)
	}
	if c.PreferredAuthMethod != AuthMethodUnknown {
		query.Add("preferred_auth_method", c.PreferredAuthMethod.String())
	}
	uri := url.URL{
		Scheme:   "serviceaccount",
		Host:     "azure",
		RawQuery: query.Encode(),
	}
	return uri.String()
}

//go:generate stringer -type=AuthMethod -trimprefix=AuthMethod

// AuthMethod is the authentication method used for the Azure API.
type AuthMethod uint32

// FromString converts a string into an AuthMethod.
func FromString(s string) AuthMethod {
	switch strings.ToLower(s) {
	case strings.ToLower(AuthMethodServicePrincipal.String()):
		return AuthMethodServicePrincipal
	case strings.ToLower(AuthMethodUserAssignedIdentity.String()):
		return AuthMethodUserAssignedIdentity
	default:
		return AuthMethodUnknown
	}
}

const (
	// AuthMethodUnknown is default value for AuthMethod.
	AuthMethodUnknown AuthMethod = iota
	// AuthMethodServicePrincipal uses a client ID and secret.
	AuthMethodServicePrincipal
	// AuthMethodUserAssignedIdentity uses a user assigned identity.
	AuthMethodUserAssignedIdentity
)

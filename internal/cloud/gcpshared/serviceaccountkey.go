/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcpshared

import (
	"fmt"
	"net/url"
)

// ServiceAccountKey is a GCP service account key.
type ServiceAccountKey struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// ServiceAccountKeyFromURI parses ServiceAccountKey from URI.
func ServiceAccountKeyFromURI(serviceAccountURI string) (ServiceAccountKey, error) {
	uri, err := url.Parse(serviceAccountURI)
	if err != nil {
		return ServiceAccountKey{}, err
	}
	if uri.Scheme != "serviceaccount" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: invalid scheme: %s", uri.Scheme)
	}
	if uri.Host != "gcp" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: invalid host: %s", uri.Host)
	}
	query := uri.Query()
	if query.Get("type") == "" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"type\": %s", uri)
	}
	if query.Get("project_id") == "" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"project_id\": %s", uri)
	}
	if query.Get("private_key_id") == "" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"private_key_id\": %s", uri)
	}
	if query.Get("private_key") == "" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"private_key\": %s", uri)
	}
	if query.Get("client_email") == "" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"client_email\": %s", uri)
	}
	if query.Get("client_id") == "" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"client_id\": %s", uri)
	}
	if query.Get("token_uri") == "" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"token_uri\": %s", uri)
	}
	if query.Get("auth_provider_x509_cert_url") == "" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"auth_provider_x509_cert_url\": %s", uri)
	}
	if query.Get("client_x509_cert_url") == "" {
		return ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"client_x509_cert_url\": %s", uri)
	}
	return ServiceAccountKey{
		Type:                    query.Get("type"),
		ProjectID:               query.Get("project_id"),
		PrivateKeyID:            query.Get("private_key_id"),
		PrivateKey:              query.Get("private_key"),
		ClientEmail:             query.Get("client_email"),
		ClientID:                query.Get("client_id"),
		AuthURI:                 query.Get("auth_uri"),
		TokenURI:                query.Get("token_uri"),
		AuthProviderX509CertURL: query.Get("auth_provider_x509_cert_url"),
		ClientX509CertURL:       query.Get("client_x509_cert_url"),
	}, nil
}

// ToCloudServiceAccountURI converts the ServiceAccountKey into a cloud service account URI.
func (k ServiceAccountKey) ToCloudServiceAccountURI() string {
	query := url.Values{}
	query.Add("type", k.Type)
	query.Add("project_id", k.ProjectID)
	query.Add("private_key_id", k.PrivateKeyID)
	query.Add("private_key", k.PrivateKey)
	query.Add("client_email", k.ClientEmail)
	query.Add("client_id", k.ClientID)
	query.Add("auth_uri", k.AuthURI)
	query.Add("token_uri", k.TokenURI)
	query.Add("auth_provider_x509_cert_url", k.AuthProviderX509CertURL)
	query.Add("client_x509_cert_url", k.ClientX509CertURL)
	uri := url.URL{
		Scheme:   "serviceaccount",
		Host:     "gcp",
		RawQuery: query.Encode(),
	}
	return uri.String()
}

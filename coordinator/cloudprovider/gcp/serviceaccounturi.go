package gcp

import (
	"fmt"
	"net/url"

	"github.com/edgelesssys/constellation/cli/gcp/client"
)

// getServiceAccountKey converts a cloudServiceAccountURI into a GCP ServiceAccountKey.
func getServiceAccountKey(cloudServiceAccountURI string) (client.ServiceAccountKey, error) {
	uri, err := url.Parse(cloudServiceAccountURI)
	if err != nil {
		return client.ServiceAccountKey{}, err
	}
	if uri.Scheme != "serviceaccount" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: invalid scheme: %s", uri.Scheme)
	}
	if uri.Host != "gcp" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: invalid host: %s", uri.Host)
	}
	query := uri.Query()
	if query.Get("type") == "" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"type\": %s", uri)
	}
	if query.Get("project_id") == "" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"project_id\": %s", uri)
	}
	if query.Get("private_key_id") == "" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"private_key_id\": %s", uri)
	}
	if query.Get("private_key") == "" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"private_key\": %s", uri)
	}
	if query.Get("client_email") == "" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"client_email\": %s", uri)
	}
	if query.Get("client_id") == "" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"client_id\": %s", uri)
	}
	if query.Get("token_uri") == "" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"token_uri\": %s", uri)
	}
	if query.Get("auth_provider_x509_cert_url") == "" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"auth_provider_x509_cert_url\": %s", uri)
	}
	if query.Get("client_x509_cert_url") == "" {
		return client.ServiceAccountKey{}, fmt.Errorf("invalid service account URI: missing parameter \"client_x509_cert_url\": %s", uri)
	}
	return client.ServiceAccountKey{
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

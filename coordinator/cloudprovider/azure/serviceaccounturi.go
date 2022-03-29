package azure

import (
	"fmt"
	"net/url"

	"github.com/edgelesssys/constellation/cli/azure/client"
)

// getApplicationCredentials converts a cloudServiceAccountURI into Azure ApplicationCredentials.
func getApplicationCredentials(cloudServiceAccountURI string) (client.ApplicationCredentials, error) {
	uri, err := url.Parse(cloudServiceAccountURI)
	if err != nil {
		return client.ApplicationCredentials{}, err
	}
	if uri.Scheme != "serviceaccount" {
		return client.ApplicationCredentials{}, fmt.Errorf("invalid service account URI: invalid scheme: %s", uri.Scheme)
	}
	if uri.Host != "azure" {
		return client.ApplicationCredentials{}, fmt.Errorf("invalid service account URI: invalid host: %s", uri.Host)
	}
	query := uri.Query()
	return client.ApplicationCredentials{
		TenantID:     query.Get("tenant_id"),
		ClientID:     query.Get("client_id"),
		ClientSecret: query.Get("client_secret"),
		Location:     query.Get("location"),
	}, nil
}

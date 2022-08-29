package azureshared

import (
	"fmt"
	"net/url"
)

// ApplicationCredentials is a set of Azure AD application credentials.
// It is the equivalent of a service account key in other cloud providers.
type ApplicationCredentials struct {
	TenantID          string
	AppClientID       string
	ClientSecretValue string
	Location          string
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
	return ApplicationCredentials{
		TenantID:          query.Get("tenant_id"),
		AppClientID:       query.Get("client_id"),
		ClientSecretValue: query.Get("client_secret"),
		Location:          query.Get("location"),
	}, nil
}

// ToCloudServiceAccountURI converts the ApplicationCredentials into a cloud service account URI.
func (c ApplicationCredentials) ToCloudServiceAccountURI() string {
	query := url.Values{}
	query.Add("tenant_id", c.TenantID)
	query.Add("client_id", c.AppClientID)
	query.Add("client_secret", c.ClientSecretValue)
	query.Add("location", c.Location)
	uri := url.URL{
		Scheme:   "serviceaccount",
		Host:     "azure",
		RawQuery: query.Encode(),
	}
	return uri.String()
}

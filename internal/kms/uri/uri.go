/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package uri provides URIs and parsing logic for KMS and storage URIs.

The URI for a KMS is of the form:

	kms://<provider>?<provider-specific-query-parameters>

The URI for a storage is of the form:

	storage://<provider>/<provider-specific-query-parameters>

A URI contains all information necessary to connect to the KMS or storage.
*/
package uri

import (
	"encoding/base64"
	"fmt"
	"net/url"
)

const (
	// DefaultCloud is the URL for the default Vault URL.
	DefaultCloud VaultBaseURL = "vault.azure.net"
	// ChinaCloud is the URL for Vaults in Azure China Cloud.
	ChinaCloud VaultBaseURL = "vault.azure.cn"
	// USGovCloud is the URL for Vaults in Azure US Government Cloud.
	USGovCloud VaultBaseURL = "vault.usgovcloudapi.net"
	// GermanCloud is the URL for Vaults in Azure German Cloud.
	GermanCloud VaultBaseURL = "vault.microsoftazure.de"
	// HSMDefaultCloud is the URL for HSM Vaults.
	HSMDefaultCloud VaultBaseURL = "managedhsm.azure.net"
)

// VaultBaseURL is the base URL of the vault.
// It defines what type of key vault is used.
type VaultBaseURL string

// Well known endpoints for KMS services.
const (
	awsKMSURI     = "kms://aws?keyPolicy=%s&kekID=%s"
	azureKMSURI   = "kms://azure?tenantID=%s&clientID=%s&clientSecret=%s&vaultName=%s&vaultType=%s&keyName=%s"
	gcpKMSURI     = "kms://gcp?project=%s&location=%s&keyRing=%s&keyName=%s&credentialsPath=%s"
	clusterKMSURI = "kms://cluster-kms?key=%s&salt=%s"
	awsS3URI      = "storage://aws?bucket=%s"
	azureBlobURI  = "storage://azure?container=%s&connectionString=%s"
	gcpStorageURI = "storage://gcp?projects=%s&bucket=%s"
	NoStoreURI    = "storage://no-store"
)

// MasterSecret holds the master key and salt for deriving keys.
type MasterSecret struct {
	Key  []byte `json:"key"`
	Salt []byte `json:"salt"`
}

// EncodeToURI returns an URI encoding the master secret.
func (m MasterSecret) EncodeToURI() string {
	return fmt.Sprintf(
		clusterKMSURI,
		base64.URLEncoding.EncodeToString(m.Key),
		base64.URLEncoding.EncodeToString(m.Salt),
	)
}

// DecodeMasterSecretFromURI decodes a master secret from an URI.
func DecodeMasterSecretFromURI(uri string) (MasterSecret, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return MasterSecret{}, err
	}

	if u.Scheme != "kms" {
		return MasterSecret{}, fmt.Errorf("invalid scheme: %q", u.Scheme)
	}
	if u.Host != "cluster-kms" {
		return MasterSecret{}, fmt.Errorf("invalid host: %q", u.Host)
	}

	q := u.Query()
	key, err := getBase64QueryParameter(q, "key")
	if err != nil {
		return MasterSecret{}, err
	}
	salt, err := getBase64QueryParameter(q, "salt")
	if err != nil {
		return MasterSecret{}, err
	}
	return MasterSecret{
		Key:  key,
		Salt: salt,
	}, nil
}

// AzureConfig is the configuration to authenticate with Azure Key Vault.
type AzureConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	KeyName      string
	VaultName    string
	VaultType    VaultBaseURL
}

// DecodeAzureConfigFromURI decodes an Azure configuration from an URI.
func DecodeAzureConfigFromURI(uri string) (AzureConfig, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return AzureConfig{}, err
	}

	if u.Scheme != "kms" {
		return AzureConfig{}, fmt.Errorf("invalid scheme: %q", u.Scheme)
	}
	if u.Host != "azure" {
		return AzureConfig{}, fmt.Errorf("invalid host: %q", u.Host)
	}

	q := u.Query()
	tenantID, err := getQueryParameter(q, "tenantID")
	if err != nil {
		return AzureConfig{}, err
	}
	clientID, err := getQueryParameter(q, "clientID")
	if err != nil {
		return AzureConfig{}, err
	}
	clientSecret, err := getQueryParameter(q, "clientSecret")
	if err != nil {
		return AzureConfig{}, err
	}
	vaultName, err := getQueryParameter(q, "vaultName")
	if err != nil {
		return AzureConfig{}, err
	}
	vaultType, err := getQueryParameter(q, "vaultType")
	if err != nil {
		return AzureConfig{}, err
	}
	keyName, err := getQueryParameter(q, "keyName")
	if err != nil {
		return AzureConfig{}, err
	}

	return AzureConfig{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		VaultName:    vaultName,
		VaultType:    VaultBaseURL(vaultType),
		KeyName:      keyName,
	}, nil
}

// EncodeToURI returns an URI encoding the Azure configuration.
func (a AzureConfig) EncodeToURI() string {
	return fmt.Sprintf(
		azureKMSURI,
		url.QueryEscape(a.TenantID),
		url.QueryEscape(a.ClientID),
		url.QueryEscape(a.ClientSecret),
		url.QueryEscape(a.VaultName),
		url.QueryEscape(string(a.VaultType)),
		url.QueryEscape(a.KeyName),
	)
}

// GCPConfig is the configuration to authenticate with GCP KMS.
type GCPConfig struct {
	// CredentialsPath is the path to a credentials file of a service account used to authorize against the KMS.
	CredentialsPath string
	ProjectID       string
	Location        string
	KeyRing         string
	KeyName         string
}

// DecodeGCPConfigFromURI decodes a GCP configuration from an URI.
func DecodeGCPConfigFromURI(uri string) (GCPConfig, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return GCPConfig{}, err
	}

	if u.Scheme != "kms" {
		return GCPConfig{}, fmt.Errorf("invalid scheme: %q", u.Scheme)
	}
	if u.Host != "gcp" {
		return GCPConfig{}, fmt.Errorf("invalid host: %q", u.Host)
	}

	q := u.Query()
	credentials, err := getQueryParameter(q, "credentials")
	if err != nil {
		return GCPConfig{}, err
	}
	projectID, err := getQueryParameter(q, "projectID")
	if err != nil {
		return GCPConfig{}, err
	}
	location, err := getQueryParameter(q, "location")
	if err != nil {
		return GCPConfig{}, err
	}
	keyRing, err := getQueryParameter(q, "keyRing")
	if err != nil {
		return GCPConfig{}, err
	}
	keyName, err := getQueryParameter(q, "keyName")
	if err != nil {
		return GCPConfig{}, err
	}

	return GCPConfig{
		CredentialsPath: credentials,
		ProjectID:       projectID,
		Location:        location,
		KeyRing:         keyRing,
		KeyName:         keyName,
	}, nil
}

// EncodeToURI returns an URI encoding the GCP configuration.
func (g GCPConfig) EncodeToURI() string {
	return fmt.Sprintf(
		gcpKMSURI,
		url.QueryEscape(g.ProjectID),
		url.QueryEscape(g.Location),
		url.QueryEscape(g.KeyRing),
		url.QueryEscape(g.KeyName),
		url.QueryEscape(g.CredentialsPath),
	)
}

// getBase64QueryParameter returns the url-base64-decoded value for the given key from the query parameters.
func getBase64QueryParameter(q url.Values, key string) ([]byte, error) {
	value, err := getQueryParameter(q, key)
	if err != nil {
		return nil, err
	}
	return base64.URLEncoding.DecodeString(value)
}

// getQueryParameter returns the unescaped value for the given key from the query parameters.
func getQueryParameter(q url.Values, key string) (string, error) {
	value := q.Get(key)
	if value == "" {
		return "", fmt.Errorf("missing query parameter %q", key)
	}

	value, err := url.QueryUnescape(value)
	if err != nil {
		return "", fmt.Errorf("failed to unescape value for key: %q", key)
	}
	return value, nil
}

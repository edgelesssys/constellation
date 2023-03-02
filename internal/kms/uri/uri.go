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
	awsKMSURI     = "kms://aws?region=%s&accessKeyID=%s&accessKey=%s&keyName=%s"
	azureKMSURI   = "kms://azure?tenantID=%s&clientID=%s&clientSecret=%s&vaultName=%s&vaultType=%s&keyName=%s"
	gcpKMSURI     = "kms://gcp?projectID=%s&location=%s&keyRing=%s&credentialsPath=%s&keyName=%s"
	clusterKMSURI = "kms://cluster-kms?key=%s&salt=%s"
	awsS3URI      = "storage://aws?bucket=%s&region=%s&accessKeyID=%s&accessKey=%s"
	azureBlobURI  = "storage://azure?account=%s&container=%s&tenantID=%s&clientID=%s&clientSecret=%s"
	gcpStorageURI = "storage://gcp?projectID=%s&bucket=%s&credentialsPath=%s"
	// NoStoreURI is a URI that indicates that no storage is used.
	// Should only be used with cluster KMS.
	NoStoreURI = "storage://no-store"
)

// MasterSecret holds the master key and salt for deriving keys.
type MasterSecret struct {
	// Key is the secret value used in HKDF to derive keys.
	Key []byte `json:"key"`
	// Salt is the salt used in HKDF to derive keys.
	Salt []byte `json:"salt"`
}

// EncodeToURI returns a URI encoding the master secret.
func (m MasterSecret) EncodeToURI() string {
	return fmt.Sprintf(
		clusterKMSURI,
		base64.URLEncoding.EncodeToString(m.Key),
		base64.URLEncoding.EncodeToString(m.Salt),
	)
}

// DecodeMasterSecretFromURI decodes a master secret from a URI.
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

// AWSConfig is the configuration to authenticate with AWS KMS.
type AWSConfig struct {
	// KeyName is the name of the key in AWS KMS.
	KeyName string
	// Region is the region of the key in AWS KMS.
	Region string
	// AccessKeyID is the ID of the access key used for authentication with the AWS API.
	AccessKeyID string
	// AccessKey is the secret value used for authentication with the AWS API.
	AccessKey string
}

// DecodeAWSConfigFromURI decodes an AWS configuration from a URI.
func DecodeAWSConfigFromURI(uri string) (AWSConfig, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return AWSConfig{}, err
	}

	if u.Scheme != "kms" {
		return AWSConfig{}, fmt.Errorf("invalid scheme: %q", u.Scheme)
	}
	if u.Host != "aws" {
		return AWSConfig{}, fmt.Errorf("invalid host: %q", u.Host)
	}

	q := u.Query()
	keyName, err := getQueryParameter(q, "keyName")
	if err != nil {
		return AWSConfig{}, err
	}
	region, err := getQueryParameter(q, "region")
	if err != nil {
		return AWSConfig{}, err
	}
	accessKeyID, err := getQueryParameter(q, "accessKeyID")
	if err != nil {
		return AWSConfig{}, err
	}
	accessKey, err := getQueryParameter(q, "accessKey")
	if err != nil {
		return AWSConfig{}, err
	}

	return AWSConfig{
		KeyName:     keyName,
		Region:      region,
		AccessKeyID: accessKeyID,
		AccessKey:   accessKey,
	}, nil
}

// EncodeToURI returns a URI encoding the AWS configuration.
func (c AWSConfig) EncodeToURI() string {
	return fmt.Sprintf(
		awsKMSURI,
		c.Region,
		c.AccessKeyID,
		c.AccessKey,
		c.KeyName,
	)
}

// AWSS3Config is the configuration to authenticate with AWS S3 storage bucket.
type AWSS3Config struct {
	// Bucket is the name of the S3 storage bucket to use.
	Bucket string
	// Region is the region storage bucket is located in.
	Region string
	// AccessKeyID is the ID of the access key used for authentication with the AWS API.
	AccessKeyID string
	// AccessKey is the secret value used for authentication with the AWS API.
	AccessKey string
}

// DecodeAWSS3ConfigFromURI decodes an S3 configuration from a URI.
func DecodeAWSS3ConfigFromURI(uri string) (AWSS3Config, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return AWSS3Config{}, err
	}

	if u.Scheme != "storage" {
		return AWSS3Config{}, fmt.Errorf("invalid scheme: %q", u.Scheme)
	}
	if u.Host != "aws" {
		return AWSS3Config{}, fmt.Errorf("invalid host: %q", u.Host)
	}

	q := u.Query()
	bucket, err := getQueryParameter(q, "bucket")
	if err != nil {
		return AWSS3Config{}, err
	}
	region, err := getQueryParameter(q, "region")
	if err != nil {
		return AWSS3Config{}, err
	}
	accessKeyID, err := getQueryParameter(q, "accessKeyID")
	if err != nil {
		return AWSS3Config{}, err
	}
	accessKey, err := getQueryParameter(q, "accessKey")
	if err != nil {
		return AWSS3Config{}, err
	}

	return AWSS3Config{
		Bucket:      bucket,
		Region:      region,
		AccessKeyID: accessKeyID,
		AccessKey:   accessKey,
	}, nil
}

// EncodeToURI returns a URI encoding the S3 configuration.
func (s AWSS3Config) EncodeToURI() string {
	return fmt.Sprintf(
		awsS3URI,
		url.QueryEscape(s.Bucket),
		url.QueryEscape(s.Region),
		url.QueryEscape(s.AccessKeyID),
		url.QueryEscape(s.AccessKey),
	)
}

// AzureConfig is the configuration to authenticate with Azure Key Vault.
type AzureConfig struct {
	// TenantID of the Azure Active Directory the Key Vault is located in.
	TenantID string
	// ClientID is the ID of the managed identity used to authenticate with the Azure API.
	ClientID string
	// ClientSecret is the secret-value/password of the managed identity used to authenticate with the Azure API.
	ClientSecret string
	// KeyName is the name of the key in Azure Key Vault.
	KeyName string
	// VaultName is the name of the vault.
	VaultName string
	// VaultType is the type of the vault.
	// This defines whether or not the Key Vault is a managed HSM.
	VaultType VaultBaseURL
}

// DecodeAzureConfigFromURI decodes an Azure configuration from a URI.
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

// EncodeToURI returns a URI encoding the Azure configuration.
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

// AzureBlobConfig is the configuration to authenticate with Azure Blob storage.
type AzureBlobConfig struct {
	// StorageAccount is the name of the storage account to use.
	StorageAccount string
	// Container is the name of the container to use.
	Container string
	// TenantID of the Azure Active Directory the Key Vault is located in.
	TenantID string
	// ClientID is the ID of the managed identity used to authenticate with the Azure API.
	ClientID string
	// ClientSecret is the secret-value/password of the managed identity used to authenticate with the Azure API.
	ClientSecret string
}

// DecodeAzureBlobConfigFromURI decodes an Azure Blob configuration from a URI.
func DecodeAzureBlobConfigFromURI(uri string) (AzureBlobConfig, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return AzureBlobConfig{}, err
	}

	if u.Scheme != "storage" {
		return AzureBlobConfig{}, fmt.Errorf("invalid scheme: %q", u.Scheme)
	}
	if u.Host != "azure" {
		return AzureBlobConfig{}, fmt.Errorf("invalid host: %q", u.Host)
	}

	q := u.Query()
	storageAccount, err := getQueryParameter(q, "account")
	if err != nil {
		return AzureBlobConfig{}, err
	}
	container, err := getQueryParameter(q, "container")
	if err != nil {
		return AzureBlobConfig{}, err
	}
	tenantID, err := getQueryParameter(q, "tenantID")
	if err != nil {
		return AzureBlobConfig{}, err
	}
	clientID, err := getQueryParameter(q, "clientID")
	if err != nil {
		return AzureBlobConfig{}, err
	}
	clientSecret, err := getQueryParameter(q, "clientSecret")
	if err != nil {
		return AzureBlobConfig{}, err
	}

	return AzureBlobConfig{
		StorageAccount: storageAccount,
		Container:      container,
		TenantID:       tenantID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
	}, nil
}

// EncodeToURI returns a URI encoding the Azure Blob configuration.
func (a AzureBlobConfig) EncodeToURI() string {
	return fmt.Sprintf(
		azureBlobURI,
		url.QueryEscape(a.StorageAccount),
		url.QueryEscape(a.Container),
		url.QueryEscape(a.TenantID),
		url.QueryEscape(a.ClientID),
		url.QueryEscape(a.ClientSecret),
	)
}

// GCPConfig is the configuration to authenticate with GCP KMS.
type GCPConfig struct {
	// CredentialsPath is the path to a credentials file of a service account used to authorize against the GCP API.
	CredentialsPath string
	// ProjectID is the name of the GCP project the KMS is located in.
	ProjectID string
	// Location is the location of the KMS.
	Location string
	// KeyRing is the name of the keyring.
	KeyRing string
	// KeyName is the name of the key in the GCP KMS.
	KeyName string
}

// DecodeGCPConfigFromURI decodes a GCP configuration from a URI.
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
	credentials, err := getQueryParameter(q, "credentialsPath")
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

// EncodeToURI returns a URI encoding the GCP configuration.
func (g GCPConfig) EncodeToURI() string {
	return fmt.Sprintf(
		gcpKMSURI,
		url.QueryEscape(g.ProjectID),
		url.QueryEscape(g.Location),
		url.QueryEscape(g.KeyRing),
		url.QueryEscape(g.CredentialsPath),
		url.QueryEscape(g.KeyName),
	)
}

// GoogleCloudStorageConfig is the configuration to authenticate with Google Cloud Storage.
type GoogleCloudStorageConfig struct {
	// CredentialsPath is the path to a credentials file of a service account used to authorize against the GCP API.
	CredentialsPath string
	// ProjectID is the name of the GCP project the storage bucket is located in.
	ProjectID string
	// Bucket is the name of the bucket to use.
	Bucket string
}

// DecodeGoogleCloudStorageConfigFromURI decodes a Google Cloud Storage configuration from a URI.
func DecodeGoogleCloudStorageConfigFromURI(uri string) (GoogleCloudStorageConfig, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return GoogleCloudStorageConfig{}, err
	}

	if u.Scheme != "storage" {
		return GoogleCloudStorageConfig{}, fmt.Errorf("invalid scheme: %q", u.Scheme)
	}
	if u.Host != "gcp" {
		return GoogleCloudStorageConfig{}, fmt.Errorf("invalid host: %q", u.Host)
	}

	q := u.Query()
	credentials, err := getQueryParameter(q, "credentialsPath")
	if err != nil {
		return GoogleCloudStorageConfig{}, err
	}
	projectID, err := getQueryParameter(q, "projectID")
	if err != nil {
		return GoogleCloudStorageConfig{}, err
	}
	bucket, err := getQueryParameter(q, "bucket")
	if err != nil {
		return GoogleCloudStorageConfig{}, err
	}

	return GoogleCloudStorageConfig{
		CredentialsPath: credentials,
		ProjectID:       projectID,
		Bucket:          bucket,
	}, nil
}

// EncodeToURI returns a URI encoding the Google Cloud Storage configuration.
func (g GoogleCloudStorageConfig) EncodeToURI() string {
	return fmt.Sprintf(
		gcpStorageURI,
		url.QueryEscape(g.ProjectID),
		url.QueryEscape(g.Bucket),
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

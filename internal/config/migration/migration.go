/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package migration contains outdated configuration formats and their migration functions.
package migration

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/semver"
)

const (
	// Version3 is the third version number for Constellation config file.
	Version3 = "v3"
)

// Config defines configuration used by CLI.
type Config struct {
	Version             string            `yaml:"version" validate:"eq=v3"`
	Image               string            `yaml:"image" validate:"required,image_compatibility"`
	Name                string            `yaml:"name" validate:"valid_name,required"`
	StateDiskSizeGB     int               `yaml:"stateDiskSizeGB" validate:"min=0"`
	KubernetesVersion   string            `yaml:"kubernetesVersion" validate:"required,supported_k8s_version"`
	MicroserviceVersion semver.Semver     `yaml:"microserviceVersion" validate:"required"`
	DebugCluster        *bool             `yaml:"debugCluster" validate:"required"`
	Provider            ProviderConfig    `yaml:"provider" validate:"dive"`
	Attestation         AttestationConfig `yaml:"attestation" validate:"dive"`
}

// ProviderConfig are cloud-provider specific configuration values used by the CLI.
// Fields should remain pointer-types so custom specific configs can nil them
// if not required.
type ProviderConfig struct {
	AWS       *AWSConfig       `yaml:"aws,omitempty" validate:"omitempty,dive"`
	Azure     *AzureConfig     `yaml:"azure,omitempty" validate:"omitempty,dive"`
	GCP       *GCPConfig       `yaml:"gcp,omitempty" validate:"omitempty,dive"`
	OpenStack *OpenStackConfig `yaml:"openstack,omitempty" validate:"omitempty,dive"`
	QEMU      *QEMUConfig      `yaml:"qemu,omitempty" validate:"omitempty,dive"`
}

// AWSConfig are AWS specific configuration values used by the CLI.
type AWSConfig struct {
	Region                 string `yaml:"region" validate:"required,aws_region"`
	Zone                   string `yaml:"zone" validate:"required,aws_zone"`
	InstanceType           string `yaml:"instanceType" validate:"lowercase,aws_instance_type"`
	StateDiskType          string `yaml:"stateDiskType" validate:"oneof=standard gp2 gp3 st1 sc1 io1"`
	IAMProfileControlPlane string `yaml:"iamProfileControlPlane" validate:"required"`
	IAMProfileWorkerNodes  string `yaml:"iamProfileWorkerNodes" validate:"required"`
	DeployCSIDriver        *bool  `yaml:"deployCSIDriver"`
}

// AzureConfig are Azure specific configuration values used by the CLI.
type AzureConfig struct {
	SubscriptionID       string `yaml:"subscription" validate:"uuid"`
	TenantID             string `yaml:"tenant" validate:"uuid"`
	Location             string `yaml:"location" validate:"required"`
	ResourceGroup        string `yaml:"resourceGroup" validate:"required"`
	UserAssignedIdentity string `yaml:"userAssignedIdentity" validate:"required"`
	InstanceType         string `yaml:"instanceType" validate:"azure_instance_type"`
	StateDiskType        string `yaml:"stateDiskType" validate:"oneof=Premium_LRS Premium_ZRS Standard_LRS StandardSSD_LRS StandardSSD_ZRS"`
	DeployCSIDriver      *bool  `yaml:"deployCSIDriver" validate:"required"`
	SecureBoot           *bool  `yaml:"secureBoot" validate:"required"`
}

// GCPConfig are GCP specific configuration values used by the CLI.
type GCPConfig struct {
	Project               string `yaml:"project" validate:"required"`
	Region                string `yaml:"region" validate:"required"`
	Zone                  string `yaml:"zone" validate:"required"`
	ServiceAccountKeyPath string `yaml:"serviceAccountKeyPath" validate:"required"`
	InstanceType          string `yaml:"instanceType" validate:"gcp_instance_type"`
	StateDiskType         string `yaml:"stateDiskType" validate:"oneof=pd-standard pd-balanced pd-ssd"`
	DeployCSIDriver       *bool  `yaml:"deployCSIDriver" validate:"required"`
}

// OpenStackConfig holds config information for OpenStack based Constellation deployments.
type OpenStackConfig struct {
	Cloud                   string `yaml:"cloud"`
	AvailabilityZone        string `yaml:"availabilityZone" validate:"required"`
	FlavorID                string `yaml:"flavorID" validate:"required"`
	FloatingIPPoolID        string `yaml:"floatingIPPoolID" validate:"required"`
	StateDiskType           string `yaml:"stateDiskType" validate:"required"`
	AuthURL                 string `yaml:"authURL" validate:"required"`
	ProjectID               string `yaml:"projectID" validate:"required"`
	ProjectName             string `yaml:"projectName" validate:"required"`
	UserDomainName          string `yaml:"userDomainName" validate:"required"`
	ProjectDomainName       string `yaml:"projectDomainName" validate:"required"`
	RegionName              string `yaml:"regionName" validate:"required"`
	Username                string `yaml:"username" validate:"required"`
	Password                string `yaml:"password"`
	DirectDownload          *bool  `yaml:"directDownload" validate:"required"`
	DeployYawolLoadBalancer *bool  `yaml:"deployYawolLoadBalancer" validate:"required"`
	YawolImageID            string `yaml:"yawolImageID"`
	YawolFlavorID           string `yaml:"yawolFlavorID"`
	DeployCSIDriver         *bool  `yaml:"deployCSIDriver" validate:"required"`
}

// QEMUConfig holds config information for QEMU based Constellation deployments.
type QEMUConfig struct {
	ImageFormat           string `yaml:"imageFormat" validate:"oneof=qcow2 raw"`
	VCPUs                 int    `yaml:"vcpus" validate:"required"`
	Memory                int    `yaml:"memory" validate:"required"`
	MetadataAPIImage      string `yaml:"metadataAPIServer" validate:"required"`
	LibvirtURI            string `yaml:"libvirtSocket"`
	LibvirtContainerImage string `yaml:"libvirtContainerImage"`
	NVRAM                 string `yaml:"nvram" validate:"required"`
	Firmware              string `yaml:"firmware"`
}

// AttestationConfig configuration values used for attestation.
// Fields should remain pointer-types so custom specific configs can nil them
// if not required.
type AttestationConfig struct {
	AWSSEVSNP          *AWSSEVSNP          `yaml:"awsSEVSNP,omitempty" validate:"omitempty,dive"`
	AWSNitroTPM        *AWSNitroTPM        `yaml:"awsNitroTPM,omitempty" validate:"omitempty,dive"`
	AzureSEVSNP        *AzureSEVSNP        `yaml:"azureSEVSNP,omitempty" validate:"omitempty,dive"`
	AzureTrustedLaunch *AzureTrustedLaunch `yaml:"azureTrustedLaunch,omitempty" validate:"omitempty,dive"`
	GCPSEVES           *GCPSEVES           `yaml:"gcpSEVES,omitempty" validate:"omitempty,dive"`
	QEMUTDX            *QEMUTDX            `yaml:"qemuTDX,omitempty" validate:"omitempty,dive"`
	QEMUVTPM           *QEMUVTPM           `yaml:"qemuVTPM,omitempty" validate:"omitempty,dive"`
}

// AWSSEVSNP is the configuration for AWS SEV-SNP attestation.
type AWSSEVSNP struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
	// TODO (derpsteb): reenable launchMeasurement once SNP is fixed on AWS.
	// description: |
	//   Expected launch measurement in SNP report.
	// LaunchMeasurement measurements.Measurement `json:"launchMeasurement" yaml:"launchMeasurement" validate:"required"`
}

// GetVariant returns aws-sev-snp as the variant.
func (AWSSEVSNP) GetVariant() variant.Variant {
	return variant.AWSSEVSNP{}
}

// GetMeasurements returns the measurements used for attestation.
func (c AWSSEVSNP) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *AWSSEVSNP) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c AWSSEVSNP) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*AWSSEVSNP)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	// TODO (derpsteb): reenable launchMeasurement once SNP is fixed on AWS.
	// if !bytes.Equal(c.LaunchMeasurement.Expected, otherCfg.LaunchMeasurement.Expected) {
	// 	return false, nil
	// }
	// if c.LaunchMeasurement.ValidationOpt != otherCfg.LaunchMeasurement.ValidationOpt {
	// 	return false, nil
	// }

	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

// AWSNitroTPM is the configuration for AWS Nitro TPM attestation.
type AWSNitroTPM struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// GetVariant returns aws-nitro-tpm as the variant.
func (AWSNitroTPM) GetVariant() variant.Variant {
	return variant.AWSNitroTPM{}
}

// GetMeasurements returns the measurements used for attestation.
func (c AWSNitroTPM) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *AWSNitroTPM) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c AWSNitroTPM) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*AWSNitroTPM)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

// SNPFirmwareSignerConfig is the configuration for validating the firmware signer.
type SNPFirmwareSignerConfig struct {
	// description: |
	//   List of accepted values for the firmware signing key digest.\nValues are enforced according to the 'enforcementPolicy'\n     - 'equal'       : Error if the reported signing key digest does not match any of the values in 'acceptedKeyDigests'\n     - 'maaFallback' : Use 'equal' checking for validation, but fallback to using Microsoft Azure Attestation (MAA) for validation if the reported digest does not match any of the values in 'acceptedKeyDigests'. See the Azure docs for more details: https://learn.microsoft.com/en-us/azure/attestation/overview#amd-sev-snp-attestation\n     - 'warnOnly'    : Same as 'equal', but only prints a warning instead of returning an error if no match is found
	AcceptedKeyDigests idkeydigest.List `json:"acceptedKeyDigests" yaml:"acceptedKeyDigests"`
	// description: |
	//   Key digest enforcement policy. One of {'equal', 'maaFallback', 'warnOnly'}
	EnforcementPolicy idkeydigest.Enforcement `json:"enforcementPolicy" yaml:"enforcementPolicy" validate:"required"`
	// description: |
	//   URL of the Microsoft Azure Attestation (MAA) instance to use for fallback validation. Only used if 'enforcementPolicy' is set to 'maaFallback'.
	MAAURL string `json:"maaURL,omitempty" yaml:"maaURL,omitempty" validate:"len=0"`
}

// EqualTo returns true if the config is equal to the given config.
func (c SNPFirmwareSignerConfig) EqualTo(other SNPFirmwareSignerConfig) bool {
	return c.AcceptedKeyDigests.EqualTo(other.AcceptedKeyDigests) && c.EnforcementPolicy == other.EnforcementPolicy && c.MAAURL == other.MAAURL
}

// GCPSEVES is the configuration for GCP SEV-ES attestation.
type GCPSEVES struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// GetVariant returns gcp-sev-es as the variant.
func (GCPSEVES) GetVariant() variant.Variant {
	return variant.GCPSEVES{}
}

// GetMeasurements returns the measurements used for attestation.
func (c GCPSEVES) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *GCPSEVES) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c GCPSEVES) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*GCPSEVES)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

// QEMUVTPM is the configuration for QEMU vTPM attestation.
type QEMUVTPM struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// GetVariant returns qemu-vtpm as the variant.
func (QEMUVTPM) GetVariant() variant.Variant {
	return variant.QEMUVTPM{}
}

// GetMeasurements returns the measurements used for attestation.
func (c QEMUVTPM) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *QEMUVTPM) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c QEMUVTPM) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*QEMUVTPM)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

// QEMUTDX is the configuration for QEMU TDX attestation.
type QEMUTDX struct {
	// description: |
	//   Expected TDX measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// GetVariant returns qemu-tdx as the variant.
func (QEMUTDX) GetVariant() variant.Variant {
	return variant.QEMUTDX{}
}

// GetMeasurements returns the measurements used for attestation.
func (c QEMUTDX) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *QEMUTDX) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c QEMUTDX) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*QEMUTDX)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

// AzureSEVSNP is the configuration for Azure SEV-SNP attestation.
type AzureSEVSNP struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
	// description: |
	//   Lowest acceptable bootloader version.
	BootloaderVersion AttestationVersion `json:"bootloaderVersion" yaml:"bootloaderVersion"`
	// description: |
	//   Lowest acceptable TEE version.
	TEEVersion AttestationVersion `json:"teeVersion" yaml:"teeVersion"`
	// description: |
	//   Lowest acceptable SEV-SNP version.
	SNPVersion AttestationVersion `json:"snpVersion" yaml:"snpVersion"`
	// description: |
	//   Lowest acceptable microcode version.
	MicrocodeVersion AttestationVersion `json:"microcodeVersion" yaml:"microcodeVersion"`
	// description: |
	//   Configuration for validating the firmware signature.
	FirmwareSignerConfig SNPFirmwareSignerConfig `json:"firmwareSignerConfig" yaml:"firmwareSignerConfig"`
	// description: |
	//   AMD Root Key certificate used to verify the SEV-SNP certificate chain.
	AMDRootKey Certificate `json:"amdRootKey" yaml:"amdRootKey"`
}

// AzureTrustedLaunch is the configuration for Azure Trusted Launch attestation.
type AzureTrustedLaunch struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// AttestationCfg is the common interface for passing attestation configs.
type AttestationCfg interface {
	// GetMeasurements returns the measurements that should be used for attestation.
	GetMeasurements() measurements.M
	// SetMeasurements updates a config's measurements using the given measurements.
	SetMeasurements(m measurements.M)
	// GetVariant returns the variant of the attestation config.
	GetVariant() variant.Variant
	// NewerThan returns true if the config is equal to the given config.
	EqualTo(AttestationCfg) (bool, error)
}

// UnmarshalAttestationConfig unmarshals the config file into the correct type.
func UnmarshalAttestationConfig(data []byte, attestVariant variant.Variant) (AttestationCfg, error) {
	switch attestVariant {
	case variant.AWSNitroTPM{}:
		return unmarshalTypedConfig[*AWSNitroTPM](data)
	case variant.AWSSEVSNP{}:
		return unmarshalTypedConfig[*AWSSEVSNP](data)
	case variant.AzureSEVSNP{}:
		return unmarshalTypedConfig[*AzureSEVSNP](data)
	case variant.AzureTrustedLaunch{}:
		return unmarshalTypedConfig[*AzureTrustedLaunch](data)
	case variant.GCPSEVES{}:
		return unmarshalTypedConfig[*GCPSEVES](data)
	case variant.QEMUVTPM{}:
		return unmarshalTypedConfig[*QEMUVTPM](data)
	case variant.QEMUTDX{}:
		return unmarshalTypedConfig[*QEMUTDX](data)
	case variant.Dummy{}:
		return unmarshalTypedConfig[*DummyCfg](data)
	default:
		return nil, fmt.Errorf("unknown variant: %s", attestVariant)
	}
}

func unmarshalTypedConfig[T AttestationCfg](data []byte) (AttestationCfg, error) {
	var cfg T
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Certificate is a wrapper around x509.Certificate allowing custom marshaling.
type Certificate x509.Certificate

// MarshalJSON marshals the certificate to PEM.
func (c Certificate) MarshalJSON() ([]byte, error) {
	pem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
	return json.Marshal(string(pem))
}

// MarshalYAML marshals the certificate to PEM.
func (c Certificate) MarshalYAML() (any, error) {
	pem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
	return string(pem), nil
}

// UnmarshalJSON unmarshals the certificate from PEM.
func (c *Certificate) UnmarshalJSON(data []byte) error {
	return c.unmarshal(func(val any) error {
		return json.Unmarshal(data, val)
	})
}

// UnmarshalYAML unmarshals the certificate from PEM.
func (c *Certificate) UnmarshalYAML(unmarshal func(any) error) error {
	return c.unmarshal(unmarshal)
}

func (c *Certificate) unmarshal(unmarshalFunc func(any) error) error {
	var pemData string
	if err := unmarshalFunc(&pemData); err != nil {
		return err
	}
	block, _ := pem.Decode([]byte(pemData))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	*c = Certificate(*cert)
	return nil
}

func mustParsePEM(data string) Certificate {
	jsonData := fmt.Sprintf("\"%s\"", data)
	var cert Certificate
	if err := json.Unmarshal([]byte(jsonData), &cert); err != nil {
		panic(err)
	}
	return cert
}

// DummyCfg is a placeholder for unknown attestation configs.
type DummyCfg struct {
	// description: |
	//   The measurements that should be used for attestation.
	Measurements measurements.M `json:"measurements,omitempty"`
}

// GetMeasurements returns the configs measurements.
func (c DummyCfg) GetMeasurements() measurements.M {
	return c.Measurements
}

// GetVariant returns a dummy variant.
func (DummyCfg) GetVariant() variant.Variant {
	return variant.Dummy{}
}

// SetMeasurements sets the configs measurements.
func (c *DummyCfg) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if measurements of the configs are equal.
func (c DummyCfg) EqualTo(other AttestationCfg) (bool, error) {
	return c.Measurements.EqualTo(other.GetMeasurements()), nil
}

const placeholderVersionValue = 0

// NewLatestPlaceholderVersion returns the latest version with a placeholder version value.
func NewLatestPlaceholderVersion() AttestationVersion {
	return AttestationVersion{
		Value:      placeholderVersionValue,
		WantLatest: true,
	}
}

// AttestationVersion is a type that represents a version of a SNP.
type AttestationVersion struct {
	Value      uint8
	WantLatest bool
}

// MarshalYAML implements a custom marshaller to resolve "latest" values.
func (v AttestationVersion) MarshalYAML() (any, error) {
	if v.WantLatest {
		return "latest", nil
	}
	return v.Value, nil
}

// UnmarshalYAML implements a custom unmarshaller to resolve "atest" values.
func (v *AttestationVersion) UnmarshalYAML(unmarshal func(any) error) error {
	var rawUnmarshal any
	if err := unmarshal(&rawUnmarshal); err != nil {
		return fmt.Errorf("raw unmarshal: %w", err)
	}

	return v.parseRawUnmarshal(rawUnmarshal)
}

// MarshalJSON implements a custom marshaller to resolve "latest" values.
func (v AttestationVersion) MarshalJSON() ([]byte, error) {
	if v.WantLatest {
		return json.Marshal("latest")
	}
	return json.Marshal(v.Value)
}

// UnmarshalJSON implements a custom unmarshaller to resolve "latest" values.
func (v *AttestationVersion) UnmarshalJSON(data []byte) (err error) {
	var rawUnmarshal any
	if err := json.Unmarshal(data, &rawUnmarshal); err != nil {
		return fmt.Errorf("raw unmarshal: %w", err)
	}
	return v.parseRawUnmarshal(rawUnmarshal)
}

func (v *AttestationVersion) parseRawUnmarshal(rawUnmarshal any) error {
	switch s := rawUnmarshal.(type) {
	case string:
		if strings.ToLower(s) == "latest" {
			v.WantLatest = true
			v.Value = placeholderVersionValue
		} else {
			return fmt.Errorf("invalid version value: %s", s)
		}
	case int:
		v.Value = uint8(s)
	// yaml spec allows "1" as float64, so version number might come as a float:  https://github.com/go-yaml/yaml/issues/430
	case float64:
		v.Value = uint8(s)
	default:
		return fmt.Errorf("invalid version value type: %s", s)
	}
	return nil
}

// DefaultForAzureSEVSNP returns the default configuration for Azure SEV-SNP attestation.
// Version numbers have placeholder values and the latest available values can be fetched using [AzureSEVSNP.FetchAndSetLatestVersionNumbers].
func DefaultForAzureSEVSNP() *AzureSEVSNP {
	return &AzureSEVSNP{
		Measurements:      measurements.DefaultsFor(cloudprovider.Azure, variant.AzureSEVSNP{}),
		BootloaderVersion: NewLatestPlaceholderVersion(),
		TEEVersion:        NewLatestPlaceholderVersion(),
		SNPVersion:        NewLatestPlaceholderVersion(),
		MicrocodeVersion:  NewLatestPlaceholderVersion(),
		FirmwareSignerConfig: SNPFirmwareSignerConfig{
			AcceptedKeyDigests: idkeydigest.DefaultList(),
			EnforcementPolicy:  idkeydigest.MAAFallback,
		},
		// AMD root key. Received from the AMD Key Distribution System API (KDS).
		AMDRootKey: mustParsePEM(`-----BEGIN CERTIFICATE-----\nMIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC\nBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS\nBgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg\nQ2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp\nY2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy\nMTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS\nBgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j\nZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG\n9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg\nW41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta\n1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2\nSzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0\n60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05\ngmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpqGd/y8KxX7jksTEzAOg\nbKAeam3lm+3yKIcTYMlsRMXPcjNbIvmsBykD//xSniusuHBkgnlENEWx1UcbQQrs\n+gVDkuVPhsnzIRNgYvM48Y+7LGiJYnrmE8xcrexekBxrva2V9TJQqnN3Q53kt5vi\nQi3+gCfmkwC0F0tirIZbLkXPrPwzZ0M9eNxhIySb2npJfgnqz55I0u33wh4r0ZNQ\neTGfw03MBUtyuzGesGkcw+loqMaq1qR4tjGbPYxCvpCq7+OgpCCoMNit2uLo9M18\nfHz10lOMT8nWAUvRZFzteXCm+7PHdYPlmQwUw3LvenJ/ILXoQPHfbkH0CyPfhl1j\nWhJFZasCAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMB0GA1UdDgQWBBSFrBrRQ/fI\nrFXUxR1BSKvVeErUUzAPBgNVHRMBAf8EBTADAQH/MDoGA1UdHwQzMDEwL6AtoCuG\nKWh0dHBzOi8va2RzaW50Zi5hbWQuY29tL3ZjZWsvdjEvTWlsYW4vY3JsMEYGCSqG\nSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZI\nAWUDBAICBQCiAwIBMKMDAgEBA4ICAQC6m0kDp6zv4Ojfgy+zleehsx6ol0ocgVel\nETobpx+EuCsqVFRPK1jZ1sp/lyd9+0fQ0r66n7kagRk4Ca39g66WGTJMeJdqYriw\nSTjjDCKVPSesWXYPVAyDhmP5n2v+BYipZWhpvqpaiO+EGK5IBP+578QeW/sSokrK\ndHaLAxG2LhZxj9aF73fqC7OAJZ5aPonw4RE299FVarh1Tx2eT3wSgkDgutCTB1Yq\nzT5DuwvAe+co2CIVIzMDamYuSFjPN0BCgojl7V+bTou7dMsqIu/TW/rPCX9/EUcp\nKGKqPQ3P+N9r1hjEFY1plBg93t53OOo49GNI+V1zvXPLI6xIFVsh+mto2RtgEX/e\npmMKTNN6psW88qg7c1hTWtN6MbRuQ0vm+O+/2tKBF2h8THb94OvvHHoFDpbCELlq\nHnIYhxy0YKXGyaW1NjfULxrrmxVW4wcn5E8GddmvNa6yYm8scJagEi13mhGu4Jqh\n3QU3sf8iUSUr09xQDwHtOQUVIqx4maBZPBtSMf+qUDtjXSSq8lfWcd8bLr9mdsUn\nJZJ0+tuPMKmBnSH860llKk+VpVQsgqbzDIvOLvD6W1Umq25boxCYJ+TuBoa4s+HH\nCViAvgT9kf/rBq1d+ivj6skkHxuzcxbk1xv6ZGxrteJxVH7KlX7YRdZ6eARKwLe4\nAFZEAwoKCQ==\n-----END CERTIFICATE-----\n`),
	}
}

// GetVariant returns azure-sev-snp as the variant.
func (AzureSEVSNP) GetVariant() variant.Variant {
	return variant.AzureSEVSNP{}
}

// GetMeasurements returns the measurements used for attestation.
func (c AzureSEVSNP) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *AzureSEVSNP) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c AzureSEVSNP) EqualTo(old AttestationCfg) (bool, error) {
	otherCfg, ok := old.(*AzureSEVSNP)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, old)
	}

	firmwareSignerCfgEqual := c.FirmwareSignerConfig.EqualTo(otherCfg.FirmwareSignerConfig)
	measurementsEqual := c.Measurements.EqualTo(otherCfg.Measurements)
	bootloaderEqual := c.BootloaderVersion == otherCfg.BootloaderVersion
	teeEqual := c.TEEVersion == otherCfg.TEEVersion
	snpEqual := c.SNPVersion == otherCfg.SNPVersion
	microcodeEqual := c.MicrocodeVersion == otherCfg.MicrocodeVersion
	rootKeyEqual := bytes.Equal(c.AMDRootKey.Raw, otherCfg.AMDRootKey.Raw)

	return firmwareSignerCfgEqual && measurementsEqual && bootloaderEqual && teeEqual && snpEqual && microcodeEqual && rootKeyEqual, nil
}

// FetchAndSetLatestVersionNumbers fetches the latest version numbers from the configapi and sets them.
func (c *AzureSEVSNP) FetchAndSetLatestVersionNumbers(ctx context.Context, fetcher attestationconfigapi.Fetcher, now time.Time) error {
	versions, err := fetcher.FetchAzureSEVSNPVersionLatest(ctx, now)
	if err != nil {
		return err
	}
	// set number and keep isLatest flag
	c.mergeWithLatestVersion(versions.AzureSEVSNPVersion)
	return nil
}

func (c *AzureSEVSNP) mergeWithLatestVersion(latest attestationconfigapi.AzureSEVSNPVersion) {
	if c.BootloaderVersion.WantLatest {
		c.BootloaderVersion.Value = latest.Bootloader
	}
	if c.TEEVersion.WantLatest {
		c.TEEVersion.Value = latest.TEE
	}
	if c.SNPVersion.WantLatest {
		c.SNPVersion.Value = latest.SNP
	}
	if c.MicrocodeVersion.WantLatest {
		c.MicrocodeVersion.Value = latest.Microcode
	}
}

// GetVariant returns azure-trusted-launch as the variant.
func (AzureTrustedLaunch) GetVariant() variant.Variant {
	return variant.AzureTrustedLaunch{}
}

// GetMeasurements returns the measurements used for attestation.
func (c AzureTrustedLaunch) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *AzureTrustedLaunch) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c AzureTrustedLaunch) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*AzureTrustedLaunch)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

// V3ToV4 converts an existing v3 config to a v4 config.
func V3ToV4(path string, fileHandler file.Handler) error {
	// Read old format
	var cfgV3 Config
	if err := fileHandler.ReadYAML(path, &cfgV3); err != nil {
		return fmt.Errorf("reading config file %s using v3 format: %w", path, err)
	}

	// Migrate to new format
	var cfgV4 config.Config
	cfgV4.Version = config.Version4
	cfgV4.Image = cfgV3.Image
	cfgV4.Name = cfgV3.Name
	cfgV4.KubernetesVersion = cfgV3.KubernetesVersion
	cfgV4.MicroserviceVersion = cfgV3.MicroserviceVersion
	cfgV4.DebugCluster = cfgV3.DebugCluster

	var zone, instanceType, stateDiskType string
	switch {
	case cfgV3.Provider.AWS != nil:
		cfgV4.Provider.AWS = &config.AWSConfig{
			Region:                 cfgV3.Provider.AWS.Region,
			Zone:                   cfgV3.Provider.AWS.Zone,
			IAMProfileControlPlane: cfgV3.Provider.AWS.IAMProfileControlPlane,
			IAMProfileWorkerNodes:  cfgV3.Provider.AWS.IAMProfileWorkerNodes,
			DeployCSIDriver:        cfgV3.Provider.AWS.DeployCSIDriver,
		}
		zone = cfgV3.Provider.AWS.Zone
		instanceType = cfgV3.Provider.AWS.InstanceType
		stateDiskType = cfgV3.Provider.AWS.StateDiskType
	case cfgV3.Provider.Azure != nil:
		cfgV4.Provider.Azure = &config.AzureConfig{
			SubscriptionID:       cfgV3.Provider.Azure.SubscriptionID,
			TenantID:             cfgV3.Provider.Azure.TenantID,
			Location:             cfgV3.Provider.Azure.Location,
			ResourceGroup:        cfgV3.Provider.Azure.ResourceGroup,
			UserAssignedIdentity: cfgV3.Provider.Azure.UserAssignedIdentity,
			DeployCSIDriver:      cfgV3.Provider.Azure.DeployCSIDriver,
			SecureBoot:           cfgV3.Provider.Azure.SecureBoot,
		}
		instanceType = cfgV3.Provider.Azure.InstanceType
		stateDiskType = cfgV3.Provider.Azure.StateDiskType
	case cfgV3.Provider.GCP != nil:
		cfgV4.Provider.GCP = &config.GCPConfig{
			Project:               cfgV3.Provider.GCP.Project,
			Region:                cfgV3.Provider.GCP.Region,
			Zone:                  cfgV3.Provider.GCP.Zone,
			ServiceAccountKeyPath: cfgV3.Provider.GCP.ServiceAccountKeyPath,
			DeployCSIDriver:       cfgV3.Provider.GCP.DeployCSIDriver,
		}
		zone = cfgV3.Provider.GCP.Zone
		instanceType = cfgV3.Provider.GCP.InstanceType
		stateDiskType = cfgV3.Provider.GCP.StateDiskType
	case cfgV3.Provider.OpenStack != nil:
		cfgV4.Provider.OpenStack = &config.OpenStackConfig{
			Cloud:                   cfgV3.Provider.OpenStack.Cloud,
			AvailabilityZone:        cfgV3.Provider.OpenStack.AvailabilityZone,
			FloatingIPPoolID:        cfgV3.Provider.OpenStack.FloatingIPPoolID,
			AuthURL:                 cfgV3.Provider.OpenStack.AuthURL,
			ProjectID:               cfgV3.Provider.OpenStack.ProjectID,
			ProjectName:             cfgV3.Provider.OpenStack.ProjectName,
			UserDomainName:          cfgV3.Provider.OpenStack.UserDomainName,
			ProjectDomainName:       cfgV3.Provider.OpenStack.ProjectDomainName,
			RegionName:              cfgV3.Provider.OpenStack.RegionName,
			Username:                cfgV3.Provider.OpenStack.Username,
			Password:                cfgV3.Provider.OpenStack.Password,
			DirectDownload:          cfgV3.Provider.OpenStack.DirectDownload,
			DeployYawolLoadBalancer: cfgV3.Provider.OpenStack.DeployYawolLoadBalancer,
			YawolImageID:            cfgV3.Provider.OpenStack.YawolImageID,
			YawolFlavorID:           cfgV3.Provider.OpenStack.YawolFlavorID,
			DeployCSIDriver:         cfgV3.Provider.OpenStack.DeployCSIDriver,
		}
		zone = cfgV3.Provider.OpenStack.AvailabilityZone
		instanceType = cfgV3.Provider.OpenStack.FlavorID
		stateDiskType = cfgV3.Provider.OpenStack.StateDiskType
	case cfgV3.Provider.QEMU != nil:
		cfgV4.Provider.QEMU = &config.QEMUConfig{
			ImageFormat:           cfgV3.Provider.QEMU.ImageFormat,
			VCPUs:                 cfgV3.Provider.QEMU.VCPUs,
			Memory:                cfgV3.Provider.QEMU.Memory,
			MetadataAPIImage:      cfgV3.Provider.QEMU.MetadataAPIImage,
			LibvirtURI:            cfgV3.Provider.QEMU.LibvirtURI,
			LibvirtContainerImage: cfgV3.Provider.QEMU.LibvirtContainerImage,
			NVRAM:                 cfgV3.Provider.QEMU.NVRAM,
			Firmware:              cfgV3.Provider.QEMU.Firmware,
		}
	}

	switch {
	case cfgV3.Attestation.AWSSEVSNP != nil:
		cfgV4.Attestation.AWSSEVSNP = &config.AWSSEVSNP{
			Measurements: cfgV3.Attestation.AWSSEVSNP.Measurements,
		}
	case cfgV3.Attestation.AWSNitroTPM != nil:
		cfgV4.Attestation.AWSNitroTPM = &config.AWSNitroTPM{
			Measurements: cfgV3.Attestation.AWSNitroTPM.Measurements,
		}
	case cfgV3.Attestation.AzureSEVSNP != nil:
		cfgV4.Attestation.AzureSEVSNP = &config.AzureSEVSNP{
			Measurements: cfgV3.Attestation.AzureSEVSNP.Measurements,
			BootloaderVersion: config.AttestationVersion{
				Value:      cfgV3.Attestation.AzureSEVSNP.BootloaderVersion.Value,
				WantLatest: cfgV3.Attestation.AzureSEVSNP.BootloaderVersion.WantLatest,
			},
			TEEVersion: config.AttestationVersion{
				Value:      cfgV3.Attestation.AzureSEVSNP.TEEVersion.Value,
				WantLatest: cfgV3.Attestation.AzureSEVSNP.TEEVersion.WantLatest,
			},
			SNPVersion: config.AttestationVersion{
				Value:      cfgV3.Attestation.AzureSEVSNP.SNPVersion.Value,
				WantLatest: cfgV3.Attestation.AzureSEVSNP.SNPVersion.WantLatest,
			},
			MicrocodeVersion: config.AttestationVersion{
				Value:      cfgV3.Attestation.AzureSEVSNP.MicrocodeVersion.Value,
				WantLatest: cfgV3.Attestation.AzureSEVSNP.MicrocodeVersion.WantLatest,
			},
			FirmwareSignerConfig: config.SNPFirmwareSignerConfig{
				AcceptedKeyDigests: cfgV3.Attestation.AzureSEVSNP.FirmwareSignerConfig.AcceptedKeyDigests,
				EnforcementPolicy:  cfgV3.Attestation.AzureSEVSNP.FirmwareSignerConfig.EnforcementPolicy,
				MAAURL:             cfgV3.Attestation.AzureSEVSNP.FirmwareSignerConfig.MAAURL,
			},
			AMDRootKey: config.Certificate(cfgV3.Attestation.AzureSEVSNP.AMDRootKey),
		}
	case cfgV3.Attestation.AzureTrustedLaunch != nil:
		cfgV4.Attestation.AzureTrustedLaunch = &config.AzureTrustedLaunch{
			Measurements: cfgV3.Attestation.AzureTrustedLaunch.Measurements,
		}
	case cfgV3.Attestation.GCPSEVES != nil:
		cfgV4.Attestation.GCPSEVES = &config.GCPSEVES{
			Measurements: cfgV3.Attestation.GCPSEVES.Measurements,
		}
	case cfgV3.Attestation.QEMUTDX != nil:
		cfgV4.Attestation.QEMUTDX = &config.QEMUTDX{
			Measurements: cfgV3.Attestation.QEMUTDX.Measurements,
		}
	case cfgV3.Attestation.QEMUVTPM != nil:
		cfgV4.Attestation.QEMUVTPM = &config.QEMUVTPM{
			Measurements: cfgV3.Attestation.QEMUVTPM.Measurements,
		}
	}

	cfgV4.NodeGroups = map[string]config.NodeGroup{
		"control_plane_default": {
			Role:            role.ControlPlane.TFString(),
			Zone:            zone,
			InstanceType:    instanceType,
			StateDiskSizeGB: cfgV3.StateDiskSizeGB,
			StateDiskType:   stateDiskType,
			InitialCount:    3, // this can be anything. When migrating, the initial count is not used.
		},
		"worker_default": {
			Role:            role.Worker.TFString(),
			Zone:            zone,
			InstanceType:    instanceType,
			StateDiskSizeGB: cfgV3.StateDiskSizeGB,
			StateDiskType:   stateDiskType,
			InitialCount:    1, // this can be anything. When migrating, the initial count is not used.
		},
	}

	// Create backup
	if err := os.Rename(path, path+".backup.v3"); err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}

	// Write migrated config
	if err := fileHandler.WriteYAML(path, cfgV4, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package migration contains outdated configuration formats and their migration functions.
package migration

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
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
	// TODO(derpsteb): reenable launchMeasurement once SNP is fixed on AWS.
	// description: |
	//   Expected launch measurement in SNP report.
	// LaunchMeasurement measurements.Measurement `json:"launchMeasurement" yaml:"launchMeasurement" validate:"required"`
}

// AWSNitroTPM is the configuration for AWS Nitro TPM attestation.
type AWSNitroTPM struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
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

// GCPSEVES is the configuration for GCP SEV-ES attestation.
type GCPSEVES struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// QEMUVTPM is the configuration for QEMU vTPM attestation.
type QEMUVTPM struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// QEMUTDX is the configuration for QEMU TDX attestation.
type QEMUTDX struct {
	// description: |
	//   Expected TDX measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
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

const placeholderVersionValue = 0

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
	cfgV4.KubernetesVersion = versions.ValidK8sVersion(cfgV3.KubernetesVersion)
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
			RegionName:              cfgV3.Provider.OpenStack.RegionName,
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
			BootloaderVersion: config.AttestationVersion[uint8]{
				Value:      cfgV3.Attestation.AzureSEVSNP.BootloaderVersion.Value,
				WantLatest: cfgV3.Attestation.AzureSEVSNP.BootloaderVersion.WantLatest,
			},
			TEEVersion: config.AttestationVersion[uint8]{
				Value:      cfgV3.Attestation.AzureSEVSNP.TEEVersion.Value,
				WantLatest: cfgV3.Attestation.AzureSEVSNP.TEEVersion.WantLatest,
			},
			SNPVersion: config.AttestationVersion[uint8]{
				Value:      cfgV3.Attestation.AzureSEVSNP.SNPVersion.Value,
				WantLatest: cfgV3.Attestation.AzureSEVSNP.SNPVersion.WantLatest,
			},
			MicrocodeVersion: config.AttestationVersion[uint8]{
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

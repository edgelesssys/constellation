/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package migration contains outdated configuration formats and their migration functions.
package migration

import (
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

const (
	// Version2 is the second version number for Constellation config file.
	Version2 = "v2"
)

// Config defines configuration used by CLI.
type Config struct {
	Version             string         `yaml:"version" validate:"eq=v2"`
	Image               string         `yaml:"image" validate:"required,version_compatibility"`
	Name                string         `yaml:"name" validate:"valid_name,required"`
	StateDiskSizeGB     int            `yaml:"stateDiskSizeGB" validate:"min=0"`
	KubernetesVersion   string         `yaml:"kubernetesVersion" validate:"required,supported_k8s_version"`
	MicroserviceVersion string         `yaml:"microserviceVersion" validate:"required,version_compatibility"`
	DebugCluster        *bool          `yaml:"debugCluster" validate:"required"`
	AttestationVariant  string         `yaml:"attestationVariant,omitempty" validate:"valid_attestation_variant"`
	Provider            ProviderConfig `yaml:"provider" validate:"dive"`
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
	Region                 string         `yaml:"region" validate:"required"`
	Zone                   string         `yaml:"zone" validate:"required"`
	InstanceType           string         `yaml:"instanceType" validate:"lowercase,aws_instance_type"`
	StateDiskType          string         `yaml:"stateDiskType" validate:"oneof=standard gp2 gp3 st1 sc1 io1"`
	IAMProfileControlPlane string         `yaml:"iamProfileControlPlane" validate:"required"`
	IAMProfileWorkerNodes  string         `yaml:"iamProfileWorkerNodes" validate:"required"`
	Measurements           measurements.M `yaml:"measurements" validate:"required,no_placeholders"`
}

// AzureConfig are Azure specific configuration values used by the CLI.
type AzureConfig struct {
	SubscriptionID       string                  `yaml:"subscription" validate:"uuid"`
	TenantID             string                  `yaml:"tenant" validate:"uuid"`
	Location             string                  `yaml:"location" validate:"required"`
	ResourceGroup        string                  `yaml:"resourceGroup" validate:"required"`
	UserAssignedIdentity string                  `yaml:"userAssignedIdentity" validate:"required"`
	AppClientID          string                  `yaml:"appClientID" validate:"uuid"`
	ClientSecretValue    string                  `yaml:"clientSecretValue" validate:"required"`
	InstanceType         string                  `yaml:"instanceType" validate:"azure_instance_type"`
	StateDiskType        string                  `yaml:"stateDiskType" validate:"oneof=Premium_LRS Premium_ZRS Standard_LRS StandardSSD_LRS StandardSSD_ZRS"`
	DeployCSIDriver      *bool                   `yaml:"deployCSIDriver" validate:"required"`
	ConfidentialVM       *bool                   `yaml:"confidentialVM,omitempty" validate:"omitempty,deprecated"`
	SecureBoot           *bool                   `yaml:"secureBoot" validate:"required"`
	IDKeyDigest          idkeydigest.List        `yaml:"idKeyDigest" validate:"required_if=EnforceIdKeyDigest true,omitempty"`
	EnforceIDKeyDigest   idkeydigest.Enforcement `yaml:"enforceIdKeyDigest" validate:"required"`
	Measurements         measurements.M          `yaml:"measurements" validate:"required,no_placeholders"`
}

// GCPConfig are GCP specific configuration values used by the CLI.
type GCPConfig struct {
	Project               string         `yaml:"project" validate:"required"`
	Region                string         `yaml:"region" validate:"required"`
	Zone                  string         `yaml:"zone" validate:"required"`
	ServiceAccountKeyPath string         `yaml:"serviceAccountKeyPath" validate:"required"`
	InstanceType          string         `yaml:"instanceType" validate:"gcp_instance_type"`
	StateDiskType         string         `yaml:"stateDiskType" validate:"oneof=pd-standard pd-balanced pd-ssd"`
	DeployCSIDriver       *bool          `yaml:"deployCSIDriver" validate:"required"`
	Measurements          measurements.M `yaml:"measurements" validate:"required,no_placeholders"`
}

// OpenStackConfig holds config information for OpenStack based Constellation deployments.
type OpenStackConfig struct {
	Cloud             string         `yaml:"cloud"`
	AvailabilityZone  string         `yaml:"availabilityZone" validate:"required"`
	FlavorID          string         `yaml:"flavorID" validate:"required"`
	FloatingIPPoolID  string         `yaml:"floatingIPPoolID" validate:"required"`
	AuthURL           string         `yaml:"authURL" validate:"required"`
	ProjectID         string         `yaml:"projectID" validate:"required"`
	ProjectName       string         `yaml:"projectName" validate:"required"`
	UserDomainName    string         `yaml:"userDomainName" validate:"required"`
	ProjectDomainName string         `yaml:"projectDomainName" validate:"required"`
	RegionName        string         `yaml:"regionName" validate:"required"`
	Username          string         `yaml:"username" validate:"required"`
	Password          string         `yaml:"password"`
	DirectDownload    *bool          `yaml:"directDownload" validate:"required"`
	Measurements      measurements.M `yaml:"measurements" validate:"required,no_placeholders"`
}

// QEMUConfig holds config information for QEMU based Constellation deployments.
type QEMUConfig struct {
	ImageFormat           string         `yaml:"imageFormat" validate:"oneof=qcow2 raw"`
	VCPUs                 int            `yaml:"vcpus" validate:"required"`
	Memory                int            `yaml:"memory" validate:"required"`
	MetadataAPIImage      string         `yaml:"metadataAPIServer" validate:"required"`
	LibvirtURI            string         `yaml:"libvirtSocket"`
	LibvirtContainerImage string         `yaml:"libvirtContainerImage"`
	NVRAM                 string         `yaml:"nvram" validate:"required"`
	Firmware              string         `yaml:"firmware"`
	Measurements          measurements.M `yaml:"measurements" validate:"required,no_placeholders"`
}

// V2ToV3 converts an existing v2 config to a v3 config.
func V2ToV3(path string, fileHandler file.Handler) error {
	// Read old format
	var cfgV2 Config
	if err := fileHandler.ReadYAML(path, &cfgV2); err != nil {
		return fmt.Errorf("reading config file %s using v2 format: %w", path, err)
	}

	// Migrate to new format
	var cfgV3 config.Config
	cfgV3.Version = config.Version3
	cfgV3.Image = cfgV2.Image
	cfgV3.Name = cfgV2.Name
	cfgV3.StateDiskSizeGB = cfgV2.StateDiskSizeGB
	cfgV3.KubernetesVersion = cfgV2.KubernetesVersion
	cfgV3.MicroserviceVersion = cfgV2.MicroserviceVersion
	cfgV3.DebugCluster = cfgV2.DebugCluster

	switch {
	case cfgV2.Provider.AWS != nil:
		cfgV3.Provider.AWS = &config.AWSConfig{
			Region:                 cfgV2.Provider.AWS.Region,
			Zone:                   cfgV2.Provider.AWS.Zone,
			InstanceType:           cfgV2.Provider.AWS.InstanceType,
			StateDiskType:          cfgV2.Provider.AWS.StateDiskType,
			IAMProfileControlPlane: cfgV2.Provider.AWS.IAMProfileControlPlane,
			IAMProfileWorkerNodes:  cfgV2.Provider.AWS.IAMProfileWorkerNodes,
		}
		cfgV3.Attestation.AWSNitroTPM = &config.AWSNitroTPM{
			Measurements: cfgV2.Provider.AWS.Measurements,
		}
	case cfgV2.Provider.Azure != nil:
		cfgV3.Provider.Azure = &config.AzureConfig{
			SubscriptionID:       cfgV2.Provider.Azure.SubscriptionID,
			TenantID:             cfgV2.Provider.Azure.TenantID,
			Location:             cfgV2.Provider.Azure.Location,
			ResourceGroup:        cfgV2.Provider.Azure.ResourceGroup,
			UserAssignedIdentity: cfgV2.Provider.Azure.UserAssignedIdentity,
			AppClientID:          cfgV2.Provider.Azure.AppClientID,
			ClientSecretValue:    cfgV2.Provider.Azure.ClientSecretValue,
			InstanceType:         cfgV2.Provider.Azure.InstanceType,
			StateDiskType:        cfgV2.Provider.Azure.StateDiskType,
			DeployCSIDriver:      cfgV2.Provider.Azure.DeployCSIDriver,
			SecureBoot:           cfgV2.Provider.Azure.SecureBoot,
		}
		if cfgV2.Provider.Azure.ConfidentialVM != nil && *cfgV2.Provider.Azure.ConfidentialVM {
			cfgV3.Attestation.AzureSEVSNP = config.DefaultForAzureSEVSNP()
			cfgV3.Attestation.AzureSEVSNP.Measurements = cfgV2.Provider.Azure.Measurements
		} else {
			cfgV3.Attestation.AzureTrustedLaunch = &config.AzureTrustedLaunch{
				Measurements: cfgV2.Provider.Azure.Measurements,
			}
		}
	case cfgV2.Provider.GCP != nil:
		cfgV3.Provider.GCP = &config.GCPConfig{
			Project:               cfgV2.Provider.GCP.Project,
			Region:                cfgV2.Provider.GCP.Region,
			Zone:                  cfgV2.Provider.GCP.Zone,
			ServiceAccountKeyPath: cfgV2.Provider.GCP.ServiceAccountKeyPath,
			InstanceType:          cfgV2.Provider.GCP.InstanceType,
			StateDiskType:         cfgV2.Provider.GCP.StateDiskType,
			DeployCSIDriver:       cfgV2.Provider.GCP.DeployCSIDriver,
		}
		cfgV3.Attestation.GCPSEVES = &config.GCPSEVES{
			Measurements: cfgV2.Provider.GCP.Measurements,
		}
	case cfgV2.Provider.OpenStack != nil:
		cfgV3.Provider.OpenStack = &config.OpenStackConfig{
			Cloud:             cfgV2.Provider.OpenStack.Cloud,
			AvailabilityZone:  cfgV2.Provider.OpenStack.AvailabilityZone,
			FlavorID:          cfgV2.Provider.OpenStack.FlavorID,
			FloatingIPPoolID:  cfgV2.Provider.OpenStack.FloatingIPPoolID,
			AuthURL:           cfgV2.Provider.OpenStack.AuthURL,
			ProjectID:         cfgV2.Provider.OpenStack.ProjectID,
			ProjectName:       cfgV2.Provider.OpenStack.ProjectName,
			UserDomainName:    cfgV2.Provider.OpenStack.UserDomainName,
			ProjectDomainName: cfgV2.Provider.OpenStack.ProjectDomainName,
			RegionName:        cfgV2.Provider.OpenStack.RegionName,
			Username:          cfgV2.Provider.OpenStack.Username,
			Password:          cfgV2.Provider.OpenStack.Password,
			DirectDownload:    cfgV2.Provider.OpenStack.DirectDownload,
		}
		cfgV3.Attestation.QEMUVTPM = &config.QEMUVTPM{
			Measurements: cfgV2.Provider.OpenStack.Measurements,
		}
	case cfgV2.Provider.QEMU != nil:
		cfgV3.Provider.QEMU = &config.QEMUConfig{
			ImageFormat:           cfgV2.Provider.QEMU.ImageFormat,
			VCPUs:                 cfgV2.Provider.QEMU.VCPUs,
			Memory:                cfgV2.Provider.QEMU.Memory,
			MetadataAPIImage:      cfgV2.Provider.QEMU.MetadataAPIImage,
			LibvirtURI:            cfgV2.Provider.QEMU.LibvirtURI,
			LibvirtContainerImage: cfgV2.Provider.QEMU.LibvirtContainerImage,
			NVRAM:                 cfgV2.Provider.QEMU.NVRAM,
			Firmware:              cfgV2.Provider.QEMU.Firmware,
		}
		cfgV3.Attestation.QEMUVTPM = &config.QEMUVTPM{
			Measurements: cfgV2.Provider.QEMU.Measurements,
		}
	}

	// Create backup
	if err := os.Rename(path, path+".backup.v2"); err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}

	// Write migrated config
	if err := fileHandler.WriteYAML(path, cfgV3, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

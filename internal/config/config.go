/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// This binary can be build from siderolabs/talos projects. Located at:
// https://github.com/siderolabs/talos/tree/master/hack/docgen
//
//go:generate docgen ./config.go ./config_doc.go Configuration

/*
Definitions for  Constellation's user config file.

The config file is used by the CLI to create and manage a Constellation cluster.

All config relevant definitions, parsing and validation functions should go here.
*/
package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"gopkg.in/yaml.v3"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config/imageversion"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/encoding"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

const (
	// Version4 is the fourth version number for Constellation config file.
	Version4 = "v4"

	defaultName = "constell"

	appRegistrationErrStr = "Azure app registrations are not supported since v2.9. Migrate to using a user assigned managed identity by following the migration guide: https://docs.edgeless.systems/constellation/reference/migration.\nPlease remove it from your config and from the Kubernetes secret in your running cluster. Ensure that the UAMI has all required permissions."
)

// Config defines configuration used by CLI.
type Config struct {
	// description: |
	//   Schema version of this configuration file.
	Version string `yaml:"version" validate:"eq=v4"`
	// description: |
	//   Machine image version used to create Constellation nodes.
	Image string `yaml:"image" validate:"required,image_compatibility"`
	// description: |
	//   Name of the cluster.
	Name string `yaml:"name" validate:"valid_name,required"`
	// description: |
	//   Kubernetes version to be installed into the cluster.
	KubernetesVersion versions.ValidK8sVersion `yaml:"kubernetesVersion" validate:"required,supported_k8s_version"`
	// description: |
	//   Microservice version to be installed into the cluster. Defaults to the version of the CLI.
	MicroserviceVersion semver.Semver `yaml:"microserviceVersion" validate:"required"`
	// description: |
	//   DON'T USE IN PRODUCTION: enable debug mode and use debug images.
	DebugCluster *bool `yaml:"debugCluster" validate:"required"`
	// description: |
	//   Optional custom endpoint (DNS name) for the Constellation API server.
	//   This can be used to point a custom dns name at the Constellation API server
	//   and is added to the Subject Alternative Name (SAN) field of the TLS certificate used by the API server.
	//   A fallback to DNS name is always available.
	CustomEndpoint string `yaml:"customEndpoint" validate:"omitempty,hostname_rfc1123"`
	// description: |
	//   Flag to enable/disable the internal load balancer. If enabled, the Constellation is only accessible from within the VPC.
	InternalLoadBalancer bool `yaml:"internalLoadBalancer" validate:"omitempty"`
	// description: |
	//   The Kubernetes Service CIDR to be used for the cluster. This value will only be used during the first initialization of the Constellation.
	ServiceCIDR string `yaml:"serviceCIDR" validate:"omitempty,cidrv4"`
	// description: |
	//   Supported cloud providers and their specific configurations.
	Provider ProviderConfig `yaml:"provider" validate:"dive"`
	// description: |
	//   Node groups to be created in the cluster.
	NodeGroups map[string]NodeGroup `yaml:"nodeGroups" validate:"required,dive"`
	// description: |
	//   Configuration for attestation validation. This configuration provides sensible defaults for the Constellation version it was created for.\nSee the docs for an overview on attestation: https://docs.edgeless.systems/constellation/architecture/attestation
	Attestation AttestationConfig `yaml:"attestation" validate:"dive"`
}

// ProviderConfig are cloud-provider specific configuration values used by the CLI.
// Fields should remain pointer-types so custom specific configs can nil them
// if not required.
type ProviderConfig struct {
	// description: |
	//   Configuration for AWS as provider.
	AWS *AWSConfig `yaml:"aws,omitempty" validate:"omitempty,dive"`
	// description: |
	//   Configuration for Azure as provider.
	Azure *AzureConfig `yaml:"azure,omitempty" validate:"omitempty,dive"`
	// description: |
	//   Configuration for Google Cloud as provider.
	GCP *GCPConfig `yaml:"gcp,omitempty" validate:"omitempty,dive"`
	// description: |
	//   Configuration for OpenStack as provider.
	OpenStack *OpenStackConfig `yaml:"openstack,omitempty" validate:"omitempty,dive"`
	// description: |
	//   Configuration for QEMU as provider.
	QEMU *QEMUConfig `yaml:"qemu,omitempty" validate:"omitempty,dive"`
}

// AWSConfig are AWS specific configuration values used by the CLI.
type AWSConfig struct {
	// description: |
	//   AWS data center region. See: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions
	Region string `yaml:"region" validate:"required,aws_region"`
	// description: |
	//   AWS data center zone name in defined region. See: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-availability-zones
	Zone string `yaml:"zone" validate:"required,aws_zone"`
	// description: |
	//   Name of the IAM profile to use for the control-plane nodes.
	IAMProfileControlPlane string `yaml:"iamProfileControlPlane" validate:"required"`
	// description: |
	//   Name of the IAM profile to use for the worker nodes.
	IAMProfileWorkerNodes string `yaml:"iamProfileWorkerNodes" validate:"required"`
	// description: |
	//   Deploy Persistent Disk CSI driver with on-node encryption. For details see: https://docs.edgeless.systems/constellation/architecture/encrypted-storage
	DeployCSIDriver *bool `yaml:"deployCSIDriver" validate:"required"`
}

// AzureConfig are Azure specific configuration values used by the CLI.
type AzureConfig struct {
	// description: |
	//   Subscription ID of the used Azure account. See: https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-subscription
	SubscriptionID string `yaml:"subscription" validate:"uuid"`
	// description: |
	//   Tenant ID of the used Azure account. See: https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-ad-tenant
	TenantID string `yaml:"tenant" validate:"uuid"`
	// description: |
	//   Azure datacenter region to be used. See: https://docs.microsoft.com/en-us/azure/availability-zones/az-overview#azure-regions-with-availability-zones
	Location string `yaml:"location" validate:"required"`
	// description: |
	//   Resource group for the cluster's resources. Must already exist.
	ResourceGroup string `yaml:"resourceGroup" validate:"required"`
	// description: |
	//   Authorize spawned VMs to access Azure API.
	UserAssignedIdentity string `yaml:"userAssignedIdentity" validate:"required"`
	// description: |
	//   Deploy Azure Disk CSI driver with on-node encryption. For details see: https://docs.edgeless.systems/constellation/architecture/encrypted-storage
	DeployCSIDriver *bool `yaml:"deployCSIDriver" validate:"required"`
	// description: |
	//   Enable secure boot for VMs. If enabled, the OS image has to include a virtual machine guest state (VMGS) blob.
	SecureBoot *bool `yaml:"secureBoot" validate:"required"`
	// description: |
	//   Use the specified Azure Marketplace image offering.
	UseMarketplaceImage *bool `yaml:"useMarketplaceImage" validate:"omitempty"`
}

// GCPConfig are GCP specific configuration values used by the CLI.
type GCPConfig struct {
	// description: |
	//   GCP project. See: https://support.google.com/googleapi/answer/7014113?hl=en
	Project string `yaml:"project" validate:"required"`
	// description: |
	//   GCP datacenter region. See: https://cloud.google.com/compute/docs/regions-zones#available
	Region string `yaml:"region" validate:"required"`
	// description: |
	//   GCP datacenter zone. See: https://cloud.google.com/compute/docs/regions-zones#available
	Zone string `yaml:"zone" validate:"required"`
	// description: |
	//   Path of service account key file. For required service account roles, see https://docs.edgeless.systems/constellation/getting-started/install#authorization
	ServiceAccountKeyPath string `yaml:"serviceAccountKeyPath" validate:"required"`
	// description: |
	//   Deploy Persistent Disk CSI driver with on-node encryption. For details see: https://docs.edgeless.systems/constellation/architecture/encrypted-storage
	DeployCSIDriver *bool `yaml:"deployCSIDriver" validate:"required"`
	// description: |
	//   Use the specified GCP Marketplace image offering.
	UseMarketplaceImage *bool `yaml:"useMarketplaceImage" validate:"omitempty"`
}

// OpenStackConfig holds config information for OpenStack based Constellation deployments.
type OpenStackConfig struct {
	// description: |
	//   OpenStack cloud name to select from "clouds.yaml". Only required if config file for OpenStack is used. Fallback authentication uses environment variables. For details see: https://docs.openstack.org/openstacksdk/latest/user/config/configuration.html.
	Cloud string `yaml:"cloud"`
	// description: |
	//   Availability zone to place the VMs in. For details see: https://docs.openstack.org/nova/latest/admin/availability-zones.html
	AvailabilityZone string `yaml:"availabilityZone" validate:"required"`
	// description: |
	//   Floating IP pool to use for the VMs. For details see: https://docs.openstack.org/ocata/user-guide/cli-manage-ip-addresses.html
	FloatingIPPoolID string `yaml:"floatingIPPoolID" validate:"required"`
	// description: |
	// AuthURL is the OpenStack Identity endpoint to use inside the cluster.
	AuthURL string `yaml:"authURL" validate:"required"`
	// description: |
	//   ProjectID is the ID of the project where a user resides.
	ProjectID string `yaml:"projectID" validate:"required"`
	// description: |
	//   ProjectName is the name of the project where a user resides.
	ProjectName string `yaml:"projectName" validate:"required"`
	// description: |
	//   UserDomainName is the name of the domain where a user resides.
	UserDomainName string `yaml:"userDomainName" validate:"required"`
	// description: |
	//   ProjectDomainName is the name of the domain where a project resides.
	ProjectDomainName string `yaml:"projectDomainName" validate:"required"`
	// description: |
	// RegionName is the name of the region to use inside the cluster.
	RegionName string `yaml:"regionName" validate:"required"`
	// description: |
	//   Username to use inside the cluster.
	Username string `yaml:"username" validate:"required"`
	// description: |
	//   Password to use inside the cluster. You can instead use the environment variable "CONSTELL_OS_PASSWORD".
	Password string `yaml:"password"`
	// description: |
	//   If enabled, downloads OS image directly from source URL to OpenStack. Otherwise, downloads image to local machine and uploads to OpenStack.
	DirectDownload *bool `yaml:"directDownload" validate:"required"`
	// description: |
	//   Deploy Yawol loadbalancer. For details see: https://github.com/stackitcloud/yawol
	DeployYawolLoadBalancer *bool `yaml:"deployYawolLoadBalancer" validate:"required"`
	// description: |
	//   OpenStack OS image used by the yawollet. For details see: https://github.com/stackitcloud/yawol
	YawolImageID string `yaml:"yawolImageID"`
	// description: |
	//   OpenStack flavor id used for yawollets. For details see: https://github.com/stackitcloud/yawol
	YawolFlavorID string `yaml:"yawolFlavorID"`
	// description: |
	//   Deploy Cinder CSI driver with on-node encryption. For details see: https://docs.edgeless.systems/constellation/architecture/encrypted-storage
	DeployCSIDriver *bool `yaml:"deployCSIDriver" validate:"required"`
}

// QEMUConfig holds config information for QEMU based Constellation deployments.
type QEMUConfig struct {
	// description: |
	//   Format of the image to use for the VMs. Should be either qcow2 or raw.
	ImageFormat string `yaml:"imageFormat" validate:"oneof=qcow2 raw"`
	// description: |
	//   vCPU count for the VMs.
	VCPUs int `yaml:"vcpus" validate:"required"`
	// description: |
	//   Amount of memory per instance (MiB).
	Memory int `yaml:"memory" validate:"required"`
	// description: |
	//   Container image to use for the QEMU metadata server.
	MetadataAPIImage string `yaml:"metadataAPIServer" validate:"required"`
	// description: |
	//   Libvirt connection URI. Leave empty to start a libvirt instance in Docker.
	LibvirtURI string `yaml:"libvirtSocket"`
	// description: |
	//   Container image to use for launching a containerized libvirt daemon. Only relevant if `libvirtSocket = ""`.
	LibvirtContainerImage string `yaml:"libvirtContainerImage"`
	// description: |
	//   NVRAM template to be used for secure boot. Can be sentinel value "production", "testing" or a path to a custom NVRAM template
	NVRAM string `yaml:"nvram" validate:"required"`
	// description: |
	//   Path to the OVMF firmware. Leave empty for auto selection.
	Firmware string `yaml:"firmware"`
}

// AttestationConfig configuration values used for attestation.
// Fields should remain pointer-types so custom specific configs can nil them
// if not required.
type AttestationConfig struct {
	// description: |
	//   AWS SEV-SNP attestation.
	AWSSEVSNP *AWSSEVSNP `yaml:"awsSEVSNP,omitempty" validate:"omitempty,dive"`
	// description: |
	//   AWS Nitro TPM attestation.
	AWSNitroTPM *AWSNitroTPM `yaml:"awsNitroTPM,omitempty" validate:"omitempty,dive"`
	// description: |
	//   Azure SEV-SNP attestation.\nFor details see: https://docs.edgeless.systems/constellation/architecture/attestation#cvm-verification
	AzureSEVSNP *AzureSEVSNP `yaml:"azureSEVSNP,omitempty" validate:"omitempty,dive"`
	// description: |
	//   Azure TDX attestation.
	AzureTDX *AzureTDX `yaml:"azureTDX,omitempty" validate:"omitempty,dive"`
	// description: |
	//   Azure TPM attestation (Trusted Launch).
	AzureTrustedLaunch *AzureTrustedLaunch `yaml:"azureTrustedLaunch,omitempty" validate:"omitempty,dive"`
	// description: |
	//   GCP SEV-ES attestation.
	GCPSEVES *GCPSEVES `yaml:"gcpSEVES,omitempty" validate:"omitempty,dive"`
	// description: |
	//   QEMU tdx attestation.
	QEMUTDX *QEMUTDX `yaml:"qemuTDX,omitempty" validate:"omitempty,dive"`
	// description: |
	//   QEMU vTPM attestation.
	QEMUVTPM *QEMUVTPM `yaml:"qemuVTPM,omitempty" validate:"omitempty,dive"`
}

// NodeGroup defines a group of nodes with the same role and configuration.
// Cloud providers use scaling groups to manage nodes of a group.
type NodeGroup struct {
	// description: |
	//   Role of the nodes in this group. Valid values are "control-plane" and "worker".
	Role string `yaml:"role" validate:"required,oneof=control-plane worker"`
	// description: |
	//   Availability zone to place the VMs in.
	Zone string `yaml:"zone" validate:"valid_zone"`
	// description: |
	//   VM instance type to use for the nodes.
	InstanceType string `yaml:"instanceType" validate:"instance_type"`
	// description: |
	//   Size (in GB) of a node's disk to store the non-volatile state.
	StateDiskSizeGB int `yaml:"stateDiskSizeGB" validate:"min=0"`
	// description: |
	//   Type of a node's state disk. The type influences boot time and I/O performance.
	StateDiskType string `yaml:"stateDiskType" validate:"disk_type"`
	// description: |
	//   Number of nodes to be initially created.
	InitialCount int `yaml:"initialCount" validate:"min=0"`
}

// Default returns a struct with the default config.
// IMPORTANT: Ensure that any state mutation is followed by a call to Validate() to ensure that the config is always in a valid state. Avoid usage outside of tests.
func Default() *Config {
	return &Config{
		Version:             Version4,
		Image:               defaultImage,
		Name:                defaultName,
		MicroserviceVersion: constants.BinaryVersion(),
		KubernetesVersion:   versions.Default,
		DebugCluster:        toPtr(false),
		ServiceCIDR:         "10.96.0.0/12",
		Provider: ProviderConfig{
			AWS: &AWSConfig{
				Region:                 "",
				IAMProfileControlPlane: "",
				IAMProfileWorkerNodes:  "",
				DeployCSIDriver:        toPtr(true),
			},
			Azure: &AzureConfig{
				SubscriptionID:       "",
				TenantID:             "",
				Location:             "",
				UserAssignedIdentity: "",
				ResourceGroup:        "",
				DeployCSIDriver:      toPtr(true),
				SecureBoot:           toPtr(false),
				UseMarketplaceImage:  toPtr(false),
			},
			GCP: &GCPConfig{
				Project:               "",
				Region:                "",
				Zone:                  "",
				ServiceAccountKeyPath: "",
				DeployCSIDriver:       toPtr(true),
				UseMarketplaceImage:   toPtr(false),
			},
			OpenStack: &OpenStackConfig{
				DirectDownload:          toPtr(true),
				DeployYawolLoadBalancer: toPtr(true),
				DeployCSIDriver:         toPtr(true),
			},
			QEMU: &QEMUConfig{
				ImageFormat:           "raw",
				VCPUs:                 2,
				Memory:                2048,
				MetadataAPIImage:      imageversion.QEMUMetadata(),
				LibvirtURI:            "",
				LibvirtContainerImage: imageversion.Libvirt(),
				NVRAM:                 "production",
			},
		},
		NodeGroups: map[string]NodeGroup{
			constants.DefaultControlPlaneGroupName: {
				Role:            "control-plane",
				Zone:            "",
				InstanceType:    "",
				StateDiskSizeGB: 30,
				StateDiskType:   "",
				InitialCount:    3,
			},
			constants.DefaultWorkerGroupName: {
				Role:            "worker",
				Zone:            "",
				InstanceType:    "",
				StateDiskSizeGB: 30,
				StateDiskType:   "",
				InitialCount:    1,
			},
		},
		// TODO(malt3): remove default attestation config as soon as one-to-one mapping is no longer possible.
		// Some problematic pairings:
		// OpenStack uses qemu-vtpm as attestation variant
		// QEMU uses qemu-vtpm as attestation variant
		// AWS uses aws-nitro-tpm as attestation variant
		// AWS will have aws-sev-snp as attestation variant
		Attestation: AttestationConfig{
			AWSSEVSNP:          DefaultForAWSSEVSNP(),
			AWSNitroTPM:        &AWSNitroTPM{Measurements: measurements.DefaultsFor(cloudprovider.AWS, variant.AWSNitroTPM{})},
			AzureSEVSNP:        DefaultForAzureSEVSNP(),
			AzureTDX:           DefaultForAzureTDX(),
			AzureTrustedLaunch: &AzureTrustedLaunch{Measurements: measurements.DefaultsFor(cloudprovider.Azure, variant.AzureTrustedLaunch{})},
			GCPSEVES:           &GCPSEVES{Measurements: measurements.DefaultsFor(cloudprovider.GCP, variant.GCPSEVES{})},
			QEMUVTPM:           &QEMUVTPM{Measurements: measurements.DefaultsFor(cloudprovider.QEMU, variant.QEMUVTPM{})},
		},
	}
}

// MiniDefault returns a default config for a mini cluster.
func MiniDefault() (*Config, error) {
	config := Default()
	config.Name = constants.MiniConstellationUID
	config.RemoveProviderAndAttestationExcept(cloudprovider.QEMU)
	for groupName, group := range config.NodeGroups {
		group.StateDiskSizeGB = 8
		group.InitialCount = 1
		config.NodeGroups[groupName] = group
	}
	// only release images (e.g. v2.7.0) use the production NVRAM
	if !config.IsReleaseImage() {
		config.Provider.QEMU.NVRAM = "testing"
	}
	return config, config.Validate(false)
}

// fromFile returns config file with `name` read from `fileHandler` by parsing
// it as YAML. You should prefer config.New to read env vars and validate
// config in a consistent manner.
func fromFile(fileHandler file.Handler, name string) (*Config, error) {
	var conf Config
	if err := fileHandler.ReadYAMLStrict(name, &conf); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("unable to find %s - use `constellation config generate` to generate it first", name)
		}
		if isAppClientIDError(err) {
			return nil, &UnsupportedAppRegistrationError{}
		}
		return nil, fmt.Errorf("could not load config from file %s: %w", name, err)
	}
	return &conf, nil
}

func isAppClientIDError(err error) bool {
	var yamlErr *yaml.TypeError
	if errors.As(err, &yamlErr) {
		for _, e := range yamlErr.Errors {
			if strings.Contains(e, "appClientID") {
				return true
			}
		}
	}
	return false
}

// UnsupportedAppRegistrationError is returned when the config contains configuration related to now unsupported app registrations.
type UnsupportedAppRegistrationError struct{}

func (e *UnsupportedAppRegistrationError) Error() string {
	return appRegistrationErrStr
}

// New creates a new config by:
// 1. Reading config file via provided fileHandler from file with name.
// 2. For "latest" version values of the attestation variants fetch the version numbers.
// 3. Read secrets from environment variables.
// 4. Validate config. If `--force` is set the version validation will be disabled and any version combination is allowed.
func New(fileHandler file.Handler, name string, fetcher attestationconfigapi.Fetcher, force bool) (*Config, error) {
	// Read config file
	c, err := fromFile(fileHandler, name)
	if err != nil {
		return nil, err
	}

	if azure := c.Attestation.AzureSEVSNP; azure != nil {
		if err := azure.FetchAndSetLatestVersionNumbers(context.Background(), fetcher); err != nil {
			return c, err
		}
	}

	if aws := c.Attestation.AWSSEVSNP; aws != nil {
		if err := aws.FetchAndSetLatestVersionNumbers(context.Background(), fetcher); err != nil {
			return c, err
		}
	}

	// Read secrets from env-vars.
	clientSecretValue := os.Getenv(constants.EnvVarAzureClientSecretValue)
	if clientSecretValue != "" && c.Provider.Azure != nil {
		fmt.Fprintf(os.Stderr, "WARNING: the environment variable %s is no longer used %s", constants.EnvVarAzureClientSecretValue, appRegistrationErrStr)
	}

	openstackPassword := os.Getenv(constants.EnvVarOpenStackPassword)
	if openstackPassword != "" && c.Provider.OpenStack != nil {
		c.Provider.OpenStack.Password = openstackPassword
	}

	return c, c.Validate(force)
}

// HasProvider checks whether the config contains the provider.
func (c *Config) HasProvider(provider cloudprovider.Provider) bool {
	switch provider {
	case cloudprovider.AWS:
		return c.Provider.AWS != nil
	case cloudprovider.Azure:
		return c.Provider.Azure != nil
	case cloudprovider.GCP:
		return c.Provider.GCP != nil
	case cloudprovider.OpenStack:
		return c.Provider.OpenStack != nil
	case cloudprovider.QEMU:
		return c.Provider.QEMU != nil
	}
	return false
}

// UpdateMeasurements overwrites measurements in config with the provided ones.
func (c *Config) UpdateMeasurements(newMeasurements measurements.M) {
	if c.Attestation.AWSSEVSNP != nil {
		c.Attestation.AWSSEVSNP.Measurements.CopyFrom(newMeasurements)
	}
	if c.Attestation.AWSNitroTPM != nil {
		c.Attestation.AWSNitroTPM.Measurements.CopyFrom(newMeasurements)
	}
	if c.Attestation.AzureSEVSNP != nil {
		c.Attestation.AzureSEVSNP.Measurements.CopyFrom(newMeasurements)
	}
	if c.Attestation.AzureTDX != nil {
		c.Attestation.AzureTDX.Measurements.CopyFrom(newMeasurements)
	}
	if c.Attestation.AzureTrustedLaunch != nil {
		c.Attestation.AzureTrustedLaunch.Measurements.CopyFrom(newMeasurements)
	}
	if c.Attestation.GCPSEVES != nil {
		c.Attestation.GCPSEVES.Measurements.CopyFrom(newMeasurements)
	}
	if c.Attestation.QEMUVTPM != nil {
		c.Attestation.QEMUVTPM.Measurements.CopyFrom(newMeasurements)
	}
}

// RemoveProviderAndAttestationExcept calls RemoveProviderExcept and sets the default attestations for the provider (only used for convenience in tests).
func (c *Config) RemoveProviderAndAttestationExcept(provider cloudprovider.Provider) {
	c.RemoveProviderExcept(provider)
	c.SetAttestation(variant.GetDefaultAttestation(provider))
}

// RemoveProviderExcept removes all provider specific configurations, i.e.,
// sets them to nil, except the one specified.
// If an unknown provider is passed, the same configuration is returned.
func (c *Config) RemoveProviderExcept(provider cloudprovider.Provider) {
	currentProviderConfigs := c.Provider
	c.Provider = ProviderConfig{}

	switch provider {
	case cloudprovider.AWS:
		c.Provider.AWS = currentProviderConfigs.AWS
	case cloudprovider.Azure:
		c.Provider.Azure = currentProviderConfigs.Azure
	case cloudprovider.GCP:
		c.Provider.GCP = currentProviderConfigs.GCP
	case cloudprovider.OpenStack:
		c.Provider.OpenStack = currentProviderConfigs.OpenStack
	case cloudprovider.QEMU:
		c.Provider.QEMU = currentProviderConfigs.QEMU
	default:
		c.Provider = currentProviderConfigs
	}
	c.SetCSPNodeGroupDefaults(provider)
}

// SetAttestation sets the attestation config for the given attestation variant and removes all other attestation configs.
func (c *Config) SetAttestation(attestation variant.Variant) {
	currentAttestationConfigs := c.Attestation
	c.Attestation = AttestationConfig{}
	switch attestation.(type) {
	case variant.AWSSEVSNP:
		c.Attestation = AttestationConfig{AWSSEVSNP: currentAttestationConfigs.AWSSEVSNP}
	case variant.AWSNitroTPM:
		c.Attestation = AttestationConfig{AWSNitroTPM: currentAttestationConfigs.AWSNitroTPM}
	case variant.AzureSEVSNP:
		c.Attestation = AttestationConfig{AzureSEVSNP: currentAttestationConfigs.AzureSEVSNP}
	case variant.AzureTDX:
		c.Attestation = AttestationConfig{AzureTDX: currentAttestationConfigs.AzureTDX}
	case variant.AzureTrustedLaunch:
		c.Attestation = AttestationConfig{AzureTrustedLaunch: currentAttestationConfigs.AzureTrustedLaunch}
	case variant.GCPSEVES:
		c.Attestation = AttestationConfig{GCPSEVES: currentAttestationConfigs.GCPSEVES}
	case variant.QEMUVTPM:
		c.Attestation = AttestationConfig{QEMUVTPM: currentAttestationConfigs.QEMUVTPM}
	}
}

// IsDebugCluster checks whether the cluster is configured as a debug cluster.
func (c *Config) IsDebugCluster() bool {
	if c.DebugCluster != nil && *c.DebugCluster {
		return true
	}
	return false
}

// IsReleaseImage checks whether image name looks like a release image.
func (c *Config) IsReleaseImage() bool {
	return strings.HasPrefix(c.Image, "v")
}

// IsNamedLikeDebugImage checks whether image name looks like a debug image.
func (c *Config) IsNamedLikeDebugImage() bool {
	v, err := versionsapi.NewVersionFromShortPath(c.Image, versionsapi.VersionKindImage)
	if err != nil {
		return false
	}
	return v.Stream() == "debug"
}

// GetProvider returns the configured cloud provider.
func (c *Config) GetProvider() cloudprovider.Provider {
	if c.Provider.AWS != nil {
		return cloudprovider.AWS
	}
	if c.Provider.Azure != nil {
		return cloudprovider.Azure
	}
	if c.Provider.GCP != nil {
		return cloudprovider.GCP
	}
	if c.Provider.OpenStack != nil {
		return cloudprovider.OpenStack
	}
	if c.Provider.QEMU != nil {
		return cloudprovider.QEMU
	}
	return cloudprovider.Unknown
}

// GetAttestationConfig returns the configured attestation config.
func (c *Config) GetAttestationConfig() AttestationCfg {
	if c.Attestation.AWSSEVSNP != nil {
		return c.Attestation.AWSSEVSNP.getToMarshallLatestWithResolvedVersions()
	}
	if c.Attestation.AWSNitroTPM != nil {
		return c.Attestation.AWSNitroTPM
	}
	if c.Attestation.AzureSEVSNP != nil {
		return c.Attestation.AzureSEVSNP.getToMarshallLatestWithResolvedVersions()
	}
	if c.Attestation.AzureTDX != nil {
		return c.Attestation.AzureTDX.getToMarshallLatestWithResolvedVersions()
	}
	if c.Attestation.AzureTrustedLaunch != nil {
		return c.Attestation.AzureTrustedLaunch
	}
	if c.Attestation.GCPSEVES != nil {
		return c.Attestation.GCPSEVES
	}
	if c.Attestation.QEMUVTPM != nil {
		return c.Attestation.QEMUVTPM
	}
	return &DummyCfg{}
}

// GetRegion returns the configured region.
func (c *Config) GetRegion() string {
	switch c.GetProvider() {
	case cloudprovider.AWS:
		return c.Provider.AWS.Region
	case cloudprovider.Azure:
		return c.Provider.Azure.Location
	case cloudprovider.GCP:
		return c.Provider.GCP.Region
	case cloudprovider.OpenStack:
		return c.Provider.OpenStack.RegionName
	case cloudprovider.QEMU:
		return ""
	}
	return ""
}

// GetZone returns the configured zone or location for providers without zone support (Azure).
func (c *Config) GetZone() string {
	switch c.GetProvider() {
	case cloudprovider.AWS:
		return c.Provider.AWS.Zone
	case cloudprovider.Azure:
		return c.Provider.Azure.Location
	case cloudprovider.GCP:
		return c.Provider.GCP.Zone
	}
	return ""
}

// UpdateMAAURL updates the MAA URL in the config.
func (c *Config) UpdateMAAURL(maaURL string) {
	if c.Attestation.AzureSEVSNP != nil {
		c.Attestation.AzureSEVSNP.FirmwareSignerConfig.MAAURL = maaURL
	}
}

// DeployCSIDriver returns whether the CSI driver should be deployed for a given cloud provider.
func (c *Config) DeployCSIDriver() bool {
	return c.Provider.Azure != nil && c.Provider.Azure.DeployCSIDriver != nil && *c.Provider.Azure.DeployCSIDriver ||
		c.Provider.AWS != nil && c.Provider.AWS.DeployCSIDriver != nil && *c.Provider.AWS.DeployCSIDriver ||
		c.Provider.GCP != nil && c.Provider.GCP.DeployCSIDriver != nil && *c.Provider.GCP.DeployCSIDriver ||
		c.Provider.OpenStack != nil && c.Provider.OpenStack.DeployCSIDriver != nil && *c.Provider.OpenStack.DeployCSIDriver
}

// DeployYawolLoadBalancer returns whether the Yawol load balancer should be deployed.
func (c *Config) DeployYawolLoadBalancer() bool {
	return c.Provider.OpenStack != nil && c.Provider.OpenStack.DeployYawolLoadBalancer != nil && *c.Provider.OpenStack.DeployYawolLoadBalancer
}

// UseMarketplaceImage returns whether a marketplace image should be used.
func (c *Config) UseMarketplaceImage() bool {
	return (c.Provider.Azure != nil && c.Provider.Azure.UseMarketplaceImage != nil && *c.Provider.Azure.UseMarketplaceImage) ||
		(c.Provider.GCP != nil && c.Provider.GCP.UseMarketplaceImage != nil && *c.Provider.GCP.UseMarketplaceImage)
}

// Validate checks the config values and returns validation errors.
func (c *Config) Validate(force bool) error {
	trans := ut.New(en.New()).GetFallback()
	validate := validator.New()
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		return err
	}

	// Register name function to return yaml name tag
	// This makes sure methods like fl.FieldName() return the yaml name tag instead of the struct field name
	// e.g. struct{DataType string `yaml:"foo,omitempty"`} will return `foo` instead of `DataType` when calling fl.FieldName()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("yaml"), ",")
		if name == "-" {
			return ""
		}
		return name
	})

	// Register AWS, Azure & GCP InstanceType validation error types
	if err := validate.RegisterTranslation("instance_type", trans, c.registerTranslateInstanceTypeError, c.translateInstanceTypeError); err != nil {
		return err
	}

	// Register Provider validation error types
	if err := validate.RegisterTranslation("no_provider", trans, registerNoProviderError, translateNoProviderError); err != nil {
		return err
	}

	if err := validate.RegisterTranslation("more_than_one_provider", trans, registerMoreThanOneProviderError, c.translateMoreThanOneProviderError); err != nil {
		return err
	}

	if err := validate.RegisterTranslation("no_placeholders", trans, registerContainsPlaceholderError, translateContainsPlaceholderError); err != nil {
		return err
	}

	if err := validate.RegisterTranslation("supported_k8s_version", trans, registerInvalidK8sVersionError, translateInvalidK8sVersionError); err != nil {
		return err
	}

	if err := validate.RegisterTranslation("image_compatibility", trans, registerImageCompatibilityError, translateImageCompatibilityError); err != nil {
		return err
	}

	if err := validate.RegisterTranslation("valid_name", trans, c.registerValidateNameError, c.translateValidateNameError); err != nil {
		return err
	}

	if err := validate.RegisterValidation("valid_name", c.validateName); err != nil {
		return err
	}

	if err := validate.RegisterValidation("no_placeholders", validateNoPlaceholder); err != nil {
		return err
	}

	// register custom validator with label supported_k8s_version to validate version based on available versionConfigs.
	if err := validate.RegisterValidation("supported_k8s_version", c.validateK8sVersion); err != nil {
		return err
	}

	versionCompatibilityValidator := validateVersionCompatibility
	if force {
		versionCompatibilityValidator = returnsTrue
	}
	if err := validate.RegisterValidation("image_compatibility", versionCompatibilityValidator); err != nil {
		return err
	}

	if err := validate.RegisterValidation("disk_type", c.validateStateDiskTypeField); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("disk_type", trans, registerTranslateDiskTypeError, c.translateDiskTypeError); err != nil {
		return err
	}

	if err := validate.RegisterValidation("instance_type", c.validateInstanceType); err != nil {
		return err
	}

	if err := validate.RegisterValidation("deprecated", warnDeprecated); err != nil {
		return err
	}

	// Register provider validation
	validate.RegisterStructValidation(validateProvider, ProviderConfig{})

	// Register NodeGroup validation error types
	if err := validate.RegisterTranslation("no_default_control_plane_group", trans, registerNoDefaultControlPlaneGroupError, translateNoDefaultControlPlaneGroupError); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("no_default_worker_group", trans, registerNoDefaultWorkerGroupError, translateNoDefaultWorkerGroupError); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("control_plane_group_initial_count", trans, registerControlPlaneGroupInitialCountError, translateControlPlaneGroupInitialCountError); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("control_plane_group_role_mismatch", trans, registerControlPlaneGroupRoleMismatchError, translateControlPlaneGroupRoleMismatchError); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("worker_group_role_mismatch", trans, registerWorkerGroupRoleMismatchError, translateWorkerGroupRoleMismatchError); err != nil {
		return err
	}

	// Register NodeGroup validation
	validate.RegisterStructValidation(validateNodeGroups, Config{})

	// Register Attestation validation error types
	if err := validate.RegisterTranslation("no_attestation", trans, registerNoAttestationError, translateNoAttestationError); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("more_than_one_attestation", trans, registerMoreThanOneAttestationError, c.translateMoreThanOneAttestationError); err != nil {
		return err
	}

	if err := validate.RegisterValidation("valid_zone", c.validateNodeGroupZoneField); err != nil {
		return err
	}
	if err := validate.RegisterValidation("aws_region", validateAWSRegionField); err != nil {
		return err
	}
	if err := validate.RegisterValidation("aws_zone", validateAWSZoneField); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("valid_zone", trans, registerValidZoneError, c.translateValidZoneError); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("aws_region", trans, registerAWSRegionError, translateAWSRegionError); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("aws_zone", trans, registerAWSZoneError, translateAWSZoneError); err != nil {
		return err
	}

	validate.RegisterStructValidation(validateMeasurement, measurements.Measurement{})
	validate.RegisterStructValidation(validateAttestation, AttestationConfig{})

	if !force {
		// Validating MicroserviceVersion separately is required since it is a custom type.
		// The validation pkg we use does not allow accessing the field name during struct validation.
		// Because of this we can't print the offending field name in the error message, resulting in
		// suboptimal UX. Adding the field name to the struct validation of Semver would make it
		// impossible to use Semver for other fields.
		if err := ValidateMicroserviceVersion(constants.BinaryVersion(), c.MicroserviceVersion); err != nil {
			msg := "microserviceVersion: " + msgFromCompatibilityError(err, constants.BinaryVersion().String(), c.MicroserviceVersion.String())
			return &ValidationError{validationErrMsgs: []string{msg}}
		}
	}

	if c.InternalLoadBalancer {
		if c.GetProvider() != cloudprovider.AWS && c.GetProvider() != cloudprovider.GCP {
			return &ValidationError{validationErrMsgs: []string{"internalLoadBalancer is only supported for AWS and GCP"}}
		}
	}

	err := validate.Struct(c)
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return err
	}

	var validationErrMsgs []string
	for _, e := range validationErrs {
		validationErrMsgs = append(validationErrMsgs, e.Translate(trans))
	}

	return &ValidationError{validationErrMsgs: validationErrMsgs}
}

// WithOpenStackProviderDefaults fills the default values for the specific OpenStack provider.
// If the provider is not supported or not an OpenStack provider, the config is returned unchanged.
func (c *Config) WithOpenStackProviderDefaults(openStackProvider string) *Config {
	switch openStackProvider {
	case "stackit":
		c.Provider.OpenStack.Cloud = "stackit"
		c.Provider.OpenStack.FloatingIPPoolID = "970ace5c-458f-484a-a660-0903bcfd91ad"
		c.Provider.OpenStack.AuthURL = "https://keystone.api.iaas.eu01.stackit.cloud/v3"
		c.Provider.OpenStack.UserDomainName = "portal_mvp"
		c.Provider.OpenStack.ProjectDomainName = "portal_mvp"
		c.Provider.OpenStack.RegionName = "RegionOne"
		c.Provider.OpenStack.DeployYawolLoadBalancer = toPtr(true)
		c.Provider.OpenStack.YawolImageID = "43d9bede-1e7a-4ca7-82c5-0a5c72388619"
		c.Provider.OpenStack.YawolFlavorID = "3b11b27e-6c73-470d-b595-1d85b95a8cdf"
		c.Provider.OpenStack.DeployCSIDriver = toPtr(true)
		c.Provider.OpenStack.DirectDownload = toPtr(true)
		for groupName, group := range c.NodeGroups {
			group.InstanceType = "2715eabe-3ffc-4c36-b02a-efa8c141a96a"
			group.StateDiskType = "storage_premium_perf6"
			c.NodeGroups[groupName] = group
		}
		return c
	}
	return c
}

// SetCSPNodeGroupDefaults sets the default values for the node groups based on the configured CSP.
func (c *Config) SetCSPNodeGroupDefaults(csp cloudprovider.Provider) {
	var instanceType, stateDiskType, zone string
	switch csp {
	case cloudprovider.AWS:
		instanceType = "m6a.xlarge"
		stateDiskType = "gp3"
		zone = c.Provider.AWS.Zone
	case cloudprovider.Azure:
		// Check attestation variant, and use different default instance type if we have TDX
		if c.GetAttestationConfig().GetVariant().Equal(variant.AzureTDX{}) {
			instanceType = "Standard_DC4as_v5"
		} else {
			instanceType = "Standard_DC4es_v5"
		}
		stateDiskType = "Premium_LRS"
	case cloudprovider.GCP:
		instanceType = "n2d-standard-4"
		stateDiskType = "pd-ssd"
		zone = c.Provider.GCP.Zone
	case cloudprovider.QEMU, cloudprovider.OpenStack:
		// empty. There are no defaults for this CSP
	}

	for groupName, group := range c.NodeGroups {
		if len(group.InstanceType) == 0 && len(instanceType) != 0 {
			group.InstanceType = instanceType
		}
		if len(group.StateDiskType) == 0 && len(stateDiskType) != 0 {
			group.StateDiskType = stateDiskType
		}
		if len(group.Zone) == 0 && len(zone) != 0 {
			group.Zone = zone
		}
		c.NodeGroups[groupName] = group
	}
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

// AWSSEVSNP is the configuration for AWS SEV-SNP attestation.
type AWSSEVSNP struct {
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
	//   AMD Root Key certificate used to verify the SEV-SNP certificate chain.
	AMDRootKey Certificate `json:"amdRootKey" yaml:"amdRootKey"`
	// description: |
	//   AMD Signing Key certificate used to verify the SEV-SNP VCEK / VLEK certificate.
	AMDSigningKey Certificate `json:"amdSigningKey,omitempty" yaml:"amdSigningKey,omitempty"`
}

// AWSNitroTPM is the configuration for AWS Nitro TPM attestation.
type AWSNitroTPM struct {
	// description: |
	//   Expected TPM measurements.
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
	// description: |
	//   AMD Signing Key certificate used to verify the SEV-SNP VCEK / VLEK certificate.
	AMDSigningKey Certificate `json:"amdSigningKey,omitempty" yaml:"amdSigningKey,omitempty" validate:"len=0"`
}

// AzureTrustedLaunch is the configuration for Azure Trusted Launch attestation.
type AzureTrustedLaunch struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// AzureTDX is the configuration for Azure TDX attestation.
type AzureTDX struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
	// description: |
	//   Minimum required QE security version number (SVN).
	QESVN uint16 `json:"qeSVN" yaml:"qeSVN"`
	// description: |
	//   Minimum required PCE security version number (SVN).
	PCESVN uint16 `json:"pceSVN" yaml:"pceSVN"`
	// description: |
	//   Component-wise minimum required 16 byte hex-encoded TEE_TCB security version number (SVN).
	TEETCBSVN encoding.HexBytes `json:"teeTCBSVN" yaml:"teeTCBSVN"`
	// description: |
	//   Expected 16 byte hex-encoded QE_VENDOR_ID field.
	QEVendorID encoding.HexBytes `json:"qeVendorID" yaml:"qeVendorID"`
	// description: |
	//   Expected 48 byte hex-encoded MR_SEAM value.
	MRSeam encoding.HexBytes `json:"mrSeam" yaml:"mrSeam"`
	// description: |
	//   Expected 8 byte hex-encoded XFAM field.
	XFAM encoding.HexBytes `json:"xfam" yaml:"xfam"`
	// description: |
	//   Intel Root Key certificate used to verify the TDX certificate chain.
	IntelRootKey Certificate `json:"intelRootKey" yaml:"intelRootKey"`
}

func toPtr[T any](v T) *T {
	return &v
}

// svnResolveMarshaller is used to marshall "latest" security version numbers with resolved versions.
type svnResolveMarshaller interface {
	// getToMarshallLatestWithResolvedVersions brings the attestation config into a state where marshalling uses the numerical version numbers for "latest" versions.
	getToMarshallLatestWithResolvedVersions() AttestationCfg
}

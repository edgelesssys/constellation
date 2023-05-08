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
	"bytes"
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

	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config/imageversion"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

const (
	// Version3 is the third version number for Constellation config file.
	Version3 = "v3"

	defaultName = "constell"
)

// Config defines configuration used by CLI.
type Config struct {
	// description: |
	//   Schema version of this configuration file.
	Version string `yaml:"version" validate:"eq=v3"`
	// description: |
	//   Machine image version used to create Constellation nodes.
	Image string `yaml:"image" validate:"required,version_compatibility"`
	// description: |
	//   Name of the cluster.
	Name string `yaml:"name" validate:"valid_name,required"`
	// description: |
	//   Size (in GB) of a node's disk to store the non-volatile state.
	StateDiskSizeGB int `yaml:"stateDiskSizeGB" validate:"min=0"`
	// description: |
	//   Kubernetes version to be installed into the cluster.
	KubernetesVersion string `yaml:"kubernetesVersion" validate:"required,supported_k8s_version"`
	// description: |
	//   Microservice version to be installed into the cluster. Defaults to the version of the CLI.
	MicroserviceVersion string `yaml:"microserviceVersion" validate:"required,version_compatibility"`
	// description: |
	//   DON'T USE IN PRODUCTION: enable debug mode and use debug images. For usage, see: https://github.com/edgelesssys/constellation/blob/main/debugd/README.md
	DebugCluster *bool `yaml:"debugCluster" validate:"required"`
	// description: |
	//   Supported cloud providers and their specific configurations.
	Provider ProviderConfig `yaml:"provider" validate:"dive"`
	// description: |
	//   Configuration for attestation validation. This configuration provides sensible defaults for the Constellation version it was created for.\nSee our docs for an overview on attestation: https://docs.edgeless.systems/constellation/architecture/attestation
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
	Region string `yaml:"region" validate:"required"`
	// description: |
	//   AWS data center zone name in defined region. See: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-availability-zones
	Zone string `yaml:"zone" validate:"required"`
	// description: |
	//   VM instance type to use for Constellation nodes. Needs to support NitroTPM. See: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enable-nitrotpm-prerequisites.html
	InstanceType string `yaml:"instanceType" validate:"lowercase,aws_instance_type"`
	// description: |
	//   Type of a node's state disk. The type influences boot time and I/O performance. See: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html
	StateDiskType string `yaml:"stateDiskType" validate:"oneof=standard gp2 gp3 st1 sc1 io1"`
	// description: |
	//   Name of the IAM profile to use for the control plane nodes.
	IAMProfileControlPlane string `yaml:"iamProfileControlPlane" validate:"required"`
	// description: |
	//   Name of the IAM profile to use for the worker nodes.
	IAMProfileWorkerNodes string `yaml:"iamProfileWorkerNodes" validate:"required"`
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
	//    Application client ID of the Active Directory app registration.
	AppClientID string `yaml:"appClientID" validate:"uuid"`
	// description: |
	//    Client secret value of the Active Directory app registration credentials. Alternatively leave empty and pass value via CONSTELL_AZURE_CLIENT_SECRET_VALUE environment variable.
	ClientSecretValue string `yaml:"clientSecretValue" validate:"required"`
	// description: |
	//   VM instance type to use for Constellation nodes.
	InstanceType string `yaml:"instanceType" validate:"azure_instance_type"`
	// description: |
	//   Type of a node's state disk. The type influences boot time and I/O performance. See: https://docs.microsoft.com/en-us/azure/virtual-machines/disks-types#disk-type-comparison
	StateDiskType string `yaml:"stateDiskType" validate:"oneof=Premium_LRS Premium_ZRS Standard_LRS StandardSSD_LRS StandardSSD_ZRS"`
	// description: |
	//   Deploy Azure Disk CSI driver with on-node encryption. For details see: https://docs.edgeless.systems/constellation/architecture/encrypted-storage
	DeployCSIDriver *bool `yaml:"deployCSIDriver" validate:"required"`
	// description: |
	//   Enable secure boot for VMs. If enabled, the OS image has to include a virtual machine guest state (VMGS) blob.
	SecureBoot *bool `yaml:"secureBoot" validate:"required"`
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
	//   VM instance type to use for Constellation nodes.
	InstanceType string `yaml:"instanceType" validate:"gcp_instance_type"`
	// description: |
	//   Type of a node's state disk. The type influences boot time and I/O performance. See: https://cloud.google.com/compute/docs/disks#disk-types
	StateDiskType string `yaml:"stateDiskType" validate:"oneof=pd-standard pd-balanced pd-ssd"`
	// description: |
	//   Deploy Persistent Disk CSI driver with on-node encryption. For details see: https://docs.edgeless.systems/constellation/architecture/encrypted-storage
	DeployCSIDriver *bool `yaml:"deployCSIDriver" validate:"required"`
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
	//   Flavor ID (machine type) to use for the VMs. For details see: https://docs.openstack.org/nova/latest/admin/flavors.html
	FlavorID string `yaml:"flavorID" validate:"required"`
	// description: |
	//   Floating IP pool to use for the VMs. For details see: https://docs.openstack.org/ocata/user-guide/cli-manage-ip-addresses.html
	FloatingIPPoolID string `yaml:"floatingIPPoolID" validate:"required"`
	// description: |
	//   Type of a node's state disk. The type influences boot time and I/O performance. Use `openstack volume type list` to get a list of available types.
	StateDiskType string `yaml:"stateDiskType" validate:"required"`
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
	//   AWS Nitro TPM attestation.
	AWSNitroTPM *AWSNitroTPM `yaml:"awsNitroTPM,omitempty" validate:"omitempty,dive"`
	// description: |
	//   Azure SEV-SNP attestation.\nSee our docs for more information on configurable values
	//   TODO(AB#3071): add link after docs are written
	AzureSEVSNP *AzureSEVSNP `yaml:"azureSEVSNP,omitempty" validate:"omitempty,dive"`
	// description: |
	//   Azure TPM attestation (Trusted Launch).
	AzureTrustedLaunch *AzureTrustedLaunch `yaml:"azureTrustedLaunch,omitempty" validate:"omitempty,dive"`
	// description: |
	//   GCP SEV-ES attestation.
	GCPSEVES *GCPSEVES `yaml:"gcpSEVES,omitempty" validate:"omitempty,dive"`
	// description: |
	//   QEMU vTPM attestation.
	QEMUVTPM *QEMUVTPM `yaml:"qemuVTPM,omitempty" validate:"omitempty,dive"`
}

// Default returns a struct with the default config.
func Default() *Config {
	return &Config{
		Version:             Version3,
		Image:               defaultImage,
		Name:                defaultName,
		MicroserviceVersion: compatibility.EnsurePrefixV(constants.VersionInfo()),
		KubernetesVersion:   string(versions.Default),
		StateDiskSizeGB:     30,
		DebugCluster:        toPtr(false),
		Provider: ProviderConfig{
			AWS: &AWSConfig{
				Region:                 "",
				InstanceType:           "m6a.xlarge",
				StateDiskType:          "gp3",
				IAMProfileControlPlane: "",
				IAMProfileWorkerNodes:  "",
			},
			Azure: &AzureConfig{
				SubscriptionID:       "",
				TenantID:             "",
				Location:             "",
				UserAssignedIdentity: "",
				ResourceGroup:        "",
				InstanceType:         "Standard_DC4as_v5",
				StateDiskType:        "Premium_LRS",
				DeployCSIDriver:      toPtr(true),
				SecureBoot:           toPtr(false),
			},
			GCP: &GCPConfig{
				Project:               "",
				Region:                "",
				Zone:                  "",
				ServiceAccountKeyPath: "",
				InstanceType:          "n2d-standard-4",
				StateDiskType:         "pd-ssd",
				DeployCSIDriver:       toPtr(true),
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
		Attestation: AttestationConfig{
			AWSNitroTPM:        &AWSNitroTPM{Measurements: measurements.DefaultsFor(cloudprovider.AWS)},
			AzureSEVSNP:        DefaultForAzureSEVSNP(),
			AzureTrustedLaunch: &AzureTrustedLaunch{Measurements: measurements.DefaultsFor(cloudprovider.Azure)},
			GCPSEVES:           &GCPSEVES{Measurements: measurements.DefaultsFor(cloudprovider.GCP)},
			QEMUVTPM:           &QEMUVTPM{Measurements: measurements.DefaultsFor(cloudprovider.QEMU)},
		},
	}
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
		return nil, fmt.Errorf("could not load config from file %s: %w", name, err)
	}
	return &conf, nil
}

// New creates a new config by:
// 1. Reading config file via provided fileHandler from file with name.
// 2. Read secrets from environment variables.
// 3. Validate config. If `--force` is set the version validation will be disabled and any version combination is allowed.
func New(fileHandler file.Handler, name string, force bool) (*Config, error) {
	// Read config file
	c, err := fromFile(fileHandler, name)
	if err != nil {
		return nil, err
	}

	// Read secrets from env-vars.
	clientSecretValue := os.Getenv(constants.EnvVarAzureClientSecretValue)
	if clientSecretValue != "" && c.Provider.Azure != nil {
		c.Provider.Azure.ClientSecretValue = clientSecretValue
	}

	openstackPassword := os.Getenv(constants.EnvVarOpenStackPassword)
	if openstackPassword != "" && c.Provider.OpenStack != nil {
		c.Provider.OpenStack.Password = openstackPassword
	}

	// Backwards compatibility: configs without the field `microserviceVersion` are valid in version 2.6.
	// In case the field is not set in an old config we prefill it with the default value.
	if c.MicroserviceVersion == "" {
		c.MicroserviceVersion = Default().MicroserviceVersion
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
	if c.Attestation.AWSNitroTPM != nil {
		c.Attestation.AWSNitroTPM.Measurements.CopyFrom(newMeasurements)
	}
	if c.Attestation.AzureSEVSNP != nil {
		c.Attestation.AzureSEVSNP.Measurements.CopyFrom(newMeasurements)
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

// RemoveProviderExcept removes all provider specific configurations, i.e.,
// sets them to nil, except the one specified.
// If an unknown provider is passed, the same configuration is returned.
func (c *Config) RemoveProviderExcept(provider cloudprovider.Provider) {
	currentProviderConfigs := c.Provider
	c.Provider = ProviderConfig{}

	// TODO(AB#2976): Replace attestation replacement
	// with custom function for attestation selection
	currentAttetationConfigs := c.Attestation
	c.Attestation = AttestationConfig{}
	switch provider {
	case cloudprovider.AWS:
		c.Provider.AWS = currentProviderConfigs.AWS
		c.Attestation.AWSNitroTPM = currentAttetationConfigs.AWSNitroTPM
	case cloudprovider.Azure:
		c.Provider.Azure = currentProviderConfigs.Azure
		c.Attestation.AzureSEVSNP = currentAttetationConfigs.AzureSEVSNP
	case cloudprovider.GCP:
		c.Provider.GCP = currentProviderConfigs.GCP
		c.Attestation.GCPSEVES = currentAttetationConfigs.GCPSEVES
	case cloudprovider.OpenStack:
		c.Provider.OpenStack = currentProviderConfigs.OpenStack
		c.Attestation.QEMUVTPM = currentAttetationConfigs.QEMUVTPM
	case cloudprovider.QEMU:
		c.Provider.QEMU = currentProviderConfigs.QEMU
		c.Attestation.QEMUVTPM = currentAttetationConfigs.QEMUVTPM
	default:
		c.Provider = currentProviderConfigs
		c.Attestation = currentAttetationConfigs
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
	if c.Attestation.AWSNitroTPM != nil {
		return c.Attestation.AWSNitroTPM
	}
	if c.Attestation.AzureSEVSNP != nil {
		return c.Attestation.AzureSEVSNP
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

// UpdateMAAURL updates the MAA URL in the config.
func (c *Config) UpdateMAAURL(maaURL string) {
	if c.Attestation.AzureSEVSNP != nil {
		c.Attestation.AzureSEVSNP.FirmwareSignerConfig.MAAURL = maaURL
	}
}

// DeployCSIDriver returns whether the CSI driver should be deployed for a given cloud provider.
func (c *Config) DeployCSIDriver() bool {
	return c.Provider.Azure != nil && c.Provider.Azure.DeployCSIDriver != nil && *c.Provider.Azure.DeployCSIDriver ||
		c.Provider.GCP != nil && c.Provider.GCP.DeployCSIDriver != nil && *c.Provider.GCP.DeployCSIDriver ||
		c.Provider.OpenStack != nil && c.Provider.OpenStack.DeployCSIDriver != nil && *c.Provider.OpenStack.DeployCSIDriver
}

// DeployYawolLoadBalancer returns whether the Yawol load balancer should be deployed.
func (c *Config) DeployYawolLoadBalancer() bool {
	return c.Provider.OpenStack != nil && c.Provider.OpenStack.DeployYawolLoadBalancer != nil && *c.Provider.OpenStack.DeployYawolLoadBalancer
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
	if err := validate.RegisterTranslation("aws_instance_type", trans, registerTranslateAWSInstanceTypeError, translateAWSInstanceTypeError); err != nil {
		return err
	}

	if err := validate.RegisterTranslation("azure_instance_type", trans, registerTranslateAzureInstanceTypeError, c.translateAzureInstanceTypeError); err != nil {
		return err
	}

	if err := validate.RegisterTranslation("gcp_instance_type", trans, registerTranslateGCPInstanceTypeError, translateGCPInstanceTypeError); err != nil {
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

	if err := validate.RegisterTranslation("version_compatibility", trans, registerVersionCompatibilityError, translateVersionCompatibilityError); err != nil {
		return err
	}

	if err := validate.RegisterTranslation("valid_name", trans, registerValidateNameError, c.translateValidateNameError); err != nil {
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
	if err := validate.RegisterValidation("version_compatibility", versionCompatibilityValidator); err != nil {
		return err
	}

	// register custom validator with label aws_instance_type to validate the AWS instance type from config input.
	if err := validate.RegisterValidation("aws_instance_type", validateAWSInstanceType); err != nil {
		return err
	}

	// register custom validator with label azure_instance_type to validate the Azure instance type from config input.
	if err := validate.RegisterValidation("azure_instance_type", c.validateAzureInstanceType); err != nil {
		return err
	}

	// register custom validator with label gcp_instance_type to validate the GCP instance type from config input.
	if err := validate.RegisterValidation("gcp_instance_type", validateGCPInstanceType); err != nil {
		return err
	}

	if err := validate.RegisterValidation("deprecated", warnDeprecated); err != nil {
		return err
	}

	// Register provider validation
	validate.RegisterStructValidation(validateProvider, ProviderConfig{})

	// Register Attestation validation error types
	if err := validate.RegisterTranslation("no_attestation", trans, registerNoAttestationError, translateNoAttestationError); err != nil {
		return err
	}
	if err := validate.RegisterTranslation("more_than_one_attestation", trans, registerMoreThanOneAttestationError, c.translateMoreThanOneAttestationError); err != nil {
		return err
	}

	validate.RegisterStructValidation(validateAttestation, AttestationConfig{})

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

// AzureSEVSNP is the configuration for Azure SEV-SNP attestation.
type AzureSEVSNP struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
	// description: |
	//   Lowest acceptable bootloader version.
	BootloaderVersion uint8 `json:"bootloaderVersion" yaml:"bootloaderVersion"`
	// description: |
	//   Lowest acceptable TEE version.
	TEEVersion uint8 `json:"teeVersion" yaml:"teeVersion"`
	// description: |
	//   Lowest acceptable SEV-SNP version.
	SNPVersion uint8 `json:"snpVersion" yaml:"snpVersion"`
	// description: |
	//   Lowest acceptable microcode version.
	MicrocodeVersion uint8 `json:"microcodeVersion" yaml:"microcodeVersion"`
	// description: |
	//   Configuration for validating the firmware signature.
	FirmwareSignerConfig SNPFirmwareSignerConfig `json:"firmwareSignerConfig" yaml:"firmwareSignerConfig"`
	// description: |
	//   AMD Root Key certificate used to verify the SEV-SNP certificate chain.
	AMDRootKey Certificate `json:"amdRootKey" yaml:"amdRootKey"`
}

// DefaultForAzureSEVSNP returns the default configuration for Azure SEV-SNP attestation.
// Version numbers are hard coded and should be updated with each new release.
// TODO(AB#3042): replace with dynamic lookup for configurable values.
func DefaultForAzureSEVSNP() *AzureSEVSNP {
	return &AzureSEVSNP{
		Measurements:      measurements.DefaultsFor(cloudprovider.Azure),
		BootloaderVersion: 2,
		TEEVersion:        0,
		SNPVersion:        6,
		MicrocodeVersion:  93,
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

// AzureTrustedLaunch is the configuration for Azure Trusted Launch attestation.
type AzureTrustedLaunch struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
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

func toPtr[T any](v T) *T {
	return &v
}

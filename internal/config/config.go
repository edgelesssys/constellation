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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Measurements is a required alias since docgen is not able to work with
// types in other packages.
type Measurements = measurements.M

// Digests is a required alias since docgen is not able to work with
// types in other packages.
type Digests = idkeydigest.IDKeyDigests

const (
	// Version2 is the second version number for Constellation config file.
	Version2 = "v2"

	defaultName = "constell"
)

// Config defines configuration used by CLI.
type Config struct {
	// description: |
	//   Schema version of this configuration file.
	Version string `yaml:"version" validate:"eq=v2"`
	// description: |
	//   Machine image version used to create Constellation nodes.
	Image string `yaml:"image" validate:"required,version_compatibility"`
	// description: |
	//   Name of the cluster.
	Name string `yaml:"name" validate:"valid_name"` // TODO: v2.7: Use "required" validation for name
	// description: |
	//   Size (in GB) of a node's disk to store the non-volatile state.
	StateDiskSizeGB int `yaml:"stateDiskSizeGB" validate:"min=0"`
	// description: |
	//   Kubernetes version to be installed into the cluster.
	KubernetesVersion string `yaml:"kubernetesVersion" validate:"required,supported_k8s_version"`
	// description: |
	//   Microservice version to be installed into the cluster. Setting this value is optional until v2.7. Defaults to the version of the CLI.
	MicroserviceVersion string `yaml:"microserviceVersion" validate:"omitempty,version_compatibility"`
	// description: |
	//   DON'T USE IN PRODUCTION: enable debug mode and use debug images. For usage, see: https://github.com/edgelesssys/constellation/blob/main/debugd/README.md
	DebugCluster *bool `yaml:"debugCluster" validate:"required"`
	// description: |
	//   Supported cloud providers and their specific configurations.
	Provider ProviderConfig `yaml:"provider" validate:"dive"`
	// description: |
	//   Configuration to apply during constellation upgrade.
	// examples:
	//   - value: 'UpgradeConfig{ Image: "", Measurements: Measurements{} }'
	Upgrade UpgradeConfig `yaml:"upgrade,omitempty" validate:"required"`
}

// UpgradeConfig defines configuration used during constellation upgrade.
type UpgradeConfig struct {
	// description: |
	//   Updated Constellation machine image to install on all nodes.
	Image string `yaml:"image"`
	// description: |
	//   Measurements of the updated image.
	Measurements Measurements `yaml:"measurements"`
	// description: |
	//   temporary field for upgrade migration
	//   TODO(AB#2654): Remove with refactoring upgrade plan command
	CSP cloudprovider.Provider `yaml:"csp"`
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
	// description: |
	//   Expected VM measurements.
	Measurements Measurements `yaml:"measurements" validate:"required,no_placeholders"`
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
	//   Use Confidential VMs. Always needs to be true.
	ConfidentialVM *bool `yaml:"confidentialVM" validate:"required"`
	// description: |
	//   Enable secure boot for VMs. If enabled, the OS image has to include a virtual machine guest state (VMGS) blob.
	SecureBoot *bool `yaml:"secureBoot" validate:"required"`
	// description: |
	//   List of accepted values for the field 'idkeydigest' in the AMD SEV-SNP attestation report. Only usable with ConfidentialVMs. See 4.6 and 7.3 in: https://www.amd.com/system/files/TechDocs/56860.pdf
	IDKeyDigest Digests `yaml:"idKeyDigest" validate:"required_if=EnforceIdKeyDigest true,omitempty"`
	// description: |
	//   Enforce the specified idKeyDigest value during remote attestation.
	EnforceIDKeyDigest *bool `yaml:"enforceIdKeyDigest" validate:"required"`
	// description: |
	//   Expected confidential VM measurements.
	Measurements Measurements `yaml:"measurements" validate:"required,no_placeholders"`
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
	// description: |
	//   Expected confidential VM measurements.
	Measurements Measurements `yaml:"measurements" validate:"required,no_placeholders"`
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
	//   If enabled, downloads OS image directly from source URL to OpenStack. Otherwise, downloads image to local machine and uploads to OpenStack.
	DirectDownload *bool `yaml:"directDownload" validate:"required"`
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
	// description: |
	//   Measurement used to enable measured boot.
	Measurements Measurements `yaml:"measurements" validate:"required,no_placeholders"`
}

// Default returns a struct with the default config.
func Default() *Config {
	return &Config{
		Version:             Version2,
		Image:               defaultImage,
		Name:                defaultName,
		MicroserviceVersion: compatibility.EnsurePrefixV(constants.VersionInfo),
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
				Measurements:           measurements.DefaultsFor(cloudprovider.AWS),
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
				IDKeyDigest:          idkeydigest.DefaultsFor(cloudprovider.Azure),
				EnforceIDKeyDigest:   toPtr(true),
				ConfidentialVM:       toPtr(true),
				SecureBoot:           toPtr(false),
				Measurements:         measurements.DefaultsFor(cloudprovider.Azure),
			},
			GCP: &GCPConfig{
				Project:               "",
				Region:                "",
				Zone:                  "",
				ServiceAccountKeyPath: "",
				InstanceType:          "n2d-standard-4",
				StateDiskType:         "pd-ssd",
				DeployCSIDriver:       toPtr(true),
				Measurements:          measurements.DefaultsFor(cloudprovider.GCP),
			},
			OpenStack: &OpenStackConfig{
				DirectDownload: toPtr(true),
			},
			QEMU: &QEMUConfig{
				ImageFormat:           "raw",
				VCPUs:                 2,
				Memory:                2048,
				MetadataAPIImage:      versions.QEMUMetadataImage,
				LibvirtURI:            "",
				LibvirtContainerImage: versions.LibvirtImage,
				NVRAM:                 "production",
				Measurements:          measurements.DefaultsFor(cloudprovider.QEMU),
			},
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

	// Backwards compatibility: configs without the field `microserviceVersion` are valid in version 2.6.
	// In case the field is not set in an old config we prefil it with the default value.
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
	case cloudprovider.QEMU:
		return c.Provider.QEMU != nil
	}
	return false
}

// UpdateMeasurements overwrites measurements in config with the provided ones.
func (c *Config) UpdateMeasurements(newMeasurements Measurements) {
	if c.Provider.AWS != nil {
		c.Provider.AWS.Measurements.CopyFrom(newMeasurements)
	}
	if c.Provider.Azure != nil {
		c.Provider.Azure.Measurements.CopyFrom(newMeasurements)
	}
	if c.Provider.GCP != nil {
		c.Provider.GCP.Measurements.CopyFrom(newMeasurements)
	}
	if c.Provider.QEMU != nil {
		c.Provider.QEMU.Measurements.CopyFrom(newMeasurements)
	}
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
}

// IsAzureNonCVM checks whether the chosen provider is azure and confidential VMs are disabled.
func (c *Config) IsAzureNonCVM() bool {
	return c.Provider.Azure != nil && c.Provider.Azure.ConfidentialVM != nil && !*c.Provider.Azure.ConfidentialVM
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

// GetMeasurements returns the configured measurements or nil if no provder is set.
func (c *Config) GetMeasurements() measurements.M {
	if c.Provider.AWS != nil {
		return c.Provider.AWS.Measurements
	}
	if c.Provider.Azure != nil {
		return c.Provider.Azure.Measurements
	}
	if c.Provider.GCP != nil {
		return c.Provider.GCP.Measurements
	}
	if c.Provider.QEMU != nil {
		return c.Provider.QEMU.Measurements
	}
	return nil
}

// EnforcesIDKeyDigest checks whether ID Key Digest should be enforced for respective cloud provider.
func (c *Config) EnforcesIDKeyDigest() bool {
	return c.Provider.Azure != nil && c.Provider.Azure.EnforceIDKeyDigest != nil && *c.Provider.Azure.EnforceIDKeyDigest
}

// EnforcedPCRs returns the list of enforced PCRs for the configured cloud provider.
func (c *Config) EnforcedPCRs() []uint32 {
	provider := c.GetProvider()
	switch provider {
	case cloudprovider.AWS:
		return c.Provider.AWS.Measurements.GetEnforced()
	case cloudprovider.Azure:
		return c.Provider.Azure.Measurements.GetEnforced()
	case cloudprovider.GCP:
		return c.Provider.GCP.Measurements.GetEnforced()
	case cloudprovider.QEMU:
		return c.Provider.QEMU.Measurements.GetEnforced()
	default:
		return nil
	}
}

// IDKeyDigests returns the ID Key Digests for the configured cloud provider.
func (c *Config) IDKeyDigests() idkeydigest.IDKeyDigests {
	if c.Provider.Azure != nil {
		return c.Provider.Azure.IDKeyDigest
	}
	return nil
}

// DeployCSIDriver returns whether the CSI driver should be deployed for a given cloud provider.
func (c *Config) DeployCSIDriver() bool {
	return c.Provider.Azure != nil && c.Provider.Azure.DeployCSIDriver != nil && *c.Provider.Azure.DeployCSIDriver ||
		c.Provider.GCP != nil && c.Provider.GCP.DeployCSIDriver != nil && *c.Provider.GCP.DeployCSIDriver
}

// Validate checks the config values and returns validation errors.
func (c *Config) Validate(force bool) error {
	trans := ut.New(en.New()).GetFallback()
	validate := validator.New()
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		return err
	}

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
	if err := validate.RegisterValidation("azure_instance_type", validateAzureInstanceType); err != nil {
		return err
	}

	// register custom validator with label gcp_instance_type to validate the GCP instance type from config input.
	if err := validate.RegisterValidation("gcp_instance_type", validateGCPInstanceType); err != nil {
		return err
	}

	// Register provider validation
	validate.RegisterStructValidation(validateProvider, ProviderConfig{})

	// register custom validator that prints a deprecation warning.
	validate.RegisterStructValidation(validateUpgradeConfig, UpgradeConfig{})

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

func toPtr[T any](v T) *T {
	return &v
}

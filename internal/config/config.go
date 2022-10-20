/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// This binary can be build from siderolabs/talos projects. Located at:
// https://github.com/siderolabs/talos/tree/master/hack/docgen
//
//go:generate docgen ./config.go ./config_doc.go Configuration
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config/instancetypes"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

const (
	Version1 = "v1"
)

var (
	azureReleaseImageRegex = regexp.MustCompile(`^\/CommunityGalleries\/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df\/Images\/constellation\/Versions\/[\d]+.[\d]+.[\d]+$`)
	gcpReleaseImageRegex   = regexp.MustCompile(`^projects\/constellation-images\/global\/images\/constellation-v[\d]+-[\d]+-[\d]+$`)
)

// Config defines configuration used by CLI.
type Config struct {
	// description: |
	//   Schema version of this configuration file.
	Version string `yaml:"version" validate:"eq=v1"`
	// description: |
	//   Size (in GB) of a node's disk to store the non-volatile state.
	StateDiskSizeGB int `yaml:"stateDiskSizeGB" validate:"min=0"`
	// description: |
	//   Kubernetes version to be installed in the cluster.
	KubernetesVersion string `yaml:"kubernetesVersion" validate:"supported_k8s_version"`
	// description: |
	//   DON'T USE IN PRODUCTION: enable debug mode and use debug images. For usage, see: https://github.com/edgelesssys/constellation/blob/main/debugd/README.md
	DebugCluster *bool `yaml:"debugCluster" validate:"required"`
	// description: |
	//   Supported cloud providers and their specific configurations.
	Provider ProviderConfig `yaml:"provider" validate:"dive"`
	// description: |
	//   Create SSH users on Constellation nodes.
	// examples:
	//   - value: '[]UserKey{ { Username:  "Alice", PublicKey: "ssh-rsa AAAAB3NzaC...5QXHKW1rufgtJeSeJ8= alice@domain.com" } }'
	SSHUsers []UserKey `yaml:"sshUsers,omitempty" validate:"dive"`
	// description: |
	//   Configuration to apply during constellation upgrade.
	// examples:
	//   - value: 'UpgradeConfig{ Image: "", Measurements: Measurements{} }'
	Upgrade UpgradeConfig `yaml:"upgrade,omitempty"`
}

// UpgradeConfig defines configuration used during constellation upgrade.
type UpgradeConfig struct {
	// description: |
	//   Updated machine image to install on all nodes.
	Image string `yaml:"image"`
	// description: |
	//   Measurements of the updated image.
	Measurements Measurements `yaml:"measurements"`
}

// UserKey describes a user that should be created with corresponding public SSH key.
type UserKey struct {
	// description: |
	//   Username of new SSH user.
	Username string `yaml:"username" validate:"required"`
	// description: |
	//   Public key of new SSH user.
	PublicKey string `yaml:"publicKey" validate:"required"`
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
	//   AMI ID of the machine image used to create Constellation nodes.
	Image string `yaml:"image" validate:"required"`
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
	Measurements Measurements `yaml:"measurements"`
	// description: |
	//   List of values that should be enforced to be equal to the ones from the measurement list. Any non-equal values not in this list will only result in a warning.
	EnforcedMeasurements []uint32 `yaml:"enforcedMeasurements"`
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
	//    Client secret value of the Active Directory app registration credentials.
	ClientSecretValue string `yaml:"clientSecretValue" validate:"required"`
	// description: |
	//   Machine image used to create Constellation nodes.
	Image string `yaml:"image" validate:"required"`
	// description: |
	//   VM instance type to use for Constellation nodes.
	InstanceType string `yaml:"instanceType" validate:"azure_instance_type"`
	// description: |
	//   Type of a node's state disk. The type influences boot time and I/O performance. See: https://docs.microsoft.com/en-us/azure/virtual-machines/disks-types#disk-type-comparison
	StateDiskType string `yaml:"stateDiskType" validate:"oneof=Premium_LRS Premium_ZRS Standard_LRS StandardSSD_LRS StandardSSD_ZRS"`
	// description: |
	//   Expected confidential VM measurements.
	Measurements Measurements `yaml:"measurements"`
	// description: |
	//   List of values that should be enforced to be equal to the ones from the measurement list. Any non-equal values not in this list will only result in a warning.
	EnforcedMeasurements []uint32 `yaml:"enforcedMeasurements"`
	// description: |
	//   Expected value for the field 'idkeydigest' in the AMD SEV-SNP attestation report. Only usable with ConfidentialVMs. See 4.6 and 7.3 in: https://www.amd.com/system/files/TechDocs/56860.pdf
	IDKeyDigest string `yaml:"idKeyDigest" validate:"required_if=EnforceIdKeyDigest true,omitempty,hexadecimal,len=96"`
	// description: |
	//   Enforce the specified idKeyDigest value during remote attestation.
	EnforceIDKeyDigest *bool `yaml:"enforceIdKeyDigest" validate:"required"`
	// description: |
	//   Use Confidential VMs. If set to false, Trusted Launch VMs are used instead. See: https://docs.microsoft.com/en-us/azure/confidential-computing/confidential-vm-overview
	ConfidentialVM *bool `yaml:"confidentialVM" validate:"required"`
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
	//   Machine image used to create Constellation nodes.
	Image string `yaml:"image" validate:"required"`
	// description: |
	//   VM instance type to use for Constellation nodes.
	InstanceType string `yaml:"instanceType" validate:"gcp_instance_type"`
	// description: |
	//   Type of a node's state disk. The type influences boot time and I/O performance. See: https://cloud.google.com/compute/docs/disks#disk-types
	StateDiskType string `yaml:"stateDiskType" validate:"oneof=pd-standard pd-balanced pd-ssd"`
	// description: |
	//   Expected confidential VM measurements.
	Measurements Measurements `yaml:"measurements"`
	// description: |
	//   List of values that should be enforced to be equal to the ones from the measurement list. Any non-equal values not in this list will only result in a warning.
	EnforcedMeasurements []uint32 `yaml:"enforcedMeasurements"`
}

type QEMUConfig struct {
	// description: |
	//   Path to the image to use for the VMs.
	Image string `yaml:"image" validate:"required"`
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
	//   Measurement used to enable measured boot.
	Measurements Measurements `yaml:"measurements"`
	// description: |
	//   List of values that should be enforced to be equal to the ones from the measurement list. Any non-equal values not in this list will only result in a warning.
	EnforcedMeasurements []uint32 `yaml:"enforcedMeasurements"`
}

// Default returns a struct with the default config.
func Default() *Config {
	return &Config{
		Version:         Version1,
		StateDiskSizeGB: 30,
		DebugCluster:    func() *bool { b := false; return &b }(),
		Provider: ProviderConfig{
			AWS: &AWSConfig{
				Region:                 "",
				Image:                  "",
				InstanceType:           "m6a.xlarge",
				StateDiskType:          "gp3",
				IAMProfileControlPlane: "",
				IAMProfileWorkerNodes:  "",
				Measurements:           copyPCRMap(awsPCRs),
				EnforcedMeasurements:   []uint32{}, // TODO: add default values
			},
			Azure: &AzureConfig{
				SubscriptionID:       "",
				TenantID:             "",
				Location:             "",
				UserAssignedIdentity: "",
				ResourceGroup:        "",
				Image:                DefaultImageAzure,
				InstanceType:         "Standard_DC4as_v5",
				StateDiskType:        "Premium_LRS",
				Measurements:         copyPCRMap(azurePCRs),
				EnforcedMeasurements: []uint32{4, 8, 9, 11, 12},
				IDKeyDigest:          "57486a447ec0f1958002a22a06b7673b9fd27d11e1c6527498056054c5fa92d23c50f9de44072760fe2b6fb89740b696",
				EnforceIDKeyDigest:   func() *bool { b := true; return &b }(),
				ConfidentialVM:       func() *bool { b := true; return &b }(),
			},
			GCP: &GCPConfig{
				Project:               "",
				Region:                "",
				Zone:                  "",
				Image:                 DefaultImageGCP,
				InstanceType:          "n2d-standard-4",
				StateDiskType:         "pd-ssd",
				ServiceAccountKeyPath: "",
				Measurements:          copyPCRMap(gcpPCRs),
				EnforcedMeasurements:  []uint32{0, 4, 8, 9, 11, 12},
			},
			QEMU: &QEMUConfig{
				ImageFormat:           "qcow2",
				VCPUs:                 2,
				Memory:                2048,
				MetadataAPIImage:      versions.QEMUMetadataImage,
				LibvirtURI:            "",
				LibvirtContainerImage: versions.LibvirtImage,
				Measurements:          copyPCRMap(qemuPCRs),
				EnforcedMeasurements:  []uint32{11, 12},
			},
		},
		KubernetesVersion: string(versions.Default),
	}
}

func validateK8sVersion(fl validator.FieldLevel) bool {
	return versions.IsSupportedK8sVersion(fl.Field().String())
}

func validateAWSInstanceType(fl validator.FieldLevel) bool {
	return validInstanceTypeForProvider(fl.Field().String(), false, cloudprovider.AWS)
}

func validateAzureInstanceType(fl validator.FieldLevel) bool {
	azureConfig := fl.Parent().Interface().(AzureConfig)
	var acceptNonCVM bool
	if azureConfig.ConfidentialVM != nil {
		// This is the inverse of the config value (acceptNonCVMs is true if confidentialVM is false).
		// We could make the validator the other way around, but this should be an explicit bypass rather than checking if CVMs are "allowed".
		acceptNonCVM = !*azureConfig.ConfidentialVM
	}
	return validInstanceTypeForProvider(fl.Field().String(), acceptNonCVM, cloudprovider.Azure)
}

func validateGCPInstanceType(fl validator.FieldLevel) bool {
	return validInstanceTypeForProvider(fl.Field().String(), false, cloudprovider.GCP)
}

// validateProvider checks if zero or more than one providers are defined in the config.
func validateProvider(sl validator.StructLevel) {
	provider := sl.Current().Interface().(ProviderConfig)
	providerCount := 0

	if provider.AWS != nil {
		providerCount++
	}
	if provider.Azure != nil {
		providerCount++
	}
	if provider.GCP != nil {
		providerCount++
	}
	if provider.QEMU != nil {
		providerCount++
	}

	if providerCount < 1 {
		sl.ReportError(provider, "Provider", "Provider", "no_provider", "")
	} else if providerCount > 1 {
		sl.ReportError(provider, "Provider", "Provider", "more_than_one_provider", "")
	}
}

// Validate checks the config values and returns validation error messages.
// The function only returns an error if the validation itself fails.
func (c *Config) Validate() ([]string, error) {
	trans := ut.New(en.New()).GetFallback()
	validate := validator.New()
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		return nil, err
	}

	// Register AWS, Azure & GCP InstanceType validation error types
	if err := validate.RegisterTranslation("aws_instance_type", trans, registerTranslateAWSInstanceTypeError, translateAWSInstanceTypeError); err != nil {
		return nil, err
	}

	if err := validate.RegisterTranslation("azure_instance_type", trans, registerTranslateAzureInstanceTypeError, c.translateAzureInstanceTypeError); err != nil {
		return nil, err
	}

	if err := validate.RegisterTranslation("gcp_instance_type", trans, registerTranslateGCPInstanceTypeError, translateGCPInstanceTypeError); err != nil {
		return nil, err
	}

	// Register Provider validation error types
	if err := validate.RegisterTranslation("no_provider", trans, registerNoProviderError, translateNoProviderError); err != nil {
		return nil, err
	}

	if err := validate.RegisterTranslation("more_than_one_provider", trans, registerMoreThanOneProviderError, c.translateMoreThanOneProviderError); err != nil {
		return nil, err
	}

	// register custom validator with label supported_k8s_version to validate version based on available versionConfigs.
	if err := validate.RegisterValidation("supported_k8s_version", validateK8sVersion); err != nil {
		return nil, err
	}

	// register custom validator with label aws_instance_type to validate the AWS instance type from config input.
	if err := validate.RegisterValidation("aws_instance_type", validateAWSInstanceType); err != nil {
		return nil, err
	}

	// register custom validator with label azure_instance_type to validate the Azure instance type from config input.
	if err := validate.RegisterValidation("azure_instance_type", validateAzureInstanceType); err != nil {
		return nil, err
	}

	// register custom validator with label gcp_instance_type to validate the GCP instance type from config input.
	if err := validate.RegisterValidation("gcp_instance_type", validateGCPInstanceType); err != nil {
		return nil, err
	}

	// Register provider validation
	validate.RegisterStructValidation(validateProvider, ProviderConfig{})

	err := validate.Struct(c)
	if err == nil {
		return nil, nil
	}

	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		return nil, err
	}

	var msgs []string
	for _, e := range errs {
		msgs = append(msgs, e.Translate(trans))
	}
	return msgs, nil
}

// Validation translation functions for Azure & GCP instance type errors.
func registerTranslateAzureInstanceTypeError(ut ut.Translator) error {
	return ut.Add("azure_instance_type", "{0} must be one of {1}", true)
}

func (c *Config) translateAzureInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	// Suggest trusted launch VMs if confidential VMs have been specifically disabled
	var t string
	if c.Provider.Azure != nil && c.Provider.Azure.ConfidentialVM != nil && !*c.Provider.Azure.ConfidentialVM {
		t, _ = ut.T("azure_instance_type", fe.Field(), fmt.Sprintf("%v", instancetypes.AzureTrustedLaunchInstanceTypes))
	} else {
		t, _ = ut.T("azure_instance_type", fe.Field(), fmt.Sprintf("%v", instancetypes.AzureCVMInstanceTypes))
	}

	return t
}

func registerTranslateAWSInstanceTypeError(ut ut.Translator) error {
	return ut.Add("aws_instance_type", fmt.Sprintf("{0} must be an instance from one of the following families types with size xlarge or higher: %v", instancetypes.AWSSupportedInstanceFamilies), true)
}

func translateAWSInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("aws_instance_type", fe.Field())

	return t
}

func registerTranslateGCPInstanceTypeError(ut ut.Translator) error {
	return ut.Add("gcp_instance_type", fmt.Sprintf("{0} must be one of %v", instancetypes.GCPInstanceTypes), true)
}

func translateGCPInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("gcp_instance_type", fe.Field())

	return t
}

// Validation translation functions for Provider errors.
func registerNoProviderError(ut ut.Translator) error {
	return ut.Add("no_provider", "{0}: No provider has been defined (requires either Azure, GCP or QEMU)", true)
}

func translateNoProviderError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("no_provider", fe.Field())

	return t
}

func registerMoreThanOneProviderError(ut ut.Translator) error {
	return ut.Add("more_than_one_provider", "{0}: Only one provider can be defined ({1} are defined)", true)
}

func (c *Config) translateMoreThanOneProviderError(ut ut.Translator, fe validator.FieldError) string {
	definedProviders := make([]string, 0)

	// c.Provider should not be nil as Provider would need to be defined for the validation to fail in this place.
	if c.Provider.AWS != nil {
		definedProviders = append(definedProviders, "AWS")
	}
	if c.Provider.Azure != nil {
		definedProviders = append(definedProviders, "Azure")
	}
	if c.Provider.GCP != nil {
		definedProviders = append(definedProviders, "GCP")
	}
	if c.Provider.QEMU != nil {
		definedProviders = append(definedProviders, "QEMU")
	}

	// Show single string if only one other provider is defined, show list with brackets if multiple are defined.
	t, _ := ut.T("more_than_one_provider", fe.Field(), strings.Join(definedProviders, ", "))

	return t
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

// Image returns OS image for the configured cloud provider.
// If multiple cloud providers are configured (which is not supported)
// only a single image is returned.
func (c *Config) Image() string {
	if c.HasProvider(cloudprovider.AWS) {
		return c.Provider.AWS.Image
	}
	if c.HasProvider(cloudprovider.Azure) {
		return c.Provider.Azure.Image
	}
	if c.HasProvider(cloudprovider.GCP) {
		return c.Provider.GCP.Image
	}
	return ""
}

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
	case cloudprovider.QEMU:
		c.Provider.QEMU = currentProviderConfigs.QEMU
	default:
		c.Provider = currentProviderConfigs
	}
}

// IsDebugImage checks whether image name looks like a release image, if not it is
// probably a debug image. In the end we do not if bootstrapper or debugd
// was put inside an image just by looking at its name.
func (c *Config) IsDebugImage() bool {
	switch {
	case c.Provider.AWS != nil:
		// TODO: Add proper image name validation for AWS when we are closer to release.
		return true
	case c.Provider.Azure != nil:
		return !azureReleaseImageRegex.MatchString(c.Provider.Azure.Image)
	case c.Provider.GCP != nil:
		return !gcpReleaseImageRegex.MatchString(c.Provider.GCP.Image)
	default:
		return false
	}
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
	if c.Provider.QEMU != nil {
		return cloudprovider.QEMU
	}
	return cloudprovider.Unknown
}

// IsAzureNonCVM checks whether the chosen provider is azure and confidential VMs are disabled.
func (c *Config) IsAzureNonCVM() bool {
	return c.Provider.Azure != nil && c.Provider.Azure.ConfidentialVM != nil && !*c.Provider.Azure.ConfidentialVM
}

func (c *Config) EnforcesIDKeyDigest() bool {
	return c.Provider.Azure != nil && c.Provider.Azure.EnforceIDKeyDigest != nil && *c.Provider.Azure.EnforceIDKeyDigest
}

// FromFile returns config file with `name` read from `fileHandler` by parsing
// it as YAML.
func FromFile(fileHandler file.Handler, name string) (*Config, error) {
	var conf Config
	if err := fileHandler.ReadYAMLStrict(name, &conf); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("unable to find %s - use `constellation config generate` to generate it first", name)
		}
		return nil, fmt.Errorf("could not load config from file %s: %w", name, err)
	}
	return &conf, nil
}

func copyPCRMap(m map[uint32][]byte) map[uint32][]byte {
	res := make(Measurements)
	res.CopyFrom(m)
	return res
}

func validInstanceTypeForProvider(insType string, acceptNonCVM bool, provider cloudprovider.Provider) bool {
	switch provider {
	case cloudprovider.AWS:
		return checkIfAWSInstanceTypeIsValid(insType)
	case cloudprovider.Azure:
		if acceptNonCVM {
			for _, instanceType := range instancetypes.AzureTrustedLaunchInstanceTypes {
				if insType == instanceType {
					return true
				}
			}
		} else {
			for _, instanceType := range instancetypes.AzureCVMInstanceTypes {
				if insType == instanceType {
					return true
				}
			}
		}
		return false
	case cloudprovider.GCP:
		for _, instanceType := range instancetypes.GCPInstanceTypes {
			if insType == instanceType {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// checkIfAWSInstanceTypeIsValid checks if an AWS instance type passed as user input is in one of the instance families supporting NitroTPM.
func checkIfAWSInstanceTypeIsValid(userInput string) bool {
	// Check if user or code does anything weird and tries to pass multiple strings as one
	if strings.Contains(userInput, " ") {
		return false
	}
	if strings.Contains(userInput, ",") {
		return false
	}
	if strings.Contains(userInput, ";") {
		return false
	}

	splitInstanceType := strings.Split(userInput, ".")

	if len(splitInstanceType) != 2 {
		return false
	}

	userDefinedFamily := splitInstanceType[0]
	userDefinedSize := splitInstanceType[1]

	// Check if instace type has at least 4 vCPUs (= contains "xlarge" in its name)
	hasEnoughVCPUs := strings.Contains(userDefinedSize, "xlarge")
	if !hasEnoughVCPUs {
		return false
	}

	// Now check if the user input is a supported family
	// Note that we cannot directly use the family split from the Graviton check above, as some instances are directly specified by their full name and not just the family in general
	for _, supportedFamily := range instancetypes.AWSSupportedInstanceFamilies {
		supportedFamilyLowercase := strings.ToLower(supportedFamily)
		if userDefinedFamily == supportedFamilyLowercase {
			return true
		}
	}

	return false
}

// IsDebugCluster checks whether the cluster is configured as a debug cluster.
func (c *Config) IsDebugCluster() bool {
	if c.DebugCluster != nil && *c.DebugCluster {
		return true
	}
	return false
}

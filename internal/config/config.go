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

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config/instancetypes"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/versions"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

const (
	Version1 = "v1"
)

// Config defines configuration used by CLI.
type Config struct {
	// description: |
	//   Schema version of this configuration file.
	Version string `yaml:"version" validate:"eq=v1"`
	// description: |
	//   Minimum number of worker nodes in autoscaling group.
	AutoscalingNodeGroupMin int `yaml:"autoscalingNodeGroupMin" validate:"min=0"`
	// description: |
	//   Maximum number of worker nodes in autoscaling group.
	AutoscalingNodeGroupMax int `yaml:"autoscalingNodeGroupMax" validate:"gtefield=AutoscalingNodeGroupMin"`
	// description: |
	//   Size (in GB) of a node's disk to store the non-volatile state.
	StateDiskSizeGB int `yaml:"stateDiskSizeGB" validate:"min=0"`
	// description: |
	//   Ingress firewall rules for node network.
	IngressFirewall Firewall `yaml:"ingressFirewall,omitempty" validate:"dive"`
	// description: |
	//   Egress firewall rules for node network.
	// examples:
	//   - value: 'Firewall{
	//     {
	//       Name: "rule#1",
	//       Description: "the first rule",
	//       Protocol: "tcp",
	//       IPRange: "0.0.0.0/0",
	//       FromPort: 443,
	//       ToPort: 443,
	//     },
	//   }'
	EgressFirewall Firewall `yaml:"egressFirewall,omitempty" validate:"dive"`
	// description: |
	//   Supported cloud providers and their specific configurations.
	Provider ProviderConfig `yaml:"provider" validate:"dive"`
	// description: |
	//   Create SSH users on Constellation nodes.
	// examples:
	//   - value: '[]UserKey{ { Username:  "Alice", PublicKey: "ssh-rsa AAAAB3NzaC...5QXHKW1rufgtJeSeJ8= alice@domain.com" } }'
	SSHUsers []UserKey `yaml:"sshUsers,omitempty" validate:"dive"`
	// description: |
	//   Kubernetes version installed in the cluster.
	KubernetesVersion string `yaml:"kubernetesVersion" validate:"supported_k8s_version"`
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

type FirewallRule struct {
	// description: |
	//   Name of rule.
	Name string `yaml:"name" validate:"required"`
	// description: |
	//   Description for rule.
	Description string `yaml:"description"`
	// description: |
	//   Protocol, such as 'udp' or 'tcp'.
	Protocol string `yaml:"protocol" validate:"required"`
	// description: |
	//   CIDR range for which this rule is applied.
	IPRange string `yaml:"iprange" validate:"required"`
	// description: |
	//   Start port of a range.
	FromPort int `yaml:"fromport" validate:"min=0,max=65535"`
	// description: |
	//   End port of a range, or 0 if a single port is given by fromport.
	ToPort int `yaml:"toport" validate:"omitempty,gtefield=FromPort,max=65535"`
}

type Firewall []FirewallRule

// ProviderConfig are cloud-provider specific configuration values used by the CLI.
// Fields should remain pointer-types so custom specific configs can nil them
// if not required.
type ProviderConfig struct {
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
	//   Machine image used to create Constellation nodes.
	Image string `yaml:"image" validate:"required"`
	// description: |
	//   Virtual machine instance type to use for Constellation nodes.
	InstanceType string `yaml:"instanceType" validate:"azure_instance_type"`
	// description: |
	//   Type of a node's state disk. The type influences boot time and I/O performance. See: https://docs.microsoft.com/en-us/azure/virtual-machines/disks-types#disk-type-comparison
	StateDiskType string `yaml:"stateDiskType" validate:"oneof=Premium_LRS Premium_ZRS Standard_LRS StandardSSD_LRS StandardSSD_ZRS"`
	// description: |
	//   Resource group to use.
	ResourceGroup string `yaml:"resourceGroup" validate:"required"`
	// description: |
	//   Authorize spawned VMs to access Azure API.
	UserAssignedIdentity string `yaml:"userAssignedIdentity" validate:"required"`
	// description: |
	//    Application client ID of the Active Directory app registration.
	AppClientID string `yaml:"appClientID" validate:"required"`
	// description: |
	//    Client secret value of the Active Directory app registration credentials.
	ClientSecretValue string `yaml:"clientSecretValue" validate:"required"`
	// description: |
	//   Expected confidential VM measurements.
	Measurements Measurements `yaml:"measurements"`
	// description: |
	//   List of values that should be enforced to be equal to the ones from the measurement list. Any non-equal values not in this list will only result in a warning.
	EnforcedMeasurements []uint32 `yaml:"enforcedMeasurements"`
	// description: |
	//   Expected value for the field 'idkeydigest' in the AMD SEV-SNP attestation report. Only usable with ConfidentialVMs. See 4.6 and 7.3 in: https://www.amd.com/system/files/TechDocs/56860.pdf
	IdKeyDigest string `yaml:"idKeyDigest" validate:"required_if=EnforceIdKeyDigest true,omitempty,hexadecimal,len=96"`
	// description: |
	//   Enforce the specified idKeyDigest value during remote attestation.
	EnforceIdKeyDigest *bool `yaml:"enforceIdKeyDigest" validate:"required"`
	// description: |
	//   Use VMs with security type Confidential VM. If set to false, Trusted Launch VMs will be used instead. See: https://docs.microsoft.com/en-us/azure/confidential-computing/confidential-vm-overview
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
	//   Machine image used to create Constellation nodes.
	Image string `yaml:"image" validate:"required"`
	// description: |
	//   Virtual machine instance type to use for Constellation nodes.
	InstanceType string `yaml:"instanceType" validate:"gcp_instance_type"`
	// description: |
	//   Type of a node's state disk. The type influences boot time and I/O performance. See: https://cloud.google.com/compute/docs/disks#disk-types
	StateDiskType string `yaml:"stateDiskType" validate:"oneof=pd-standard pd-balanced pd-ssd"`
	// description: |
	//   Path of service account key file. For needed service account roles, see https://constellation-docs.edgeless.systems/constellation/getting-started/install#authorization
	ServiceAccountKeyPath string `yaml:"serviceAccountKeyPath"`
	// description: |
	//   Expected confidential VM measurements.
	Measurements Measurements `yaml:"measurements"`
	// description: |
	//   List of values that should be enforced to be equal to the ones from the measurement list. Any non-equal values not in this list will only result in a warning.
	EnforcedMeasurements []uint32 `yaml:"enforcedMeasurements"`
}

type QEMUConfig struct {
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
		Version:                 Version1,
		AutoscalingNodeGroupMin: 1,
		AutoscalingNodeGroupMax: 10,
		StateDiskSizeGB:         30,
		IngressFirewall: Firewall{
			{
				Name:        "bootstrapper",
				Description: "bootstrapper default port",
				Protocol:    "tcp",
				IPRange:     "0.0.0.0/0",
				FromPort:    constants.BootstrapperPort,
			},
			{
				Name:        "ssh",
				Description: "SSH",
				Protocol:    "tcp",
				IPRange:     "0.0.0.0/0",
				FromPort:    constants.SSHPort,
			},
			{
				Name:        "nodeport",
				Description: "NodePort",
				Protocol:    "tcp",
				IPRange:     "0.0.0.0/0",
				FromPort:    constants.NodePortFrom,
				ToPort:      constants.NodePortTo,
			},
			{
				Name:        "kubernetes",
				Description: "Kubernetes",
				Protocol:    "tcp",
				IPRange:     "0.0.0.0/0",
				FromPort:    constants.KubernetesPort,
			},
		},
		Provider: ProviderConfig{
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
				EnforcedMeasurements: []uint32{8, 9, 11, 12},
				IdKeyDigest:          "57486a447ec0f1958002a22a06b7673b9fd27d11e1c6527498056054c5fa92d23c50f9de44072760fe2b6fb89740b696",
				EnforceIdKeyDigest:   func() *bool { b := true; return &b }(),
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
				EnforcedMeasurements:  []uint32{0, 8, 9, 11, 12},
			},
			QEMU: &QEMUConfig{
				Measurements:         copyPCRMap(qemuPCRs),
				EnforcedMeasurements: []uint32{11, 12},
			},
		},
		KubernetesVersion: string(versions.Latest),
	}
}

func validateK8sVersion(fl validator.FieldLevel) bool {
	return versions.IsSupportedK8sVersion(fl.Field().String())
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

// Validate checks the config values and returns validation error messages.
// The function only returns an error if the validation itself fails.
func (c *Config) Validate() ([]string, error) {
	trans := ut.New(en.New()).GetFallback()
	validate := validator.New()
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		return nil, err
	}

	// Register Azure & GCP InstanceType validation error types
	if err := validate.RegisterTranslation("azure_instance_type", trans, c.registerTranslateAzureInstanceTypeError, translateAzureInstanceTypeError); err != nil {
		return nil, err
	}

	if err := validate.RegisterTranslation("gcp_instance_type", trans, registerTranslateGCPInstanceTypeError, translateGCPInstanceTypeError); err != nil {
		return nil, err
	}

	// register custom validator with label supported_k8s_version to validate version based on available versionConfigs.
	if err := validate.RegisterValidation("supported_k8s_version", validateK8sVersion); err != nil {
		return nil, err
	}

	// register custom validator with label azure_instance_type to validate version based on available versionConfigs.
	if err := validate.RegisterValidation("azure_instance_type", validateAzureInstanceType); err != nil {
		return nil, err
	}

	// register custom validator with label azure_instance_type to validate version based on available versionConfigs.
	if err := validate.RegisterValidation("gcp_instance_type", validateGCPInstanceType); err != nil {
		return nil, err
	}

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

// Validation translation functions for Azure & GCP instance type error functions.
func (c *Config) registerTranslateAzureInstanceTypeError(ut ut.Translator) error {
	// Suggest trusted launch VMs if confidential VMs have been specifically disabled
	if c.Provider.Azure != nil && c.Provider.Azure.ConfidentialVM != nil && !*c.Provider.Azure.ConfidentialVM {
		return ut.Add("azure_instance_type", fmt.Sprintf("{0} must be one of %v", instancetypes.AzureTrustedLaunchInstanceTypes), true)
	}
	// Otherwise suggest CVMs
	return ut.Add("azure_instance_type", fmt.Sprintf("{0} must be one of %v", instancetypes.AzureCVMInstanceTypes), true)
}

func translateAzureInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("azure_instance_type", fe.Field())

	return t
}

func registerTranslateGCPInstanceTypeError(ut ut.Translator) error {
	return ut.Add("gcp_instance_type", fmt.Sprintf("{0} must be one of %v", instancetypes.GCPInstanceTypes), true)
}

func translateGCPInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("gcp_instance_type", fe.Field())

	return t
}

// HasProvider checks whether the config contains the provider.
func (c *Config) HasProvider(provider cloudprovider.Provider) bool {
	switch provider {
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
	if c.HasProvider(cloudprovider.Azure) {
		return c.Provider.Azure.Image
	}
	if c.HasProvider(cloudprovider.GCP) {
		return c.Provider.GCP.Image
	}
	return ""
}

func (c *Config) UpdateMeasurements(newMeasurements Measurements) {
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

// IsImageDebug checks whether image name looks like a release image, if not it is
// probably a debug image. In the end we do not if bootstrapper or debugd
// was put inside an image just by looking at its name.
func (c *Config) IsImageDebug() bool {
	switch {
	case c.Provider.GCP != nil:
		gcpRegex := regexp.MustCompile(`^projects\/constellation-images\/global\/images\/constellation-v[\d]+-[\d]+-[\d]+$`)
		return !gcpRegex.MatchString(c.Provider.GCP.Image)
	case c.Provider.Azure != nil:
		azureRegex := regexp.MustCompile(`^\/CommunityGalleries\/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df\/Images\/constellation\/Versions\/[\d]+.[\d]+.[\d]+$`)
		return !azureRegex.MatchString(c.Provider.Azure.Image)
	default:
		return false
	}
}

// GetProvider returns the configured cloud provider.
func (c *Config) GetProvider() cloudprovider.Provider {
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

func (c *Config) EnforcesIdKeyDigest() bool {
	return c.Provider.Azure != nil && c.Provider.Azure.EnforceIdKeyDigest != nil && *c.Provider.Azure.EnforceIdKeyDigest
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
	case cloudprovider.GCP:
		for _, instanceType := range instancetypes.GCPInstanceTypes {
			if insType == instanceType {
				return true
			}
		}
		return false
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
	default:
		return false
	}
}

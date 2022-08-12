// This binary can be build from siderolabs/talos projects. Located at:
// https://github.com/siderolabs/talos/tree/master/hack/docgen
//
//go:generate docgen ./config.go ./config_doc.go Configuration
package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
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
	//   Type of a node's state disk. The type influences boot time and I/O performance. See: https://docs.microsoft.com/en-us/azure/virtual-machines/disks-types#disk-type-comparison
	StateDiskType string `yaml:"stateDiskType" validate:"oneof=Premium_LRS Premium_ZRS Standard_LRS StandardSSD_LRS StandardSSD_ZRS"`
	// description: |
	//   Expected confidential VM measurements.
	Measurements Measurements `yaml:"measurements"`
	// description: |
	//   List of values that should be enforced to be equal to the ones from the measurement list. Any non-equal values not in this list will only result in a warning.
	EnforcedMeasurements []uint32 `yaml:"enforcedMeasurements"`
	// description: |
	//   Authorize spawned VMs to access Azure API. See: https://constellation-docs.edgeless.systems/6c320851-bdd2-41d5-bf10-e27427398692/#/getting-started/install?id=azure
	UserAssignedIdentity string `yaml:"userAssignedIdentity" validate:"required"`
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
	//   Type of a node's state disk. The type influences boot time and I/O performance. See: https://cloud.google.com/compute/docs/disks#disk-types
	StateDiskType string `yaml:"stateDiskType" validate:"oneof=pd-standard pd-balanced pd-ssd"`
	// description: |
	//   Roles added to service account.
	ServiceAccountRoles []string `yaml:"serviceAccountRoles"`
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
			// TODO remove our subscriptions from the default config
			Azure: &AzureConfig{
				SubscriptionID:       "",
				TenantID:             "",
				Location:             "",
				UserAssignedIdentity: "",
				Image:                "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1659453699",
				StateDiskType:        "StandardSSD_LRS", // TODO: Replace with Premium_LRS when we replace the default VM size (Standard_D2a_v4) since the size does not support Premium_LRS
				Measurements:         copyPCRMap(azurePCRs),
				EnforcedMeasurements: []uint32{8, 9, 11, 12},
			},
			GCP: &GCPConfig{
				Project: "",
				Region:  "",
				Zone:    "",
				Image:   "projects/constellation-images/global/images/constellation-coreos-1659453699",
				ServiceAccountRoles: []string{
					"roles/compute.instanceAdmin.v1",
					"roles/compute.networkAdmin",
					"roles/compute.securityAdmin",
					"roles/storage.admin",
					"roles/iam.serviceAccountUser",
				},
				StateDiskType:        "pd-ssd",
				Measurements:         copyPCRMap(gcpPCRs),
				EnforcedMeasurements: []uint32{0, 8, 9, 11, 12},
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

// Validate checks the config values and returns validation error messages.
// The function only returns an error if the validation itself fails.
func (c *Config) Validate() ([]string, error) {
	trans := ut.New(en.New()).GetFallback()
	validate := validator.New()
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		return nil, err
	}

	// register custom validator with label supported_k8s_version to validate version based on available versionConfigs.
	if err := validate.RegisterValidation("supported_k8s_version", validateK8sVersion); err != nil {
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

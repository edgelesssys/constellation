package config

import (
	"fmt"
	"strconv"

	azureClient "github.com/edgelesssys/constellation/cli/azure/client"
	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/cli/ec2"
	awsClient "github.com/edgelesssys/constellation/cli/ec2/client"
	"github.com/edgelesssys/constellation/cli/file"
	gcpClient "github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/constants"
	"google.golang.org/protobuf/proto"
)

var (
	// gcpPCRs is a map of the expected PCR values for a GCP Constellation node.
	// TODO: Get a full list once we have stable releases.
	gcpPCRs = map[uint32][]byte{
		0:                              {0x0F, 0x35, 0xC2, 0x14, 0x60, 0x8D, 0x93, 0xC7, 0xA6, 0xE6, 0x8A, 0xE7, 0x35, 0x9B, 0x4A, 0x8B, 0xE5, 0xA0, 0xE9, 0x9E, 0xEA, 0x91, 0x07, 0xEC, 0xE4, 0x27, 0xC4, 0xDE, 0xA4, 0xE4, 0x39, 0xCF},
		uint32(vtpm.PCRIndexOwnerID):   {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		uint32(vtpm.PCRIndexClusterID): {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}

	// azurePCRs is a map of the expected PCR values for an Azure Constellation node.
	// TODO: Get a full list once we have a working setup with stable releases.
	azurePCRs = map[uint32][]byte{
		uint32(vtpm.PCRIndexOwnerID):   {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		uint32(vtpm.PCRIndexClusterID): {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}

	qemuPCRs = map[uint32][]byte{
		uint32(vtpm.PCRIndexOwnerID):   {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		uint32(vtpm.PCRIndexClusterID): {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
)

// Config defines a configuration used by the CLI.
// All fields in this struct and its child structs have pointer types
// to ensure the default values of the actual type is not confused with an omitted value.
type Config struct {
	CoordinatorPort          *string         `json:"coordinatorport,omitempty"`
	AutoscalingNodeGroupsMin *int            `json:"autoscalingnodegroupsmin,omitempty"`
	AutoscalingNodeGroupsMax *int            `json:"autoscalingnodegroupsmax,omitempty"`
	StateDiskSizeGB          *int            `json:"statedisksizegb,omitempty"`
	Provider                 *ProviderConfig `json:"provider,omitempty"`
}

// Default returns a struct with the default config.
func Default() *Config {
	return &Config{
		CoordinatorPort:          proto.String(strconv.Itoa(constants.CoordinatorPort)),
		AutoscalingNodeGroupsMin: intPtr(1),
		AutoscalingNodeGroupsMax: intPtr(10),
		StateDiskSizeGB:          intPtr(30),
		Provider: &ProviderConfig{
			EC2: &EC2Config{
				Image: proto.String("ami-07d3864beb84157d3"),
				Tags: &[]ec2.Tag{
					{
						Key:   "responsible",
						Value: "cli",
					},
					{
						Key:   "Name",
						Value: "Constellation",
					},
				},
				SecurityGroupInput: &awsClient.SecurityGroupInput{
					Inbound: cloudtypes.Firewall{
						{
							Description: "Coordinator default port",
							Protocol:    "TCP",
							IPRange:     "0.0.0.0/0",
							FromPort:    constants.CoordinatorPort,
						},
						{
							Description: "Enclave SSH",
							Protocol:    "TCP",
							IPRange:     "0.0.0.0/0",
							FromPort:    constants.EnclaveSSHPort,
						},
						{
							Description: "WireGuard default port",
							Protocol:    "UDP",
							IPRange:     "0.0.0.0/0",
							FromPort:    constants.WireguardPort,
						},
						{
							Description: "SSH",
							Protocol:    "TCP",
							IPRange:     "0.0.0.0/0",
							FromPort:    constants.SSHPort,
						},
						{
							Description: "NVMe over TCP",
							Protocol:    "TCP",
							IPRange:     "0.0.0.0/0",
							FromPort:    constants.NVMEOverTCPPort,
						},
						{
							Description: "NodePort",
							Protocol:    "TCP",
							IPRange:     "0.0.0.0/0",
							FromPort:    constants.NodePortFrom,
							ToPort:      constants.NodePortTo,
						},
					},
				},
			},
			Azure: &AzureConfig{
				SubscriptionID: proto.String("0d202bbb-4fa7-4af8-8125-58c269a05435"),
				TenantID:       proto.String("adb650a8-5da3-4b15-b4b0-3daf65ff7626"),
				Location:       proto.String("North Europe"),
				Image:          proto.String("/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1651150807"),
				NetworkSecurityGroupInput: &azureClient.NetworkSecurityGroupInput{
					Ingress: cloudtypes.Firewall{
						{
							Name:        "coordinator",
							Description: "Coordinator default port",
							Protocol:    "tcp",
							IPRange:     "0.0.0.0/0",
							FromPort:    constants.CoordinatorPort,
						},
						{
							Name:        "wireguard",
							Description: "WireGuard default port",
							Protocol:    "udp",
							IPRange:     "0.0.0.0/0",
							FromPort:    constants.WireguardPort,
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
					},
				},
				PCRs:                 pcrPtr(azurePCRs),
				UserAssignedIdentity: proto.String("/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.ManagedIdentity/userAssignedIdentities/constellation-dev-identity"),
			},
			GCP: &GCPConfig{
				Project: proto.String("constellation-331613"),
				Region:  proto.String("europe-west3"),
				Zone:    proto.String("europe-west3-b"),
				Image:   proto.String("constellation-coreos-1651150807"),
				FirewallInput: &gcpClient.FirewallInput{
					Ingress: cloudtypes.Firewall{
						{
							Name:        "coordinator",
							Description: "Coordinator default port",
							Protocol:    "tcp",
							FromPort:    constants.CoordinatorPort,
						},
						{
							Name:        "wireguard",
							Description: "WireGuard default port",
							Protocol:    "udp",
							FromPort:    constants.WireguardPort,
						},
						{
							Name:        "ssh",
							Description: "SSH",
							Protocol:    "tcp",
							FromPort:    constants.SSHPort,
						},
						{
							Name:        "nodeport",
							Description: "NodePort",
							Protocol:    "tcp",
							FromPort:    constants.NodePortFrom,
							ToPort:      constants.NodePortTo,
						},
					},
				},
				VPCsInput: &gcpClient.VPCsInput{
					SubnetCIDR:    "192.168.178.0/24",
					SubnetExtCIDR: "10.10.0.0/16",
				},
				ServiceAccountRoles: &[]string{
					"roles/compute.instanceAdmin.v1",
					"roles/compute.networkAdmin",
					"roles/compute.securityAdmin",
					"roles/storage.admin",
					"roles/iam.serviceAccountUser",
				},
				PCRs: pcrPtr(gcpPCRs),
			},
			QEMU: &QEMUConfig{
				PCRs: pcrPtr(qemuPCRs),
			},
		},
	}
}

// FromFile returns a default config that has been merged with a config file.
// If name is empty, the defaults are returned.
func FromFile(fileHandler file.Handler, name string) (*Config, error) {
	conf := Default()
	if name == "" {
		return conf, nil
	}

	if err := fileHandler.ReadJSON(name, conf); err != nil {
		return nil, fmt.Errorf("could not load config from file %s: %w", name, err)
	}
	return conf, nil
}

// ProviderConfig are cloud-provider specific configuration values used by the CLI.
type ProviderConfig struct {
	EC2   *EC2Config   `json:"ec2config,omitempty"`
	Azure *AzureConfig `json:"azureconfig,omitempty"`
	GCP   *GCPConfig   `json:"gcpconfig,omitempty"`
	QEMU  *QEMUConfig  `json:"qemuconfig,omitempty"`
}

// EC2Config are AWS EC2 specific configuration values used by the CLI.
type EC2Config struct {
	Image              *string                       `json:"image,omitempty"`
	Tags               *[]ec2.Tag                    `json:"tags,omitempty"`
	SecurityGroupInput *awsClient.SecurityGroupInput `json:"securitygroupinput,omitempty"`
}

// AzureConfig are Azure specific configuration values used by the CLI.
type AzureConfig struct {
	SubscriptionID            *string                                `json:"subscription,omitempty"` // TODO: This will be user input
	TenantID                  *string                                `json:"tenant,omitempty"`       // TODO: This will be user input
	Location                  *string                                `json:"location,omitempty"`     // TODO: This will be user input
	Image                     *string                                `json:"image,omitempty"`
	NetworkSecurityGroupInput *azureClient.NetworkSecurityGroupInput `json:"networksecuritygroupinput,omitempty"`
	PCRs                      *map[uint32][]byte                     `json:"pcrs,omitempty"`
	UserAssignedIdentity      *string                                `json:"userassignedidentity,omitempty"`
}

// GCPConfig are GCP specific configuration values used by the CLI.
type GCPConfig struct {
	Project             *string                  `json:"project,omitempty"` // TODO: This will be user input
	Region              *string                  `json:"region,omitempty"`  // TODO: This will be user input
	Zone                *string                  `json:"zone,omitempty"`    // TODO: This will be user input
	Image               *string                  `json:"image,omitempty"`
	FirewallInput       *gcpClient.FirewallInput `json:"firewallinput,omitempty"`
	VPCsInput           *gcpClient.VPCsInput     `json:"vpcsinput,omitempty"`
	ServiceAccountRoles *[]string                `json:"serviceaccountroles,omitempty"`
	PCRs                *map[uint32][]byte       `json:"pcrs,omitempty"`
}

type QEMUConfig struct {
	PCRs *map[uint32][]byte `json:"pcrs,omitempty"`
}

func pcrPtr(pcrs map[uint32][]byte) *map[uint32][]byte {
	return &pcrs
}

// intPtr returns a pointer to the copied value of in.
func intPtr(in int) *int {
	return &in
}

package state

import (
	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/edgelesssys/constellation/cli/gcp"
)

// ConstellationState is the state of a Constellation.
type ConstellationState struct {
	Name          string `json:"name,omitempty"`
	UID           string `json:"uid,omitempty"`
	CloudProvider string `json:"cloudprovider,omitempty"`

	EC2Instances     ec2.Instances `json:"ec2instances,omitempty"`
	EC2SecurityGroup string        `json:"ec2securitygroup,omitempty"`

	GCPNodes                       gcp.Instances `json:"gcpnodes,omitempty"`
	GCPCoordinators                gcp.Instances `json:"gcpcoordinators,omitempty"`
	GCPNodeInstanceGroup           string        `json:"gcpnodeinstancegroup,omitempty"`
	GCPCoordinatorInstanceGroup    string        `json:"gcpcoordinatorinstancegroup,omitempty"`
	GCPNodeInstanceTemplate        string        `json:"gcpnodeinstancetemplate,omitempty"`
	GCPCoordinatorInstanceTemplate string        `json:"gcpcoordinatorinstancetemplate,omitempty"`
	GCPNetwork                     string        `json:"gcpnetwork,omitempty"`
	GCPSubnetwork                  string        `json:"gcpsubnetwork,omitempty"`
	GCPFirewalls                   []string      `json:"gcpfirewalls,omitempty"`
	GCPProject                     string        `json:"gcpproject,omitempty"`
	GCPZone                        string        `json:"gcpzone,omitempty"`
	GCPRegion                      string        `json:"gcpregion,omitempty"`
	GCPServiceAccount              string        `json:"gcpserviceaccount,omitempty"`

	AzureNodes                azure.Instances `json:"azurenodes,omitempty"`
	AzureCoordinators         azure.Instances `json:"azurecoordinators,omitempty"`
	AzureResourceGroup        string          `json:"azureresourcegroup,omitempty"`
	AzureLocation             string          `json:"azurelocation,omitempty"`
	AzureSubscription         string          `json:"azuresubscription,omitempty"`
	AzureTenant               string          `json:"azuretenant,omitempty"`
	AzureSubnet               string          `json:"azuresubnet,omitempty"`
	AzureNetworkSecurityGroup string          `json:"azurenetworksecuritygroup,omitempty"`
	AzureNodesScaleSet        string          `json:"azurenodesscaleset,omitempty"`
	AzureCoordinatorsScaleSet string          `json:"azurecoordinatorsscaleset,omitempty"`
	AzureADAppObjectID        string          `json:"azureadappobjectid,omitempty"`
}

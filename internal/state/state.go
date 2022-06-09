package state

import (
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
)

// ConstellationState is the state of a Constellation.
type ConstellationState struct {
	Name          string `json:"name,omitempty"`
	UID           string `json:"uid,omitempty"`
	CloudProvider string `json:"cloudprovider,omitempty"`

	GCPNodes                       cloudtypes.Instances `json:"gcpnodes,omitempty"`
	GCPCoordinators                cloudtypes.Instances `json:"gcpcoordinators,omitempty"`
	GCPNodeInstanceGroup           string               `json:"gcpnodeinstancegroup,omitempty"`
	GCPCoordinatorInstanceGroup    string               `json:"gcpcoordinatorinstancegroup,omitempty"`
	GCPNodeInstanceTemplate        string               `json:"gcpnodeinstancetemplate,omitempty"`
	GCPCoordinatorInstanceTemplate string               `json:"gcpcoordinatorinstancetemplate,omitempty"`
	GCPNetwork                     string               `json:"gcpnetwork,omitempty"`
	GCPSubnetwork                  string               `json:"gcpsubnetwork,omitempty"`
	GCPFirewalls                   []string             `json:"gcpfirewalls,omitempty"`
	GCPBackendService              string               `json:"gcpbackendservice,omitempty"`
	GCPHealthCheck                 string               `json:"gcphealthcheck,omitempty"`
	GCPForwardingRule              string               `json:"gcpforwardingrule,omitempty"`
	GCPProject                     string               `json:"gcpproject,omitempty"`
	GCPZone                        string               `json:"gcpzone,omitempty"`
	GCPRegion                      string               `json:"gcpregion,omitempty"`
	GCPServiceAccount              string               `json:"gcpserviceaccount,omitempty"`

	AzureNodes                cloudtypes.Instances `json:"azurenodes,omitempty"`
	AzureCoordinators         cloudtypes.Instances `json:"azurecoordinators,omitempty"`
	AzureResourceGroup        string               `json:"azureresourcegroup,omitempty"`
	AzureLocation             string               `json:"azurelocation,omitempty"`
	AzureSubscription         string               `json:"azuresubscription,omitempty"`
	AzureTenant               string               `json:"azuretenant,omitempty"`
	AzureSubnet               string               `json:"azuresubnet,omitempty"`
	AzureNetworkSecurityGroup string               `json:"azurenetworksecuritygroup,omitempty"`
	AzureNodesScaleSet        string               `json:"azurenodesscaleset,omitempty"`
	AzureCoordinatorsScaleSet string               `json:"azurecoordinatorsscaleset,omitempty"`
	AzureADAppObjectID        string               `json:"azureadappobjectid,omitempty"`

	QEMUNodes        cloudtypes.Instances `json:"qemunodes,omitempty"`
	QEMUCoordinators cloudtypes.Instances `json:"qemucoordinators,omitempty"`
}

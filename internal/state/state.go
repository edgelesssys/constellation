package state

import (
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
)

// ConstellationState is the state of a Constellation.
type ConstellationState struct {
	Name             string `json:"name,omitempty"`
	UID              string `json:"uid,omitempty"`
	CloudProvider    string `json:"cloudprovider,omitempty"`
	BootstrapperHost string `json:"bootstrapperhost,omitempty"`

	GCPWorkers                      cloudtypes.Instances `json:"gcpworkers,omitempty"`
	GCPControlPlanes                cloudtypes.Instances `json:"gcpcontrolplanes,omitempty"`
	GCPWorkerInstanceGroup          string               `json:"gcpworkerinstancegroup,omitempty"`
	GCPControlPlaneInstanceGroup    string               `json:"gcpcontrolplaneinstancegroup,omitempty"`
	GCPWorkerInstanceTemplate       string               `json:"gcpworkerinstancetemplate,omitempty"`
	GCPControlPlaneInstanceTemplate string               `json:"gcpcontrolplaneinstancetemplate,omitempty"`
	GCPNetwork                      string               `json:"gcpnetwork,omitempty"`
	GCPSubnetwork                   string               `json:"gcpsubnetwork,omitempty"`
	GCPFirewalls                    []string             `json:"gcpfirewalls,omitempty"`
	GCPBackendService               string               `json:"gcpbackendservice,omitempty"`
	GCPHealthCheck                  string               `json:"gcphealthcheck,omitempty"`
	GCPForwardingRule               string               `json:"gcpforwardingrule,omitempty"`
	GCPProject                      string               `json:"gcpproject,omitempty"`
	GCPZone                         string               `json:"gcpzone,omitempty"`
	GCPRegion                       string               `json:"gcpregion,omitempty"`
	GCPServiceAccount               string               `json:"gcpserviceaccount,omitempty"`

	AzureWorkers               cloudtypes.Instances `json:"azureworkers,omitempty"`
	AzureControlPlane          cloudtypes.Instances `json:"azurecontrolplanes,omitempty"`
	AzureResourceGroup         string               `json:"azureresourcegroup,omitempty"`
	AzureLocation              string               `json:"azurelocation,omitempty"`
	AzureSubscription          string               `json:"azuresubscription,omitempty"`
	AzureTenant                string               `json:"azuretenant,omitempty"`
	AzureSubnet                string               `json:"azuresubnet,omitempty"`
	AzureNetworkSecurityGroup  string               `json:"azurenetworksecuritygroup,omitempty"`
	AzureWorkersScaleSet       string               `json:"azureworkersscaleset,omitempty"`
	AzureControlPlanesScaleSet string               `json:"azurecontrolplanesscaleset,omitempty"`
	AzureADAppObjectID         string               `json:"azureadappobjectid,omitempty"`

	QEMUWorkers      cloudtypes.Instances `json:"qemuworkers,omitempty"`
	QEMUControlPlane cloudtypes.Instances `json:"qemucontrolplanes,omitempty"`
}

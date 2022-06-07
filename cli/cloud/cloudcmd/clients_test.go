package cloudcmd

import (
	"context"
	"errors"
	"strconv"

	"github.com/edgelesssys/constellation/cli/azure"
	azurecl "github.com/edgelesssys/constellation/cli/azure/client"
	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
	gcpcl "github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/state"
)

type fakeAzureClient struct {
	nodes        azure.Instances
	coordinators azure.Instances

	resourceGroup        string
	name                 string
	uid                  string
	location             string
	subscriptionID       string
	tenantID             string
	subnetID             string
	loadBalancerName     string
	coordinatorsScaleSet string
	nodesScaleSet        string
	networkSecurityGroup string
	adAppObjectID        string
}

func (c *fakeAzureClient) GetState() (state.ConstellationState, error) {
	stat := state.ConstellationState{
		CloudProvider:             cloudprovider.Azure.String(),
		AzureNodes:                c.nodes,
		AzureCoordinators:         c.coordinators,
		Name:                      c.name,
		UID:                       c.uid,
		AzureResourceGroup:        c.resourceGroup,
		AzureLocation:             c.location,
		AzureSubscription:         c.subscriptionID,
		AzureTenant:               c.tenantID,
		AzureSubnet:               c.subnetID,
		AzureNetworkSecurityGroup: c.networkSecurityGroup,
		AzureNodesScaleSet:        c.nodesScaleSet,
		AzureCoordinatorsScaleSet: c.coordinatorsScaleSet,
		AzureADAppObjectID:        c.adAppObjectID,
	}
	return stat, nil
}

func (c *fakeAzureClient) SetState(stat state.ConstellationState) error {
	c.nodes = stat.AzureNodes
	c.coordinators = stat.AzureCoordinators
	c.name = stat.Name
	c.uid = stat.UID
	c.resourceGroup = stat.AzureResourceGroup
	c.location = stat.AzureLocation
	c.subscriptionID = stat.AzureSubscription
	c.tenantID = stat.AzureTenant
	c.subnetID = stat.AzureSubnet
	c.networkSecurityGroup = stat.AzureNetworkSecurityGroup
	c.nodesScaleSet = stat.AzureNodesScaleSet
	c.coordinatorsScaleSet = stat.AzureCoordinatorsScaleSet
	c.adAppObjectID = stat.AzureADAppObjectID
	return nil
}

func (c *fakeAzureClient) CreateResourceGroup(ctx context.Context) error {
	c.resourceGroup = "resource-group"
	return nil
}

func (c *fakeAzureClient) CreateVirtualNetwork(ctx context.Context) error {
	c.subnetID = "subnet"
	return nil
}

func (c *fakeAzureClient) CreateExternalLoadBalancer(ctx context.Context) error {
	c.loadBalancerName = "loadBalancer"
	return nil
}

func (c *fakeAzureClient) CreateSecurityGroup(ctx context.Context, input azurecl.NetworkSecurityGroupInput) error {
	c.networkSecurityGroup = "network-security-group"
	return nil
}

func (c *fakeAzureClient) CreateInstances(ctx context.Context, input azurecl.CreateInstancesInput) error {
	c.coordinatorsScaleSet = "coordinators-scale-set"
	c.nodesScaleSet = "nodes-scale-set"
	c.nodes = make(azure.Instances)
	for i := 0; i < input.CountNodes; i++ {
		id := "id-" + strconv.Itoa(i)
		c.nodes[id] = azure.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.coordinators = make(azure.Instances)
	for i := 0; i < input.CountCoordinators; i++ {
		id := "id-" + strconv.Itoa(i)
		c.coordinators[id] = azure.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	return nil
}

// TODO: deprecate as soon as scale sets are available.
func (c *fakeAzureClient) CreateInstancesVMs(ctx context.Context, input azurecl.CreateInstancesInput) error {
	c.nodes = make(azure.Instances)
	for i := 0; i < input.CountNodes; i++ {
		id := "id-" + strconv.Itoa(i)
		c.nodes[id] = azure.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.coordinators = make(azure.Instances)
	for i := 0; i < input.CountCoordinators; i++ {
		id := "id-" + strconv.Itoa(i)
		c.coordinators[id] = azure.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	return nil
}

func (c *fakeAzureClient) CreateServicePrincipal(ctx context.Context) (string, error) {
	c.adAppObjectID = "00000000-0000-0000-0000-000000000001"
	return azurecl.ApplicationCredentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}.ConvertToCloudServiceAccountURI(), nil
}

func (c *fakeAzureClient) TerminateResourceGroup(ctx context.Context) error {
	if c.resourceGroup == "" {
		return nil
	}
	c.nodes = nil
	c.coordinators = nil
	c.resourceGroup = ""
	c.subnetID = ""
	c.networkSecurityGroup = ""
	c.nodesScaleSet = ""
	c.coordinatorsScaleSet = ""
	return nil
}

func (c *fakeAzureClient) TerminateServicePrincipal(ctx context.Context) error {
	if c.adAppObjectID == "" {
		return nil
	}
	c.adAppObjectID = ""
	return nil
}

type stubAzureClient struct {
	terminateResourceGroupCalled    bool
	terminateServicePrincipalCalled bool

	getStateErr                  error
	setStateErr                  error
	createResourceGroupErr       error
	createVirtualNetworkErr      error
	createSecurityGroupErr       error
	createLoadBalancerErr        error
	createInstancesErr           error
	createServicePrincipalErr    error
	terminateResourceGroupErr    error
	terminateServicePrincipalErr error
}

func (c *stubAzureClient) GetState() (state.ConstellationState, error) {
	return state.ConstellationState{}, c.getStateErr
}

func (c *stubAzureClient) SetState(state.ConstellationState) error {
	return c.setStateErr
}

func (c *stubAzureClient) CreateExternalLoadBalancer(ctx context.Context) error {
	return c.createLoadBalancerErr
}

func (c *stubAzureClient) CreateResourceGroup(ctx context.Context) error {
	return c.createResourceGroupErr
}

func (c *stubAzureClient) CreateVirtualNetwork(ctx context.Context) error {
	return c.createVirtualNetworkErr
}

func (c *stubAzureClient) CreateSecurityGroup(ctx context.Context, input azurecl.NetworkSecurityGroupInput) error {
	return c.createSecurityGroupErr
}

func (c *stubAzureClient) CreateInstances(ctx context.Context, input azurecl.CreateInstancesInput) error {
	return c.createInstancesErr
}

// TODO: deprecate as soon as scale sets are available.
func (c *stubAzureClient) CreateInstancesVMs(ctx context.Context, input azurecl.CreateInstancesInput) error {
	return c.createInstancesErr
}

func (c *stubAzureClient) CreateServicePrincipal(ctx context.Context) (string, error) {
	return azurecl.ApplicationCredentials{
		ClientID:     "00000000-0000-0000-0000-000000000000",
		ClientSecret: "secret",
	}.ConvertToCloudServiceAccountURI(), c.createServicePrincipalErr
}

func (c *stubAzureClient) TerminateResourceGroup(ctx context.Context) error {
	c.terminateResourceGroupCalled = true
	return c.terminateResourceGroupErr
}

func (c *stubAzureClient) TerminateServicePrincipal(ctx context.Context) error {
	c.terminateServicePrincipalCalled = true
	return c.terminateServicePrincipalErr
}

type fakeGcpClient struct {
	nodes        cloudtypes.Instances
	coordinators cloudtypes.Instances

	nodesInstanceGroup       string
	coordinatorInstanceGroup string
	coordinatorTemplate      string
	nodeTemplate             string
	network                  string
	subnetwork               string
	firewalls                []string
	project                  string
	uid                      string
	name                     string
	zone                     string
	serviceAccount           string
}

func (c *fakeGcpClient) GetState() (state.ConstellationState, error) {
	stat := state.ConstellationState{
		CloudProvider:                  cloudprovider.GCP.String(),
		GCPNodes:                       c.nodes,
		GCPCoordinators:                c.coordinators,
		GCPNodeInstanceGroup:           c.nodesInstanceGroup,
		GCPCoordinatorInstanceGroup:    c.coordinatorInstanceGroup,
		GCPNodeInstanceTemplate:        c.nodeTemplate,
		GCPCoordinatorInstanceTemplate: c.coordinatorTemplate,
		GCPNetwork:                     c.network,
		GCPSubnetwork:                  c.subnetwork,
		GCPFirewalls:                   c.firewalls,
		GCPProject:                     c.project,
		Name:                           c.name,
		UID:                            c.uid,
		GCPZone:                        c.zone,
		GCPServiceAccount:              c.serviceAccount,
	}
	return stat, nil
}

func (c *fakeGcpClient) SetState(stat state.ConstellationState) error {
	c.nodes = stat.GCPNodes
	c.coordinators = stat.GCPCoordinators
	c.nodesInstanceGroup = stat.GCPNodeInstanceGroup
	c.coordinatorInstanceGroup = stat.GCPCoordinatorInstanceGroup
	c.nodeTemplate = stat.GCPNodeInstanceTemplate
	c.coordinatorTemplate = stat.GCPCoordinatorInstanceTemplate
	c.network = stat.GCPNetwork
	c.subnetwork = stat.GCPSubnetwork
	c.firewalls = stat.GCPFirewalls
	c.project = stat.GCPProject
	c.name = stat.Name
	c.uid = stat.UID
	c.zone = stat.GCPZone
	c.serviceAccount = stat.GCPServiceAccount
	return nil
}

func (c *fakeGcpClient) CreateVPCs(ctx context.Context) error {
	c.network = "network"
	c.subnetwork = "subnetwork"
	return nil
}

func (c *fakeGcpClient) CreateFirewall(ctx context.Context, input gcpcl.FirewallInput) error {
	if c.network == "" {
		return errors.New("client has not network")
	}
	for _, rule := range input.Ingress {
		c.firewalls = append(c.firewalls, rule.Name)
	}
	return nil
}

func (c *fakeGcpClient) CreateInstances(ctx context.Context, input gcpcl.CreateInstancesInput) error {
	c.coordinatorInstanceGroup = "coordinator-group"
	c.nodesInstanceGroup = "nodes-group"
	c.nodeTemplate = "node-template"
	c.coordinatorTemplate = "coordinator-template"
	c.nodes = make(cloudtypes.Instances)
	for i := 0; i < input.CountNodes; i++ {
		id := "id-" + strconv.Itoa(i)
		c.nodes[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.coordinators = make(cloudtypes.Instances)
	for i := 0; i < input.CountCoordinators; i++ {
		id := "id-" + strconv.Itoa(i)
		c.coordinators[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	return nil
}

func (c *fakeGcpClient) CreateServiceAccount(ctx context.Context, input gcpcl.ServiceAccountInput) (string, error) {
	c.serviceAccount = "service-account@" + c.project + ".iam.gserviceaccount.com"
	return gcpcl.ServiceAccountKey{
		Type:                    "service_account",
		ProjectID:               c.project,
		PrivateKeyID:            "key-id",
		PrivateKey:              "-----BEGIN PRIVATE KEY-----\nprivate-key\n-----END PRIVATE KEY-----\n",
		ClientEmail:             c.serviceAccount,
		ClientID:                "client-id",
		AuthURI:                 "https://accounts.google.com/o/oauth2/auth",
		TokenURI:                "https://accounts.google.com/o/oauth2/token",
		AuthProviderX509CertURL: "https://www.googleapis.com/oauth2/v1/certs",
		ClientX509CertURL:       "https://www.googleapis.com/robot/v1/metadata/x509/service-account-email",
	}.ConvertToCloudServiceAccountURI(), nil
}

func (c *fakeGcpClient) TerminateFirewall(ctx context.Context) error {
	if len(c.firewalls) == 0 {
		return nil
	}
	c.firewalls = nil
	return nil
}

func (c *fakeGcpClient) TerminateVPCs(context.Context) error {
	if len(c.firewalls) != 0 {
		return errors.New("client has firewalls, which must be deleted first")
	}
	c.network = ""
	c.subnetwork = ""
	return nil
}

func (c *fakeGcpClient) TerminateInstances(context.Context) error {
	c.nodeTemplate = ""
	c.coordinatorTemplate = ""
	c.nodesInstanceGroup = ""
	c.coordinatorInstanceGroup = ""
	c.nodes = nil
	c.coordinators = nil
	return nil
}

func (c *fakeGcpClient) TerminateServiceAccount(context.Context) error {
	c.serviceAccount = ""
	return nil
}

func (c *fakeGcpClient) Close() error {
	return nil
}

type stubGcpClient struct {
	terminateFirewallCalled       bool
	terminateInstancesCalled      bool
	terminateVPCsCalled           bool
	terminateServiceAccountCalled bool
	closeCalled                   bool

	getStateErr                error
	setStateErr                error
	createVPCsErr              error
	createFirewallErr          error
	createInstancesErr         error
	createServiceAccountErr    error
	terminateFirewallErr       error
	terminateVPCsErr           error
	terminateInstancesErr      error
	terminateServiceAccountErr error
	closeErr                   error
}

func (c *stubGcpClient) GetState() (state.ConstellationState, error) {
	return state.ConstellationState{}, c.getStateErr
}

func (c *stubGcpClient) SetState(state.ConstellationState) error {
	return c.setStateErr
}

func (c *stubGcpClient) CreateVPCs(ctx context.Context) error {
	return c.createVPCsErr
}

func (c *stubGcpClient) CreateFirewall(ctx context.Context, input gcpcl.FirewallInput) error {
	return c.createFirewallErr
}

func (c *stubGcpClient) CreateInstances(ctx context.Context, input gcpcl.CreateInstancesInput) error {
	return c.createInstancesErr
}

func (c *stubGcpClient) CreateServiceAccount(ctx context.Context, input gcpcl.ServiceAccountInput) (string, error) {
	return gcpcl.ServiceAccountKey{}.ConvertToCloudServiceAccountURI(), c.createServiceAccountErr
}

func (c *stubGcpClient) TerminateFirewall(ctx context.Context) error {
	c.terminateFirewallCalled = true
	return c.terminateFirewallErr
}

func (c *stubGcpClient) TerminateVPCs(context.Context) error {
	c.terminateVPCsCalled = true
	return c.terminateVPCsErr
}

func (c *stubGcpClient) TerminateInstances(context.Context) error {
	c.terminateInstancesCalled = true
	return c.terminateInstancesErr
}

func (c *stubGcpClient) TerminateServiceAccount(context.Context) error {
	c.terminateServiceAccountCalled = true
	return c.terminateServiceAccountErr
}

func (c *stubGcpClient) Close() error {
	c.closeCalled = true
	return c.closeErr
}

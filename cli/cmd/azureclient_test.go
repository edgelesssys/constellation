package cmd

import (
	"context"
	"strconv"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/azure/client"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
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

func (c *fakeAzureClient) CreateSecurityGroup(ctx context.Context, input client.NetworkSecurityGroupInput) error {
	c.networkSecurityGroup = "network-security-group"
	return nil
}

func (c *fakeAzureClient) CreateInstances(ctx context.Context, input client.CreateInstancesInput) error {
	c.coordinatorsScaleSet = "coordinators-scale-set"
	c.nodesScaleSet = "nodes-scale-set"
	c.nodes = make(azure.Instances)
	for i := 0; i < input.Count-1; i++ {
		id := strconv.Itoa(i)
		c.nodes[id] = azure.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.coordinators = make(azure.Instances)
	c.coordinators["0"] = azure.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	return nil
}

// TODO: deprecate as soon as scale sets are available.
func (c *fakeAzureClient) CreateInstancesVMs(ctx context.Context, input client.CreateInstancesInput) error {
	c.nodes = make(azure.Instances)
	for i := 0; i < input.Count-1; i++ {
		id := strconv.Itoa(i)
		c.nodes[id] = azure.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.coordinators = make(azure.Instances)
	c.coordinators["0"] = azure.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	return nil
}

func (c *fakeAzureClient) CreateServicePrincipal(ctx context.Context) (string, error) {
	c.adAppObjectID = "00000000-0000-0000-0000-000000000001"
	return client.ApplicationCredentials{
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
	terminateResourceGroupCalled bool

	getStateErr                  error
	setStateErr                  error
	createResourceGroupErr       error
	createVirtualNetworkErr      error
	createSecurityGroupErr       error
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

func (c *stubAzureClient) CreateResourceGroup(ctx context.Context) error {
	return c.createResourceGroupErr
}

func (c *stubAzureClient) CreateVirtualNetwork(ctx context.Context) error {
	return c.createVirtualNetworkErr
}

func (c *stubAzureClient) CreateSecurityGroup(ctx context.Context, input client.NetworkSecurityGroupInput) error {
	return c.createSecurityGroupErr
}

func (c *stubAzureClient) CreateInstances(ctx context.Context, input client.CreateInstancesInput) error {
	return c.createInstancesErr
}

// TODO: deprecate as soon as scale sets are available.
func (c *stubAzureClient) CreateInstancesVMs(ctx context.Context, input client.CreateInstancesInput) error {
	return c.createInstancesErr
}

func (c *stubAzureClient) CreateServicePrincipal(ctx context.Context) (string, error) {
	return client.ApplicationCredentials{
		ClientID:     "00000000-0000-0000-0000-000000000000",
		ClientSecret: "secret",
	}.ConvertToCloudServiceAccountURI(), c.createServicePrincipalErr
}

func (c *stubAzureClient) TerminateResourceGroup(ctx context.Context) error {
	c.terminateResourceGroupCalled = true
	return c.terminateResourceGroupErr
}

func (c *stubAzureClient) TerminateServicePrincipal(ctx context.Context) error {
	return c.terminateServicePrincipalErr
}

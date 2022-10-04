/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"strconv"
	"testing"

	azurecl "github.com/edgelesssys/constellation/v2/cli/internal/azure/client"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

type fakeAzureClient struct {
	workers       cloudtypes.Instances
	controlPlanes cloudtypes.Instances

	resourceGroup        string
	name                 string
	uid                  string
	location             string
	subscriptionID       string
	tenantID             string
	subnetID             string
	loadBalancerName     string
	controlPlaneScaleSet string
	workerScaleSet       string
	networkSecurityGroup string
	adAppObjectID        string
}

func (c *fakeAzureClient) GetState() state.ConstellationState {
	return state.ConstellationState{
		CloudProvider:              cloudprovider.Azure.String(),
		AzureWorkerInstances:       c.workers,
		AzureControlPlaneInstances: c.controlPlanes,
		Name:                       c.name,
		UID:                        c.uid,
		AzureResourceGroup:         c.resourceGroup,
		AzureLocation:              c.location,
		AzureSubscription:          c.subscriptionID,
		AzureTenant:                c.tenantID,
		AzureSubnet:                c.subnetID,
		AzureNetworkSecurityGroup:  c.networkSecurityGroup,
		AzureWorkerScaleSet:        c.workerScaleSet,
		AzureControlPlaneScaleSet:  c.controlPlaneScaleSet,
		AzureADAppObjectID:         c.adAppObjectID,
	}
}

func (c *fakeAzureClient) SetState(stat state.ConstellationState) {
	c.workers = stat.AzureWorkerInstances
	c.controlPlanes = stat.AzureControlPlaneInstances
	c.name = stat.Name
	c.uid = stat.UID
	c.resourceGroup = stat.AzureResourceGroup
	c.location = stat.AzureLocation
	c.subscriptionID = stat.AzureSubscription
	c.tenantID = stat.AzureTenant
	c.subnetID = stat.AzureSubnet
	c.networkSecurityGroup = stat.AzureNetworkSecurityGroup
	c.workerScaleSet = stat.AzureWorkerScaleSet
	c.controlPlaneScaleSet = stat.AzureControlPlaneScaleSet
	c.adAppObjectID = stat.AzureADAppObjectID
}

func (c *fakeAzureClient) CreateApplicationInsight(ctx context.Context) error {
	return nil
}

func (c *fakeAzureClient) CreateVirtualNetwork(ctx context.Context) error {
	c.subnetID = "subnet"
	return nil
}

func (c *fakeAzureClient) CreateExternalLoadBalancer(ctx context.Context, isDebugCluster bool) error {
	c.loadBalancerName = "loadBalancer"
	return nil
}

func (c *fakeAzureClient) CreateSecurityGroup(ctx context.Context, input azurecl.NetworkSecurityGroupInput) error {
	c.networkSecurityGroup = "network-security-group"
	return nil
}

func (c *fakeAzureClient) CreateInstances(ctx context.Context, input azurecl.CreateInstancesInput) error {
	c.controlPlaneScaleSet = "controlplanes-scale-set"
	c.workerScaleSet = "workers-scale-set"
	c.workers = make(cloudtypes.Instances)
	for i := 0; i < input.CountWorkers; i++ {
		id := "id-" + strconv.Itoa(i)
		c.workers[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.controlPlanes = make(cloudtypes.Instances)
	for i := 0; i < input.CountControlPlanes; i++ {
		id := "id-" + strconv.Itoa(i)
		c.controlPlanes[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	return nil
}

func (c *fakeAzureClient) CreateServicePrincipal(ctx context.Context) (string, error) {
	c.adAppObjectID = "00000000-0000-0000-0000-000000000001"
	return azureshared.ApplicationCredentials{
		AppClientID:       "client-id",
		ClientSecretValue: "client-secret",
	}.ToCloudServiceAccountURI(), nil
}

func (c *fakeAzureClient) TerminateResourceGroupResources(ctx context.Context) error {
	// TODO(katexochen)
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
	terminateResourceGroupResourcesCalled bool
	terminateServicePrincipalCalled       bool

	createApplicationInsightErr        error
	createVirtualNetworkErr            error
	createSecurityGroupErr             error
	createLoadBalancerErr              error
	createInstancesErr                 error
	createServicePrincipalErr          error
	terminateResourceGroupResourcesErr error
	terminateServicePrincipalErr       error
}

func (c *stubAzureClient) GetState() state.ConstellationState {
	return state.ConstellationState{}
}

func (c *stubAzureClient) SetState(state.ConstellationState) {
}

func (c *stubAzureClient) CreateExternalLoadBalancer(ctx context.Context, isDebugCluster bool) error {
	return c.createLoadBalancerErr
}

func (c *stubAzureClient) CreateApplicationInsight(ctx context.Context) error {
	return c.createApplicationInsightErr
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

func (c *stubAzureClient) CreateServicePrincipal(ctx context.Context) (string, error) {
	return azureshared.ApplicationCredentials{
		AppClientID:       "00000000-0000-0000-0000-000000000000",
		ClientSecretValue: "secret",
	}.ToCloudServiceAccountURI(), c.createServicePrincipalErr
}

func (c *stubAzureClient) TerminateResourceGroupResources(ctx context.Context) error {
	c.terminateResourceGroupResourcesCalled = true
	return c.terminateResourceGroupResourcesErr
}

func (c *stubAzureClient) TerminateServicePrincipal(ctx context.Context) error {
	c.terminateServicePrincipalCalled = true
	return c.terminateServicePrincipalErr
}

type stubTerraformClient struct {
	state                  state.ConstellationState
	cleanUpWorkspaceCalled bool
	removeInstallerCalled  bool
	destroyClusterCalled   bool
	createClusterErr       error
	destroyClusterErr      error
	cleanUpWorkspaceErr    error
}

func (c *stubTerraformClient) GetState() state.ConstellationState {
	return c.state
}

func (c *stubTerraformClient) CreateCluster(ctx context.Context, name string, input terraform.Variables) error {
	return c.createClusterErr
}

func (c *stubTerraformClient) DestroyCluster(ctx context.Context) error {
	c.destroyClusterCalled = true
	return c.destroyClusterErr
}

func (c *stubTerraformClient) CleanUpWorkspace() error {
	c.cleanUpWorkspaceCalled = true
	return c.cleanUpWorkspaceErr
}

func (c *stubTerraformClient) RemoveInstaller() {
	c.removeInstallerCalled = true
}

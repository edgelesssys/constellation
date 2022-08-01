package client

import (
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestSetGetState(t *testing.T) {
	state := state.ConstellationState{
		CloudProvider: cloudprovider.Azure.String(),
		AzureWorkerInstances: cloudtypes.Instances{
			"0": {PublicIP: "ip1", PrivateIP: "ip2"},
		},
		AzureControlPlaneInstances: cloudtypes.Instances{
			"0": {PublicIP: "ip3", PrivateIP: "ip4"},
		},
		Name:                      "name",
		UID:                       "uid",
		LoadBalancerIP:            "bootstrapper-host",
		AzureResourceGroup:        "resource-group",
		AzureLocation:             "location",
		AzureSubscription:         "subscription",
		AzureTenant:               "tenant",
		AzureSubnet:               "azure-subnet",
		AzureNetworkSecurityGroup: "network-security-group",
		AzureWorkerScaleSet:       "worker-scale-set",
		AzureControlPlaneScaleSet: "controlplane-scale-set",
	}

	t.Run("SetState", func(t *testing.T) {
		assert := assert.New(t)

		client := Client{}
		client.SetState(state)
		assert.Equal(state.AzureWorkerInstances, client.workers)
		assert.Equal(state.AzureControlPlaneInstances, client.controlPlanes)
		assert.Equal(state.Name, client.name)
		assert.Equal(state.UID, client.uid)
		assert.Equal(state.AzureResourceGroup, client.resourceGroup)
		assert.Equal(state.AzureLocation, client.location)
		assert.Equal(state.AzureSubscription, client.subscriptionID)
		assert.Equal(state.AzureTenant, client.tenantID)
		assert.Equal(state.AzureSubnet, client.subnetID)
		assert.Equal(state.AzureNetworkSecurityGroup, client.networkSecurityGroup)
		assert.Equal(state.AzureWorkerScaleSet, client.workerScaleSet)
		assert.Equal(state.AzureControlPlaneScaleSet, client.controlPlaneScaleSet)
	})

	t.Run("GetState", func(t *testing.T) {
		assert := assert.New(t)

		client := Client{
			workers:              state.AzureWorkerInstances,
			controlPlanes:        state.AzureControlPlaneInstances,
			name:                 state.Name,
			uid:                  state.UID,
			loadBalancerPubIP:    state.LoadBalancerIP,
			resourceGroup:        state.AzureResourceGroup,
			location:             state.AzureLocation,
			subscriptionID:       state.AzureSubscription,
			tenantID:             state.AzureTenant,
			subnetID:             state.AzureSubnet,
			networkSecurityGroup: state.AzureNetworkSecurityGroup,
			workerScaleSet:       state.AzureWorkerScaleSet,
			controlPlaneScaleSet: state.AzureControlPlaneScaleSet,
		}
		stat := client.GetState()
		assert.Equal(state, stat)
	})
}

func TestInit(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	client := Client{}
	require.NoError(client.init("location", "name"))
	assert.Equal("location", client.location)
	assert.Equal("name", client.name)
	assert.NotEmpty(client.uid)
}

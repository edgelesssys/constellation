package client

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/stretchr/testify/assert"
)

func TestCreateResourceGroup(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		resourceGroupAPI resourceGroupAPI
		wantErr          bool
	}{
		"successful create": {
			resourceGroupAPI: stubResourceGroupAPI{},
		},
		"failed create": {
			resourceGroupAPI: stubResourceGroupAPI{createErr: someErr},
			wantErr:          true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				location:         "location",
				name:             "name",
				uid:              "uid",
				resourceGroupAPI: tc.resourceGroupAPI,
				workers:          make(cloudtypes.Instances),
				controlPlanes:    make(cloudtypes.Instances),
			}

			if tc.wantErr {
				assert.Error(client.CreateResourceGroup(ctx))
			} else {
				assert.NoError(client.CreateResourceGroup(ctx))
				assert.Equal(client.name+"-"+client.uid, client.resourceGroup)
			}
		})
	}
}

func TestTerminateResourceGroup(t *testing.T) {
	someErr := errors.New("failed")
	clientWithResourceGroup := Client{
		resourceGroup:        "name",
		location:             "location",
		name:                 "name",
		uid:                  "uid",
		subnetID:             "subnet",
		workerScaleSet:       "node-scale-set",
		controlPlaneScaleSet: "controlplane-scale-set",
		workers: cloudtypes.Instances{
			"0": {
				PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1",
			},
		},
		controlPlanes: cloudtypes.Instances{
			"0": {
				PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1",
			},
		},
	}
	testCases := map[string]struct {
		resourceGroup    string
		resourceGroupAPI resourceGroupAPI
		client           Client
		wantErr          bool
	}{
		"successful terminate": {
			resourceGroupAPI: stubResourceGroupAPI{},
			client:           clientWithResourceGroup,
		},
		"no resource group to terminate": {
			resourceGroupAPI: stubResourceGroupAPI{},
			client:           Client{},
			resourceGroup:    "",
		},
		"failed terminate": {
			resourceGroupAPI: stubResourceGroupAPI{terminateErr: someErr},
			client:           clientWithResourceGroup,
			wantErr:          true,
		},
		"failed to poll terminate response": {
			resourceGroupAPI: stubResourceGroupAPI{pollErr: someErr},
			client:           clientWithResourceGroup,
			wantErr:          true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			tc.client.resourceGroupAPI = tc.resourceGroupAPI
			ctx := context.Background()

			if tc.wantErr {
				assert.Error(tc.client.TerminateResourceGroup(ctx))
				return
			}
			assert.NoError(tc.client.TerminateResourceGroup(ctx))
			assert.Empty(tc.client.resourceGroup)
			assert.Empty(tc.client.subnetID)
			assert.Empty(tc.client.workers)
			assert.Empty(tc.client.controlPlanes)
			assert.Empty(tc.client.workerScaleSet)
			assert.Empty(tc.client.controlPlaneScaleSet)
		})
	}
}

func TestCreateInstances(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		publicIPAddressesAPI publicIPAddressesAPI
		networkInterfacesAPI networkInterfacesAPI
		scaleSetsAPI         scaleSetsAPI
		resourceGroupAPI     resourceGroupAPI
		roleAssignmentsAPI   roleAssignmentsAPI
		createInstancesInput CreateInstancesInput
		wantErr              bool
	}{
		"successful create": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			scaleSetsAPI: stubScaleSetsAPI{
				stubResponse: armcomputev2.VirtualMachineScaleSetsClientCreateOrUpdateResponse{
					VirtualMachineScaleSet: armcomputev2.VirtualMachineScaleSet{Identity: &armcomputev2.VirtualMachineScaleSetIdentity{PrincipalID: to.Ptr("principal-id")}},
				},
			},
			resourceGroupAPI:   newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountControlPlanes:   3,
				CountWorkers:         3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
				ConfidentialVM:       true,
			},
		},
		"error when creating scale set": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			scaleSetsAPI:         stubScaleSetsAPI{createErr: someErr},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountControlPlanes:   3,
				CountWorkers:         3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
				ConfidentialVM:       true,
			},
			wantErr: true,
		},
		"error when polling create scale set response": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			scaleSetsAPI:         stubScaleSetsAPI{getErr: someErr},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountControlPlanes:   3,
				CountWorkers:         3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
				ConfidentialVM:       true,
			},
			wantErr: true,
		},
		"error when retrieving private IPs": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{getErr: someErr},
			scaleSetsAPI:         stubScaleSetsAPI{},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountWorkers:         3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
				ConfidentialVM:       true,
			},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				location:             "location",
				name:                 "name",
				uid:                  "uid",
				resourceGroup:        "name",
				publicIPAddressesAPI: tc.publicIPAddressesAPI,
				networkInterfacesAPI: tc.networkInterfacesAPI,
				scaleSetsAPI:         tc.scaleSetsAPI,
				resourceGroupAPI:     tc.resourceGroupAPI,
				roleAssignmentsAPI:   tc.roleAssignmentsAPI,
				workers:              make(cloudtypes.Instances),
				controlPlanes:        make(cloudtypes.Instances),
				loadBalancerPubIP:    "lbip",
			}

			if tc.wantErr {
				assert.Error(client.CreateInstances(ctx, tc.createInstancesInput))
			} else {
				assert.NoError(client.CreateInstances(ctx, tc.createInstancesInput))
				assert.Equal(tc.createInstancesInput.CountControlPlanes, len(client.controlPlanes))
				assert.Equal(tc.createInstancesInput.CountWorkers, len(client.workers))
				assert.NotEmpty(client.workers["0"].PrivateIP)
				assert.NotEmpty(client.workers["0"].PublicIP)
				assert.NotEmpty(client.controlPlanes["0"].PrivateIP)
				assert.NotEmpty(client.controlPlanes["0"].PublicIP)
			}
		})
	}
}

func newSuccessfulResourceGroupStub() *stubResourceGroupAPI {
	return &stubResourceGroupAPI{
		getResourceGroup: armresources.ResourceGroup{
			ID: to.Ptr("resource-group-id"),
		},
	}
}

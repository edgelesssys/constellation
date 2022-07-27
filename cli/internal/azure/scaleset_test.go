package azure

import (
	"testing"

	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirewallPermissions(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	scaleSet := ScaleSet{
		Name:                 "name",
		NamePrefix:           "constellation-",
		Location:             "UK South",
		InstanceType:         "Standard_D2s_v3",
		Count:                3,
		Username:             "constellation",
		SubnetID:             "subnet-id",
		NetworkSecurityGroup: "network-security-group",
		Password:             "password",
		Image:                "image",
		UserAssignedIdentity: "user-identity",
	}

	scaleSetAzure := scaleSet.Azure()

	require.NotNil(scaleSetAzure.Name)
	assert.Equal(scaleSet.Name, *scaleSetAzure.Name)
	require.NotNil(scaleSetAzure.Location)
	assert.Equal(scaleSet.Location, *scaleSetAzure.Location)

	require.NotNil(scaleSetAzure.SKU)
	require.NotNil(scaleSetAzure.SKU.Name)
	assert.Equal(scaleSet.InstanceType, *scaleSetAzure.SKU.Name)

	require.NotNil(scaleSetAzure.SKU.Capacity)
	assert.Equal(scaleSet.Count, *scaleSetAzure.SKU.Capacity)

	require.NotNil(scaleSetAzure.Properties)
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile)
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.OSProfile)

	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.OSProfile.ComputerNamePrefix)
	assert.Equal(scaleSet.NamePrefix, *scaleSetAzure.Properties.VirtualMachineProfile.OSProfile.ComputerNamePrefix)
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.OSProfile.AdminUsername)
	assert.Equal(scaleSet.Username, *scaleSetAzure.Properties.VirtualMachineProfile.OSProfile.AdminUsername)
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.OSProfile.AdminPassword)
	assert.Equal(scaleSet.Password, *scaleSetAzure.Properties.VirtualMachineProfile.OSProfile.AdminPassword)

	// Verify image
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.StorageProfile)
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.StorageProfile.ImageReference)

	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.StorageProfile.ImageReference.ID)
	assert.Equal(scaleSet.Image, *scaleSetAzure.Properties.VirtualMachineProfile.StorageProfile.ImageReference.ID)

	// Verify network
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.NetworkProfile)
	require.Len(scaleSetAzure.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations, 1)
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations[0])

	networkConfig := scaleSetAzure.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations[0]

	require.NotNil(networkConfig.Name)
	assert.Equal(scaleSet.Name, *networkConfig.Name)

	require.NotNil(networkConfig.Properties)
	require.Len(networkConfig.Properties.IPConfigurations, 1)
	require.NotNil(networkConfig.Properties.IPConfigurations[0])

	ipConfig := networkConfig.Properties.IPConfigurations[0]

	require.NotNil(ipConfig.Name)
	assert.Equal(scaleSet.Name, *ipConfig.Name)

	require.NotNil(ipConfig.Properties)
	require.NotNil(ipConfig.Properties.Subnet)

	require.NotNil(ipConfig.Properties.Subnet.ID)
	assert.Equal(scaleSet.SubnetID, *ipConfig.Properties.Subnet.ID)

	require.NotNil(networkConfig.Properties.NetworkSecurityGroup)
	assert.Equal(scaleSet.NetworkSecurityGroup, *networkConfig.Properties.NetworkSecurityGroup.ID)

	// Verify vTPM
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.SecurityProfile)
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.SecurityProfile.SecurityType)
	assert.Equal(armcomputev2.SecurityTypesTrustedLaunch, *scaleSetAzure.Properties.VirtualMachineProfile.SecurityProfile.SecurityType)
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.SecurityProfile.UefiSettings)
	require.NotNil(scaleSetAzure.Properties.VirtualMachineProfile.SecurityProfile.UefiSettings.VTpmEnabled)
	assert.True(*scaleSetAzure.Properties.VirtualMachineProfile.SecurityProfile.UefiSettings.VTpmEnabled)

	// Verify UserAssignedIdentity
	require.NotNil(scaleSetAzure.Identity)
	require.NotNil(scaleSetAzure.Identity.Type)
	assert.Equal(armcomputev2.ResourceIdentityTypeUserAssigned, *scaleSetAzure.Identity.Type)
	require.Len(scaleSetAzure.Identity.UserAssignedIdentities, 1)
	assert.Contains(scaleSetAzure.Identity.UserAssignedIdentities, scaleSet.UserAssignedIdentity)
}

func TestGeneratePassword(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	pw, err := GeneratePassword()
	require.NoError(err)
	assert.Len(pw, 20)
}

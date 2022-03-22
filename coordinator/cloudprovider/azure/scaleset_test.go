package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetScaleSetVM(t *testing.T) {
	expectedInstance := core.Instance{
		Name:       "scale-set-name-instance-id",
		ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
		IPs:        []string{"192.0.2.0"},
		SSHKeys:    map[string][]string{"user": {"key-data"}},
	}
	testCases := map[string]struct {
		providerID                   string
		networkInterfacesAPI         networkInterfacesAPI
		virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
		expectErr                    bool
		expectedInstance             core.Instance
	}{
		"getVM for scale set instance works": {
			providerID:                   "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			expectedInstance:             expectedInstance,
		},
		"getVM for individual instance must fail": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			expectErr:  true,
		},
		"Get fails": {
			providerID:                   "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			virtualMachineScaleSetVMsAPI: newFailingGetScaleSetVirtualMachinesStub(),
			expectErr:                    true,
		},
		"retrieving interfaces fails": {
			providerID:                   "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			networkInterfacesAPI:         newFailingNetworkInterfacesStub(),
			expectErr:                    true,
		},
		"conversion fails": {
			providerID:                   "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			virtualMachineScaleSetVMsAPI: newGetInvalidScaleSetVirtualMachinesStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			expectErr:                    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				networkInterfacesAPI:         tc.networkInterfacesAPI,
				virtualMachineScaleSetVMsAPI: tc.virtualMachineScaleSetVMsAPI,
			}
			instance, err := metadata.getScaleSetVM(context.Background(), tc.providerID)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedInstance, instance)
		})
	}
}

func TestListScaleSetVMs(t *testing.T) {
	expectedInstances := []core.Instance{
		{
			Name:       "scale-set-name-instance-id",
			ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			IPs:        []string{"192.0.2.0"},
			SSHKeys:    map[string][]string{"user": {"key-data"}},
		},
	}
	testCases := map[string]struct {
		imdsAPI                      imdsAPI
		networkInterfacesAPI         networkInterfacesAPI
		virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
		scaleSetsAPI                 scaleSetsAPI
		expectErr                    bool
		expectedInstances            []core.Instance
	}{
		"listVMs works": {
			imdsAPI:                      newIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			expectedInstances:            expectedInstances,
		},
		"invalid scale sets are skipped": {
			imdsAPI:                      newIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			scaleSetsAPI:                 newListContainingNilScaleSetStub(),
			expectedInstances:            expectedInstances,
		},
		"listVMs can return 0 VMs": {
			imdsAPI:                      newIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: &stubVirtualMachineScaleSetVMsAPI{},
			scaleSetsAPI:                 newScaleSetsStub(),
			expectedInstances:            []core.Instance{},
		},
		"can skip nil in VM list": {
			imdsAPI:                      newIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newListContainingNilScaleSetVirtualMachinesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			expectedInstances:            expectedInstances,
		},
		"retrieving network interfaces fails": {
			imdsAPI:                      newIMDSStub(),
			networkInterfacesAPI:         newFailingNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			expectErr:                    true,
		},
		"converting instance fails": {
			imdsAPI:                      newIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newListContainingInvalidScaleSetVirtualMachinesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			expectErr:                    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI:                      tc.imdsAPI,
				networkInterfacesAPI:         tc.networkInterfacesAPI,
				virtualMachineScaleSetVMsAPI: tc.virtualMachineScaleSetVMsAPI,
				scaleSetsAPI:                 tc.scaleSetsAPI,
			}
			instances, err := metadata.listScaleSetVMs(context.Background(), "resource-group")

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.expectedInstances, instances)
		})
	}
}

func TestSplitScaleSetProviderID(t *testing.T) {
	testCases := map[string]struct {
		providerID             string
		expectErr              bool
		expectedSubscriptionID string
		expectedResourceGroup  string
		expectedScaleSet       string
		expectedInstanceID     string
	}{
		"providerID for scale set instance works": {
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			expectedSubscriptionID: "subscription-id",
			expectedResourceGroup:  "resource-group",
			expectedScaleSet:       "scale-set-name",
			expectedInstanceID:     "instance-id",
		},
		"providerID for individual instance must fail": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			expectErr:  true,
		},
		"providerID is malformed": {
			providerID: "malformed-provider-id",
			expectErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			subscriptionID, resourceGroup, scaleSet, instanceID, err := splitScaleSetProviderID(tc.providerID)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedSubscriptionID, subscriptionID)
			assert.Equal(tc.expectedResourceGroup, resourceGroup)
			assert.Equal(tc.expectedScaleSet, scaleSet)
			assert.Equal(tc.expectedInstanceID, instanceID)
		})
	}
}

func TestConvertScaleSetVMToCoreInstance(t *testing.T) {
	testCases := map[string]struct {
		inVM                 armcompute.VirtualMachineScaleSetVM
		inInterfaceIPConfigs []*armnetwork.InterfaceIPConfiguration
		expectErr            bool
		expectedInstance     core.Instance
	}{
		"conversion works": {
			inVM: armcompute.VirtualMachineScaleSetVM{
				Name: to.StringPtr("scale-set-name_instance-id"),
				ID:   to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
				Tags: map[string]*string{"tag-key": to.StringPtr("tag-value")},
				Properties: &armcompute.VirtualMachineScaleSetVMProperties{
					OSProfile: &armcompute.OSProfile{
						ComputerName: to.StringPtr("scale-set-name-instance-id"),
					},
				},
			},
			inInterfaceIPConfigs: []*armnetwork.InterfaceIPConfiguration{
				{
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAddress: to.StringPtr("192.0.2.0"),
					},
				},
			},
			expectedInstance: core.Instance{
				Name:       "scale-set-name-instance-id",
				ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
				IPs:        []string{"192.0.2.0"},
				SSHKeys:    map[string][]string{},
			},
		},
		"invalid instance": {
			inVM:      armcompute.VirtualMachineScaleSetVM{},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			instance, err := convertScaleSetVMToCoreInstance("scale-set", tc.inVM, tc.inInterfaceIPConfigs)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedInstance, instance)
		})
	}
}

func TestExtractScaleSetVMRole(t *testing.T) {
	testCases := map[string]struct {
		scaleSet     string
		expectedRole role.Role
	}{
		"coordinator role": {
			scaleSet:     "constellation-scale-set-coordinators-abcd123",
			expectedRole: role.Coordinator,
		},
		"node role": {
			scaleSet:     "constellation-scale-set-nodes-abcd123",
			expectedRole: role.Node,
		},
		"unknown role": {
			scaleSet:     "unknown",
			expectedRole: role.Unknown,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			role := extractScaleSetVMRole(tc.scaleSet)

			assert.Equal(tc.expectedRole, role)
		})
	}
}

func newFailingGetScaleSetVirtualMachinesStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		getErr: errors.New("get err"),
	}
}

func newGetInvalidScaleSetVirtualMachinesStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		getVM: armcompute.VirtualMachineScaleSetVM{},
	}
}

func newListContainingNilScaleSetVirtualMachinesStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		listPages: [][]*armcompute.VirtualMachineScaleSetVM{
			{
				nil,
				{
					Name:       to.StringPtr("scale-set-name_instance-id"),
					ID:         to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
					InstanceID: to.StringPtr("instance-id"),
					Tags: map[string]*string{
						"tag-key": to.StringPtr("tag-value"),
					},
					Properties: &armcompute.VirtualMachineScaleSetVMProperties{
						NetworkProfile: &armcompute.NetworkProfile{
							NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
								{
									ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
								},
							},
						},
						OSProfile: &armcompute.OSProfile{
							ComputerName: to.StringPtr("scale-set-name-instance-id"),
							LinuxConfiguration: &armcompute.LinuxConfiguration{
								SSH: &armcompute.SSHConfiguration{
									PublicKeys: []*armcompute.SSHPublicKey{
										{
											KeyData: to.StringPtr("key-data"),
											Path:    to.StringPtr("/home/user/.ssh/authorized_keys"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func newListContainingInvalidScaleSetVirtualMachinesStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		listPages: [][]*armcompute.VirtualMachineScaleSetVM{
			{
				{
					Name:       nil,
					ID:         nil,
					InstanceID: to.StringPtr("instance-id"),
					Properties: &armcompute.VirtualMachineScaleSetVMProperties{
						OSProfile: &armcompute.OSProfile{
							ComputerName: nil,
						},
						NetworkProfile: &armcompute.NetworkProfile{
							NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
								{
									ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func newListContainingNilScaleSetStub() *stubScaleSetsAPI {
	return &stubScaleSetsAPI{
		listPages: [][]*armcompute.VirtualMachineScaleSet{
			{
				nil,
				{Name: to.StringPtr("scale-set-name")},
			},
		},
	}
}

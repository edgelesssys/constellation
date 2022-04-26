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

func TestList(t *testing.T) {
	wantInstances := []core.Instance{
		{
			Name:       "instance-name",
			ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			IPs:        []string{"192.0.2.0"},
			SSHKeys:    map[string][]string{"user": {"key-data"}},
		},
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
		scaleSetsAPI                 scaleSetsAPI
		virtualMachinesAPI           virtualMachinesAPI
		virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
		tagsAPI                      tagsAPI
		wantErr                      bool
		wantInstances                []core.Instance
	}{
		"List works": {
			imdsAPI:                      newIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			virtualMachinesAPI:           newVirtualMachinesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			tagsAPI:                      newTagsStub(),
			wantInstances:                wantInstances,
		},
		"providerID cannot be retrieved": {
			imdsAPI: &stubIMDSAPI{retrieveErr: errors.New("imds err")},
			wantErr: true,
		},
		"providerID cannot be parsed": {
			imdsAPI: newInvalidIMDSStub(),
			wantErr: true,
		},
		"listVMs fails": {
			imdsAPI:            newIMDSStub(),
			virtualMachinesAPI: newFailingListsVirtualMachinesStub(),
			wantErr:            true,
		},
		"listScaleSetVMs fails": {
			imdsAPI:                      newIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			virtualMachinesAPI:           newVirtualMachinesStub(),
			virtualMachineScaleSetVMsAPI: newFailingListsVirtualMachineScaleSetsVMsStub(),
			tagsAPI:                      newTagsStub(),
			wantErr:                      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI:                      tc.imdsAPI,
				networkInterfacesAPI:         tc.networkInterfacesAPI,
				scaleSetsAPI:                 tc.scaleSetsAPI,
				virtualMachinesAPI:           tc.virtualMachinesAPI,
				virtualMachineScaleSetVMsAPI: tc.virtualMachineScaleSetVMsAPI,
				tagsAPI:                      tc.tagsAPI,
			}
			instances, err := metadata.List(context.Background())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantInstances, instances)
		})
	}
}

func TestSelf(t *testing.T) {
	wantVMInstance := core.Instance{
		Name:       "instance-name",
		ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
		IPs:        []string{"192.0.2.0"},
		SSHKeys:    map[string][]string{"user": {"key-data"}},
	}
	wantScaleSetInstance := core.Instance{
		Name:       "scale-set-name-instance-id",
		ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
		IPs:        []string{"192.0.2.0"},
		SSHKeys:    map[string][]string{"user": {"key-data"}},
	}
	testCases := map[string]struct {
		imdsAPI                      imdsAPI
		networkInterfacesAPI         networkInterfacesAPI
		virtualMachinesAPI           virtualMachinesAPI
		virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
		wantErr                      bool
		wantInstance                 core.Instance
	}{
		"self for individual instance works": {
			imdsAPI:                      newIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachinesAPI:           newVirtualMachinesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			wantInstance:                 wantVMInstance,
		},
		"self for scale set instance works": {
			imdsAPI:                      newScaleSetIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachinesAPI:           newVirtualMachinesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			wantInstance:                 wantScaleSetInstance,
		},
		"providerID cannot be retrieved": {
			imdsAPI: &stubIMDSAPI{retrieveErr: errors.New("imds err")},
			wantErr: true,
		},
		"GetInstance fails": {
			imdsAPI:            newIMDSStub(),
			virtualMachinesAPI: newFailingGetVirtualMachinesStub(),
			wantErr:            true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI:                      tc.imdsAPI,
				networkInterfacesAPI:         tc.networkInterfacesAPI,
				virtualMachinesAPI:           tc.virtualMachinesAPI,
				virtualMachineScaleSetVMsAPI: tc.virtualMachineScaleSetVMsAPI,
			}
			instance, err := metadata.Self(context.Background())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantInstance, instance)
		})
	}
}

func TestSignalRole(t *testing.T) {
	testCases := map[string]struct {
		imdsAPI imdsAPI
		tagsAPI tagsAPI
		wantErr bool
	}{
		"SignalRole works": {
			imdsAPI: newIMDSStub(),
			tagsAPI: newTagsStub(),
		},
		"SignalRole is not attempted on scale set vm": {
			imdsAPI: newScaleSetIMDSStub(),
		},
		"providerID cannot be retrieved": {
			imdsAPI: &stubIMDSAPI{retrieveErr: errors.New("imds err")},
			wantErr: true,
		},
		"setting tag fails": {
			imdsAPI: newIMDSStub(),
			tagsAPI: newFailingTagsStub(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI: tc.imdsAPI,
				tagsAPI: tc.tagsAPI,
			}
			err := metadata.SignalRole(context.Background(), role.Coordinator)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestSetVPNIP(t *testing.T) {
	assert := assert.New(t)
	metadata := Metadata{}
	assert.NoError(metadata.SetVPNIP(context.Background(), "192.0.2.0"))
}

func TestMetadataSupported(t *testing.T) {
	assert := assert.New(t)
	metadata := Metadata{}
	assert.True(metadata.Supported())
}

func TestProviderID(t *testing.T) {
	testCases := map[string]struct {
		imdsAPI        imdsAPI
		wantErr        bool
		wantProviderID string
	}{
		"providerID for individual instance works": {
			imdsAPI:        newIMDSStub(),
			wantProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
		},
		"providerID for scale set instance works": {
			imdsAPI:        newScaleSetIMDSStub(),
			wantProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
		},
		"imds retrieval fails": {
			imdsAPI: &stubIMDSAPI{retrieveErr: errors.New("imds err")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI: tc.imdsAPI,
			}
			providerID, err := metadata.providerID(context.Background())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantProviderID, providerID)
		})
	}
}

func TestExtractBasicsFromProviderID(t *testing.T) {
	testCases := map[string]struct {
		providerID         string
		wantErr            bool
		wantSubscriptionID string
		wantResourceGroup  string
	}{
		"providerID for individual instance works": {
			providerID:         "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			wantSubscriptionID: "subscription-id",
			wantResourceGroup:  "resource-group",
		},
		"providerID for scale set instance works": {
			providerID:         "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			wantSubscriptionID: "subscription-id",
			wantResourceGroup:  "resource-group",
		},
		"providerID is malformed": {
			providerID: "malformed-provider-id",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			subscriptionID, resourceGroup, err := extractBasicsFromProviderID(tc.providerID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantSubscriptionID, subscriptionID)
			assert.Equal(tc.wantResourceGroup, resourceGroup)
		})
	}
}

func TestExtractInstanceTags(t *testing.T) {
	testCases := map[string]struct {
		in       map[string]*string
		wantTags map[string]string
	}{
		"tags are extracted": {
			in:       map[string]*string{"key": to.StringPtr("value")},
			wantTags: map[string]string{"key": "value"},
		},
		"nil values are skipped": {
			in:       map[string]*string{"key": nil},
			wantTags: map[string]string{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			tags := extractInstanceTags(tc.in)

			assert.Equal(tc.wantTags, tags)
		})
	}
}

func TestExtractSSHKeys(t *testing.T) {
	testCases := map[string]struct {
		in       armcompute.SSHConfiguration
		wantKeys map[string][]string
	}{
		"ssh key is extracted": {
			in: armcompute.SSHConfiguration{
				PublicKeys: []*armcompute.SSHPublicKey{
					{
						KeyData: to.StringPtr("key-data"),
						Path:    to.StringPtr("/home/user/.ssh/authorized_keys"),
					},
				},
			},
			wantKeys: map[string][]string{"user": {"key-data"}},
		},
		"invalid path is skipped": {
			in: armcompute.SSHConfiguration{
				PublicKeys: []*armcompute.SSHPublicKey{
					{
						KeyData: to.StringPtr("key-data"),
						Path:    to.StringPtr("invalid-path"),
					},
				},
			},
			wantKeys: map[string][]string{},
		},
		"key data is nil": {
			in: armcompute.SSHConfiguration{
				PublicKeys: []*armcompute.SSHPublicKey{
					{
						Path: to.StringPtr("/home/user/.ssh/authorized_keys"),
					},
				},
			},
			wantKeys: map[string][]string{},
		},
		"path is nil": {
			in: armcompute.SSHConfiguration{
				PublicKeys: []*armcompute.SSHPublicKey{
					{
						KeyData: to.StringPtr("key-data"),
					},
				},
			},
			wantKeys: map[string][]string{},
		},
		"public keys are nil": {
			in:       armcompute.SSHConfiguration{},
			wantKeys: map[string][]string{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			keys := extractSSHKeys(tc.in)

			assert.Equal(tc.wantKeys, keys)
		})
	}
}

func newIMDSStub() *stubIMDSAPI {
	return &stubIMDSAPI{
		res: metadataResponse{Compute: struct {
			ResourceID string `json:"resourceId,omitempty"`
		}{"/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"}},
	}
}

func newScaleSetIMDSStub() *stubIMDSAPI {
	return &stubIMDSAPI{
		res: metadataResponse{Compute: struct {
			ResourceID string `json:"resourceId,omitempty"`
		}{"/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"}},
	}
}

func newInvalidIMDSStub() *stubIMDSAPI {
	return &stubIMDSAPI{
		res: metadataResponse{Compute: struct {
			ResourceID string `json:"resourceId,omitempty"`
		}{"invalid-resource-id"}},
	}
}

func newFailingIMDSStub() *stubIMDSAPI {
	return &stubIMDSAPI{
		retrieveErr: errors.New("imds retrieve error"),
	}
}

func newNetworkInterfacesStub() *stubNetworkInterfacesAPI {
	return &stubNetworkInterfacesAPI{
		getInterface: armnetwork.Interface{
			Name: to.StringPtr("interface-name"),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAddress: to.StringPtr("192.0.2.0"),
						},
					},
				},
			},
		},
	}
}

func newScaleSetsStub() *stubScaleSetsAPI {
	return &stubScaleSetsAPI{
		listPages: [][]*armcompute.VirtualMachineScaleSet{
			{
				&armcompute.VirtualMachineScaleSet{
					Name: to.StringPtr("scale-set-name"),
				},
			},
		},
	}
}

func newVirtualMachinesStub() *stubVirtualMachinesAPI {
	return &stubVirtualMachinesAPI{
		getVM: armcompute.VirtualMachine{
			Name: to.StringPtr("instance-name"),
			ID:   to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"),
			Properties: &armcompute.VirtualMachineProperties{
				NetworkProfile: &armcompute.NetworkProfile{
					NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
						{
							ID: to.StringPtr("/subscriptions/subscription/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
						},
					},
				},
				OSProfile: &armcompute.OSProfile{
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
		listPages: [][]*armcompute.VirtualMachine{
			{
				{
					Name: to.StringPtr("instance-name"),
					ID:   to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"),
					Properties: &armcompute.VirtualMachineProperties{
						NetworkProfile: &armcompute.NetworkProfile{
							NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
								{
									ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
								},
							},
						},
						OSProfile: &armcompute.OSProfile{
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

func newFailingListsVirtualMachinesStub() *stubVirtualMachinesAPI {
	return &stubVirtualMachinesAPI{
		listPages: [][]*armcompute.VirtualMachine{
			{
				{},
			},
		},
	}
}

func newFailingGetVirtualMachinesStub() *stubVirtualMachinesAPI {
	return &stubVirtualMachinesAPI{
		getErr: errors.New("get err"),
	}
}

func newVirtualMachineScaleSetsVMsStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		getVM: armcompute.VirtualMachineScaleSetVM{
			Name:       to.StringPtr("scale-set-name_instance-id"),
			InstanceID: to.StringPtr("instance-id"),
			ID:         to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
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
		listPages: [][]*armcompute.VirtualMachineScaleSetVM{
			{
				&armcompute.VirtualMachineScaleSetVM{
					Name:       to.StringPtr("scale-set-name_instance-id"),
					InstanceID: to.StringPtr("instance-id"),
					ID:         to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
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

func newFailingListsVirtualMachineScaleSetsVMsStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		listPages: [][]*armcompute.VirtualMachineScaleSetVM{
			{
				{
					InstanceID: to.StringPtr("invalid-instance-id"),
				},
			},
		},
	}
}

func newTagsStub() *stubTagsAPI {
	return &stubTagsAPI{}
}

func newFailingTagsStub() *stubTagsAPI {
	return &stubTagsAPI{
		createOrUpdateAtScopeErr: errors.New("createOrUpdateErr"),
		updateAtScopeErr:         errors.New("updateErr"),
	}
}

package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	wantInstances := []metadata.InstanceMetadata{
		{
			Name:       "scale-set-name-instance-id",
			ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			VPCIP:      "192.0.2.0",
			SSHKeys:    map[string][]string{"user": {"key-data"}},
		},
	}
	testCases := map[string]struct {
		imdsAPI                      imdsAPI
		networkInterfacesAPI         networkInterfacesAPI
		scaleSetsAPI                 scaleSetsAPI
		virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
		tagsAPI                      tagsAPI
		wantErr                      bool
		wantInstances                []metadata.InstanceMetadata
	}{
		"List works": {
			imdsAPI:                      newScaleSetIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
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
		"listScaleSetVMs fails": {
			imdsAPI:                      newScaleSetIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			virtualMachineScaleSetVMsAPI: newFailingListsVirtualMachineScaleSetsVMsStub(),
			tagsAPI:                      newTagsStub(),
			wantErr:                      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			azureMetadata := Metadata{
				imdsAPI:                      tc.imdsAPI,
				networkInterfacesAPI:         tc.networkInterfacesAPI,
				scaleSetsAPI:                 tc.scaleSetsAPI,
				virtualMachineScaleSetVMsAPI: tc.virtualMachineScaleSetVMsAPI,
				tagsAPI:                      tc.tagsAPI,
			}
			instances, err := azureMetadata.List(context.Background())

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
	wantScaleSetInstance := metadata.InstanceMetadata{
		Name:       "scale-set-name-instance-id",
		ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
		VPCIP:      "192.0.2.0",
		SSHKeys:    map[string][]string{"user": {"key-data"}},
	}
	testCases := map[string]struct {
		imdsAPI                      imdsAPI
		networkInterfacesAPI         networkInterfacesAPI
		virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
		wantErr                      bool
		wantInstance                 metadata.InstanceMetadata
	}{
		"self for scale set instance works": {
			imdsAPI:                      newScaleSetIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			wantInstance:                 wantScaleSetInstance,
		},
		"providerID cannot be retrieved": {
			imdsAPI: &stubIMDSAPI{retrieveErr: errors.New("imds err")},
			wantErr: true,
		},
		"GetInstance fails": {
			imdsAPI:                      newScaleSetIMDSStub(),
			virtualMachineScaleSetVMsAPI: &stubVirtualMachineScaleSetVMsAPI{getErr: errors.New("failed")},
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

func TestGetNetworkSecurityGroupName(t *testing.T) {
	name := "network-security-group-name"
	testCases := map[string]struct {
		securityGroupsAPI securityGroupsAPI
		imdsAPI           imdsAPI
		wantName          string
		wantErr           bool
	}{
		"GetNetworkSecurityGroupName works": {
			imdsAPI: newScaleSetIMDSStub(),
			securityGroupsAPI: &stubSecurityGroupsAPI{
				listPages: [][]*armnetwork.SecurityGroup{
					{
						{
							Name: to.StringPtr(name),
						},
					},
				},
			},
			wantName: name,
		},
		"no security group": {
			imdsAPI:           newScaleSetIMDSStub(),
			securityGroupsAPI: &stubSecurityGroupsAPI{},
			wantErr:           true,
		},
		"missing name in security group struct": {
			imdsAPI:           newScaleSetIMDSStub(),
			securityGroupsAPI: &stubSecurityGroupsAPI{listPages: [][]*armnetwork.SecurityGroup{{{}}}},
			wantErr:           true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI:           tc.imdsAPI,
				securityGroupsAPI: tc.securityGroupsAPI,
			}
			name, err := metadata.GetNetworkSecurityGroupName(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantName, name)
		})
	}
}

func TestGetSubnetworkCIDR(t *testing.T) {
	subnetworkCIDR := "192.0.2.0/24"
	name := "name"
	testCases := map[string]struct {
		virtualNetworksAPI virtualNetworksAPI
		imdsAPI            imdsAPI
		wantNetworkCIDR    string
		wantErr            bool
	}{
		"GetSubnetworkCIDR works": {
			imdsAPI: newScaleSetIMDSStub(),
			virtualNetworksAPI: &stubVirtualNetworksAPI{listPages: [][]*armnetwork.VirtualNetwork{
				{
					{
						Name: to.StringPtr(name),
						Properties: &armnetwork.VirtualNetworkPropertiesFormat{
							Subnets: []*armnetwork.Subnet{
								{Properties: &armnetwork.SubnetPropertiesFormat{AddressPrefix: to.StringPtr(subnetworkCIDR)}},
							},
						},
					},
				},
			}},
			wantNetworkCIDR: subnetworkCIDR,
		},
		"no virtual networks found": {
			imdsAPI: newScaleSetIMDSStub(),
			virtualNetworksAPI: &stubVirtualNetworksAPI{listPages: [][]*armnetwork.VirtualNetwork{
				{},
			}},
			wantErr:         true,
			wantNetworkCIDR: subnetworkCIDR,
		},
		"malformed network struct": {
			imdsAPI: newScaleSetIMDSStub(),
			virtualNetworksAPI: &stubVirtualNetworksAPI{listPages: [][]*armnetwork.VirtualNetwork{
				{
					{},
				},
			}},
			wantErr:         true,
			wantNetworkCIDR: subnetworkCIDR,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI:            tc.imdsAPI,
				virtualNetworksAPI: tc.virtualNetworksAPI,
			}
			subnetworkCIDR, err := metadata.GetSubnetworkCIDR(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantNetworkCIDR, subnetworkCIDR)
		})
	}
}

func TestGetLoadBalancerName(t *testing.T) {
	loadBalancerName := "load-balancer-name"
	testCases := map[string]struct {
		loadBalancerAPI loadBalancerAPI
		imdsAPI         imdsAPI
		wantName        string
		wantErr         bool
	}{
		"GetLoadBalancerName works": {
			imdsAPI: newScaleSetIMDSStub(),
			loadBalancerAPI: &stubLoadBalancersAPI{
				listPages: [][]*armnetwork.LoadBalancer{
					{
						{
							Name:       to.StringPtr(loadBalancerName),
							Properties: &armnetwork.LoadBalancerPropertiesFormat{},
						},
					},
				},
			},
			wantName: loadBalancerName,
		},
		"invalid load balancer struct": {
			imdsAPI:         newScaleSetIMDSStub(),
			loadBalancerAPI: &stubLoadBalancersAPI{listPages: [][]*armnetwork.LoadBalancer{{{}}}},
			wantErr:         true,
		},
		"invalid missing name": {
			imdsAPI: newScaleSetIMDSStub(),
			loadBalancerAPI: &stubLoadBalancersAPI{listPages: [][]*armnetwork.LoadBalancer{{{
				Properties: &armnetwork.LoadBalancerPropertiesFormat{},
			}}}},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI:         tc.imdsAPI,
				loadBalancerAPI: tc.loadBalancerAPI,
			}
			loadbalancerName, err := metadata.GetLoadBalancerName(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantName, loadbalancerName)
		})
	}
}

func TestGetLoadBalancerIP(t *testing.T) {
	loadBalancerName := "load-balancer-name"
	publicIP := "192.0.2.1"
	correctPublicIPID := "/subscriptions/subscription/resourceGroups/resourceGroup/providers/Microsoft.Network/publicIPAddresses/pubIPName"
	someErr := errors.New("some error")
	testCases := map[string]struct {
		loadBalancerAPI      loadBalancerAPI
		publicIPAddressesAPI publicIPAddressesAPI
		imdsAPI              imdsAPI
		wantIP               string
		wantErr              bool
	}{
		"GetLoadBalancerIP works": {
			imdsAPI: newScaleSetIMDSStub(),
			loadBalancerAPI: &stubLoadBalancersAPI{
				listPages: [][]*armnetwork.LoadBalancer{
					{
						{
							Name: to.StringPtr(loadBalancerName),
							Properties: &armnetwork.LoadBalancerPropertiesFormat{
								FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
									{
										Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
											PublicIPAddress: &armnetwork.PublicIPAddress{
												ID: &correctPublicIPID,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: armnetwork.PublicIPAddressesClientGetResponse{
				PublicIPAddressesClientGetResult: armnetwork.PublicIPAddressesClientGetResult{
					PublicIPAddress: armnetwork.PublicIPAddress{
						Properties: &armnetwork.PublicIPAddressPropertiesFormat{
							IPAddress: &publicIP,
						},
					},
				},
			}},
			wantIP: publicIP,
		},
		"no load balancer": {
			imdsAPI: newScaleSetIMDSStub(),
			loadBalancerAPI: &stubLoadBalancersAPI{
				listPages: [][]*armnetwork.LoadBalancer{
					{},
				},
			},
			wantErr: true,
		},
		"load balancer missing public IP reference": {
			imdsAPI: newScaleSetIMDSStub(),
			loadBalancerAPI: &stubLoadBalancersAPI{
				listPages: [][]*armnetwork.LoadBalancer{
					{
						{
							Name: to.StringPtr(loadBalancerName),
							Properties: &armnetwork.LoadBalancerPropertiesFormat{
								FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"public IP reference has wrong format": {
			imdsAPI: newScaleSetIMDSStub(),
			loadBalancerAPI: &stubLoadBalancersAPI{
				listPages: [][]*armnetwork.LoadBalancer{
					{
						{
							Name: to.StringPtr(loadBalancerName),
							Properties: &armnetwork.LoadBalancerPropertiesFormat{
								FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
									{
										Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
											PublicIPAddress: &armnetwork.PublicIPAddress{
												ID: to.StringPtr("wrong-format"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"no public IP address found": {
			imdsAPI: newScaleSetIMDSStub(),
			loadBalancerAPI: &stubLoadBalancersAPI{
				listPages: [][]*armnetwork.LoadBalancer{
					{
						{
							Name: to.StringPtr(loadBalancerName),
							Properties: &armnetwork.LoadBalancerPropertiesFormat{
								FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
									{
										Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
											PublicIPAddress: &armnetwork.PublicIPAddress{
												ID: &correctPublicIPID,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getErr: someErr},
			wantErr:              true,
		},
		"found public IP has no address field": {
			imdsAPI: newScaleSetIMDSStub(),
			loadBalancerAPI: &stubLoadBalancersAPI{
				listPages: [][]*armnetwork.LoadBalancer{
					{
						{
							Name: to.StringPtr(loadBalancerName),
							Properties: &armnetwork.LoadBalancerPropertiesFormat{
								FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
									{
										Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
											PublicIPAddress: &armnetwork.PublicIPAddress{
												ID: &correctPublicIPID,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: armnetwork.PublicIPAddressesClientGetResponse{
				PublicIPAddressesClientGetResult: armnetwork.PublicIPAddressesClientGetResult{
					PublicIPAddress: armnetwork.PublicIPAddress{
						Properties: &armnetwork.PublicIPAddressPropertiesFormat{},
					},
				},
			}},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI:              tc.imdsAPI,
				loadBalancerAPI:      tc.loadBalancerAPI,
				publicIPAddressesAPI: tc.publicIPAddressesAPI,
			}
			loadbalancerName, err := metadata.GetLoadBalancerIP(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantIP, loadbalancerName)
		})
	}
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

func newNetworkInterfacesStub() *stubNetworkInterfacesAPI {
	return &stubNetworkInterfacesAPI{
		getInterface: armnetwork.Interface{
			Name: to.StringPtr("interface-name"),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAddress: to.StringPtr("192.0.2.0"),
							Primary:          to.BoolPtr(true),
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

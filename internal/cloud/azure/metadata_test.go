/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	wantInstances := []metadata.InstanceMetadata{
		{
			Name:       "scale-set-name-instance-id",
			ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			Role:       role.Worker,
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
			imdsAPI:                      &stubIMDSAPI{resourceGroup: "resourceGroup"},
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			tagsAPI:                      newTagsStub(),
			wantInstances:                wantInstances,
		},
		"imds resource group fails": {
			imdsAPI: &stubIMDSAPI{resourceGroupErr: errors.New("failed")},
			wantErr: true,
		},
		"listScaleSetVMs fails": {
			imdsAPI:                      &stubIMDSAPI{resourceGroup: "resourceGroup"},
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
		Role:       role.Worker,
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
			imdsAPI:                      &stubIMDSAPI{providerID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"},
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			wantInstance:                 wantScaleSetInstance,
		},
		"providerID cannot be retrieved": {
			imdsAPI: &stubIMDSAPI{providerIDErr: errors.New("failed")},
			wantErr: true,
		},
		"GetInstance fails": {
			imdsAPI:                      &stubIMDSAPI{providerID: wantScaleSetInstance.ProviderID},
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
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			securityGroupsAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{{Name: to.Ptr(name)}},
				},
			},
			wantName: name,
		},
		"no security group": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			securityGroupsAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{},
			},
			wantErr: true,
		},
		"missing name in security group struct": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			securityGroupsAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{{}},
				},
			},
			wantErr: true,
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
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			virtualNetworksAPI: &stubVirtualNetworksAPI{
				pager: &stubVirtualNetworksClientListPager{
					list: []armnetwork.VirtualNetwork{{
						Name: to.Ptr(name),
						Properties: &armnetwork.VirtualNetworkPropertiesFormat{
							Subnets: []*armnetwork.Subnet{
								{Properties: &armnetwork.SubnetPropertiesFormat{AddressPrefix: to.Ptr(subnetworkCIDR)}},
							},
						},
					}},
				},
			},
			wantNetworkCIDR: subnetworkCIDR,
		},
		"no virtual networks found": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			virtualNetworksAPI: &stubVirtualNetworksAPI{
				pager: &stubVirtualNetworksClientListPager{},
			},
			wantErr:         true,
			wantNetworkCIDR: subnetworkCIDR,
		},
		"malformed network struct": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			virtualNetworksAPI: &stubVirtualNetworksAPI{
				pager: &stubVirtualNetworksClientListPager{list: []armnetwork.VirtualNetwork{{}}},
			},
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
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{{
						Name:       to.Ptr(loadBalancerName),
						Properties: &armnetwork.LoadBalancerPropertiesFormat{},
					}},
				},
			},
			wantName: loadBalancerName,
		},
		"invalid load balancer struct": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{list: []armnetwork.LoadBalancer{{}}},
			},
			wantErr: true,
		},
		"invalid missing name": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{list: []armnetwork.LoadBalancer{{
					Properties: &armnetwork.LoadBalancerPropertiesFormat{},
				}}},
			},
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

func TestGetLoadBalancerEndpoint(t *testing.T) {
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
		"GetLoadBalancerEndpoint works": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{{
						Name: to.Ptr(loadBalancerName),
						Properties: &armnetwork.LoadBalancerPropertiesFormat{
							FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
								{
									Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
										PublicIPAddress: &armnetwork.PublicIPAddress{ID: &correctPublicIPID},
									},
								},
							},
						},
					}},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: armnetwork.PublicIPAddressesClientGetResponse{
				PublicIPAddress: armnetwork.PublicIPAddress{
					Properties: &armnetwork.PublicIPAddressPropertiesFormat{
						IPAddress: &publicIP,
					},
				},
			}},
			wantIP: publicIP,
		},
		"no load balancer": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{},
			},
			wantErr: true,
		},
		"load balancer missing public IP reference": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{{
						Name: to.Ptr(loadBalancerName),
						Properties: &armnetwork.LoadBalancerPropertiesFormat{
							FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{},
						},
					}},
				},
			},
			wantErr: true,
		},
		"public IP reference has wrong format": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{{
						Name: to.Ptr(loadBalancerName),
						Properties: &armnetwork.LoadBalancerPropertiesFormat{
							FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
								{
									Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
										PublicIPAddress: &armnetwork.PublicIPAddress{
											ID: to.Ptr("wrong-format"),
										},
									},
								},
							},
						},
					}},
				},
			},
			wantErr: true,
		},
		"no public IP address found": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{{
						Name: to.Ptr(loadBalancerName),
						Properties: &armnetwork.LoadBalancerPropertiesFormat{
							FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
								{
									Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
										PublicIPAddress: &armnetwork.PublicIPAddress{ID: &correctPublicIPID},
									},
								},
							},
						},
					}},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getErr: someErr},
			wantErr:              true,
		},
		"found public IP has no address field": {
			imdsAPI: &stubIMDSAPI{resourceGroup: "resourceGroup"},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{{
						Name: to.Ptr(loadBalancerName),
						Properties: &armnetwork.LoadBalancerPropertiesFormat{
							FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
								{
									Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
										PublicIPAddress: &armnetwork.PublicIPAddress{ID: &correctPublicIPID},
									},
								},
							},
						},
					}},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: armnetwork.PublicIPAddressesClientGetResponse{
				PublicIPAddress: armnetwork.PublicIPAddress{
					Properties: &armnetwork.PublicIPAddressPropertiesFormat{},
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
			loadbalancerName, err := metadata.GetLoadBalancerEndpoint(context.Background())
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
			imdsAPI:        &stubIMDSAPI{providerID: "provider-id"},
			wantProviderID: "azure://provider-id",
		},
		"imds providerID fails": {
			imdsAPI: &stubIMDSAPI{providerIDErr: errors.New("failed")},
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

func TestUID(t *testing.T) {
	testCases := map[string]struct {
		imdsAPI imdsAPI
		wantErr bool
		wantUID string
	}{
		"success": {
			imdsAPI: &stubIMDSAPI{uid: "uid"},
			wantUID: "uid",
		},
		"imds uid error": {
			imdsAPI: &stubIMDSAPI{uidErr: errors.New("failed")},
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
			uid, err := metadata.UID(context.Background())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantUID, uid)
		})
	}
}

func TestExtractInstanceTags(t *testing.T) {
	testCases := map[string]struct {
		in       map[string]*string
		wantTags map[string]string
	}{
		"tags are extracted": {
			in:       map[string]*string{"key": to.Ptr("value")},
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
		in       armcomputev2.SSHConfiguration
		wantKeys map[string][]string
	}{
		"ssh key is extracted": {
			in: armcomputev2.SSHConfiguration{
				PublicKeys: []*armcomputev2.SSHPublicKey{
					{
						KeyData: to.Ptr("key-data"),
						Path:    to.Ptr("/home/user/.ssh/authorized_keys"),
					},
				},
			},
			wantKeys: map[string][]string{"user": {"key-data"}},
		},
		"invalid path is skipped": {
			in: armcomputev2.SSHConfiguration{
				PublicKeys: []*armcomputev2.SSHPublicKey{
					{
						KeyData: to.Ptr("key-data"),
						Path:    to.Ptr("invalid-path"),
					},
				},
			},
			wantKeys: map[string][]string{},
		},
		"key data is nil": {
			in: armcomputev2.SSHConfiguration{
				PublicKeys: []*armcomputev2.SSHPublicKey{
					{
						Path: to.Ptr("/home/user/.ssh/authorized_keys"),
					},
				},
			},
			wantKeys: map[string][]string{},
		},
		"path is nil": {
			in: armcomputev2.SSHConfiguration{
				PublicKeys: []*armcomputev2.SSHPublicKey{
					{
						KeyData: to.Ptr("key-data"),
					},
				},
			},
			wantKeys: map[string][]string{},
		},
		"public keys are nil": {
			in:       armcomputev2.SSHConfiguration{},
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

func newNetworkInterfacesStub() *stubNetworkInterfacesAPI {
	return &stubNetworkInterfacesAPI{
		getInterface: armnetwork.Interface{
			Name: to.Ptr("interface-name"),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAddress: to.Ptr("192.0.2.0"),
							Primary:          to.Ptr(true),
						},
					},
				},
			},
		},
	}
}

func newScaleSetsStub() *stubScaleSetsAPI {
	return &stubScaleSetsAPI{
		pager: &stubVirtualMachineScaleSetsClientListPager{
			list: []armcomputev2.VirtualMachineScaleSet{{
				Name: to.Ptr("scale-set-name"),
				Tags: map[string]*string{
					"constellation-uid": to.Ptr("uid"),
					"role":              to.Ptr("worker"),
				},
			}},
		},
	}
}

func newVirtualMachineScaleSetsVMsStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		getVM: armcomputev2.VirtualMachineScaleSetVM{
			Name:       to.Ptr("scale-set-name_instance-id"),
			InstanceID: to.Ptr("instance-id"),
			ID:         to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
			Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
				NetworkProfile: &armcomputev2.NetworkProfile{
					NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
						{
							ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
						},
					},
				},
				OSProfile: &armcomputev2.OSProfile{
					ComputerName: to.Ptr("scale-set-name-instance-id"),
					LinuxConfiguration: &armcomputev2.LinuxConfiguration{
						SSH: &armcomputev2.SSHConfiguration{
							PublicKeys: []*armcomputev2.SSHPublicKey{
								{
									KeyData: to.Ptr("key-data"),
									Path:    to.Ptr("/home/user/.ssh/authorized_keys"),
								},
							},
						},
					},
				},
			},
			Tags: map[string]*string{
				"constellation-uid": to.Ptr("uid"),
				"role":              to.Ptr("worker"),
			},
		},
		pager: &stubVirtualMachineScaleSetVMPager{
			list: []armcomputev2.VirtualMachineScaleSetVM{
				{
					Name:       to.Ptr("scale-set-name_instance-id"),
					InstanceID: to.Ptr("instance-id"),
					ID:         to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
					Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
						NetworkProfile: &armcomputev2.NetworkProfile{
							NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
								{
									ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
								},
							},
						},
						OSProfile: &armcomputev2.OSProfile{
							ComputerName: to.Ptr("scale-set-name-instance-id"),
							LinuxConfiguration: &armcomputev2.LinuxConfiguration{
								SSH: &armcomputev2.SSHConfiguration{
									PublicKeys: []*armcomputev2.SSHPublicKey{
										{
											KeyData: to.Ptr("key-data"),
											Path:    to.Ptr("/home/user/.ssh/authorized_keys"),
										},
									},
								},
							},
						},
					},
					Tags: map[string]*string{
						"constellation-uid": to.Ptr("uid"),
						"role":              to.Ptr("worker"),
					},
				},
			},
		},
	}
}

func newFailingListsVirtualMachineScaleSetsVMsStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		pager: &stubVirtualMachineScaleSetVMPager{
			list: []armcomputev2.VirtualMachineScaleSetVM{{
				InstanceID: to.Ptr("invalid-instance-id"),
			}},
		},
	}
}

func newTagsStub() *stubTagsAPI {
	return &stubTagsAPI{}
}

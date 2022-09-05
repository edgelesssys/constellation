/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"errors"
	"testing"

	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVMInterfaces(t *testing.T) {
	wantNetworkInterfaces := []armnetwork.Interface{
		{
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
	vm := armcomputev2.VirtualMachine{
		Properties: &armcomputev2.VirtualMachineProperties{
			NetworkProfile: &armcomputev2.NetworkProfile{
				NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
					{
						ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
					},
				},
			},
		},
	}
	testCases := map[string]struct {
		vm                    armcomputev2.VirtualMachine
		networkInterfacesAPI  networkInterfacesAPI
		wantErr               bool
		wantNetworkInterfaces []armnetwork.Interface
	}{
		"retrieval works": {
			vm: vm,
			networkInterfacesAPI: &stubNetworkInterfacesAPI{
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
			},
			wantNetworkInterfaces: wantNetworkInterfaces,
		},
		"vm can have 0 interfaces": {
			vm: armcomputev2.VirtualMachine{},
			networkInterfacesAPI: &stubNetworkInterfacesAPI{
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
			},
			wantNetworkInterfaces: []armnetwork.Interface{},
		},
		"interface retrieval fails": {
			vm: vm,
			networkInterfacesAPI: &stubNetworkInterfacesAPI{
				getErr: errors.New("get err"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				networkInterfacesAPI: tc.networkInterfacesAPI,
			}
			vmNetworkInteraces, err := metadata.getVMInterfaces(context.Background(), tc.vm, "resource-group")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantNetworkInterfaces, vmNetworkInteraces)
		})
	}
}

func TestGetScaleSetVMInterfaces(t *testing.T) {
	wantNetworkInterfaces := []armnetwork.Interface{
		{
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
	vm := armcomputev2.VirtualMachineScaleSetVM{
		Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
			NetworkProfile: &armcomputev2.NetworkProfile{
				NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
					{
						ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
					},
				},
			},
		},
	}
	testCases := map[string]struct {
		vm                    armcomputev2.VirtualMachineScaleSetVM
		networkInterfacesAPI  networkInterfacesAPI
		wantErr               bool
		wantNetworkInterfaces []armnetwork.Interface
	}{
		"retrieval works": {
			vm: vm,
			networkInterfacesAPI: &stubNetworkInterfacesAPI{
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
			},
			wantNetworkInterfaces: wantNetworkInterfaces,
		},
		"vm can have 0 interfaces": {
			vm: armcomputev2.VirtualMachineScaleSetVM{},
			networkInterfacesAPI: &stubNetworkInterfacesAPI{
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
			},
			wantNetworkInterfaces: []armnetwork.Interface{},
		},
		"interface retrieval fails": {
			vm: vm,
			networkInterfacesAPI: &stubNetworkInterfacesAPI{
				getErr: errors.New("get err"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				networkInterfacesAPI: tc.networkInterfacesAPI,
			}
			configs, err := metadata.getScaleSetVMInterfaces(context.Background(), tc.vm, "resource-group", "scale-set-name", "instance-id")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantNetworkInterfaces, configs)
		})
	}
}

func TestGetScaleSetVMPublicIPAddresses(t *testing.T) {
	someErr := errors.New("some err")
	newNetworkInterfaces := func() []armnetwork.Interface {
		return []armnetwork.Interface{{
			Name: to.StringPtr("interface-name"),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.StringPtr("ip-config-name"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							Primary: to.BoolPtr(true),
							PublicIPAddress: &armnetwork.PublicIPAddress{
								ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/publicIPAddresses/public-ip-name"),
							},
						},
					},
				},
			},
		}, {
			Name: to.StringPtr("interface-name2"),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.StringPtr("ip-config-name2"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PublicIPAddress: &armnetwork.PublicIPAddress{
								ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/publicIPAddresses/public-ip-name2"),
							},
						},
					},
				},
			},
		}}
	}

	testCases := map[string]struct {
		networkInterfacesMutator func(*[]armnetwork.Interface)
		networkInterfaces        []armnetwork.Interface
		publicIPAddressesAPI     publicIPAddressesAPI
		wantIP                   string
		wantErr                  bool
	}{
		"retrieval works": {
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getVirtualMachineScaleSetPublicIPAddressResponse: armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse{
				PublicIPAddress: armnetwork.PublicIPAddress{
					Properties: &armnetwork.PublicIPAddressPropertiesFormat{
						IPAddress: to.StringPtr("192.0.2.1"),
					},
				},
			}},
			networkInterfaces: newNetworkInterfaces(),
			wantIP:            "192.0.2.1",
		},
		"retrieval works for no valid interfaces": {
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getVirtualMachineScaleSetPublicIPAddressResponse: armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse{
				PublicIPAddress: armnetwork.PublicIPAddress{
					Properties: &armnetwork.PublicIPAddressPropertiesFormat{
						IPAddress: to.StringPtr("192.0.2.1"),
					},
				},
			}},
			networkInterfaces: newNetworkInterfaces(),
			networkInterfacesMutator: func(nets *[]armnetwork.Interface) {
				(*nets)[0].Properties.IPConfigurations = []*armnetwork.InterfaceIPConfiguration{nil}
				(*nets)[1] = armnetwork.Interface{Name: nil}
			},
		},
		"fail to get public IP": {
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getErr: someErr},
			networkInterfaces:    newNetworkInterfaces(),
			wantErr:              true,
		},
		"fail to parse IPv4 address of public IP": {
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getVirtualMachineScaleSetPublicIPAddressResponse: armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse{
				PublicIPAddress: armnetwork.PublicIPAddress{},
			}},
			networkInterfaces: newNetworkInterfaces(),
			wantErr:           true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			if tc.networkInterfacesMutator != nil {
				tc.networkInterfacesMutator(&tc.networkInterfaces)
			}

			metadata := Metadata{
				publicIPAddressesAPI: tc.publicIPAddressesAPI,
			}

			ips, err := metadata.getScaleSetVMPublicIPAddress(context.Background(), "resource-group", "scale-set-name", "instance-id", tc.networkInterfaces)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantIP, ips)
		})
	}
}

func TestExtractPrivateIPs(t *testing.T) {
	testCases := map[string]struct {
		networkInterfaces []armnetwork.Interface
		wantIP            string
	}{
		"extraction works": {
			networkInterfaces: []armnetwork.Interface{
				{
					Properties: &armnetwork.InterfacePropertiesFormat{
						IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
							{
								Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
									Primary:          to.BoolPtr(true),
									PrivateIPAddress: to.StringPtr("192.0.2.0"),
								},
							},
							{
								Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
									PrivateIPAddress: to.StringPtr("192.0.2.1"),
								},
							},
						},
					},
				},
			},
			wantIP: "192.0.2.0",
		},
		"can be empty": {
			networkInterfaces: []armnetwork.Interface{},
		},
		"invalid interface is skipped": {
			networkInterfaces: []armnetwork.Interface{{}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ip := extractVPCIP(tc.networkInterfaces)
			assert.Equal(tc.wantIP, ip)
		})
	}
}

func TestExtractInterfaceNamesFromInterfaceReferences(t *testing.T) {
	testCases := map[string]struct {
		references []*armcomputev2.NetworkInterfaceReference
		wantNames  []string
	}{
		"extraction with individual interface reference works": {
			references: []*armcomputev2.NetworkInterfaceReference{
				{
					ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
				},
			},
			wantNames: []string{"interface-name"},
		},
		"extraction with scale set interface reference works": {
			references: []*armcomputev2.NetworkInterfaceReference{
				{
					ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
				},
			},
			wantNames: []string{"interface-name"},
		},
		"can be empty": {
			references: []*armcomputev2.NetworkInterfaceReference{},
		},
		"interface reference containing nil fields is skipped": {
			references: []*armcomputev2.NetworkInterfaceReference{
				{},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			names := extractInterfaceNamesFromInterfaceReferences(tc.references)

			assert.ElementsMatch(tc.wantNames, names)
		})
	}
}

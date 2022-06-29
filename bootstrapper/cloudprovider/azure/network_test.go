package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
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
	vm := armcompute.VirtualMachine{
		Properties: &armcompute.VirtualMachineProperties{
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
					},
				},
			},
		},
	}
	testCases := map[string]struct {
		vm                    armcompute.VirtualMachine
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
			vm: armcompute.VirtualMachine{},
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
	vm := armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
					},
				},
			},
		},
	}
	testCases := map[string]struct {
		vm                    armcompute.VirtualMachineScaleSetVM
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
			vm: armcompute.VirtualMachineScaleSetVM{},
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
		wantIPs                  []string
		wantErr                  bool
	}{
		"retrieval works": {
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getVirtualMachineScaleSetPublicIPAddressResponse: armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse{
				PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResult: armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResult{
					PublicIPAddress: armnetwork.PublicIPAddress{
						Properties: &armnetwork.PublicIPAddressPropertiesFormat{
							IPAddress: to.StringPtr("192.0.2.1"),
						},
					},
				},
			}},
			networkInterfaces: newNetworkInterfaces(),
			wantIPs:           []string{"192.0.2.1", "192.0.2.1"},
		},
		"retrieval works for no valid interfaces": {
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getVirtualMachineScaleSetPublicIPAddressResponse: armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse{
				PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResult: armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResult{
					PublicIPAddress: armnetwork.PublicIPAddress{
						Properties: &armnetwork.PublicIPAddressPropertiesFormat{
							IPAddress: to.StringPtr("192.0.2.1"),
						},
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
				PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResult: armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResult{
					PublicIPAddress: armnetwork.PublicIPAddress{},
				},
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

			ips, err := metadata.getScaleSetVMPublicIPAddresses(context.Background(), "resource-group", "scale-set-name", "instance-id", tc.networkInterfaces)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantIPs, ips)
		})
	}
}

func TestExtractPrivateIPs(t *testing.T) {
	testCases := map[string]struct {
		networkInterfaces []armnetwork.Interface
		wantIPs           []string
	}{
		"extraction works": {
			networkInterfaces: []armnetwork.Interface{
				{
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
			wantIPs: []string{"192.0.2.0"},
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

			ips := extractPrivateIPs(tc.networkInterfaces)

			assert.ElementsMatch(tc.wantIPs, ips)
		})
	}
}

func TestExtractInterfaceNamesFromInterfaceReferences(t *testing.T) {
	testCases := map[string]struct {
		references []*armcompute.NetworkInterfaceReference
		wantNames  []string
	}{
		"extraction with individual interface reference works": {
			references: []*armcompute.NetworkInterfaceReference{
				{
					ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
				},
			},
			wantNames: []string{"interface-name"},
		},
		"extraction with scale set interface reference works": {
			references: []*armcompute.NetworkInterfaceReference{
				{
					ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
				},
			},
			wantNames: []string{"interface-name"},
		},
		"can be empty": {
			references: []*armcompute.NetworkInterfaceReference{},
		},
		"interface reference containing nil fields is skipped": {
			references: []*armcompute.NetworkInterfaceReference{
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

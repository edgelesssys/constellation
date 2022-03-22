package azure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// getVMInterfaces retrieves all network interfaces referenced by a virtual machine.
func (m *Metadata) getVMInterfaces(ctx context.Context, vm armcompute.VirtualMachine, resourceGroup string) ([]*armnetwork.InterfaceIPConfiguration, error) {
	if vm.Properties == nil || vm.Properties.NetworkProfile == nil {
		return []*armnetwork.InterfaceIPConfiguration{}, nil
	}
	interfaceNames := extractInterfaceNamesFromInterfaceReferences(vm.Properties.NetworkProfile.NetworkInterfaces)
	interfaceIPConfigurations := []*armnetwork.InterfaceIPConfiguration{}
	for _, interfaceName := range interfaceNames {
		networkInterfacesResp, err := m.networkInterfacesAPI.Get(ctx, resourceGroup, interfaceName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve network interface %v: %w", interfaceName, err)
		}
		if networkInterfacesResp.Interface.Properties == nil || networkInterfacesResp.Interface.Properties.IPConfigurations == nil {
			return nil, errors.New("retrieved network interface has invalid ip configuration")
		}
		interfaceIPConfigurations = append(interfaceIPConfigurations, networkInterfacesResp.Properties.IPConfigurations...)
	}
	return interfaceIPConfigurations, nil
}

// getScaleSetVMInterfaces retrieves all network interfaces referenced by a scale set virtual machine.
func (m *Metadata) getScaleSetVMInterfaces(ctx context.Context, vm armcompute.VirtualMachineScaleSetVM, resourceGroup, scaleSet, instanceID string) ([]*armnetwork.InterfaceIPConfiguration, error) {
	if vm.Properties == nil || vm.Properties.NetworkProfile == nil {
		return []*armnetwork.InterfaceIPConfiguration{}, nil
	}
	interfaceNames := extractInterfaceNamesFromInterfaceReferences(vm.Properties.NetworkProfile.NetworkInterfaces)
	interfaceIPConfigurations := []*armnetwork.InterfaceIPConfiguration{}
	for _, interfaceName := range interfaceNames {
		networkInterfacesResp, err := m.networkInterfacesAPI.GetVirtualMachineScaleSetNetworkInterface(ctx, resourceGroup, scaleSet, instanceID, interfaceName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve network interface %v: %w", interfaceName, err)
		}
		if networkInterfacesResp.Interface.Properties == nil || networkInterfacesResp.Interface.Properties.IPConfigurations == nil {
			return nil, errors.New("retrieved network interface has invalid ip configuration")
		}
		interfaceIPConfigurations = append(interfaceIPConfigurations, networkInterfacesResp.Properties.IPConfigurations...)
	}
	return interfaceIPConfigurations, nil
}

// extractPrivateIPs extracts private IPs from a list of network interface IP configurations.
func extractPrivateIPs(interfaceIPConfigs []*armnetwork.InterfaceIPConfiguration) []string {
	addresses := []string{}
	for _, config := range interfaceIPConfigs {
		if config == nil || config.Properties == nil || config.Properties.PrivateIPAddress == nil {
			continue
		}
		addresses = append(addresses, *config.Properties.PrivateIPAddress)
	}
	return addresses
}

// extractInterfaceNamesFromInterfaceReferences extracts the name of a network interface from a reference id.
// Format:
// - "/subscriptions/<subscription>/resourceGroups/<resource-group>/providers/Microsoft.Network/networkInterfaces/<interface-name>"
// - "/subscriptions/<subscription>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachineScaleSets/<scale-set-name>/virtualMachines/<instanceID>/networkInterfaces/<interface-name>".
func extractInterfaceNamesFromInterfaceReferences(references []*armcompute.NetworkInterfaceReference) []string {
	interfaceNames := []string{}
	for _, interfaceReference := range references {
		if interfaceReference == nil || interfaceReference.ID == nil {
			continue
		}
		interfaceIDParts := strings.Split(*interfaceReference.ID, "/")
		if len(interfaceIDParts) < 1 {
			continue
		}
		interfaceName := interfaceIDParts[len(interfaceIDParts)-1]
		interfaceNames = append(interfaceNames, interfaceName)
	}
	return interfaceNames
}

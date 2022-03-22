package azure

import (
	"context"
	"errors"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/role"
)

// getScaleSetVM tries to get an azure vm belonging to a scale set.
func (m *Metadata) getScaleSetVM(ctx context.Context, providerID string) (core.Instance, error) {
	_, resourceGroup, scaleSet, instanceID, err := splitScaleSetProviderID(providerID)
	if err != nil {
		return core.Instance{}, err
	}
	vmResp, err := m.virtualMachineScaleSetVMsAPI.Get(ctx, resourceGroup, scaleSet, instanceID, nil)
	if err != nil {
		return core.Instance{}, err
	}
	interfaceIPConfigurations, err := m.getScaleSetVMInterfaces(ctx, vmResp.VirtualMachineScaleSetVM, resourceGroup, scaleSet, instanceID)
	if err != nil {
		return core.Instance{}, err
	}

	return convertScaleSetVMToCoreInstance(scaleSet, vmResp.VirtualMachineScaleSetVM, interfaceIPConfigurations)
}

// listScaleSetVMs lists all scale set VMs in the current resource group.
func (m *Metadata) listScaleSetVMs(ctx context.Context, resourceGroup string) ([]core.Instance, error) {
	instances := []core.Instance{}
	scaleSetPager := m.scaleSetsAPI.List(resourceGroup, nil)
	for scaleSetPager.NextPage(ctx) {
		for _, scaleSet := range scaleSetPager.PageResponse().Value {
			if scaleSet == nil || scaleSet.Name == nil {
				continue
			}
			vmPager := m.virtualMachineScaleSetVMsAPI.List(resourceGroup, *scaleSet.Name, nil)
			for vmPager.NextPage(ctx) {
				for _, vm := range vmPager.PageResponse().Value {
					if vm == nil || vm.InstanceID == nil {
						continue
					}
					interfaces, err := m.getScaleSetVMInterfaces(ctx, *vm, resourceGroup, *scaleSet.Name, *vm.InstanceID)
					if err != nil {
						return nil, err
					}
					instance, err := convertScaleSetVMToCoreInstance(*scaleSet.Name, *vm, interfaces)
					if err != nil {
						return nil, err
					}
					instances = append(instances, instance)
				}
			}
		}
	}
	return instances, nil
}

// splitScaleSetProviderID splits a provider's id belonging to an azure scaleset into core components.
// A providerID for scale set VMs is build after the following schema:
// - 'azure:///subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachineScaleSets/<scale-set-name>/virtualMachines/<instance-id>'
func splitScaleSetProviderID(providerID string) (subscriptionID, resourceGroup, scaleSet, instanceID string, err error) {
	// providerIDregex is a regex matching an azure scaleset vm providerID with each part of the URI being a submatch.
	providerIDregex := regexp.MustCompile(`^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachineScaleSets/([^/]+)/virtualMachines/([^/]+)$`)
	matches := providerIDregex.FindStringSubmatch(providerID)
	if len(matches) != 5 {
		return "", "", "", "", errors.New("error splitting providerID")
	}
	return matches[1], matches[2], matches[3], matches[4], nil
}

// convertScaleSetVMToCoreInstance converts an azure scale set virtual machine with interface configurations into a core.Instance.
func convertScaleSetVMToCoreInstance(scaleSet string, vm armcompute.VirtualMachineScaleSetVM, interfaceIPConfigs []*armnetwork.InterfaceIPConfiguration) (core.Instance, error) {
	if vm.ID == nil {
		return core.Instance{}, errors.New("retrieving instance from armcompute API client returned no instance ID")
	}
	if vm.Properties == nil || vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil {
		return core.Instance{}, errors.New("retrieving instance from armcompute API client returned no computer name")
	}
	var sshKeys map[string][]string
	if vm.Properties.OSProfile.LinuxConfiguration == nil || vm.Properties.OSProfile.LinuxConfiguration.SSH == nil {
		sshKeys = map[string][]string{}
	} else {
		sshKeys = extractSSHKeys(*vm.Properties.OSProfile.LinuxConfiguration.SSH)
	}
	return core.Instance{
		Name:       *vm.Properties.OSProfile.ComputerName,
		ProviderID: "azure://" + *vm.ID,
		Role:       extractScaleSetVMRole(scaleSet),
		IPs:        extractPrivateIPs(interfaceIPConfigs),
		SSHKeys:    sshKeys,
	}, nil
}

// extractScaleSetVMRole extracts the constellation role of a scale set using its name.
func extractScaleSetVMRole(scaleSet string) role.Role {
	coordinatorScaleSetRegexp := regexp.MustCompile(`constellation-scale-set-coordinators-[0-9a-zA-Z]+$`)
	nodeScaleSetRegexp := regexp.MustCompile(`constellation-scale-set-nodes-[0-9a-zA-Z]+$`)
	if coordinatorScaleSetRegexp.MatchString(scaleSet) {
		return role.Coordinator
	}
	if nodeScaleSetRegexp.MatchString(scaleSet) {
		return role.Node
	}
	return role.Unknown
}

package azure

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/edgelesssys/constellation/coordinator/cloudprovider"
	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
)

var azureVMProviderIDRegexp = regexp.MustCompile(`^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachines/([^/]+)$`)

// getVM tries to get a single azure vm.
func (m *Metadata) getVM(ctx context.Context, providerID string) (cloudtypes.Instance, error) {
	_, resourceGroup, instanceName, err := splitVMProviderID(providerID)
	if err != nil {
		return cloudtypes.Instance{}, err
	}
	vmResp, err := m.virtualMachinesAPI.Get(ctx, resourceGroup, instanceName, nil)
	if err != nil {
		return cloudtypes.Instance{}, err
	}
	interfaceIPConfigurations, err := m.getVMInterfaces(ctx, vmResp.VirtualMachine, resourceGroup)
	if err != nil {
		return cloudtypes.Instance{}, err
	}
	return convertVMToCoreInstance(vmResp.VirtualMachine, interfaceIPConfigurations)
}

// listVMs lists all individual VMs in the current resource group.
func (m *Metadata) listVMs(ctx context.Context, resourceGroup string) ([]cloudtypes.Instance, error) {
	instances := []cloudtypes.Instance{}
	pager := m.virtualMachinesAPI.List(resourceGroup, nil)
	for pager.NextPage(ctx) {
		for _, vm := range pager.PageResponse().Value {
			if vm == nil {
				continue
			}
			interfaces, err := m.getVMInterfaces(ctx, *vm, resourceGroup)
			if err != nil {
				return nil, err
			}
			instance, err := convertVMToCoreInstance(*vm, interfaces)
			if err != nil {
				return nil, err
			}
			instances = append(instances, instance)
		}
	}
	return instances, nil
}

// setTag merges key-value pair into VM tags.
func (m *Metadata) setTag(ctx context.Context, key, value string) error {
	instanceMetadata, err := m.imdsAPI.Retrieve(ctx)
	if err != nil {
		return err
	}
	_, err = m.tagsAPI.UpdateAtScope(ctx, instanceMetadata.Compute.ResourceID, armresources.TagsPatchResource{
		Operation: armresources.TagsPatchOperationMerge.ToPtr(),
		Properties: &armresources.Tags{
			Tags: map[string]*string{
				key: to.StringPtr(value),
			},
		},
	}, nil)
	return err
}

// splitVMProviderID splits a provider's id belonging to a single azure instance into core components.
// A providerID  for individual VMs is build after the following schema:
// - 'azure:///subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachines/<instance-name>'
func splitVMProviderID(providerID string) (subscriptionID, resourceGroup, instanceName string, err error) {
	matches := azureVMProviderIDRegexp.FindStringSubmatch(providerID)
	if len(matches) != 4 {
		return "", "", "", errors.New("error splitting providerID")
	}
	return matches[1], matches[2], matches[3], nil
}

// convertVMToCoreInstance converts an azure virtual machine with interface configurations into a cloudtypes.Instance.
func convertVMToCoreInstance(vm armcompute.VirtualMachine, networkInterfaces []armnetwork.Interface) (cloudtypes.Instance, error) {
	if vm.Name == nil || vm.ID == nil {
		return cloudtypes.Instance{}, fmt.Errorf("retrieving instance from armcompute API client returned invalid instance Name (%v) or ID (%v)", vm.Name, vm.ID)
	}
	var sshKeys map[string][]string
	if vm.Properties == nil || vm.Properties.OSProfile == nil || vm.Properties.OSProfile.LinuxConfiguration == nil || vm.Properties.OSProfile.LinuxConfiguration.SSH == nil {
		sshKeys = map[string][]string{}
	} else {
		sshKeys = extractSSHKeys(*vm.Properties.OSProfile.LinuxConfiguration.SSH)
	}
	metadata := extractInstanceTags(vm.Tags)
	return cloudtypes.Instance{
		Name:       *vm.Name,
		ProviderID: "azure://" + *vm.ID,
		Role:       cloudprovider.ExtractRole(metadata),
		PrivateIPs: extractPrivateIPs(networkInterfaces),
		SSHKeys:    sshKeys,
	}, nil
}

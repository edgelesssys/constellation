package azure

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/azureshared"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
)

var (
	controlPlaneScaleSetRegexp = regexp.MustCompile(`constellation-scale-set-controlplanes-[0-9a-zA-Z]+$`)
	workerScaleSetRegexp       = regexp.MustCompile(`constellation-scale-set-workers-[0-9a-zA-Z]+$`)
)

// getScaleSetVM tries to get an azure vm belonging to a scale set.
func (m *Metadata) getScaleSetVM(ctx context.Context, providerID string) (metadata.InstanceMetadata, error) {
	_, resourceGroup, scaleSet, instanceID, err := azureshared.ScaleSetInformationFromProviderID(providerID)
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}
	vmResp, err := m.virtualMachineScaleSetVMsAPI.Get(ctx, resourceGroup, scaleSet, instanceID, nil)
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}
	networkInterfaces, err := m.getScaleSetVMInterfaces(ctx, vmResp.VirtualMachineScaleSetVM, resourceGroup, scaleSet, instanceID)
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}
	publicIPAddress, err := m.getScaleSetVMPublicIPAddress(ctx, resourceGroup, scaleSet, instanceID, networkInterfaces)
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}

	return convertScaleSetVMToCoreInstance(scaleSet, vmResp.VirtualMachineScaleSetVM, networkInterfaces, publicIPAddress)
}

// listScaleSetVMs lists all scale set VMs in the current resource group.
func (m *Metadata) listScaleSetVMs(ctx context.Context, resourceGroup string) ([]metadata.InstanceMetadata, error) {
	instances := []metadata.InstanceMetadata{}
	scaleSetPager := m.scaleSetsAPI.NewListPager(resourceGroup, nil)
	for scaleSetPager.More() {
		page, err := scaleSetPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving scale sets: %w", err)
		}
		for _, scaleSet := range page.Value {
			if scaleSet == nil || scaleSet.Name == nil {
				continue
			}
			vmPager := m.virtualMachineScaleSetVMsAPI.NewListPager(resourceGroup, *scaleSet.Name, nil)
			for vmPager.More() {
				vmPage, err := vmPager.NextPage(ctx)
				if err != nil {
					return nil, fmt.Errorf("retrieving vms: %w", err)
				}
				for _, vm := range vmPage.Value {
					if vm == nil || vm.InstanceID == nil {
						continue
					}
					interfaces, err := m.getScaleSetVMInterfaces(ctx, *vm, resourceGroup, *scaleSet.Name, *vm.InstanceID)
					if err != nil {
						return nil, err
					}
					instance, err := convertScaleSetVMToCoreInstance(*scaleSet.Name, *vm, interfaces, "")
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

// convertScaleSetVMToCoreInstance converts an azure scale set virtual machine with interface configurations into a core.Instance.
func convertScaleSetVMToCoreInstance(scaleSet string, vm armcomputev2.VirtualMachineScaleSetVM, networkInterfaces []armnetwork.Interface, publicIPAddress string) (metadata.InstanceMetadata, error) {
	if vm.ID == nil {
		return metadata.InstanceMetadata{}, errors.New("retrieving instance from armcompute API client returned no instance ID")
	}
	if vm.Properties == nil || vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil {
		return metadata.InstanceMetadata{}, errors.New("retrieving instance from armcompute API client returned no computer name")
	}
	var sshKeys map[string][]string
	if vm.Properties.OSProfile.LinuxConfiguration == nil || vm.Properties.OSProfile.LinuxConfiguration.SSH == nil {
		sshKeys = map[string][]string{}
	} else {
		sshKeys = extractSSHKeys(*vm.Properties.OSProfile.LinuxConfiguration.SSH)
	}
	return metadata.InstanceMetadata{
		Name:       *vm.Properties.OSProfile.ComputerName,
		ProviderID: "azure://" + *vm.ID,
		Role:       extractScaleSetVMRole(scaleSet),
		VPCIP:      extractVPCIP(networkInterfaces),
		PublicIP:   publicIPAddress,
		SSHKeys:    sshKeys,
	}, nil
}

// extractScaleSetVMRole extracts the constellation role of a scale set using its name.
func extractScaleSetVMRole(scaleSet string) role.Role {
	if controlPlaneScaleSetRegexp.MatchString(scaleSet) {
		return role.ControlPlane
	}
	if workerScaleSetRegexp.MatchString(scaleSet) {
		return role.Worker
	}
	return role.Unknown
}

// ImageReferenceFromImage sets the `ID` or `CommunityGalleryImageID` field
// of `ImageReference` depending on the provided `img`.
func ImageReferenceFromImage(img string) *armcomputev2.ImageReference {
	ref := &armcomputev2.ImageReference{}

	if strings.HasPrefix(img, "/CommunityGalleries") {
		ref.CommunityGalleryImageID = to.Ptr(img)
	} else {
		ref.ID = to.Ptr(img)
	}

	return ref
}

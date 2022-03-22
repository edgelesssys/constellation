package azure

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/role"
)

// Metadata implements azure metadata APIs.
type Metadata struct {
	imdsAPI
	networkInterfacesAPI
	scaleSetsAPI
	virtualMachinesAPI
	virtualMachineScaleSetVMsAPI
	tagsAPI
}

// NewMetadata creates a new Metadata.
func NewMetadata(ctx context.Context) (*Metadata, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	// The default http client may use a system-wide proxy and it is recommended to disable the proxy explicitly:
	// https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service?tabs=linux#proxies
	// See also: https://github.com/microsoft/azureimds/blob/master/imdssample.go#L10
	imdsAPI := imdsClient{
		client: &http.Client{Transport: &http.Transport{Proxy: nil}},
	}
	instanceMetadata, err := imdsAPI.Retrieve(ctx)
	if err != nil {
		return nil, err
	}
	subscriptionID, _, err := extractBasicsFromProviderID("azure://" + instanceMetadata.Compute.ResourceID)
	if err != nil {
		return nil, err
	}
	networkInterfacesAPI := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
	scaleSetsAPI := armcompute.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)
	virtualMachinesAPI := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	virtualMachineScaleSetVMsAPI := armcompute.NewVirtualMachineScaleSetVMsClient(subscriptionID, cred, nil)
	tagsAPI := armresources.NewTagsClient(subscriptionID, cred, nil)

	return &Metadata{
		imdsAPI:                      &imdsAPI,
		networkInterfacesAPI:         &networkInterfacesClient{networkInterfacesAPI},
		scaleSetsAPI:                 &scaleSetsClient{scaleSetsAPI},
		virtualMachinesAPI:           &virtualMachinesClient{virtualMachinesAPI},
		virtualMachineScaleSetVMsAPI: &virtualMachineScaleSetVMsClient{virtualMachineScaleSetVMsAPI},
		tagsAPI:                      &tagsClient{tagsAPI},
	}, nil
}

// List retrieves all instances belonging to the current constellation.
func (m *Metadata) List(ctx context.Context) ([]core.Instance, error) {
	providerID, err := m.providerID(ctx)
	if err != nil {
		return nil, err
	}
	_, resourceGroup, err := extractBasicsFromProviderID(providerID)
	if err != nil {
		return nil, err
	}
	singleInstances, err := m.listVMs(ctx, resourceGroup)
	if err != nil {
		return nil, err
	}
	scaleSetInstances, err := m.listScaleSetVMs(ctx, resourceGroup)
	if err != nil {
		return nil, err
	}
	instances := make([]core.Instance, 0, len(singleInstances)+len(scaleSetInstances))
	instances = append(instances, singleInstances...)
	instances = append(instances, scaleSetInstances...)
	return instances, nil
}

// Self retrieves the current instance.
func (m *Metadata) Self(ctx context.Context) (core.Instance, error) {
	providerID, err := m.providerID(ctx)
	if err != nil {
		return core.Instance{}, err
	}
	return m.GetInstance(ctx, providerID)
}

// GetInstance retrieves an instance using its providerID.
func (m *Metadata) GetInstance(ctx context.Context, providerID string) (core.Instance, error) {
	instance, singleErr := m.getVM(ctx, providerID)
	if singleErr == nil {
		return instance, nil
	}
	instance, scaleSetErr := m.getScaleSetVM(ctx, providerID)
	if scaleSetErr == nil {
		return instance, nil
	}
	return core.Instance{}, fmt.Errorf("could not retrieve instance given providerID %v as either single vm or scale set vm: %v %v", providerID, singleErr, scaleSetErr)
}

// SignalRole signals the constellation role via cloud provider metadata.
// On single VMs, the role is stored in tags, on scale set VMs, the role is inferred from the scale set and not signalied explicitly.
func (m *Metadata) SignalRole(ctx context.Context, role role.Role) error {
	providerID, err := m.providerID(ctx)
	if err != nil {
		return err
	}
	if _, _, _, _, err := splitScaleSetProviderID(providerID); err == nil {
		// scale set instances cannot store tags and role can be inferred from scale set name.
		return nil
	}
	return m.setTag(ctx, core.RoleMetadataKey, role.String())
}

// SetVPNIP stores the internally used VPN IP in cloud provider metadata (not required on azure).
func (m *Metadata) SetVPNIP(ctx context.Context, vpnIP string) error {
	return nil
}

// Supported is used to determine if metadata API is implemented for this cloud provider.
func (m *Metadata) Supported() bool {
	return true
}

// providerID retrieves the current instances providerID.
func (m *Metadata) providerID(ctx context.Context) (string, error) {
	instanceMetadata, err := m.imdsAPI.Retrieve(ctx)
	if err != nil {
		return "", err
	}
	return "azure://" + instanceMetadata.Compute.ResourceID, nil
}

// extractBasicsFromProviderID extracts subscriptionID and resourceGroup from both types of valid azure providerID.
func extractBasicsFromProviderID(providerID string) (subscriptionID, resourceGroup string, err error) {
	subscriptionID, resourceGroup, _, err = splitVMProviderID(providerID)
	if err == nil {
		return subscriptionID, resourceGroup, nil
	}
	subscriptionID, resourceGroup, _, _, err = splitScaleSetProviderID(providerID)
	if err == nil {
		return subscriptionID, resourceGroup, nil
	}
	return "", "", fmt.Errorf("providerID %v is malformatted", providerID)
}

// extractInstanceTags converts azure tags into metadata key-value pairs.
func extractInstanceTags(tags map[string]*string) map[string]string {
	metadataMap := map[string]string{}
	for key, value := range tags {
		if value == nil {
			continue
		}
		metadataMap[key] = *value
	}
	return metadataMap
}

// extractSSHKeys extracts SSH public keys from azure instance OS Profile.
func extractSSHKeys(sshConfig armcompute.SSHConfiguration) map[string][]string {
	keyPathRegexp := regexp.MustCompile(`^\/home\/([^\/]+)\/\.ssh\/authorized_keys$`)
	sshKeys := map[string][]string{}
	for _, key := range sshConfig.PublicKeys {
		if key == nil || key.Path == nil || key.KeyData == nil {
			continue
		}
		matches := keyPathRegexp.FindStringSubmatch(*key.Path)
		if len(matches) != 2 {
			continue
		}
		sshKeys[matches[1]] = append(sshKeys[matches[1]], *key.KeyData)
	}
	return sshKeys
}

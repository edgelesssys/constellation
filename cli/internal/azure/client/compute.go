package client

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/edgelesssys/constellation/cli/internal/azure"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
)

func (c *Client) CreateInstances(ctx context.Context, input CreateInstancesInput) error {
	// Create worker scale set
	createWorkerInput := CreateScaleSetInput{
		Name:                           "constellation-scale-set-workers-" + c.uid,
		NamePrefix:                     c.name + "-worker-" + c.uid + "-",
		Count:                          input.CountWorkers,
		InstanceType:                   input.InstanceType,
		StateDiskSizeGB:                int32(input.StateDiskSizeGB),
		Image:                          input.Image,
		UserAssingedIdentity:           input.UserAssingedIdentity,
		LoadBalancerBackendAddressPool: azure.BackendAddressPoolWorkerName + "-" + c.uid,
	}

	if err := c.createScaleSet(ctx, createWorkerInput); err != nil {
		return err
	}

	c.workerScaleSet = createWorkerInput.Name

	// Create control plane scale set
	createControlPlaneInput := CreateScaleSetInput{
		Name:                           "constellation-scale-set-controlplanes-" + c.uid,
		NamePrefix:                     c.name + "-control-plane-" + c.uid + "-",
		Count:                          input.CountControlPlanes,
		InstanceType:                   input.InstanceType,
		StateDiskSizeGB:                int32(input.StateDiskSizeGB),
		Image:                          input.Image,
		UserAssingedIdentity:           input.UserAssingedIdentity,
		LoadBalancerBackendAddressPool: azure.BackendAddressPoolControlPlaneName + "-" + c.uid,
	}

	if err := c.createScaleSet(ctx, createControlPlaneInput); err != nil {
		return err
	}

	// Get worker IPs
	instances, err := c.getInstanceIPs(ctx, createWorkerInput.Name, createWorkerInput.Count)
	if err != nil {
		return err
	}
	c.workers = instances

	// Get control plane IPs
	c.controlPlaneScaleSet = createControlPlaneInput.Name
	instances, err = c.getInstanceIPs(ctx, createControlPlaneInput.Name, createControlPlaneInput.Count)
	if err != nil {
		return err
	}
	c.controlPlanes = instances

	return nil
}

// CreateInstancesInput is the input for a CreateInstances operation.
type CreateInstancesInput struct {
	CountWorkers         int
	CountControlPlanes   int
	InstanceType         string
	StateDiskSizeGB      int
	Image                string
	UserAssingedIdentity string
}

// CreateInstancesVMs creates instances based on standalone VMs.
// TODO: deprecate as soon as scale sets are available.
func (c *Client) CreateInstancesVMs(ctx context.Context, input CreateInstancesInput) error {
	pw, err := azure.GeneratePassword()
	if err != nil {
		return err
	}

	for i := 0; i < input.CountControlPlanes; i++ {
		vm := azure.VMInstance{
			Name:         c.name + "-control-plane-" + c.uid + "-" + strconv.Itoa(i),
			Username:     "constell",
			Password:     pw,
			Location:     c.location,
			InstanceType: input.InstanceType,
			Image:        input.Image,
		}
		instance, err := c.createInstanceVM(ctx, vm)
		if err != nil {
			return err
		}
		c.controlPlanes[strconv.Itoa(i)] = instance
	}

	for i := 0; i < input.CountWorkers; i++ {
		vm := azure.VMInstance{
			Name:         c.name + "-worker-" + c.uid + "-" + strconv.Itoa(i),
			Username:     "constell",
			Password:     pw,
			Location:     c.location,
			InstanceType: input.InstanceType,
			Image:        input.Image,
		}
		instance, err := c.createInstanceVM(ctx, vm)
		if err != nil {
			return err
		}
		c.workers[strconv.Itoa(i)] = instance
	}

	return nil
}

// createInstanceVM creates a single VM with a public IP address
// and a network interface.
// TODO: deprecate as soon as scale sets are available.
func (c *Client) createInstanceVM(ctx context.Context, input azure.VMInstance) (cloudtypes.Instance, error) {
	pubIPName := input.Name + "-pubIP"
	pubIP, err := c.createPublicIPAddress(ctx, pubIPName)
	if err != nil {
		return cloudtypes.Instance{}, err
	}

	nicName := input.Name + "-NIC"
	privIP, nicID, err := c.createNIC(ctx, nicName, *pubIP.ID)
	if err != nil {
		return cloudtypes.Instance{}, err
	}

	input.NIC = nicID

	poller, err := c.virtualMachinesAPI.BeginCreateOrUpdate(ctx, c.resourceGroup, input.Name, input.Azure(), nil)
	if err != nil {
		return cloudtypes.Instance{}, err
	}

	vm, err := poller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return cloudtypes.Instance{}, err
	}

	if vm.Identity == nil || vm.Identity.PrincipalID == nil {
		return cloudtypes.Instance{}, errors.New("virtual machine was created without system managed identity")
	}

	if err := c.assignResourceGroupRole(ctx, *vm.Identity.PrincipalID, virtualMachineContributorRoleDefinitionID); err != nil {
		return cloudtypes.Instance{}, err
	}

	res, err := c.publicIPAddressesAPI.Get(ctx, c.resourceGroup, pubIPName, nil)
	if err != nil {
		return cloudtypes.Instance{}, err
	}

	return cloudtypes.Instance{PublicIP: *res.PublicIPAddressesClientGetResult.PublicIPAddress.Properties.IPAddress, PrivateIP: privIP}, nil
}

func (c *Client) createScaleSet(ctx context.Context, input CreateScaleSetInput) error {
	// TODO: Generating a random password to be able
	// to create the scale set. This is a temporary fix.
	// We need to think about azure access at some point.
	pw, err := azure.GeneratePassword()
	if err != nil {
		return err
	}
	scaleSet := azure.ScaleSet{
		Name:                           input.Name,
		NamePrefix:                     input.NamePrefix,
		Location:                       c.location,
		InstanceType:                   input.InstanceType,
		StateDiskSizeGB:                input.StateDiskSizeGB,
		Count:                          int64(input.Count),
		Username:                       "constellation",
		SubnetID:                       c.subnetID,
		NetworkSecurityGroup:           c.networkSecurityGroup,
		Image:                          input.Image,
		Password:                       pw,
		UserAssignedIdentity:           input.UserAssingedIdentity,
		Subscription:                   c.subscriptionID,
		ResourceGroup:                  c.resourceGroup,
		LoadBalancerName:               c.loadBalancerName,
		LoadBalancerBackendAddressPool: input.LoadBalancerBackendAddressPool,
	}.Azure()

	poller, err := c.scaleSetsAPI.BeginCreateOrUpdate(
		ctx, c.resourceGroup, input.Name,
		scaleSet,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) getInstanceIPs(ctx context.Context, scaleSet string, count int) (cloudtypes.Instances, error) {
	instances := cloudtypes.Instances{}
	for i := 0; i < count; i++ {
		// get public ip address
		var publicIPAddress string
		pager := c.publicIPAddressesAPI.ListVirtualMachineScaleSetVMPublicIPAddresses(
			c.resourceGroup, scaleSet, strconv.Itoa(i), scaleSet, scaleSet, nil)

		// We always need one pager.NextPage, since calling
		// pager.PageResponse() directly return no result.
		// We expect to get one page with one entry for each VM.
		for pager.NextPage(ctx) {
			for _, v := range pager.PageResponse().Value {
				if v.Properties != nil && v.Properties.IPAddress != nil {
					publicIPAddress = *v.Properties.IPAddress
					break
				}
			}
		}

		// get private ip address
		var privateIPAddress string
		res, err := c.networkInterfacesAPI.GetVirtualMachineScaleSetNetworkInterface(
			ctx, c.resourceGroup, scaleSet, strconv.Itoa(i), scaleSet, nil)
		if err != nil {
			return nil, err
		}
		configs := res.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResult.Interface.Properties.IPConfigurations
		for _, config := range configs {
			privateIPAddress = *config.Properties.PrivateIPAddress
			break
		}

		instance := cloudtypes.Instance{
			PrivateIP: privateIPAddress,
			PublicIP:  publicIPAddress,
		}
		instances[strconv.Itoa(i)] = instance
	}
	return instances, nil
}

// CreateScaleSetInput is the input for a CreateScaleSet operation.
type CreateScaleSetInput struct {
	Name                           string
	NamePrefix                     string
	Count                          int
	InstanceType                   string
	StateDiskSizeGB                int32
	Image                          string
	UserAssingedIdentity           string
	LoadBalancerBackendAddressPool string
}

// CreateResourceGroup creates a resource group.
func (c *Client) CreateResourceGroup(ctx context.Context) error {
	_, err := c.resourceGroupAPI.CreateOrUpdate(ctx, c.name+"-"+c.uid,
		armresources.ResourceGroup{
			Location: &c.location,
		}, nil)
	if err != nil {
		return err
	}
	c.resourceGroup = c.name + "-" + c.uid
	return nil
}

// TerminateResourceGroup terminates a resource group.
func (c *Client) TerminateResourceGroup(ctx context.Context) error {
	if c.resourceGroup == "" {
		return nil
	}

	poller, err := c.resourceGroupAPI.BeginDelete(ctx, c.resourceGroup, nil)
	if err != nil {
		return err
	}

	if _, err = poller.PollUntilDone(ctx, 30*time.Second); err != nil {
		return err
	}
	c.workers = nil
	c.controlPlanes = nil
	c.resourceGroup = ""
	c.subnetID = ""
	c.networkSecurityGroup = ""
	c.workerScaleSet = ""
	c.controlPlaneScaleSet = ""
	return nil
}

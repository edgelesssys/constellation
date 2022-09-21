/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/edgelesssys/constellation/v2/cli/internal/azure"
	"github.com/edgelesssys/constellation/v2/cli/internal/azure/internal/poller"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
)

const (
	// scaleSetCreateTimeout maximum timeout to wait for scale set creation.
	scaleSetCreateTimeout = 5 * time.Minute
	powerStateStarting    = "PowerState/starting"
	powerStateRunning     = "PowerState/running"
)

func (c *Client) CreateInstances(ctx context.Context, input CreateInstancesInput) error {
	// Create worker scale set
	createWorkerInput := CreateScaleSetInput{
		Name:                           "constellation-scale-set-workers-" + c.uid,
		NamePrefix:                     c.name + "-worker-" + c.uid + "-",
		Count:                          input.CountWorkers,
		InstanceType:                   input.InstanceType,
		StateDiskSizeGB:                int32(input.StateDiskSizeGB),
		StateDiskType:                  input.StateDiskType,
		Image:                          input.Image,
		UserAssingedIdentity:           input.UserAssingedIdentity,
		LoadBalancerBackendAddressPool: azure.BackendAddressPoolWorkerName + "-" + c.uid,
		ConfidentialVM:                 input.ConfidentialVM,
	}

	// Create control plane scale set
	createControlPlaneInput := CreateScaleSetInput{
		Name:                           "constellation-scale-set-controlplanes-" + c.uid,
		NamePrefix:                     c.name + "-control-plane-" + c.uid + "-",
		Count:                          input.CountControlPlanes,
		InstanceType:                   input.InstanceType,
		StateDiskSizeGB:                int32(input.StateDiskSizeGB),
		StateDiskType:                  input.StateDiskType,
		Image:                          input.Image,
		UserAssingedIdentity:           input.UserAssingedIdentity,
		LoadBalancerBackendAddressPool: azure.BackendAddressPoolControlPlaneName + "-" + c.uid,
		ConfidentialVM:                 input.ConfidentialVM,
	}

	var wg sync.WaitGroup
	var controlPlaneErr, workerErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		workerErr = c.createScaleSet(ctx, createWorkerInput)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		controlPlaneErr = c.createScaleSet(ctx, createControlPlaneInput)
	}()

	wg.Wait()
	if controlPlaneErr != nil {
		return fmt.Errorf("creating control-plane scaleset: %w", controlPlaneErr)
	}
	if workerErr != nil {
		return fmt.Errorf("creating worker scaleset: %w", workerErr)
	}

	// TODO: Remove getInstanceIPs calls after init has been refactored to not use node IPs
	// Get worker IPs
	c.workerScaleSet = createWorkerInput.Name
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
	StateDiskType        string
	Image                string
	UserAssingedIdentity string
	ConfidentialVM       bool
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
		UID:                            c.uid,
		Location:                       c.location,
		InstanceType:                   input.InstanceType,
		StateDiskSizeGB:                input.StateDiskSizeGB,
		StateDiskType:                  input.StateDiskType,
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
		ConfidentialVM:                 input.ConfidentialVM,
	}.Azure()

	_, err = c.scaleSetsAPI.BeginCreateOrUpdate(
		ctx, c.resourceGroup, input.Name,
		scaleSet,
		nil,
	)
	if err != nil {
		return err
	}

	// use custom poller to wait for resource creation but skip waiting for OS provisioning.
	// OS provisioning does not work reliably without the azure guest agent installed.
	poller := poller.New[bool](&scaleSetCreationPollingHandler{
		resourceGroup:                c.resourceGroup,
		scaleSet:                     input.Name,
		scaleSetsAPI:                 c.scaleSetsAPI,
		virtualMachineScaleSetVMsAPI: c.virtualMachineScaleSetVMsAPI,
	})

	pollCtx, cancel := context.WithTimeout(ctx, scaleSetCreateTimeout)
	defer cancel()
	_, err = poller.PollUntilDone(pollCtx, nil)
	return err
}

func (c *Client) getInstanceIPs(ctx context.Context, scaleSet string, count int) (cloudtypes.Instances, error) {
	instances := cloudtypes.Instances{}
	for i := 0; i < count; i++ {
		// get public ip address
		var publicIPAddress string
		pager := c.publicIPAddressesAPI.NewListVirtualMachineScaleSetVMPublicIPAddressesPager(
			c.resourceGroup, scaleSet, strconv.Itoa(i), scaleSet, scaleSet, nil)

		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return cloudtypes.Instances{}, err
			}
			for _, v := range page.Value {
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
		configs := res.Interface.Properties.IPConfigurations
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
	StateDiskType                  string
	Image                          string
	UserAssingedIdentity           string
	LoadBalancerBackendAddressPool string
	ConfidentialVM                 bool
}

// scaleSetCreationPollingHandler is a custom poller used to check if a scale set was created successfully.
type scaleSetCreationPollingHandler struct {
	done                         bool
	instanceIDOffset             int
	resourceGroup                string
	scaleSet                     string
	scaleSetsAPI                 scaleSetsAPI
	virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
}

// Done returns true if the condition is met.
func (h *scaleSetCreationPollingHandler) Done() bool {
	return h.done
}

// Poll checks if the scale set resource was created successfully and every VM is starting or running.
func (h *scaleSetCreationPollingHandler) Poll(ctx context.Context) error {
	// check if scale set can be retrieved from API
	scaleSet, err := h.scaleSetsAPI.Get(ctx, h.resourceGroup, h.scaleSet, nil)
	if err != nil {
		return ignoreNotFoundError(err)
	}
	if scaleSet.SKU == nil || scaleSet.SKU.Capacity == nil {
		return errors.New("invalid scale set capacity")
	}
	// check if every VM in the scale set has power state starting or running
	for i := h.instanceIDOffset; i < int(*scaleSet.SKU.Capacity); i++ {
		instanceView, err := h.virtualMachineScaleSetVMsAPI.GetInstanceView(ctx, h.resourceGroup, h.scaleSet, strconv.Itoa(i), nil)
		if err != nil {
			return ignoreNotFoundError(err)
		}
		if !vmIsStartingOrRunning(instanceView.Statuses) {
			return nil
		}
		h.instanceIDOffset = i + 1 // skip this VM in the next Poll() invocation
	}
	h.done = true
	return nil
}

// Result returns the result of the poller if the condition is met.
// If the condition is not met, an error is returned.
func (h *scaleSetCreationPollingHandler) Result(ctx context.Context, out *bool) error {
	if !h.done {
		return fmt.Errorf("failed to create scale set")
	}
	*out = h.done
	return nil
}

func ignoreNotFoundError(err error) error {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) && respErr.StatusCode == http.StatusNotFound {
		// resource does not exist yet - retry later
		return nil
	}
	return err
}

func vmIsStartingOrRunning(statuses []*armcomputev2.InstanceViewStatus) bool {
	for _, status := range statuses {
		if status == nil || status.Code == nil {
			continue
		}
		switch *status.Code {
		case powerStateStarting:
			return true
		case powerStateRunning:
			return true
		}
	}
	return false
}

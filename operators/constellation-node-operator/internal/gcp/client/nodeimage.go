/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

// GetNodeImage returns the image name of the node.
func (c *Client) GetNodeImage(ctx context.Context, providerID string) (string, error) {
	project, zone, instanceName, err := splitProviderID(providerID)
	if err != nil {
		return "", err
	}
	project, err = c.canonicalProjectID(ctx, project)
	if err != nil {
		return "", err
	}
	instance, err := c.instanceAPI.Get(ctx, &computepb.GetInstanceRequest{
		Instance: instanceName,
		Project:  project,
		Zone:     zone,
	})
	if err != nil {
		return "", err
	}
	// first disk is always the boot disk
	if len(instance.Disks) < 1 {
		return "", fmt.Errorf("instance %v has no disks", instanceName)
	}
	if instance.Disks[0] == nil || instance.Disks[0].Source == nil {
		return "", fmt.Errorf("instance %q has invalid disk", instanceName)
	}
	diskReq, err := diskSourceToDiskReq(*instance.Disks[0].Source)
	if err != nil {
		return "", err
	}
	disk, err := c.diskAPI.Get(ctx, diskReq)
	if err != nil {
		return "", err
	}
	if disk.SourceImage == nil {
		return "", fmt.Errorf("disk %q has no source image", diskReq.Disk)
	}
	return uriNormalize(*disk.SourceImage), nil
}

// GetScalingGroupID returns the scaling group ID of the node.
func (c *Client) GetScalingGroupID(ctx context.Context, providerID string) (string, error) {
	project, zone, instanceName, err := splitProviderID(providerID)
	if err != nil {
		return "", err
	}
	instance, err := c.instanceAPI.Get(ctx, &computepb.GetInstanceRequest{
		Instance: instanceName,
		Project:  project,
		Zone:     zone,
	})
	if err != nil {
		return "", fmt.Errorf("getting instance %q: %w", instanceName, err)
	}
	scalingGroupID := getMetadataByKey(instance.Metadata, "created-by")
	if scalingGroupID == "" {
		return "", fmt.Errorf("instance %q has no created-by metadata", instanceName)
	}
	scalingGroupID, err = c.canonicalInstanceGroupID(ctx, scalingGroupID)
	if err != nil {
		return "", err
	}
	return scalingGroupID, nil
}

// CreateNode creates a node in the specified scaling group.
func (c *Client) CreateNode(ctx context.Context, scalingGroupID string) (nodeName, providerID string, err error) {
	project, region, instanceGroupName, err := splitInstanceGroupID(scalingGroupID)
	if err != nil {
		return "", "", err
	}
	project, err = c.canonicalProjectID(ctx, project)
	if err != nil {
		return "", "", err
	}
	instanceGroupManager, err := c.regionInstanceGroupManagersAPI.Get(ctx, &computepb.GetRegionInstanceGroupManagerRequest{
		InstanceGroupManager: instanceGroupName,
		Project:              project,
		Region:               region,
	})
	if err != nil {
		return "", "", err
	}
	if instanceGroupManager.BaseInstanceName == nil {
		return "", "", fmt.Errorf("instance group manager %q has no base instance name", instanceGroupName)
	}
	instanceName := generateInstanceName(*instanceGroupManager.BaseInstanceName, c.prng)
	op, err := c.regionInstanceGroupManagersAPI.CreateInstances(ctx, &computepb.CreateInstancesRegionInstanceGroupManagerRequest{
		InstanceGroupManager: instanceGroupName,
		Project:              project,
		Region:               region,
		RegionInstanceGroupManagersCreateInstancesRequestResource: &computepb.RegionInstanceGroupManagersCreateInstancesRequest{
			Instances: []*computepb.PerInstanceConfig{
				{Name: proto.String(instanceName)},
			},
		},
	})
	if err != nil {
		return "", "", err
	}
	if err := op.Wait(ctx); err != nil {
		return "", "", err
	}

	managedInstanceIter := c.regionInstanceGroupManagersAPI.ListManagedInstances(ctx,
		&computepb.ListManagedInstancesRegionInstanceGroupManagersRequest{
			InstanceGroupManager: instanceGroupName,
			Project:              project,
			Region:               region,
			Filter:               proto.String(fmt.Sprintf("instance eq '.*%s'", instanceName)),
		},
	)
	managedInstance, err := managedInstanceIter.Next()
	if err != nil {
		return "", "", fmt.Errorf("getting managed instance %q: %w", instanceName, err)
	}
	if _, err := managedInstanceIter.Next(); err != iterator.Done {
		return "", "", fmt.Errorf("expected 1 managed instance with name %q but found multiple", instanceName)
	}
	if managedInstance.Instance == nil {
		return "", "", errors.New("ListManagedInstances returned managedInstance with empty instance field")
	}

	_, zone, _, err := splitInstanceID(uriNormalize(*managedInstance.Instance))
	if err != nil {
		return "", "", fmt.Errorf("parsing zone of managed instance %q: %w", instanceName, err)
	}

	return instanceName, joinProviderID(project, zone, instanceName), nil
}

// DeleteNode deletes a node specified by its provider ID.
func (c *Client) DeleteNode(ctx context.Context, providerID string) error {
	_, zone, instanceName, err := splitProviderID(providerID)
	if err != nil {
		return err
	}
	scalingGroupID, err := c.GetScalingGroupID(ctx, providerID)
	if err != nil {
		return err
	}
	instanceGroupProject, instanceGroupRegion, instanceGroupName, err := splitInstanceGroupID(scalingGroupID)
	if err != nil {
		return err
	}
	instanceID := joinInstanceID(zone, instanceName)
	op, err := c.regionInstanceGroupManagersAPI.DeleteInstances(ctx, &computepb.DeleteInstancesRegionInstanceGroupManagerRequest{
		InstanceGroupManager: instanceGroupName,
		Project:              instanceGroupProject,
		Region:               instanceGroupRegion,
		RegionInstanceGroupManagersDeleteInstancesRequestResource: &computepb.RegionInstanceGroupManagersDeleteInstancesRequest{
			Instances: []string{instanceID},
		},
	})
	if err != nil {
		return fmt.Errorf("deleting instance %q from instance group manager %q: %w", instanceID, scalingGroupID, err)
	}
	return op.Wait(ctx)
}

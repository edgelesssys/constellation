/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/apiv1/computepb"
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
	project, zone, instanceGroupName, err := splitInstanceGroupID(scalingGroupID)
	if err != nil {
		return "", "", err
	}
	project, err = c.canonicalProjectID(ctx, project)
	if err != nil {
		return "", "", err
	}
	instanceGroupManager, err := c.instanceGroupManagersAPI.Get(ctx, &computepb.GetInstanceGroupManagerRequest{
		InstanceGroupManager: instanceGroupName,
		Project:              project,
		Zone:                 zone,
	})
	if err != nil {
		return "", "", err
	}
	if instanceGroupManager.BaseInstanceName == nil {
		return "", "", fmt.Errorf("instance group manager %q has no base instance name", instanceGroupName)
	}
	instanceName := generateInstanceName(*instanceGroupManager.BaseInstanceName, c.prng)
	op, err := c.instanceGroupManagersAPI.CreateInstances(ctx, &computepb.CreateInstancesInstanceGroupManagerRequest{
		InstanceGroupManager: instanceGroupName,
		Project:              project,
		Zone:                 zone,
		InstanceGroupManagersCreateInstancesRequestResource: &computepb.InstanceGroupManagersCreateInstancesRequest{
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
	instanceGroupProject, instanceGroupZone, instanceGroupName, err := splitInstanceGroupID(scalingGroupID)
	if err != nil {
		return err
	}
	instanceID := joinInstanceID(zone, instanceName)
	op, err := c.instanceGroupManagersAPI.DeleteInstances(ctx, &computepb.DeleteInstancesInstanceGroupManagerRequest{
		InstanceGroupManager: instanceGroupName,
		Project:              instanceGroupProject,
		Zone:                 instanceGroupZone,
		InstanceGroupManagersDeleteInstancesRequestResource: &computepb.InstanceGroupManagersDeleteInstancesRequest{
			Instances: []string{instanceID},
			SkipInstancesOnValidationError: toPtr(true),
		},
	})
	if err != nil {
		return fmt.Errorf("deleting instance %q from instance group manager %q: %w", instanceID, scalingGroupID, err)
	}
	return op.Wait(ctx)
}

func toPtr[T any](v T) *T {
	return &v
}

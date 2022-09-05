/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

// GetScalingGroupImage returns the image URI of the scaling group.
func (c *Client) GetScalingGroupImage(ctx context.Context, scalingGroupID string) (string, error) {
	instanceTemplate, err := c.getScalingGroupTemplate(ctx, scalingGroupID)
	if err != nil {
		return "", err
	}
	return instanceTemplateSourceImage(instanceTemplate)
}

// SetScalingGroupImage sets the image URI of the scaling group.
func (c *Client) SetScalingGroupImage(ctx context.Context, scalingGroupID, imageURI string) error {
	project, zone, instanceGroupName, err := splitInstanceGroupID(scalingGroupID)
	if err != nil {
		return err
	}
	// get current template
	instanceTemplate, err := c.getScalingGroupTemplate(ctx, scalingGroupID)
	if err != nil {
		return err
	}
	// check if template already uses the same image
	oldImageURI, err := instanceTemplateSourceImage(instanceTemplate)
	if err != nil {
		return err
	}
	if oldImageURI == imageURI {
		return nil
	}

	// clone template with desired image
	if instanceTemplate.Name == nil {
		return fmt.Errorf("instance template of scaling group %q has no name", scalingGroupID)
	}
	instanceTemplate.Properties.Disks[0].InitializeParams.SourceImage = &imageURI
	newTemplateName, err := generateInstanceTemplateName(*instanceTemplate.Name)
	if err != nil {
		return err
	}
	instanceTemplate.Name = &newTemplateName
	op, err := c.instanceTemplateAPI.Insert(ctx, &computepb.InsertInstanceTemplateRequest{
		Project:                  project,
		InstanceTemplateResource: instanceTemplate,
	})
	if err != nil {
		return fmt.Errorf("cloning instance template: %w", err)
	}
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("waiting for cloned instance template: %w", err)
	}

	newTemplateURI := joinInstanceTemplateURI(project, newTemplateName)
	// update instance group manager to use new template
	op, err = c.instanceGroupManagersAPI.SetInstanceTemplate(ctx, &computepb.SetInstanceTemplateInstanceGroupManagerRequest{
		InstanceGroupManager: instanceGroupName,
		Project:              project,
		Zone:                 zone,
		InstanceGroupManagersSetInstanceTemplateRequestResource: &computepb.InstanceGroupManagersSetInstanceTemplateRequest{
			InstanceTemplate: &newTemplateURI,
		},
	})
	if err != nil {
		return fmt.Errorf("setting instance template: %w", err)
	}
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("waiting for setting instance template: %w", err)
	}
	return nil
}

// GetScalingGroupName retrieves the name of a scaling group.
func (c *Client) GetScalingGroupName(ctx context.Context, scalingGroupID string) (string, error) {
	_, _, instanceGroupName, err := splitInstanceGroupID(scalingGroupID)
	if err != nil {
		return "", fmt.Errorf("getting scaling group name: %w", err)
	}
	return strings.ToLower(instanceGroupName), nil
}

// ListScalingGroups retrieves a list of scaling groups for the cluster.
func (c *Client) ListScalingGroups(ctx context.Context, uid string) (controlPlaneGroupIDs []string, workerGroupIDs []string, err error) {
	iter := c.instanceGroupManagersAPI.AggregatedList(ctx, &computepb.AggregatedListInstanceGroupManagersRequest{
		Filter:  proto.String(fmt.Sprintf("name eq \".+-.+-%s\"", uid)), // filter by constellation UID
		Project: c.projectID,
	})
	for instanceGroupManagerScopedListPair, err := iter.Next(); ; instanceGroupManagerScopedListPair, err = iter.Next() {
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("listing instance group managers: %w", err)
		}
		if instanceGroupManagerScopedListPair.Value == nil {
			continue
		}
		for _, instanceGroupManager := range instanceGroupManagerScopedListPair.Value.InstanceGroupManagers {
			if instanceGroupManager == nil || instanceGroupManager.Name == nil || instanceGroupManager.SelfLink == nil {
				continue
			}
			groupID, err := c.canonicalInstanceGroupID(ctx, *instanceGroupManager.SelfLink)
			if err != nil {
				return nil, nil, fmt.Errorf("normalizing instance group ID: %w", err)
			}

			if isControlPlaneInstanceGroup(*instanceGroupManager.Name) {
				controlPlaneGroupIDs = append(controlPlaneGroupIDs, groupID)
			} else if isWorkerInstanceGroup(*instanceGroupManager.Name) {
				workerGroupIDs = append(workerGroupIDs, groupID)
			}
		}
	}
	return controlPlaneGroupIDs, workerGroupIDs, nil
}

func (c *Client) getScalingGroupTemplate(ctx context.Context, scalingGroupID string) (*computepb.InstanceTemplate, error) {
	project, zone, instanceGroupName, err := splitInstanceGroupID(scalingGroupID)
	if err != nil {
		return nil, err
	}
	instanceGroupManager, err := c.instanceGroupManagersAPI.Get(ctx, &computepb.GetInstanceGroupManagerRequest{
		InstanceGroupManager: instanceGroupName,
		Project:              project,
		Zone:                 zone,
	})
	if err != nil {
		return nil, fmt.Errorf("getting instance group manager %q: %w", instanceGroupName, err)
	}
	if instanceGroupManager.InstanceTemplate == nil {
		return nil, fmt.Errorf("instance group manager %q has no instance template", instanceGroupName)
	}
	instanceTemplateProject, instanceTemplateName, err := splitInstanceTemplateID(uriNormalize(*instanceGroupManager.InstanceTemplate))
	if err != nil {
		return nil, fmt.Errorf("splitting instance template name: %w", err)
	}
	instanceTemplate, err := c.instanceTemplateAPI.Get(ctx, &computepb.GetInstanceTemplateRequest{
		InstanceTemplate: instanceTemplateName,
		Project:          instanceTemplateProject,
	})
	if err != nil {
		return nil, fmt.Errorf("getting instance template %q: %w", instanceTemplateName, err)
	}
	return instanceTemplate, nil
}

func instanceTemplateSourceImage(instanceTemplate *computepb.InstanceTemplate) (string, error) {
	if instanceTemplate.Properties == nil ||
		len(instanceTemplate.Properties.Disks) == 0 ||
		instanceTemplate.Properties.Disks[0].InitializeParams == nil ||
		instanceTemplate.Properties.Disks[0].InitializeParams.SourceImage == nil {
		return "", errors.New("instance template has no source image")
	}
	return uriNormalize(*instanceTemplate.Properties.Disks[0].InitializeParams.SourceImage), nil
}

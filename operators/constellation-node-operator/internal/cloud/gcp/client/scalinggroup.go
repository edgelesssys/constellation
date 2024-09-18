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

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	cspapi "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/api"
	computeREST "google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
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
	if instanceTemplate.Name == "" {
		return fmt.Errorf("instance template of scaling group %q has no name", scalingGroupID)
	}
	instanceTemplate.Properties.Disks[0].InitializeParams.SourceImage = imageURI
	newTemplateName, err := generateInstanceTemplateName(instanceTemplate.Name)
	if err != nil {
		return err
	}
	instanceTemplate.Name = newTemplateName
	if _, err := c.instanceTemplateAPI.Insert(project, instanceTemplate); err != nil {
		return fmt.Errorf("cloning instance template: %w", err)
	}

	newTemplateURI := joinInstanceTemplateURI(project, newTemplateName)
	// update instance group manager to use new template
	op, err := c.instanceGroupManagersAPI.SetInstanceTemplate(ctx, &computepb.SetInstanceTemplateInstanceGroupManagerRequest{
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
// This keeps the casing of the original name, but Kubernetes requires the name to be lowercase,
// so use strings.ToLower() on the result if using the name in a Kubernetes context.
func (c *Client) GetScalingGroupName(scalingGroupID string) (string, error) {
	_, _, instanceGroupName, err := splitInstanceGroupID(scalingGroupID)
	if err != nil {
		return "", fmt.Errorf("getting scaling group name: %w", err)
	}
	return instanceGroupName, nil
}

// GetAutoscalingGroupName retrieves the name of a scaling group as needed by the cluster-autoscaler.
func (c *Client) GetAutoscalingGroupName(scalingGroupID string) (string, error) {
	project, zone, instanceGroupName, err := splitInstanceGroupID(scalingGroupID)
	if err != nil {
		return "", fmt.Errorf("getting autoscaling scaling group name: %w", err)
	}
	return ensureURIPrefixed(fmt.Sprintf("projects/%s/zones/%s/instanceGroups/%s", project, zone, instanceGroupName)), nil
}

// ListScalingGroups retrieves a list of scaling groups for the cluster.
func (c *Client) ListScalingGroups(ctx context.Context, uid string) ([]cspapi.ScalingGroup, error) {
	results := []cspapi.ScalingGroup{}
	iter := c.instanceGroupManagersAPI.AggregatedList(ctx, &computepb.AggregatedListInstanceGroupManagersRequest{
		Project: c.projectID,
	})
	for instanceGroupManagerScopedListPair, err := iter.Next(); ; instanceGroupManagerScopedListPair, err = iter.Next() {
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("listing instance group managers: %w", err)
		}
		if instanceGroupManagerScopedListPair.Value == nil {
			continue
		}
		for _, grpManager := range instanceGroupManagerScopedListPair.Value.InstanceGroupManagers {
			if grpManager == nil || grpManager.Name == nil || grpManager.SelfLink == nil || grpManager.InstanceTemplate == nil {
				continue
			}

			templateURI := strings.Split(*grpManager.InstanceTemplate, "/")
			if len(templateURI) < 1 {
				continue // invalid template URI
			}
			template, err := c.instanceTemplateAPI.Get(c.projectID, templateURI[len(templateURI)-1])
			if err != nil {
				return nil, fmt.Errorf("getting instance template: %w", err)
			}
			if template.Properties == nil || template.Properties.Labels == nil {
				continue
			}
			if template.Properties.Labels["constellation-uid"] != uid {
				continue
			}

			groupID, err := c.canonicalInstanceGroupID(ctx, *grpManager.SelfLink)
			if err != nil {
				return nil, fmt.Errorf("normalizing instance group ID: %w", err)
			}

			role := updatev1alpha1.NodeRoleFromString(template.Properties.Labels["constellation-role"])

			name, err := c.GetScalingGroupName(groupID)
			if err != nil {
				return nil, fmt.Errorf("getting scaling group name: %w", err)
			}

			nodeGroupName := template.Properties.Labels["constellation-node-group"]
			// fallback for legacy clusters
			// TODO(malt3): remove this fallback once we can assume all clusters have the correct labels
			if nodeGroupName == "" {
				switch role {
				case updatev1alpha1.ControlPlaneRole:
					nodeGroupName = constants.ControlPlaneDefault
				case updatev1alpha1.WorkerRole:
					nodeGroupName = constants.WorkerDefault
				}
			}

			autoscalerGroupName, err := c.GetAutoscalingGroupName(groupID)
			if err != nil {
				return nil, fmt.Errorf("getting autoscaling group name: %w", err)
			}

			results = append(results, cspapi.ScalingGroup{
				Name:                 name,
				NodeGroupName:        nodeGroupName,
				GroupID:              groupID,
				AutoscalingGroupName: autoscalerGroupName,
				Role:                 role,
			})
		}
	}
	return results, nil
}

func (c *Client) getScalingGroupTemplate(ctx context.Context, scalingGroupID string) (*computeREST.InstanceTemplate, error) {
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
	instanceTemplate, err := c.instanceTemplateAPI.Get(instanceTemplateProject, instanceTemplateName)
	if err != nil {
		return nil, fmt.Errorf("getting instance template %q: %w", instanceTemplateName, err)
	}
	return instanceTemplate, nil
}

func instanceTemplateSourceImage(instanceTemplate *computeREST.InstanceTemplate) (string, error) {
	if instanceTemplate.Properties == nil ||
		len(instanceTemplate.Properties.Disks) == 0 ||
		instanceTemplate.Properties.Disks[0].InitializeParams == nil ||
		instanceTemplate.Properties.Disks[0].InitializeParams.SourceImage == "" {
		return "", errors.New("instance template has no source image")
	}
	return uriNormalize(instanceTemplate.Properties.Disks[0].InitializeParams.SourceImage), nil
}

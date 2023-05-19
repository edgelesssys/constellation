/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	scalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// GetScalingGroupImage returns the image URI of the scaling group.
func (c *Client) GetScalingGroupImage(ctx context.Context, scalingGroupID string) (string, error) {
	launchTemplate, err := c.getScalingGroupTemplate(ctx, scalingGroupID)
	if err != nil {
		return "", err
	}

	if launchTemplate.LaunchTemplateData == nil {
		return "", fmt.Errorf("launch template data is nil for scaling group %q", scalingGroupID)
	}

	if launchTemplate.LaunchTemplateData.ImageId == nil {
		return "", fmt.Errorf("image ID is nil for scaling group %q", scalingGroupID)
	}

	return *launchTemplate.LaunchTemplateData.ImageId, nil
}

// SetScalingGroupImage sets the image URI of the scaling group.
func (c *Client) SetScalingGroupImage(ctx context.Context, scalingGroupID, imageURI string) error {
	launchTemplate, err := c.getScalingGroupTemplate(ctx, scalingGroupID)
	if err != nil {
		return fmt.Errorf("failed to get launch template for scaling group %q: %w", scalingGroupID, err)
	}

	if launchTemplate.VersionNumber == nil {
		return fmt.Errorf("version number is nil for scaling group %q", scalingGroupID)
	}

	createLaunchTemplateOut, err := c.ec2Client.CreateLaunchTemplateVersion(
		ctx,
		&ec2.CreateLaunchTemplateVersionInput{
			LaunchTemplateData: &ec2types.RequestLaunchTemplateData{
				ImageId: &imageURI,
			},
			LaunchTemplateId: launchTemplate.LaunchTemplateId,
			SourceVersion:    toPtr(fmt.Sprintf("%d", *launchTemplate.VersionNumber)),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create launch template version: %w", err)
	}

	if createLaunchTemplateOut == nil {
		return fmt.Errorf("create launch template version output is nil")
	}
	if createLaunchTemplateOut.LaunchTemplateVersion == nil {
		return fmt.Errorf("created launch template version is nil")
	}
	if createLaunchTemplateOut.LaunchTemplateVersion.VersionNumber == nil {
		return fmt.Errorf("created launch template version number is nil")
	}

	// set created version as default
	_, err = c.ec2Client.ModifyLaunchTemplate(
		ctx,
		&ec2.ModifyLaunchTemplateInput{
			LaunchTemplateId: launchTemplate.LaunchTemplateId,
			DefaultVersion:   toPtr(fmt.Sprintf("%d", createLaunchTemplateOut.LaunchTemplateVersion.VersionNumber)),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to modify launch template: %w", err)
	}

	return nil
}

func (c *Client) getScalingGroupTemplate(ctx context.Context, scalingGroupID string) (ec2types.LaunchTemplateVersion, error) {
	groupOutput, err := c.scalingClient.DescribeAutoScalingGroups(
		ctx,
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []string{scalingGroupID},
		},
	)
	if err != nil {
		return ec2types.LaunchTemplateVersion{}, fmt.Errorf("failed to describe scaling group %q: %w", scalingGroupID, err)
	}

	if len(groupOutput.AutoScalingGroups) != 1 {
		return ec2types.LaunchTemplateVersion{}, fmt.Errorf("expected exactly one scaling group, got %d", len(groupOutput.AutoScalingGroups))
	}

	if groupOutput.AutoScalingGroups[0].LaunchTemplate == nil {
		return ec2types.LaunchTemplateVersion{}, fmt.Errorf("launch template is nil for scaling group %q", scalingGroupID)
	}

	if groupOutput.AutoScalingGroups[0].LaunchTemplate.LaunchTemplateId == nil {
		return ec2types.LaunchTemplateVersion{}, fmt.Errorf("launch template ID is nil for scaling group %q", scalingGroupID)
	}

	launchTemplateID := groupOutput.AutoScalingGroups[0].LaunchTemplate.LaunchTemplateId

	launchTemplateOutput, err := c.ec2Client.DescribeLaunchTemplateVersions(
		ctx,
		&ec2.DescribeLaunchTemplateVersionsInput{
			LaunchTemplateId: launchTemplateID,
			Versions:         []string{"$Latest"},
		},
	)
	if err != nil {
		return ec2types.LaunchTemplateVersion{}, fmt.Errorf("failed to describe launch template %q: %w", *launchTemplateID, err)
	}

	if len(launchTemplateOutput.LaunchTemplateVersions) != 1 {
		return ec2types.LaunchTemplateVersion{}, fmt.Errorf("expected exactly one launch template, got %d", len(launchTemplateOutput.LaunchTemplateVersions))
	}
	return launchTemplateOutput.LaunchTemplateVersions[0], nil
}

// GetScalingGroupName retrieves the name of a scaling group.
// This keeps the casing of the original name, but Kubernetes requires the name to be lowercase,
// so use strings.ToLower() on the result if using the name in a Kubernetes context.
func (c *Client) GetScalingGroupName(scalingGroupID string) (string, error) {
	return strings.ToLower(scalingGroupID), nil
}

// GetAutoscalingGroupName retrieves the name of a scaling group as needed by the cluster-autoscaler.
func (c *Client) GetAutoscalingGroupName(scalingGroupID string) (string, error) {
	return scalingGroupID, nil
}

// ListScalingGroups retrieves a list of scaling groups for the cluster.
func (c *Client) ListScalingGroups(ctx context.Context, uid string) (controlPlaneGroupIDs []string, workerGroupIDs []string, err error) {
	output, err := c.scalingClient.DescribeAutoScalingGroups(
		ctx,
		&autoscaling.DescribeAutoScalingGroupsInput{
			Filters: []scalingtypes.Filter{
				{
					Name:   toPtr("tag:constellation-uid"),
					Values: []string{uid},
				},
			},
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to describe scaling groups: %w", err)
	}

	for _, group := range output.AutoScalingGroups {
		if group.Tags == nil {
			continue
		}
		for _, tag := range group.Tags {
			if *tag.Key == "constellation-role" {
				if *tag.Value == "control-plane" {
					controlPlaneGroupIDs = append(controlPlaneGroupIDs, *group.AutoScalingGroupName)
				} else if *tag.Value == "worker" {
					workerGroupIDs = append(workerGroupIDs, *group.AutoScalingGroupName)
				}
			}
		}
	}
	return controlPlaneGroupIDs, workerGroupIDs, nil
}

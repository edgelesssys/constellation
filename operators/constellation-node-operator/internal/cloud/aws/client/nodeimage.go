/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// GetNodeImage returns the image name of the node.
func (c *Client) GetNodeImage(ctx context.Context, providerID string) (string, error) {
	instanceName, err := getInstanceNameFromProviderID(providerID)
	if err != nil {
		return "", fmt.Errorf("failed to get instance name from providerID: %w", err)
	}

	params := &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			instanceName,
		},
	}

	resp, err := c.ec2Client.DescribeInstances(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to describe instances: %w", err)
	}

	if len(resp.Reservations) == 0 {
		return "", fmt.Errorf("no reservations for instance %q", instanceName)
	}

	if len(resp.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("no instances for instance %q", instanceName)
	}

	if resp.Reservations[0].Instances[0].ImageId == nil {
		return "", fmt.Errorf("no image for instance %q", instanceName)
	}

	return *resp.Reservations[0].Instances[0].ImageId, nil
}

// GetScalingGroupID returns the scaling group ID of the node.
func (c *Client) GetScalingGroupID(ctx context.Context, providerID string) (string, error) {
	instanceName, err := getInstanceNameFromProviderID(providerID)
	if err != nil {
		return "", fmt.Errorf("failed to get instance name from providerID: %w", err)
	}
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			instanceName,
		},
	}

	resp, err := c.ec2Client.DescribeInstances(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to describe instances: %w", err)
	}

	if len(resp.Reservations) == 0 {
		return "", fmt.Errorf("no reservations for instance %q", instanceName)
	}

	if len(resp.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("no instances for instance %q", instanceName)
	}

	if resp.Reservations[0].Instances[0].Tags == nil {
		return "", fmt.Errorf("no tags for instance %q", instanceName)
	}

	for _, tag := range resp.Reservations[0].Instances[0].Tags {
		if tag.Key == nil || tag.Value == nil {
			continue
		}
		if *tag.Key == "aws:autoscaling:groupName" {
			return *tag.Value, nil
		}
	}

	return "", fmt.Errorf("node %q does not have valid tags", providerID)
}

// CreateNode creates a node in the specified scaling group.
func (c *Client) CreateNode(ctx context.Context, scalingGroupID string) (nodeName, providerID string, err error) {
	containsInstance := func(instances []types.Instance, target types.Instance) bool {
		for _, i := range instances {
			if i.InstanceId == nil || target.InstanceId == nil {
				continue
			}
			if *i.InstanceId == *target.InstanceId {
				return true
			}
		}
		return false
	}

	// get current capacity
	groups, err := c.scalingClient.DescribeAutoScalingGroups(
		ctx,
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []string{scalingGroupID},
		},
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to describe autoscaling group: %w", err)
	}

	if len(groups.AutoScalingGroups) != 1 {
		return "", "", fmt.Errorf("expected exactly one autoscaling group, got %d", len(groups.AutoScalingGroups))
	}

	if groups.AutoScalingGroups[0].DesiredCapacity == nil {
		return "", "", fmt.Errorf("desired capacity is nil")
	}
	currentCapacity := int(*groups.AutoScalingGroups[0].DesiredCapacity)

	// check for int32 overflow
	if currentCapacity >= int(^uint32(0)>>1) {
		return "", "", fmt.Errorf("current capacity is at maximum")
	}

	// get current list of instances
	previousInstances := groups.AutoScalingGroups[0].Instances

	// create new instance by increasing capacity by 1
	_, err = c.scalingClient.SetDesiredCapacity(
		ctx,
		&autoscaling.SetDesiredCapacityInput{
			AutoScalingGroupName: &scalingGroupID,
			DesiredCapacity:      toPtr(int32(currentCapacity + 1)),
		},
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to set desired capacity: %w", err)
	}

	// poll until new instance is created with 30 second timeout
	newInstance := types.Instance{}
	for i := 0; i < 30; i++ {
		groups, err := c.scalingClient.DescribeAutoScalingGroups(
			ctx,
			&autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []string{scalingGroupID},
			},
		)
		if err != nil {
			return "", "", fmt.Errorf("failed to describe autoscaling group: %w", err)
		}

		if len(groups.AutoScalingGroups) != 1 {
			return "", "", fmt.Errorf("expected exactly one autoscaling group, got %d", len(groups.AutoScalingGroups))
		}

		for _, instance := range groups.AutoScalingGroups[0].Instances {
			if !containsInstance(previousInstances, instance) {
				newInstance = instance
				break
			}
		}

		// break if new instance is found
		if newInstance.InstanceId != nil {
			break
		}

		// wait 1 second
		select {
		case <-ctx.Done():
			return "", "", fmt.Errorf("context cancelled")
		case <-time.After(1 * time.Second):
		}
	}

	if newInstance.InstanceId == nil {
		return "", "", fmt.Errorf("timed out waiting for new instance")
	}

	if newInstance.AvailabilityZone == nil {
		return "", "", fmt.Errorf("new instance %s does not have availability zone", *newInstance.InstanceId)
	}

	// return new instance
	return *newInstance.InstanceId, fmt.Sprintf("aws:///%s/%s", *newInstance.AvailabilityZone, *newInstance.InstanceId), nil
}

// DeleteNode deletes a node from the specified scaling group.
func (c *Client) DeleteNode(ctx context.Context, providerID string) error {
	instanceID, err := getInstanceNameFromProviderID(providerID)
	if err != nil {
		return fmt.Errorf("failed to get instance name from providerID: %w", err)
	}

	_, err = c.scalingClient.TerminateInstanceInAutoScalingGroup(
		ctx,
		&autoscaling.TerminateInstanceInAutoScalingGroupInput{
			InstanceId:                     &instanceID,
			ShouldDecrementDesiredCapacity: toPtr(true),
		},
	)
	if err != nil && !isInstanceNotFoundError(err) {
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	return nil
}

func toPtr[T any](v T) *T {
	return &v
}

func isInstanceNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Instance Id not found")
}

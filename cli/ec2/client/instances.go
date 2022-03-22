package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/edgelesssys/constellation/cli/ec2"
)

// CreateInstances creates the instances defined in input.
//
// An existing security group is needed to create instances.
func (c *Client) CreateInstances(ctx context.Context, input CreateInput) error {
	if c.securityGroup == "" {
		return errors.New("no security group set")
	}
	input.securityGroupIds = []string{c.securityGroup}

	if err := c.createDryRun(ctx, input); err != nil {
		return err
	}

	resp, err := c.api.RunInstances(ctx, input.AWS())
	if err != nil {
		return fmt.Errorf("failed to create instances: %w", err)
	}

	for _, instance := range resp.Instances {
		id := instance.InstanceId
		if id == nil {
			return errors.New("instanceId is nil pointer")
		}
		c.instances[*id] = ec2.Instance{}
	}

	if err := c.waitStateRunning(ctx); err != nil {
		return err
	}

	if err := c.tagInstances(ctx, input.Tags); err != nil {
		return err
	}

	if err := c.getInstanceIPs(ctx); err != nil {
		return err
	}
	return nil
}

// TerminateInstances terminates all instances of a Client.
func (c *Client) TerminateInstances(ctx context.Context) error {
	if len(c.instances) == 0 {
		return nil
	}

	input := &awsec2.TerminateInstancesInput{
		InstanceIds: c.instances.IDs(),
	}
	if err := c.terminateDryRun(ctx, *input); err != nil {
		return err
	}

	if _, err := c.api.TerminateInstances(ctx, input); err != nil {
		return err
	}

	if err := c.waitStateTerminated(ctx); err != nil {
		return err
	}
	c.instances = ec2.Instances{}
	return nil
}

// waitStateRunning waits until all the client's instances reached the running state.
//
// A set of instances is also considered to be running if at least one of the
// instances' state is 'running' and the other instances have a nil state.
func (c *Client) waitStateRunning(ctx context.Context) error {
	if len(c.instances) == 0 {
		return errors.New("client has no instances")
	}
	describeInput := &awsec2.DescribeInstancesInput{
		InstanceIds: c.instances.IDs(),
	}
	waiter := awsec2.NewInstanceRunningWaiter(c.api)
	return waiter.Wait(ctx, describeInput, c.timeout)
}

// waitStateTerminated waits until all the client's instances reached the terminated state.
//
// A set of instances is also considered to be terminated if at least one of the
// instances' state is 'terminated' and the other instances have a nil state.
func (c *Client) waitStateTerminated(ctx context.Context) error {
	if len(c.instances) == 0 {
		return errors.New("client has no instances")
	}

	describeInput := &awsec2.DescribeInstancesInput{
		InstanceIds: c.instances.IDs(),
	}
	waiter := awsec2.NewInstanceTerminatedWaiter(c.api)
	return waiter.Wait(ctx, describeInput, c.timeout)
}

// tagInstances tags all instances of a client with a given set of tags.
func (c *Client) tagInstances(ctx context.Context, tags ec2.Tags) error {
	if len(c.instances) == 0 {
		return errors.New("client has no instances")
	}

	tagInput := &awsec2.CreateTagsInput{
		Resources: c.instances.IDs(),
		Tags:      tags.AWS(),
	}
	if _, err := c.api.CreateTags(ctx, tagInput); err != nil {
		return fmt.Errorf("failed to tag instances: %w", err)
	}
	return nil
}

// createDryRun checks if user has the privilege to create the instances
// which were defined in input.
func (c *Client) createDryRun(ctx context.Context, input CreateInput) error {
	runInput := input.AWS()
	runInput.DryRun = aws.Bool(true)
	_, err := c.api.RunInstances(ctx, runInput)
	return checkDryRunError(err)
}

// terminateDryRun checks if user has the privilege to terminate the instances
// which were defined in input.
func (c *Client) terminateDryRun(ctx context.Context, input awsec2.TerminateInstancesInput) error {
	input.DryRun = aws.Bool(true)
	_, err := c.api.TerminateInstances(ctx, &input)
	return checkDryRunError(err)
}

// getInstanceIPs queries the private and public IP addresses
// and adds the information to each instance.
//
// The instances must be in 'running' state.
func (c *Client) getInstanceIPs(ctx context.Context) error {
	describeInput := &awsec2.DescribeInstancesInput{
		InstanceIds: c.instances.IDs(),
	}
	paginator := awsec2.NewDescribeInstancesPaginator(c.api, describeInput)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, reservation := range output.Reservations {
			for _, instanceDescription := range reservation.Instances {
				if instanceDescription.InstanceId == nil {
					return errors.New("instanceId is nil pointer")
				}
				if instanceDescription.PublicIpAddress == nil {
					return errors.New("publicIpAddress is nil pointer")
				}
				if instanceDescription.PrivateIpAddress == nil {
					return errors.New("privateIpAddress is nil pointer")
				}
				instance, ok := c.instances[*instanceDescription.InstanceId]
				if !ok {
					return errors.New("got an instance description to an unknown instanceId")
				}
				instance.PublicIP = *instanceDescription.PublicIpAddress
				instance.PrivateIP = *instanceDescription.PrivateIpAddress
				c.instances[*instanceDescription.InstanceId] = instance
			}
		}
	}
	return nil
}

// CreateInput defines the propertis of the instances to create.
type CreateInput struct {
	ImageId          string
	InstanceType     string
	Count            int
	Tags             ec2.Tags
	securityGroupIds []string
}

// AWS creates a AWS ec2.RunInstancesInput from an CreateInput.
func (ci *CreateInput) AWS() *awsec2.RunInstancesInput {
	return &awsec2.RunInstancesInput{
		ImageId:          aws.String(ci.ImageId),
		InstanceType:     ec2.InstanceTypes[ci.InstanceType],
		MaxCount:         aws.Int32(int32(ci.Count)),
		MinCount:         aws.Int32(int32(ci.Count)),
		EnclaveOptions:   &types.EnclaveOptionsRequest{Enabled: aws.Bool(true)},
		SecurityGroupIds: ci.securityGroupIds,
	}
}

package cmd

import (
	"context"
	"errors"
	"strconv"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/edgelesssys/constellation/cli/ec2/client"
	"github.com/edgelesssys/constellation/internal/state"
)

type fakeEc2Client struct {
	instances     ec2.Instances
	securityGroup string
	ec2state      []fakeEc2Instance
}

func (c *fakeEc2Client) GetState() (state.ConstellationState, error) {
	if len(c.instances) == 0 {
		return state.ConstellationState{}, errors.New("client has no instances")
	}
	stat := state.ConstellationState{
		CloudProvider:    cloudprovider.AWS.String(),
		EC2Instances:     c.instances,
		EC2SecurityGroup: c.securityGroup,
	}
	for id, instance := range c.instances {
		instance.PrivateIP = "192.0.2.1"
		instance.PublicIP = "192.0.2.2"
		c.instances[id] = instance
	}
	return stat, nil
}

func (c *fakeEc2Client) SetState(stat state.ConstellationState) error {
	if len(stat.EC2Instances) == 0 {
		return errors.New("state has no instances")
	}
	c.instances = stat.EC2Instances
	c.securityGroup = stat.EC2SecurityGroup
	return nil
}

func (c *fakeEc2Client) CreateInstances(_ context.Context, input client.CreateInput) error {
	if c.securityGroup == "" {
		return errors.New("client has no security group")
	}
	if c.instances == nil {
		c.instances = make(ec2.Instances)
	}
	for i := 0; i < input.Count; i++ {
		id := "id-" + strconv.Itoa(len(c.ec2state))
		c.ec2state = append(c.ec2state, fakeEc2Instance{
			state:         running,
			instanceID:    id,
			securityGroup: c.securityGroup,
			tags:          input.Tags,
		})
		c.instances[id] = ec2.Instance{}
	}
	return nil
}

func (c *fakeEc2Client) TerminateInstances(_ context.Context) error {
	if len(c.instances) == 0 {
		return nil
	}
	for _, instance := range c.ec2state {
		instance.state = terminated
	}
	return nil
}

func (c *fakeEc2Client) CreateSecurityGroup(_ context.Context, input client.SecurityGroupInput) error {
	if c.securityGroup != "" {
		return errors.New("client already has a security group")
	}
	c.securityGroup = "sg-test"
	return nil
}

func (c *fakeEc2Client) DeleteSecurityGroup(_ context.Context) error {
	c.securityGroup = ""
	return nil
}

type ec2InstanceState int

const (
	running = iota
	terminated
)

type fakeEc2Instance struct {
	state         ec2InstanceState
	instanceID    string
	tags          ec2.Tags
	securityGroup string
}

type stubEc2Client struct {
	terminateInstancesCalled  bool
	deleteSecurityGroupCalled bool

	getStateErr            error
	setStateErr            error
	createInstancesErr     error
	terminateInstancesErr  error
	createSecurityGroupErr error
	deleteSecurityGroupErr error
}

func (c *stubEc2Client) GetState() (state.ConstellationState, error) {
	return state.ConstellationState{}, c.getStateErr
}

func (c *stubEc2Client) SetState(stat state.ConstellationState) error {
	return c.setStateErr
}

func (c *stubEc2Client) CreateInstances(_ context.Context, input client.CreateInput) error {
	return c.createInstancesErr
}

func (c *stubEc2Client) TerminateInstances(_ context.Context) error {
	c.terminateInstancesCalled = true
	return c.terminateInstancesErr
}

func (c *stubEc2Client) CreateSecurityGroup(_ context.Context, input client.SecurityGroupInput) error {
	return c.createSecurityGroupErr
}

func (c *stubEc2Client) DeleteSecurityGroup(_ context.Context) error {
	c.deleteSecurityGroupCalled = true
	return c.deleteSecurityGroupErr
}

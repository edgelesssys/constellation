package client

import (
	"context"
	"errors"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/state"
)

// Client for the AWS EC2 API.
type Client struct {
	api           api
	instances     ec2.Instances
	securityGroup string
	timeout       time.Duration
}

func newClient(api api) (*Client, error) {
	return &Client{
		api:       api,
		instances: make(map[string]ec2.Instance),
		timeout:   2 * time.Minute,
	}, nil
}

// NewFromDefault creates a Client from the default config.
func NewFromDefault(ctx context.Context) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return newClient(awsec2.NewFromConfig(cfg))
}

// GetState returns the current configuration of the Constellation,
// which can be stored and used through later CLI commands.
func (c *Client) GetState() (state.ConstellationState, error) {
	if len(c.instances) == 0 {
		return state.ConstellationState{}, errors.New("client has no instances")
	}
	if c.securityGroup == "" {
		return state.ConstellationState{}, errors.New("client has no security group")
	}
	return state.ConstellationState{
		CloudProvider:    cloudprovider.AWS.String(),
		EC2Instances:     c.instances,
		EC2SecurityGroup: c.securityGroup,
	}, nil
}

// SetState sets a Client to an existing configuration.
func (c *Client) SetState(stat state.ConstellationState) error {
	if stat.CloudProvider != cloudprovider.AWS.String() {
		return errors.New("state is not aws state")
	}
	if len(stat.EC2Instances) == 0 {
		return errors.New("state has no instances")
	}
	if stat.EC2SecurityGroup == "" {
		return errors.New("state has no security group")
	}
	c.instances = stat.EC2Instances
	c.securityGroup = stat.EC2SecurityGroup
	return nil
}

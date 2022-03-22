package client

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
	"github.com/google/uuid"
)

// CreateSecurityGroup creates a AWS security group with the handed properties.
func (c *Client) CreateSecurityGroup(ctx context.Context, input SecurityGroupInput) error {
	if c.securityGroup != "" {
		return errors.New("client already has a security group")
	}

	id := uuid.New()
	createInput := &awsec2.CreateSecurityGroupInput{
		Description: aws.String("Security group of Constellation. This group was generated through the Constellation CLI."),
		GroupName:   aws.String("Constellation-" + id.String()),
		DryRun:      aws.Bool(true),
	}

	// DryRun
	_, err := c.api.CreateSecurityGroup(ctx, createInput)
	if err := checkDryRunError(err); err != nil {
		return err
	}
	createInput.DryRun = aws.Bool(false)

	// Create
	out, err := c.api.CreateSecurityGroup(ctx, createInput)
	if err != nil {
		return err
	}
	if out.GroupId == nil {
		return errors.New("security group creation didn't return an id")
	}
	c.securityGroup = *out.GroupId

	// Authorize.
	return c.authorizeSecurityGroup(ctx, input)
}

// DeleteSecurityGroup deletes the security group of the client.
// The deletion is idempotent, no error is returned if the client has
// an empty securityGroupID.
func (c *Client) DeleteSecurityGroup(ctx context.Context) error {
	if c.securityGroup == "" {
		return nil
	}

	input := &awsec2.DeleteSecurityGroupInput{
		GroupId: aws.String(c.securityGroup),
		DryRun:  aws.Bool(true),
	}

	// DryRun
	_, err := c.api.DeleteSecurityGroup(ctx, input)
	if err := checkDryRunError(err); err != nil {
		return err
	}
	input.DryRun = aws.Bool(false)

	// Delete
	if _, err := c.api.DeleteSecurityGroup(ctx, input); err != nil {
		return err
	}
	c.securityGroup = ""
	return nil
}

func (c *Client) authorizeSecurityGroup(ctx context.Context, input SecurityGroupInput) error {
	if c.securityGroup == "" {
		return errors.New("client hasn't got a security group id")
	}

	if err := c.authorizeSecurityGroupIngress(ctx, input.Inbound); err != nil {
		return err
	}

	return c.authorizeSecurityGroupEgress(ctx, input.Outbound)
}

func (c *Client) authorizeSecurityGroupIngress(ctx context.Context, perms cloudtypes.Firewall) error {
	if len(perms) == 0 {
		return nil
	}

	authInput := &awsec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       aws.String(c.securityGroup),
		IpPermissions: perms.AWS(),
		DryRun:        aws.Bool(true),
	}

	// DryRun
	_, err := c.api.AuthorizeSecurityGroupIngress(ctx, authInput)
	if err := checkDryRunError(err); err != nil {
		return err
	}
	authInput.DryRun = aws.Bool(false)

	// Authorize
	_, err = c.api.AuthorizeSecurityGroupIngress(ctx, authInput)
	return err
}

func (c *Client) authorizeSecurityGroupEgress(ctx context.Context, perms cloudtypes.Firewall) error {
	if len(perms) == 0 {
		return nil
	}

	authInput := &awsec2.AuthorizeSecurityGroupEgressInput{
		GroupId:       aws.String(c.securityGroup),
		IpPermissions: perms.AWS(),
		DryRun:        aws.Bool(true),
	}

	// DryRun
	_, err := c.api.AuthorizeSecurityGroupEgress(ctx, authInput)
	if err := checkDryRunError(err); err != nil {
		return err
	}
	authInput.DryRun = aws.Bool(false)

	// Authorize
	_, err = c.api.AuthorizeSecurityGroupEgress(ctx, authInput)
	return err
}

type SecurityGroupInput struct {
	Inbound  cloudtypes.Firewall
	Outbound cloudtypes.Firewall
}

package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// api collects used functions of AWS' ec2.Client as interfaces to enable testing.
type api interface {
	ec2.DescribeInstancesAPIClient

	// Instances
	RunInstances(ctx context.Context,
		params *ec2.RunInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)

	TerminateInstances(ctx context.Context,
		params *ec2.TerminateInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)

	CreateTags(ctx context.Context,
		params *ec2.CreateTagsInput,
		optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)

	// SecurityGroup
	CreateSecurityGroup(ctx context.Context,
		params *ec2.CreateSecurityGroupInput,
		optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error)

	DeleteSecurityGroup(ctx context.Context,
		params *ec2.DeleteSecurityGroupInput,
		optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)

	AuthorizeSecurityGroupIngress(ctx context.Context,
		params *ec2.AuthorizeSecurityGroupIngressInput,
		optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error)

	AuthorizeSecurityGroupEgress(ctx context.Context,
		params *ec2.AuthorizeSecurityGroupEgressInput,
		optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupEgressOutput, error)
}

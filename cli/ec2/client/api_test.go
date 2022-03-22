package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

// stubAPI is a stub ec2 api for testing.
type stubAPI struct {
	instances     []types.Instance
	securityGroup types.SecurityGroup

	describeInstancesErr                   error
	runInstancesErr                        error
	runInstancesDryRunErr                  *error
	terminateInstancesErr                  error
	terminateInstancesDryRunErr            *error
	createTagsErr                          error
	createSecurityGroupErr                 error
	createSecurityGroupDryRunErr           *error
	deleteSecurityGroupErr                 error
	deleteSecurityGroupDryRunErr           *error
	authorizeSecurityGroupIngressErr       error
	authorizeSecurityGroupIngressDryRunErr *error
	authorizeSecurityGroupEgressErr        error
	authorizeSecurityGroupEgressDryRunErr  *error
}

func (a stubAPI) DescribeInstances(ctx context.Context,
	params *ec2.DescribeInstancesInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{Instances: a.instances},
		},
	}, a.describeInstancesErr
}

func (a stubAPI) RunInstances(ctx context.Context,
	params *ec2.RunInstancesInput,
	optFns ...func(*ec2.Options),
) (*ec2.RunInstancesOutput, error) {
	if err := getDryRunErr(params.DryRun, a.runInstancesDryRunErr); err != nil {
		return nil, err
	}

	return &ec2.RunInstancesOutput{Instances: a.instances}, a.runInstancesErr
}

func (a stubAPI) CreateTags(ctx context.Context,
	params *ec2.CreateTagsInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateTagsOutput, error) {
	return nil, a.createTagsErr
}

func (a stubAPI) TerminateInstances(ctx context.Context,
	params *ec2.TerminateInstancesInput,
	optFns ...func(*ec2.Options),
) (*ec2.TerminateInstancesOutput, error) {
	if err := getDryRunErr(params.DryRun, a.terminateInstancesDryRunErr); err != nil {
		return nil, err
	}

	return nil, a.terminateInstancesErr
}

func (a stubAPI) CreateSecurityGroup(ctx context.Context,
	params *ec2.CreateSecurityGroupInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateSecurityGroupOutput, error) {
	if err := getDryRunErr(params.DryRun, a.createSecurityGroupDryRunErr); err != nil {
		return nil, err
	}

	return &ec2.CreateSecurityGroupOutput{
		GroupId: a.securityGroup.GroupId,
	}, a.createSecurityGroupErr
}

func (a stubAPI) DeleteSecurityGroup(ctx context.Context,
	params *ec2.DeleteSecurityGroupInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteSecurityGroupOutput, error) {
	if err := getDryRunErr(params.DryRun, a.deleteSecurityGroupDryRunErr); err != nil {
		return nil, err
	}

	return nil, a.deleteSecurityGroupErr
}

func (a stubAPI) AuthorizeSecurityGroupIngress(ctx context.Context,
	params *ec2.AuthorizeSecurityGroupIngressInput,
	optFns ...func(*ec2.Options),
) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	if err := getDryRunErr(params.DryRun, a.authorizeSecurityGroupIngressDryRunErr); err != nil {
		return nil, err
	}

	return nil, a.authorizeSecurityGroupIngressErr
}

func (a stubAPI) AuthorizeSecurityGroupEgress(ctx context.Context,
	params *ec2.AuthorizeSecurityGroupEgressInput,
	optFns ...func(*ec2.Options),
) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	if err := getDryRunErr(params.DryRun, a.authorizeSecurityGroupEgressDryRunErr); err != nil {
		return nil, err
	}

	return nil, a.authorizeSecurityGroupEgressErr
}

func getDryRunErr(dryRun *bool, stubErr *error) error {
	if dryRun == nil || !*dryRun {
		return nil
	}
	if stubErr != nil {
		return *stubErr
	}
	return &smithy.GenericAPIError{Code: "DryRunOperation"}
}

var stateRunning = types.InstanceState{
	Code: aws.Int32(int32(16)),
	Name: types.InstanceStateNameRunning,
}

var stateTerminated = types.InstanceState{
	Code: aws.Int32(48),
	Name: types.InstanceStateNameTerminated,
}

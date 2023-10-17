/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Implements interaction with the AWS API.

Instance metadata is retrieved from the [AWS IMDS API].

Retrieving metadata of other instances is done by using the AWS compute API, and requires AWS credentials.

[AWS IMDS API]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
*/
package aws

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elasticloadbalancingv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	tagType "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

const (
	tagName = "Name"
)

type resourceAPI interface {
	GetResources(context.Context, *resourcegroupstaggingapi.GetResourcesInput, ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error)
}

type loadbalancerAPI interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput,
		optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
}

type ec2API interface {
	DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeAddresses(context.Context, *ec2.DescribeAddressesInput, ...func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error)
}

type imdsAPI interface {
	GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error)
}

// Cloud provides AWS metadata and API access.
type Cloud struct {
	ec2               ec2API
	imds              imdsAPI
	loadbalancer      loadbalancerAPI
	resourceapiClient resourceAPI
}

// New initializes a new AWS Metadata client using instance default credentials.
// Default region is set up using the AWS imds api.
func New(ctx context.Context) (*Cloud, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithEC2IMDSRegion())
	if err != nil {
		return nil, err
	}
	return &Cloud{
		ec2:               ec2.NewFromConfig(cfg),
		imds:              imds.New(imds.Options{}),
		loadbalancer:      elasticloadbalancingv2.NewFromConfig(cfg),
		resourceapiClient: resourcegroupstaggingapi.NewFromConfig(cfg),
	}, nil
}

// List retrieves all instances belonging to the current Constellation.
func (c *Cloud) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	uid, err := c.readInstanceTag(ctx, cloud.TagUID)
	if err != nil {
		return nil, fmt.Errorf("retrieving uid tag: %w", err)
	}
	ec2Instances, err := c.getAllInstancesInGroup(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances: %w", err)
	}

	return c.convertToMetadataInstance(ec2Instances)
}

// Self retrieves the current instance.
func (c *Cloud) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	identity, err := c.imds.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving instance identity: %w", err)
	}

	instanceRole, err := c.readInstanceTag(ctx, cloud.TagRole)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving role tag: %w", err)
	}

	return metadata.InstanceMetadata{
		Name:       identity.InstanceID,
		ProviderID: fmt.Sprintf("aws:///%s/%s", identity.AvailabilityZone, identity.InstanceID),
		Role:       role.FromString(instanceRole),
		VPCIP:      identity.PrivateIP,
	}, nil
}

// UID returns the UID of the Constellation.
func (c *Cloud) UID(ctx context.Context) (string, error) {
	return c.readInstanceTag(ctx, cloud.TagUID)
}

// InitSecretHash returns the InitSecretHash of the current instance.
func (c *Cloud) InitSecretHash(ctx context.Context) ([]byte, error) {
	initSecretHash, err := c.readInstanceTag(ctx, cloud.TagInitSecretHash)
	if err != nil {
		return nil, fmt.Errorf("retrieving init secret hash tag: %w", err)
	}
	return []byte(initSecretHash), nil
}

// GetLoadBalancerEndpoint returns the endpoint of the load balancer.
func (c *Cloud) GetLoadBalancerEndpoint(ctx context.Context) (host, port string, err error) {
	hostname, err := c.getLoadBalancerDNSName(ctx)
	if err != nil {
		return "", "", fmt.Errorf("retrieving load balancer dns name: %w", err)
	}
	return hostname, strconv.FormatInt(constants.KubernetesPort, 10), nil
}

func (c *Cloud) getLoadBalancerDNSName(ctx context.Context) (string, error) {
	loadbalancer, err := c.getLoadBalancer(ctx)
	if err != nil {
		return "", fmt.Errorf("finding Constellation load balancer: %w", err)
	}
	if loadbalancer.DNSName == nil {
		return "", errors.New("load balancer dns name missing")
	}
	return *loadbalancer.DNSName, nil
}

func (c *Cloud) getLoadBalancer(ctx context.Context) (*elasticloadbalancingv2types.LoadBalancer, error) {
	uid, err := c.readInstanceTag(ctx, cloud.TagUID)
	if err != nil {
		return nil, fmt.Errorf("retrieving uid tag: %w", err)
	}
	arns, err := c.getARNsByTag(ctx, uid, "elasticloadbalancing:loadbalancer")
	if err != nil {
		return nil, fmt.Errorf("retrieving load balancer ARNs: %w", err)
	}
	if len(arns) != 1 {
		return nil, fmt.Errorf("%d load balancers found", len(arns))
	}

	output, err := c.loadbalancer.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
		LoadBalancerArns: arns,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving load balancer: %w", err)
	}
	if len(output.LoadBalancers) != 1 {
		return nil, fmt.Errorf("%d load balancers found; expected 1", len(output.LoadBalancers))
	}
	return &output.LoadBalancers[0], nil
}

// getARNsByTag returns a list of ARNs that have the given tag.
func (c *Cloud) getARNsByTag(ctx context.Context, uid, resourceType string) ([]string, error) {
	var ARNs []string
	resourcesReq := &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []tagType.TagFilter{
			{
				Key:    aws.String(cloud.TagUID),
				Values: []string{uid},
			},
		},
		ResourceTypeFilters: []string{resourceType},
	}

	for out, err := c.resourceapiClient.GetResources(ctx, resourcesReq); ; out, err = c.resourceapiClient.GetResources(ctx, resourcesReq) {
		if err != nil {
			return nil, fmt.Errorf("retrieving resources: %w", err)
		}

		for _, resource := range out.ResourceTagMappingList {
			if resource.ResourceARN != nil {
				ARNs = append(ARNs, *resource.ResourceARN)
			}
		}

		if out.PaginationToken == nil || *out.PaginationToken == "" {
			return ARNs, nil
		}
		resourcesReq.PaginationToken = out.PaginationToken
	}
}

func (c *Cloud) getAllInstancesInGroup(ctx context.Context, uid string) ([]ec2Types.Instance, error) {
	var instances []ec2Types.Instance
	instanceReq := &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("tag:" + cloud.TagUID),
				Values: []string{uid},
			},
		},
	}

	for out, err := c.ec2.DescribeInstances(ctx, instanceReq); ; out, err = c.ec2.DescribeInstances(ctx, instanceReq) {
		if err != nil {
			return nil, fmt.Errorf("retrieving instances: %w", err)
		}

		for _, reservation := range out.Reservations {
			instances = append(instances, reservation.Instances...)
		}

		if out.NextToken == nil {
			return instances, nil
		}
		instanceReq.NextToken = out.NextToken
	}
}

func (c *Cloud) convertToMetadataInstance(ec2Instances []ec2Types.Instance) ([]metadata.InstanceMetadata, error) {
	var instances []metadata.InstanceMetadata
	for _, ec2Instance := range ec2Instances {
		// ignore not running instances
		if ec2Instance.State == nil || ec2Instance.State.Name != ec2Types.InstanceStateNameRunning {
			continue
		}

		// sanity checks to avoid panics
		if ec2Instance.InstanceId == nil {
			return nil, errors.New("instance id is nil")
		}
		if ec2Instance.PrivateIpAddress == nil {
			return nil, fmt.Errorf("instance %s has no private IP address", *ec2Instance.InstanceId)
		}

		newInstance := metadata.InstanceMetadata{
			VPCIP: *ec2Instance.PrivateIpAddress,
			Name:  *ec2Instance.InstanceId,
		}

		instanceRole, err := findTag(ec2Instance.Tags, cloud.TagRole)
		if err != nil {
			return nil, fmt.Errorf("retrieving tag for instance %s: %w", *ec2Instance.InstanceId, err)
		}
		newInstance.Role = role.FromString(instanceRole)

		// Set ProviderID
		if ec2Instance.Placement != nil {
			// set to aws:///<region>/<instance-id>
			newInstance.ProviderID = fmt.Sprintf("aws:///%s/%s", *ec2Instance.Placement.AvailabilityZone, *ec2Instance.InstanceId)
		} else {
			// fallback to aws:///<instance-id>
			newInstance.ProviderID = fmt.Sprintf("aws:///%s", *ec2Instance.InstanceId)
		}

		instances = append(instances, newInstance)
	}

	return instances, nil
}

func (c *Cloud) readInstanceTag(ctx context.Context, tag string) (string, error) {
	identity, err := c.imds.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
	if err != nil {
		return "", fmt.Errorf("retrieving instance identity: %w", err)
	}

	if identity == nil {
		return "", errors.New("instance identity is nil")
	}

	out, err := c.ec2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{identity.InstanceID},
	})
	if err != nil {
		return "", fmt.Errorf("descibing instances: %w", err)
	}

	if len(out.Reservations) != 1 || len(out.Reservations[0].Instances) != 1 {
		return "", fmt.Errorf("expected 1 instance, got %d", len(out.Reservations[0].Instances))
	}

	return findTag(out.Reservations[0].Instances[0].Tags, tag)
}

func findTag(tags []ec2Types.Tag, wantKey string) (string, error) {
	for _, tag := range tags {
		if tag.Key == nil || tag.Value == nil {
			continue
		}
		if *tag.Key == wantKey {
			return *tag.Value, nil
		}
	}
	return "", fmt.Errorf("tag %q not found", wantKey)
}

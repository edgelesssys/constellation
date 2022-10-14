/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

const (
	tagName = "Name"
)

type ec2API interface {
	DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

type imdsAPI interface {
	GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error)
	GetMetadata(context.Context, *imds.GetMetadataInput, ...func(*imds.Options)) (*imds.GetMetadataOutput, error)
}

// Metadata implements core.ProviderMetadata interface for AWS.
type Metadata struct {
	ec2  ec2API
	imds imdsAPI
}

// New initializes a new AWS Metadata client using instance default credentials.
// Default region is set up using the AWS imds api.
func New(ctx context.Context) (*Metadata, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithEC2IMDSRegion())
	if err != nil {
		return nil, err
	}
	return &Metadata{
		ec2:  ec2.NewFromConfig(cfg),
		imds: imds.New(imds.Options{}),
	}, nil
}

// Supported is used to determine if metadata API is implemented for this cloud provider.
func (m *Metadata) Supported() bool {
	return true
}

// List retrieves all instances belonging to the current Constellation.
func (m *Metadata) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	uid, err := readInstanceTag(ctx, m.imds, cloud.TagUID)
	if err != nil {
		return nil, fmt.Errorf("retrieving uid tag: %w", err)
	}
	ec2Instances, err := m.getAllInstancesInGroup(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances: %w", err)
	}

	return m.convertToMetadataInstance(ec2Instances)
}

// Self retrieves the current instance.
func (m *Metadata) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	identity, err := m.imds.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving instance identity: %w", err)
	}

	name, err := readInstanceTag(ctx, m.imds, tagName)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving name tag: %w", err)
	}
	instanceRole, err := readInstanceTag(ctx, m.imds, cloud.TagRole)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving role tag: %w", err)
	}

	return metadata.InstanceMetadata{
		Name:       name,
		ProviderID: fmt.Sprintf("aws:///%s/%s", identity.AvailabilityZone, identity.InstanceID),
		Role:       role.FromString(instanceRole),
		VPCIP:      identity.PrivateIP,
	}, nil
}

func (m *Metadata) getAllInstancesInGroup(ctx context.Context, uid string) ([]types.Instance, error) {
	var instances []types.Instance
	instanceReq := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:" + cloud.TagUID),
				Values: []string{uid},
			},
		},
	}

	for out, err := m.ec2.DescribeInstances(ctx, instanceReq); ; out, err = m.ec2.DescribeInstances(ctx, instanceReq) {
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

func (m *Metadata) convertToMetadataInstance(ec2Instances []types.Instance) ([]metadata.InstanceMetadata, error) {
	var instances []metadata.InstanceMetadata
	for _, ec2Instance := range ec2Instances {
		// ignore not running instances
		if ec2Instance.State == nil || ec2Instance.State.Name != types.InstanceStateNameRunning {
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
		}

		name, err := findTag(ec2Instance.Tags, tagName)
		if err != nil {
			return nil, fmt.Errorf("retrieving tag for instance %s: %w", *ec2Instance.InstanceId, err)
		}
		newInstance.Name = name

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

func readInstanceTag(ctx context.Context, api imdsAPI, tag string) (string, error) {
	reader, err := api.GetMetadata(ctx, &imds.GetMetadataInput{
		Path: "/tags/instance/" + tag,
	})
	if err != nil {
		return "", err
	}
	defer reader.Content.Close()
	instanceTag, err := io.ReadAll(reader.Content)
	return string(instanceTag), err
}

func findTag(tags []types.Tag, wantKey string) (string, error) {
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

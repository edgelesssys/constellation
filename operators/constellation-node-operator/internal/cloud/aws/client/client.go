/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Client is a client for the AWS Cloud.
type Client struct {
	ec2Client     ec2API
	scalingClient scalingAPI
}

// New creates a client with initialized clients.
func New(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	// get region from ec2metadata
	imdsClient := imds.NewFromConfig(cfg)
	regionOut, err := imdsClient.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get region from ec2metadata: %w", err)
	}
	return NewWithRegion(ctx, regionOut.Region)
}

// NewWithRegion creates a client with initialized clients and a given region.
func NewWithRegion(ctx context.Context, region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	cfg.Region = region

	ec2Client := ec2.NewFromConfig(cfg)
	scalingClient := autoscaling.NewFromConfig(cfg)
	return &Client{
		ec2Client:     ec2Client,
		scalingClient: scalingClient,
	}, nil
}

func getInstanceNameFromProviderID(providerID string) (string, error) {
	// aws:///us-east-2a/i-06888991e7138ed4e
	providerIDParts := strings.Split(providerID, "/")
	if len(providerIDParts) != 5 {
		return "", fmt.Errorf("invalid providerID: %s", providerID)
	}
	return providerIDParts[4], nil
}

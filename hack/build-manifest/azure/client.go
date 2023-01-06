/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

// Client for Azure Gallery API.
type Client struct {
	log           *logger.Logger
	opts          Options
	versionClient *armcompute.GalleryImageVersionsClient
}

// NewClient creates a new Client.
func NewClient(log *logger.Logger, opts Options) *Client {
	log = log.Named("azure-client")

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("unable to create default credentials: %v", err)
	}

	versionClient, err := armcompute.NewGalleryImageVersionsClient(opts.SubscriptionID, cred, &arm.ClientOptions{})
	if err != nil {
		log.Fatalf("unable to create version client: %v", err)
	}

	return &Client{
		log:           log,
		opts:          opts,
		versionClient: versionClient,
	}
}

// FetchImages for the given client options.
func (c *Client) FetchImages(ctx context.Context) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	imageVersionPager := c.versionClient.NewListByGalleryImagePager(
		c.opts.ResourceGroupName,
		c.opts.GalleryName,
		c.opts.ImageDefinition,
		&armcompute.GalleryImageVersionsClientListByGalleryImageOptions{},
	)

	images := map[string]string{}

	for imageVersionPager.More() {
		imageVersionPage, err := imageVersionPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, imageVersion := range imageVersionPage.Value {
			imageName := "v" + *imageVersion.Name
			images[imageName] = *imageVersion.ID
		}
	}

	return images, nil
}

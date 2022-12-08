/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"google.golang.org/api/iterator"
)

// Client for GCP Image API.
type Client struct {
	client *compute.ImagesClient
	log    *logger.Logger
	opts   Options
}

// NewClient creates a new Client.
func NewClient(ctx context.Context, log *logger.Logger, opts Options) *Client {
	client, err := compute.NewImagesRESTClient(ctx)
	if err != nil {
		log.Fatalf("Unable to create GCP client: %v", err)
	}

	return &Client{
		client: client,
		log:    log,
		opts:   opts,
	}
}

// Close the GCP client.
func (c *Client) Close() error {
	return c.client.Close()
}

// FetchImages for the given client options.
func (c *Client) FetchImages(ctx context.Context) (map[string]string, error) {
	imgIterator := c.client.List(ctx, &computepb.ListImagesRequest{
		Project: c.opts.ProjectID,
	})

	images := map[string]string{}

	for {
		img, err := imgIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.log.Fatalf("unable to request image: %v", err)
		}
		if img == nil || *img.Family != c.opts.ImageFamily {
			continue
		}
		imgReference := strings.TrimPrefix(*img.SelfLink, "https://www.googleapis.com/compute/v1/")
		imgVersion, err := c.opts.Filter(imgReference)
		if err != nil {
			continue
		}
		images[imgVersion] = imgReference
	}

	return images, nil
}

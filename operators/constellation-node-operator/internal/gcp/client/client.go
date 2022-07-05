package client

import (
	"context"
	"math/rand"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"go.uber.org/multierr"
)

// Client is a client for the Google Compute Engine.
type Client struct {
	instanceAPI
	instanceTemplateAPI
	instanceGroupManagersAPI
	diskAPI
	// prng is a pseudo-random number generator seeded with time. Not used for security.
	prng
}

// New creates a new client for the Google Compute Engine.
func New(ctx context.Context) (*Client, error) {
	var closers []closer
	insAPI, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	closers = append(closers, insAPI)
	templAPI, err := compute.NewInstanceTemplatesRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, templAPI)
	groupAPI, err := compute.NewInstanceGroupManagersRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, groupAPI)
	diskAPI, err := compute.NewDisksRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}

	return &Client{
		instanceAPI:              insAPI,
		instanceTemplateAPI:      &instanceTemplateClient{templAPI},
		instanceGroupManagersAPI: &instanceGroupManagersClient{groupAPI},
		diskAPI:                  diskAPI,
		prng:                     rand.New(rand.NewSource(int64(time.Now().Nanosecond()))),
	}, nil
}

// Close closes the client's connection.
func (c *Client) Close() error {
	closers := []closer{
		c.instanceAPI,
		c.instanceTemplateAPI,
		c.instanceGroupManagersAPI,
		c.diskAPI,
	}
	return closeAll(closers)
}

type closer interface {
	Close() error
}

// closeAll closes all closers, even if an error occurs.
//
// Errors are collected and a composed error is returned.
func closeAll(closers []closer) error {
	// Since this function is intended to be deferred, it will always call all
	// close operations, even if a previous operation failed.
	var errs error
	for _, closer := range closers {
		errs = multierr.Append(errs, closer.Close())
	}
	return errs
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"math/rand"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/spf13/afero"
	computeREST "google.golang.org/api/compute/v1"
)

// Client is a client for the Google Compute Engine.
type Client struct {
	projectID string
	projectAPI
	instanceAPI
	instanceTemplateAPI
	instanceGroupManagersAPI
	diskAPI
	// prng is a pseudo-random number generator seeded with time. Not used for security.
	prng
}

// New creates a new client for the Google Compute Engine.
func New(ctx context.Context, configPath string) (*Client, error) {
	projectID, err := loadProjectID(afero.NewOsFs(), configPath)
	if err != nil {
		return nil, err
	}

	var closers []closer
	projectAPI, err := compute.NewProjectsRESTClient(ctx)
	if err != nil {
		return nil, err
	}

	closers = append(closers, projectAPI)
	insAPI, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, insAPI)

	// TODO(msanft): Go back to protobuf-based API when it supports setting
	// a confidential instance type.
	// See https://github.com/googleapis/google-cloud-go/issues/10873 for the current status.
	restClient, err := computeREST.NewService(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	templAPI := computeREST.NewInstanceTemplatesService(restClient)

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
		projectID:                projectID,
		projectAPI:               projectAPI,
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
		c.projectAPI,
		c.instanceAPI,
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
		errs = errors.Join(errs, closer.Close())
	}
	return errs
}

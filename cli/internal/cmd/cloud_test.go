/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/iamid"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

type stubCloudCreator struct {
	createCalled bool
	id           clusterid.File
	createErr    error
}

func (c *stubCloudCreator) Create(
	ctx context.Context,
	provider cloudprovider.Provider,
	config *config.Config,
	insType string,
	coordCount, nodeCount int,
) (clusterid.File, error) {
	c.createCalled = true
	c.id.CloudProvider = provider
	return c.id, c.createErr
}

type stubCloudTerminator struct {
	called       bool
	terminateErr error
}

func (c *stubCloudTerminator) Terminate(context.Context) error {
	c.called = true
	return c.terminateErr
}

func (c *stubCloudTerminator) Called() bool {
	return c.called
}

type stubIAMCreator struct {
	createCalled bool
	id           iamid.File
	createErr    error
}

func (c *stubIAMCreator) Create(
	ctx context.Context,
	provider cloudprovider.Provider,
	iamConfig *cloudcmd.IAMConfig,
) (iamid.File, error) {
	c.createCalled = true
	c.id.CloudProvider = provider
	return c.id, c.createErr
}

type stubIAMDestroyer struct {
	destroyCalled bool
	destroyErr    error
}

func (d *stubIAMDestroyer) DestroyIAMUser(ctx context.Context) error {
	d.destroyCalled = true
	return d.destroyErr
}

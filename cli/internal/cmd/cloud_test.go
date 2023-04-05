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
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
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
	_ context.Context,
	opts cloudcmd.CreateOptions,
) (clusterid.File, error) {
	c.createCalled = true
	c.id.CloudProvider = opts.Provider
	return c.id, c.createErr
}

type stubCloudTerminator struct {
	called       bool
	terminateErr error
}

func (c *stubCloudTerminator) Terminate(_ context.Context, _ terraform.LogLevel) error {
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
	_ context.Context,
	provider cloudprovider.Provider,
	_ *cloudcmd.IAMConfigOptions,
) (iamid.File, error) {
	c.createCalled = true
	c.id.CloudProvider = provider
	return c.id, c.createErr
}

type stubIAMDestroyer struct {
	destroyCalled       bool
	getTfstateKeyCalled bool
	gcpSaKey            gcpshared.ServiceAccountKey
	destroyErr          error
	getTfstateKeyErr    error
}

func (d *stubIAMDestroyer) DestroyIAMConfiguration(_ context.Context, _ terraform.LogLevel) error {
	d.destroyCalled = true
	return d.destroyErr
}

func (d *stubIAMDestroyer) GetTfstateServiceAccountKey(_ context.Context) (gcpshared.ServiceAccountKey, error) {
	d.getTfstateKeyCalled = true
	return d.gcpSaKey, d.getTfstateKeyErr
}

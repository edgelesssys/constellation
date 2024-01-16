/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
	)
}

type stubCloudCreator struct {
	state               state.Infrastructure
	planCalled          bool
	planDiff            bool
	planErr             error
	applyCalled         bool
	applyErr            error
	restoreErr          error
	workspaceIsEmpty    bool
	workspaceIsEmptyErr error
}

func (c *stubCloudCreator) Plan(_ context.Context, _ *config.Config) (bool, error) {
	c.planCalled = true
	return c.planDiff, c.planErr
}

func (c *stubCloudCreator) Apply(_ context.Context, _ cloudprovider.Provider, _ cloudcmd.RollbackBehavior) (state.Infrastructure, error) {
	c.applyCalled = true
	return c.state, c.applyErr
}

func (c *stubCloudCreator) RestoreWorkspace() error {
	return c.restoreErr
}

func (c *stubCloudCreator) WorkingDirIsEmpty() (bool, error) {
	return c.workspaceIsEmpty, c.workspaceIsEmptyErr
}

type stubCloudTerminator struct {
	called       bool
	terminateErr error
}

func (c *stubCloudTerminator) Terminate(_ context.Context, _ string, _ terraform.LogLevel) error {
	c.called = true
	return c.terminateErr
}

func (c *stubCloudTerminator) Called() bool {
	return c.called
}

type stubIAMCreator struct {
	createCalled bool
	id           cloudcmd.IAMOutput
	createErr    error
}

func (c *stubIAMCreator) Create(
	_ context.Context,
	provider cloudprovider.Provider,
	_ *cloudcmd.IAMConfigOptions,
) (cloudcmd.IAMOutput, error) {
	c.createCalled = true
	c.id.CloudProvider = provider
	return c.id, c.createErr
}

type stubIAMDestroyer struct {
	destroyCalled       bool
	getTfStateKeyCalled bool
	gcpSaKey            gcpshared.ServiceAccountKey
	destroyErr          error
	getTfStateKeyErr    error
}

func (d *stubIAMDestroyer) DestroyIAMConfiguration(_ context.Context, _ string, _ terraform.LogLevel) error {
	d.destroyCalled = true
	return d.destroyErr
}

func (d *stubIAMDestroyer) GetTfStateServiceAccountKey(_ context.Context, _ string) (gcpshared.ServiceAccountKey, error) {
	d.getTfStateKeyCalled = true
	return d.gcpSaKey, d.getTfStateKeyErr
}

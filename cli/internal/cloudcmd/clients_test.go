/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

type stubTerraformClient struct {
	ip                     string
	cleanUpWorkspaceCalled bool
	removeInstallerCalled  bool
	destroyClusterCalled   bool
	createClusterErr       error
	destroyClusterErr      error
	cleanUpWorkspaceErr    error
}

func (c *stubTerraformClient) CreateCluster(ctx context.Context, name string, input terraform.Variables) (string, error) {
	return c.ip, c.createClusterErr
}

func (c *stubTerraformClient) DestroyCluster(ctx context.Context) error {
	c.destroyClusterCalled = true
	return c.destroyClusterErr
}

func (c *stubTerraformClient) CleanUpWorkspace() error {
	c.cleanUpWorkspaceCalled = true
	return c.cleanUpWorkspaceErr
}

func (c *stubTerraformClient) RemoveInstaller() {
	c.removeInstallerCalled = true
}

type stubLibvirtRunner struct {
	startCalled bool
	stopCalled  bool
	startErr    error
	stopErr     error
}

func (r *stubLibvirtRunner) Start(_ context.Context, _, _ string) error {
	r.startCalled = true
	return r.startErr
}

func (r *stubLibvirtRunner) Stop(context.Context) error {
	r.stopCalled = true
	return r.stopErr
}

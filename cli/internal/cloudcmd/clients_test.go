/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"io"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
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

type stubTerraformClient struct {
	ip                     string
	initSecret             string
	iamOutput              terraform.IAMOutput
	uid                    string
	cleanUpWorkspaceCalled bool
	removeInstallerCalled  bool
	destroyCalled          bool
	createClusterErr       error
	destroyErr             error
	prepareWorkspaceErr    error
	cleanUpWorkspaceErr    error
	iamOutputErr           error
}

func (c *stubTerraformClient) CreateCluster(ctx context.Context) (terraform.CreateOutput, error) {
	return terraform.CreateOutput{
		IP:     c.ip,
		Secret: c.initSecret,
		UID:    c.uid,
	}, c.createClusterErr
}

func (c *stubTerraformClient) CreateIAMConfig(ctx context.Context, provider cloudprovider.Provider) (terraform.IAMOutput, error) {
	return c.iamOutput, c.iamOutputErr
}

func (c *stubTerraformClient) PrepareWorkspace(path string, input terraform.Variables) error {
	return c.prepareWorkspaceErr
}

func (c *stubTerraformClient) Destroy(ctx context.Context) error {
	c.destroyCalled = true
	return c.destroyErr
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

type stubImageFetcher struct {
	reference         string
	fetchReferenceErr error
}

func (f *stubImageFetcher) FetchReference(_ context.Context, _ *config.Config) (string, error) {
	return f.reference, f.fetchReferenceErr
}

type stubRawDownloader struct {
	destination string
	downloadErr error
}

func (d *stubRawDownloader) Download(_ context.Context, _ io.Writer, _ bool, _ string, _ string) (string, error) {
	return d.destination, d.downloadErr
}

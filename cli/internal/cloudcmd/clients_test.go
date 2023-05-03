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
	tfjson "github.com/hashicorp/terraform-json"

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
	attestationURL         string
	tfjsonState            *tfjson.State
	cleanUpWorkspaceCalled bool
	removeInstallerCalled  bool
	destroyCalled          bool
	showCalled             bool
	createClusterErr       error
	destroyErr             error
	prepareWorkspaceErr    error
	cleanUpWorkspaceErr    error
	iamOutputErr           error
	showErr                error
}

func (c *stubTerraformClient) CreateCluster(_ context.Context, _ terraform.LogLevel, _ ...string) (terraform.CreateOutput, error) {
	return terraform.CreateOutput{
		IP:             c.ip,
		Secret:         c.initSecret,
		UID:            c.uid,
		AttestationURL: c.attestationURL,
	}, c.createClusterErr
}

func (c *stubTerraformClient) CreateIAMConfig(_ context.Context, _ cloudprovider.Provider, _ terraform.LogLevel) (terraform.IAMOutput, error) {
	return c.iamOutput, c.iamOutputErr
}

func (c *stubTerraformClient) PrepareWorkspace(_ string, _ terraform.Variables) error {
	return c.prepareWorkspaceErr
}

func (c *stubTerraformClient) Destroy(_ context.Context, _ terraform.LogLevel) error {
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

func (c *stubTerraformClient) Show(_ context.Context) (*tfjson.State, error) {
	c.showCalled = true
	return c.tfjsonState, c.showErr
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

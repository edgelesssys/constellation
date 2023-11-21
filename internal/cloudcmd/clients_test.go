/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"io"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/terraform"

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
	infraState             state.Infrastructure
	cleanUpWorkspaceCalled bool
	removeInstallerCalled  bool
	destroyCalled          bool
	showCalled             bool
	applyClusterErr        error
	destroyErr             error
	prepareWorkspaceErr    error
	cleanUpWorkspaceErr    error
	iamOutputErr           error
	showInfrastructureErr  error
	showIAMErr             error
	planDiff               bool
	planErr                error
	showPlanErr            error
}

func (c *stubTerraformClient) ApplyCluster(_ context.Context, _ cloudprovider.Provider, _ terraform.LogLevel) (state.Infrastructure, error) {
	return state.Infrastructure{
		ClusterEndpoint: c.ip,
		InitSecret:      []byte(c.initSecret),
		UID:             c.uid,
		Azure: &state.Azure{
			AttestationURL: c.attestationURL,
		},
	}, c.applyClusterErr
}

func (c *stubTerraformClient) ApplyIAM(_ context.Context, _ cloudprovider.Provider, _ terraform.LogLevel) (terraform.IAMOutput, error) {
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

func (c *stubTerraformClient) ShowInfrastructure(_ context.Context, _ cloudprovider.Provider) (state.Infrastructure, error) {
	c.showCalled = true
	return c.infraState, c.showInfrastructureErr
}

func (c *stubTerraformClient) ShowIAM(_ context.Context, _ cloudprovider.Provider) (terraform.IAMOutput, error) {
	c.showCalled = true
	return c.iamOutput, c.showIAMErr
}

func (c *stubTerraformClient) Plan(_ context.Context, _ terraform.LogLevel) (bool, error) {
	return c.planDiff, c.planErr
}

func (c *stubTerraformClient) ShowPlan(_ context.Context, _ terraform.LogLevel, _ io.Writer) error {
	return c.showPlanErr
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

func (f *stubImageFetcher) FetchReference(_ context.Context,
	_ cloudprovider.Provider, _ variant.Variant,
	_, _ string,
) (string, error) {
	return f.reference, f.fetchReferenceErr
}

type stubRawDownloader struct {
	destination string
	downloadErr error
}

func (d *stubRawDownloader) Download(_ context.Context, _ io.Writer, _ bool, _ string, _ string) (string, error) {
	return d.destination, d.downloadErr
}

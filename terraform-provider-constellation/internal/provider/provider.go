/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// The provider package implements the Constellation Terraform provider's
// "provider" resource, which is the main entrypoint for Terraform to
// interact with the provider.
package provider

import (
	"context"

	datastruct "github.com/edgelesssys/constellation/v2/terraform-provider-constellation/internal/data"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Perform interface cast to ensure ConstellationProvider satisfies various provider interfaces.
var _ provider.Provider = &ConstellationProvider{}

// ConstellationProviderModel is the provider data model.
type ConstellationProviderModel struct{}

// ConstellationProvider is the provider implementation.
type ConstellationProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// New creates a new provider, based on a version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ConstellationProvider{
			version: version,
		}
	}
}

// Metadata returns the Providers name and version upon request.
func (p *ConstellationProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "constellation"
	resp.Version = p.version
}

// Schema defines the HCL schema of the provider, i.e. what attributes it has and what they are used for.
func (p *ConstellationProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "The Constellation provider manages Constellation clusters.",
		MarkdownDescription: `The Constellation provider manages Constellation clusters.`, // TODO(msanft): Provide a more sophisticated description.
	}
}

// Configure is called when the provider block is initialized, and conventionally
// used to setup any API clients or other resources required for the provider.
func (p *ConstellationProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Populate the provider configuration model with what the user supplied when
	// declaring the provider block. No-op for now, as no attributes are defined.
	var data ConstellationProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO(msanft): Initialize persistent clients here.
	config := datastruct.ProviderData{}

	// Make the clients available during data source and resource "Configure" methods.
	resp.DataSourceData = config
	resp.ResourceData = config
}

// Resources lists the resources implemented by the provider.
func (p *ConstellationProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
	}
}

// DataSources lists the data sources implemented by the provider.
func (p *ConstellationProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewImageDataSource, NewAttestationDataSource,
	}
}

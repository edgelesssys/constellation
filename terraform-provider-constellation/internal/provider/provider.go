/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Perform interface cast to ensure ConstellationProvider satisfies various provider interfaces.
var _ provider.Provider = &ConstellationProvider{}

// ConstellationProvider is the provider implementation.
type ConstellationProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ConstellationProviderModel is the provider data model.
type ConstellationProviderModel struct {
	ExampleValue types.String `tfsdk:"example_value"`
}

// Metadata returns the Providers name and version upon request.
func (p *ConstellationProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "constellation"
	resp.Version = p.version
}

// Schema defines the HCL schema of the provider, i.e. what attributes it has and what they are used for.
func (p *ConstellationProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"example_value": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
		},
	}
}

// Configure is called when the provider block is initialized, and conventionally
// used to setup any API clients or other resources required for the provider.
func (p *ConstellationProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Populate the provider configuration model with what the user supplied when
	// declaring the provider block.
	var data ConstellationProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

// Resources lists the resources implemented by the provider.
func (p *ConstellationProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

// DataSources lists the data sources implemented by the provider.
func (p *ConstellationProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

// New creates a new provider, based on a version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ConstellationProvider{
			version: version,
		}
	}
}

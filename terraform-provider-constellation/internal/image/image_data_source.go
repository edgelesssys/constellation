/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package image

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/terraform-provider-constellation/internal/data"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ImageDataSource{}

// NewImageDataSource creates a new examplary data source.
func NewImageDataSource() datasource.DataSource {
	return &ImageDataSource{}
}

// ImageDataSource defines the data source implementation.
type ImageDataSource struct {
	imageFetcher data.ImageFetcher
}

// ImageDataSourceModel defines the image data source's data model.
type ImageDataSourceModel struct {
	AttestationVariant types.String `tfsdk:"attestation_variant"`
	ImageVersion       types.String `tfsdk:"image_version"`
	CSP                types.String `tfsdk:"csp"`
	Region             types.String `tfsdk:"region"`
	Reference          types.String `tfsdk:"reference"`
}

// Metadata returns the metadata for the image data source.
func (d *ImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

// Schema returns the schema for the image data source.
func (d *ImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Data source to retrieve the Constellation OS image reference for a given CSP and Attestation Variant.",
		MarkdownDescription: "Data source to retrieve the Constellation OS image reference for a given CSP and Attestation Variant.",
		Attributes: map[string]schema.Attribute{
			"attestation_variant": schema.StringAttribute{
				Description: "Attestation variant the image should work with. (e.g. `azure-sev-snp`)",
				MarkdownDescription: "Attestation variant the image should work with. Can be one of:\n" +
					"  * `aws-sev-snp`\n" +
					"  * `aws-nitro-tpm`\n" +
					"  * `azure-sev-snp`\n" +
					"  * `gcp-sev-es`\n",
				Required: true,
				// TODO(msanft): Add validators.
			},
			"image_version": schema.StringAttribute{
				Description:         "Version of the Constellation OS image to use. (e.g. `v2.13.0`)",
				MarkdownDescription: "Version of the Constellation OS image to use. (e.g. `v2.13.0`)",
				Required:            true, // TODO(msanft): Make this optional to support "lockstep" mode.
				// TODO(msanft): Add validators.
			},
			"csp": schema.StringAttribute{
				Description: "CSP (Cloud Service Provider) to use. (e.g. `azure`)",
				MarkdownDescription: "CSP (Cloud Service Provider) to use. (e.g. `azure`)\n" +
					"See the [full list of CSPs](https://docs.edgeless.systems/constellation/overview/clouds) that Constellation supports.",
				Required: true,
				// TODO(msanft): Add validators.
			},
			"region": schema.StringAttribute{
				Description: "Region to retrieve the image for. Only required for AWS.",
				MarkdownDescription: "Region to retrieve the image for. Only required for AWS.\n" +
					"The Constellation OS image must be [replicated to the region](https://docs.edgeless.systems/constellation/workflows/config)," +
					"and the region must [support AMD SEV-SNP](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snp-requirements.html), if it is used for Attestation.",
				Required: true,
				// TODO(msanft): Add validators.
			},
			"reference": schema.StringAttribute{
				Description:         "CSP-specific reference to the image.",
				MarkdownDescription: "CSP-specific reference to the image.",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source.
func (d *ImageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// check if the right type is passed down.
	providerData, ok := req.ProviderData.(data.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *data.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.imageFetcher = providerData.ImageFetcher
}

// Read reads from the data source.
func (d *ImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve the configuration values for this data source instance.
	var data ImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Check configuration for errors.
	csp := cloudprovider.FromString(data.CSP.ValueString())
	if csp == cloudprovider.Unknown {
		resp.Diagnostics.AddError(
			"Invalid CSP",
			fmt.Sprintf("Invalid CSP: %s", data.CSP.ValueString()),
		)
		return
	}

	attestationVariant, err := variant.FromString(data.AttestationVariant.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Attestation Variant",
			fmt.Sprintf("When parsing the Attestation Variant (%s), an error occured: %s", data.AttestationVariant.ValueString(), err),
		)
		return
	}

	// Retrieve Image Reference
	imageRef, err := d.imageFetcher.FetchReference(ctx, csp, attestationVariant, data.ImageVersion.ValueString(), data.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching Image Reference",
			fmt.Sprintf("When fetching the image reference, an error occured: %s", err),
		)
		return
	}

	// Save data into Terraform state
	data.Reference = types.StringValue(imageRef)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ImageDataSource{}

// NewImageDataSource creates a new data source for fetching Constellation OS images
// from the Versions-API.
func NewImageDataSource() datasource.DataSource {
	return &ImageDataSource{}
}

// ImageDataSource defines the data source implementation for the image data source.
// It is used to retrieve the Constellation OS image reference for a given CSP and Attestation Variant.
type ImageDataSource struct {
	imageFetcher imageFetcher
}

// imageFetcher gets an image reference from the versionsapi.
type imageFetcher interface {
	FetchReference(ctx context.Context,
		provider cloudprovider.Provider, attestationVariant variant.Variant,
		image, region string,
	) (string, error)
}

// ImageDataSourceModel defines the image data source's data model.
type ImageDataSourceModel struct {
	ID                 types.String `tfsdk:"id"` // Required for testing.
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
			"id": schema.StringAttribute{
				Computed: true,
			},
			"attestation_variant": schema.StringAttribute{
				Description: "Attestation variant the image should work with. (e.g. `azure-sev-snp`)",
				MarkdownDescription: "Attestation variant the image should work with. Can be one of:\n" +
					"  * `aws-sev-snp`\n" +
					"  * `aws-nitro-tpm`\n" +
					"  * `azure-sev-snp`\n" +
					"  * `gcp-sev-es`\n",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("aws-sev-snp", "aws-nitro-tpm", "azure-sev-snp", "gcp-sev-es"),
				},
			},
			"image_version": schema.StringAttribute{
				Description:         "Version of the Constellation OS image to use. (e.g. `v2.13.0`)",
				MarkdownDescription: "Version of the Constellation OS image to use. (e.g. `v2.13.0`)",
				Required:            true, // TODO(msanft): Make this optional to support "lockstep" mode.
			},
			"csp": schema.StringAttribute{
				Description: "CSP (Cloud Service Provider) to use. (e.g. `azure`)",
				MarkdownDescription: "CSP (Cloud Service Provider) to use. (e.g. `azure`)\n" +
					"See the [full list of CSPs](https://docs.edgeless.systems/constellation/overview/clouds) that Constellation supports.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("aws", "azure", "gcp"),
				},
			},
			"region": schema.StringAttribute{
				Description: "Region to retrieve the image for. Only required for AWS.",
				MarkdownDescription: "Region to retrieve the image for. Only required for AWS.\n" +
					"The Constellation OS image must be [replicated to the region](https://docs.edgeless.systems/constellation/workflows/config)," +
					"and the region must [support AMD SEV-SNP](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snp-requirements.html), if it is used for Attestation.",
				Optional: true,
			},
			"reference": schema.StringAttribute{
				Description:         "CSP-specific reference to the image.",
				MarkdownDescription: "CSP-specific reference to the image.",
				Computed:            true,
			},
		},
	}
}

// TODO(msanft): Possibly implement more complex validation for inter-dependencies between attributes.
// E.g., region should be required if, and only if, AWS is used.

// Configure configures the data source.
func (d *ImageDataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	// Create the image-fetcher client.
	d.imageFetcher = imagefetcher.New()
}

// Read reads from the data source.
func (d *ImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve the configuration values for this data source instance.
	var data ImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Check configuration for errors.
	csp := cloudprovider.FromString(data.CSP.ValueString())
	if csp == cloudprovider.Unknown {
		resp.Diagnostics.AddAttributeError(
			path.Root("csp"),
			"Invalid CSP",
			fmt.Sprintf("Invalid CSP: %s", data.CSP.ValueString()),
		)
	}

	attestationVariant, err := variant.FromString(data.AttestationVariant.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("attestation_variant"),
			"Invalid Attestation Variant",
			fmt.Sprintf("When parsing the Attestation Variant (%s), an error occurred: %s", data.AttestationVariant.ValueString(), err),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve Image Reference
	imageRef, err := d.imageFetcher.FetchReference(ctx, csp, attestationVariant, data.ImageVersion.ValueString(), data.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching Image Reference",
			fmt.Sprintf("When fetching the image reference, an error occurred: %s", err),
		)
		return
	}

	// Save data into Terraform state
	data.Reference = types.StringValue(imageRef)
	// Use a placeholder ID for testing, as per https://developer.hashicorp.com/terraform/plugin/framework/acctests#no-id-found-in-attributes
	data.ID = types.StringValue("placeholder")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

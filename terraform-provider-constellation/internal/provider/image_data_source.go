/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/terraform-provider-constellation/internal/data"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	// Ensure provider defined types fully satisfy framework interfaces.
	_                                       datasource.DataSource = &ImageDataSource{}
	caseInsensitiveCommunityGalleriesRegexp                       = regexp.MustCompile(`(?i)\/communitygalleries\/`)
	caseInsensitiveImagesRegExp                                   = regexp.MustCompile(`(?i)\/images\/`)
	caseInsensitiveVersionsRegExp                                 = regexp.MustCompile(`(?i)\/versions\/`)
)

// NewImageDataSource creates a new data source for fetching Constellation OS images
// from the Versions-API.
func NewImageDataSource() datasource.DataSource {
	return &ImageDataSource{}
}

// ImageDataSource defines the data source implementation for the image data source.
// It is used to retrieve the Constellation OS image reference for a given CSP and Attestation Variant.
type ImageDataSource struct {
	imageFetcher imageFetcher
	version      string
}

// imageFetcher gets an image reference from the versionsapi.
type imageFetcher interface {
	FetchReference(ctx context.Context,
		provider cloudprovider.Provider, attestationVariant variant.Variant,
		image, region string, useMarketplaceImage bool,
	) (string, error)
}

// ImageDataSourceModel defines the image data source's data model.
type ImageDataSourceModel struct {
	AttestationVariant types.String `tfsdk:"attestation_variant"`
	ImageVersion       types.String `tfsdk:"image_version"`
	CSP                types.String `tfsdk:"csp"`
	MarketplaceImage   types.Bool   `tfsdk:"marketplace_image"`
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
		Description:         "The data source to resolve the CSP-specific OS image reference for a given version and attestation variant.",
		MarkdownDescription: "Data source to resolve the CSP-specific OS image reference for a given version and attestation variant. The `reference` output of this data source is needed as `image` input for the `constellation_cluster` resource.",
		Attributes: map[string]schema.Attribute{
			"attestation_variant": newAttestationVariantAttribute(attributeInput),
			"image_version": schema.StringAttribute{
				Description:         "Version of the Constellation OS image to use. (e.g. `v2.13.0`)",
				MarkdownDescription: "Version of the Constellation OS image to use. (e.g. `v2.13.0`)",
				Optional:            true,
			},
			"csp": newCSPAttribute(),
			"marketplace_image": schema.BoolAttribute{
				Description:         "Whether a marketplace image should be used. Currently only supported for Azure.",
				MarkdownDescription: "Whether a marketplace image should be used. Currently only supported for Azure.",
				Optional:            true,
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

// ValidateConfig validates the configuration for the image data source.
func (d *ImageDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data ImageDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.CSP.Equal(types.StringValue("aws")) && data.Region.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Region must be set for AWS", "When CSP is set to AWS, 'region' must be specified.",
		)
		return
	}
}

// Configure configures the data source.
func (d *ImageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured. is necessary!
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(data.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *data.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.version = providerData.Version
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

	imageVersion := data.ImageVersion.ValueString()
	if imageVersion == "" {
		imageVersion = d.version // Use provider version as default.
	}

	// Retrieve Image Reference
	imageRef, err := d.imageFetcher.FetchReference(ctx, csp, attestationVariant,
		imageVersion, data.Region.ValueString(), data.MarketplaceImage.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching Image Reference",
			fmt.Sprintf("When fetching the image reference, an error occurred: %s", err),
		)
		return
	}

	// Do adjustments for Azure casing
	if csp == cloudprovider.Azure {
		imageRef = caseInsensitiveCommunityGalleriesRegexp.ReplaceAllString(imageRef, "/communityGalleries/")
		imageRef = caseInsensitiveImagesRegExp.ReplaceAllString(imageRef, "/images/")
		imageRef = caseInsensitiveVersionsRegExp.ReplaceAllString(imageRef, "/versions/")
	}

	// Save data into Terraform state
	data.Reference = types.StringValue(imageRef)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

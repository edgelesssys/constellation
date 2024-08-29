/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/terraform-provider-constellation/internal/data"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	// Ensure provider defined types fully satisfy framework interfaces.
	_ datasource.DataSource                   = &AttestationDataSource{}
	_ datasource.DataSourceWithValidateConfig = &AttestationDataSource{}
	_ datasource.DataSourceWithConfigure      = &AttestationDataSource{}
)

// NewAttestationDataSource creates a new attestation data source.
func NewAttestationDataSource() datasource.DataSource {
	return &AttestationDataSource{}
}

// AttestationDataSource defines the data source implementation.
type AttestationDataSource struct {
	client  *http.Client
	fetcher attestationconfigapi.Fetcher
	rekor   *sigstore.Rekor
	version string
}

// AttestationDataSourceModel describes the data source data model.
type AttestationDataSourceModel struct {
	CSP                types.String `tfsdk:"csp"`
	AttestationVariant types.String `tfsdk:"attestation_variant"`
	Image              types.Object `tfsdk:"image"`
	MaaURL             types.String `tfsdk:"maa_url"`
	Insecure           types.Bool   `tfsdk:"insecure"`
	Attestation        types.Object `tfsdk:"attestation"`
}

// Configure configures the data source.
func (d *AttestationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured. is necessary!
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(data.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected data.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.version = providerData.Version.String()

	d.client = http.DefaultClient
	d.fetcher = attestationconfigapi.NewFetcher()
	rekor, err := sigstore.NewRekor()
	if err != nil {
		resp.Diagnostics.AddError("constructing rekor client", err.Error())
		return
	}
	d.rekor = rekor
}

// Metadata returns the metadata for the data source.
func (d *AttestationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attestation"
}

// Schema returns the schema for the data source.
func (d *AttestationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Data source to fetch an attestation configuration for a given cloud service provider, attestation variant, and OS image.",
		MarkdownDescription: "Data source to fetch an attestation configuration for a given cloud service provider, attestation variant, and OS image.",

		Attributes: map[string]schema.Attribute{
			"csp":                 newCSPAttributeSchema(),
			"attestation_variant": newAttestationVariantAttributeSchema(attributeInput),
			"image":               newImageAttributeSchema(attributeInput),
			"maa_url": schema.StringAttribute{
				MarkdownDescription: `For Azure only, the URL of the Microsoft Azure Attestation service. The MAA's policy needs to be patched manually to work with Constellation OS images.
See the [Constellation documentation](https://docs.edgeless.systems/constellation/workflows/terraform-provider#quick-setup) for more information.`,
				Optional: true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "DON'T USE IN PRODUCTION Skip the signature verification when fetching measurements for the image.",
				Optional:            true,
			},
			"attestation": newAttestationConfigAttributeSchema(attributeOutput),
		},
	}
}

// ValidateConfig validates the configuration for the image data source.
func (d *AttestationDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data AttestationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.AttestationVariant.Equal(types.StringValue("azure-sev-snp")) && !data.MaaURL.IsNull() {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("maa_url"),
			"MAA URL should only be set for Azure SEV-SNP", "Only when attestation_variant is set to 'azure-sev-snp', 'maa_url' should be specified.",
		)
		return
	}

	if !data.MaaURL.IsNull() {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("maa_url"),
			"Ensure that the MAA's policy is patched", "When MAA is used, please ensure the MAA's policy is patche properly for use within Constellation. See https://docs.edgeless.systems/constellation/workflows/terraform-provider#quick-setup for more information.",
		)
		return
	}

	if data.AttestationVariant.Equal(types.StringValue("azure-sev-snp")) && data.MaaURL.IsNull() {
		tflog.Info(ctx, "MAA URL not set, MAA fallback will be unavailable")
	}
}

// Read reads from the data source.
func (d *AttestationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AttestationDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	csp := cloudprovider.FromString(data.CSP.ValueString())
	if csp == cloudprovider.Unknown {
		resp.Diagnostics.AddAttributeError(
			path.Root("csp"),
			"Invalid CSP",
			fmt.Sprintf("Invalid CSP: %s", data.CSP.ValueString()),
		)
		return
	}
	attestationVariant, err := variant.FromString(data.AttestationVariant.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("attestation_variant"),
			"Invalid Attestation Variant",
			fmt.Sprintf("Invalid attestation variant: %s", data.AttestationVariant.ValueString()),
		)
		return
	}

	insecureFetch := data.Insecure.ValueBool()

	latestVersions := attestationconfigapi.Entry{}
	if attestationVariant.Equal(variant.AWSSEVSNP{}) ||
		attestationVariant.Equal(variant.AzureSEVSNP{}) ||
		attestationVariant.Equal(variant.AzureTDX{}) ||
		attestationVariant.Equal(variant.GCPSEVSNP{}) {
		latestVersions, err = d.fetcher.FetchLatestVersion(ctx, attestationVariant)
		if err != nil {
			resp.Diagnostics.AddError("Fetching SNP Version numbers", err.Error())
			return
		}
	}
	tfAttestation, err := convertToTfAttestation(attestationVariant, latestVersions)
	if err != nil {
		resp.Diagnostics.AddError("Converting attestation", err.Error())
	}
	verifyFetcher := measurements.NewVerifyFetcher(sigstore.NewCosignVerifier, d.rekor, d.client)

	// parse OS image version
	var image imageAttribute
	convertDiags := data.Image.As(ctx, &image, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(convertDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fetchedMeasurements, err := verifyFetcher.FetchAndVerifyMeasurements(ctx, image.ShortPath,
		csp, attestationVariant, insecureFetch)
	if err != nil {
		var rekErr *measurements.RekorError
		if errors.As(err, &rekErr) {
			resp.Diagnostics.AddWarning("Ignoring Rekor related error", err.Error())
		} else {
			resp.Diagnostics.AddError("fetching and verifying measurements", err.Error())
			return
		}
	}
	tfAttestation.Measurements = convertToTfMeasurements(fetchedMeasurements)

	diags := resp.State.SetAttribute(ctx, path.Root("attestation"), tfAttestation)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "read constellation attestation data source")
}

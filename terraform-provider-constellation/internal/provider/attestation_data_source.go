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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AttestationDataSource{}

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
	ImageVersion       types.String `tfsdk:"image_version"`
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
	d.version = providerData.Version

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
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The data source to fetch measurements from a configured cloud provider and image.",
		Attributes: map[string]schema.Attribute{
			"csp":                 newCSPAttribute(),
			"attestation_variant": newAttestationVariantAttribute(attributeInput),
			"image_version": schema.StringAttribute{
				MarkdownDescription: "The image version to use. If not set, the provider version value is used.",
				Optional:            true,
			},
			"maa_url": schema.StringAttribute{
				MarkdownDescription: "For Azure only, the URL of the Microsoft Azure Attestation service",
				Optional:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "DON'T USE IN PRODUCTION Skip the signature verification when fetching measurements for the image.",
				Optional:            true,
			},
			"attestation": newAttestationConfigAttribute(attributeOutput),
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
	if data.AttestationVariant.Equal(types.StringValue("azure-sev-snp")) && data.MaaURL.IsNull() {
		tflog.Info(ctx, "MAA URL not set, MAA fallback will be unavaiable")
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

	snpVersions := attestationconfigapi.SEVSNPVersionAPI{}
	if attestationVariant.Equal(variant.AzureSEVSNP{}) || attestationVariant.Equal(variant.AWSSEVSNP{}) {
		snpVersions, err = d.fetcher.FetchSEVSNPVersionLatest(ctx, attestationVariant)
		if err != nil {
			resp.Diagnostics.AddError("Fetching SNP Version numbers", err.Error())
			return
		}
	}
	tfAttestation, err := convertToTfAttestation(attestationVariant, snpVersions)
	if err != nil {
		resp.Diagnostics.AddError("Converting SNP attestation", err.Error())
	}
	verifyFetcher := measurements.NewVerifyFetcher(sigstore.NewCosignVerifier, d.rekor, d.client)

	imageVersion := data.ImageVersion.ValueString()
	if imageVersion == "" {
		tflog.Info(ctx, fmt.Sprintf("No image version specified, using provider version %s", d.version))
		imageVersion = d.version // Use provider version as default.
	}
	fetchedMeasurements, err := verifyFetcher.FetchAndVerifyMeasurements(ctx, imageVersion,
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

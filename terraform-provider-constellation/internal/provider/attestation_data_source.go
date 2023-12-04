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
}

// AttestationDataSourceModel describes the data source data model.
type AttestationDataSourceModel struct {
	CSP                types.String `tfsdk:"csp"`
	AttestationVariant types.String `tfsdk:"attestation_variant"`
	ImageVersion       types.String `tfsdk:"image_version"`
	MaaURL             types.String `tfsdk:"maa_url"`
	Measurements       types.Map    `tfsdk:"measurements"`
	Attestation        types.Object `tfsdk:"attestation"`
}

// Configure configures the data source.
func (d *AttestationDataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
			"attestation_variant": newAttesationVariantAttribute(),
			"image_version": schema.StringAttribute{
				MarkdownDescription: "The image version to use",
				Required:            true,
			},
			"maa_url": schema.StringAttribute{
				MarkdownDescription: "For Azure only, the URL of the Microsoft Azure Attestation service",
				Optional:            true,
			},
			"measurements": schema.MapNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"expected": schema.StringAttribute{
							Computed: true,
						},
						"warn_only": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
			"attestation": newAttestationConfigAttribute(false),
		},
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
	if attestationVariant.Equal(variant.AzureSEVSNP{}) || attestationVariant.Equal(variant.AWSSEVSNP{}) {
		snpVersions, err := d.fetcher.FetchSEVSNPVersionLatest(ctx, attestationVariant)
		if err != nil {
			resp.Diagnostics.AddError("Fetching SNP Version numbers", err.Error())
			return
		}
		tfSnpAttestation, err := convertSNPAttestationTfStateCompatible(attestationVariant, snpVersions)
		if err != nil {
			resp.Diagnostics.AddError("Converting SNP attestation", err.Error())
		}
		diags := resp.State.SetAttribute(ctx, path.Root("attestation"), tfSnpAttestation)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	verifyFetcher := measurements.NewVerifyFetcher(sigstore.NewCosignVerifier, d.rekor, d.client)
	fetchedMeasurements, err := verifyFetcher.FetchAndVerifyMeasurements(ctx, data.ImageVersion.ValueString(),
		csp, attestationVariant, false)
	if err != nil {
		var rekErr *measurements.RekorError
		if errors.As(err, &rekErr) {
			resp.Diagnostics.AddWarning("Ignoring Rekor related error", err.Error())
		} else {
			resp.Diagnostics.AddError("fetching and verifying measurements", err.Error())
			return
		}
	}
	tfMeasurements := convertMeasurementsTfStateCompatible(fetchedMeasurements)
	diags := resp.State.SetAttribute(ctx, path.Root("measurements"), tfMeasurements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "read constellation attestation data source")
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	ID                 types.String `tfsdk:"id"`
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
			"csp": schema.StringAttribute{
				Description: "CSP (Cloud Service Provider) to use. (e.g. `azure`)",
				MarkdownDescription: "CSP (Cloud Service Provider) to use. (e.g. `azure`)\n" +
					"See the [full list of CSPs](https://docs.edgeless.systems/constellation/overview/clouds) that Constellation supports.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("aws", "azure", "gcp"),
				},
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
				MarkdownDescription: "The image version to use",
				Required:            true,
			},
			"maa_url": schema.StringAttribute{
				MarkdownDescription: "For Azure only, the URL of the Microsoft Azure Attestation service",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the data source",
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
			"attestation": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Only relevant for SEV-SNP.",
				Description:         "The values provide sensible defaults. See the docs for advanced usage.", // TODO(elchead): AB#3568
				Attributes: map[string]schema.Attribute{
					"bootloader_version": schema.Int64Attribute{
						Computed: true,
					},
					"tee_version": schema.Int64Attribute{
						Computed: true,
					},
					"snp_version": schema.Int64Attribute{
						Computed: true,
					},
					"microcode_version": schema.Int64Attribute{
						Computed: true,
					},
					"azure_firmware_signer_config": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"accepted_key_digests": schema.ListAttribute{
								Computed:    true,
								ElementType: types.StringType,
							},
							"enforcement_policy": schema.StringAttribute{
								Computed: true,
							},
							"maa_url": schema.StringAttribute{
								Computed: true,
							},
						},
					},
					"amd_root_key": schema.StringAttribute{
						Computed: true,
					},
				},
			},
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
			fmt.Sprintf("Invalid attestation variant: %s", data.CSP.ValueString()),
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

func convertSNPAttestationTfStateCompatible(attestationVariant variant.Variant,
	snpVersions attestationconfigapi.SEVSNPVersionAPI,
) (tfSnpAttestation sevSnpAttestation, err error) {
	var cert config.Certificate
	switch attestationVariant.(type) {
	case variant.AWSSEVSNP:
		cert = config.DefaultForAWSSEVSNP().AMDRootKey
	case variant.AzureSEVSNP:
		cert = config.DefaultForAzureSEVSNP().AMDRootKey
	}
	certBytes, err := cert.MarshalJSON()
	if err != nil {
		return tfSnpAttestation, err
	}
	tfSnpAttestation = sevSnpAttestation{
		BootloaderVersion: snpVersions.Bootloader,
		TEEVersion:        snpVersions.TEE,
		SNPVersion:        snpVersions.SNP,
		MicrocodeVersion:  snpVersions.Microcode,
		AMDRootKey:        string(certBytes),
	}
	if attestationVariant.Equal(variant.AzureSEVSNP{}) {
		firmwareCfg := config.DefaultForAzureSEVSNP().FirmwareSignerConfig
		keyDigestAny, err := firmwareCfg.AcceptedKeyDigests.MarshalYAML()
		if err != nil {
			return tfSnpAttestation, err
		}
		keyDigest, ok := keyDigestAny.([]string)
		if !ok {
			return tfSnpAttestation, errors.New("reading Accepted Key Digests: could not convert to []string")
		}
		tfSnpAttestation.AzureSNPFirmwareSignerConfig = azureSnpFirmwareSignerConfig{
			AcceptedKeyDigests: keyDigest,
			EnforcementPolicy:  firmwareCfg.EnforcementPolicy.String(),
			MAAURL:             firmwareCfg.MAAURL,
		}
	}
	return tfSnpAttestation, nil
}

func convertMeasurementsTfStateCompatible(m measurements.M) map[string]measurement {
	tfMeasurements := map[string]measurement{}
	for key, value := range m {
		keyStr := strconv.FormatUint(uint64(key), 10)
		tfMeasurements[keyStr] = measurement{
			Expected: hex.EncodeToString(value.Expected[:]),
			WarnOnly: bool(value.ValidationOpt),
		}
	}
	return tfMeasurements
}

type measurement struct {
	Expected string `tfsdk:"expected"`
	WarnOnly bool   `tfsdk:"warn_only"`
}

type sevSnpAttestation struct {
	BootloaderVersion            uint8                        `tfsdk:"bootloader_version"`
	TEEVersion                   uint8                        `tfsdk:"tee_version"`
	SNPVersion                   uint8                        `tfsdk:"snp_version"`
	MicrocodeVersion             uint8                        `tfsdk:"microcode_version"`
	AMDRootKey                   string                       `tfsdk:"amd_root_key"`
	AzureSNPFirmwareSignerConfig azureSnpFirmwareSignerConfig `tfsdk:"azure_firmware_signer_config"`
}

type azureSnpFirmwareSignerConfig struct {
	AcceptedKeyDigests []string `tfsdk:"accepted_key_digests"`
	EnforcementPolicy  string   `tfsdk:"enforcement_policy"`
	MAAURL             string   `tfsdk:"maa_url"`
}

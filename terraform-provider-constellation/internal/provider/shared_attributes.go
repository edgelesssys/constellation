/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// attributeInput is the attribute type used for input variables.
	attributeInput attributeType = true
	// attributeOutput is the attribute type used for output variables.
	attributeOutput attributeType = false
)

type attributeType bool

func newAttestationVariantAttributeSchema(t attributeType) schema.Attribute {
	isInput := bool(t)
	return schema.StringAttribute{
		Description: "Attestation variant the image should work with. (e.g. `azure-sev-snp`)",
		MarkdownDescription: "Attestation variant the image should work with. Can be one of:\n" +
			"  * `aws-sev-snp`\n" +
			"  * `aws-nitro-tpm`\n" +
			"  * `azure-sev-snp`\n" +
			"  * `gcp-sev-es`\n",
		Required: isInput,
		Computed: !isInput,
		Validators: []validator.String{
			stringvalidator.OneOf("aws-sev-snp", "aws-nitro-tpm", "azure-sev-snp", "gcp-sev-es"),
		},
	}
}

func newCSPAttributeSchema() schema.Attribute {
	return schema.StringAttribute{
		Description: "CSP (Cloud Service Provider) to use. (e.g. `azure`)",
		MarkdownDescription: "CSP (Cloud Service Provider) to use. (e.g. `azure`)\n" +
			"See the [full list of CSPs](https://docs.edgeless.systems/constellation/overview/clouds) that Constellation supports.",
		Required: true,
		Validators: []validator.String{
			stringvalidator.OneOf("aws", "azure", "gcp"),
		},
	}
}

func newMeasurementsAttributeSchema(t attributeType) schema.Attribute {
	isInput := bool(t)
	return schema.MapNestedAttribute{
		Computed: !isInput,
		Required: isInput,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"expected": schema.StringAttribute{
					Required: isInput,
					Computed: !isInput,
				},
				"warn_only": schema.BoolAttribute{
					Required: isInput,
					Computed: !isInput,
				},
			},
		},
	}
}

// measurementAttribute is the measurement attribute's data model.
type measurementAttribute struct {
	Expected string `tfsdk:"expected"`
	WarnOnly bool   `tfsdk:"warn_only"`
}

func newAttestationConfigAttributeSchema(t attributeType) schema.Attribute {
	isInput := bool(t)
	var additionalDescription string
	if isInput {
		additionalDescription = " The output of the [constellation_attestation](../data-sources/attestation.md) data source provides sensible defaults.  "
	}
	return schema.SingleNestedAttribute{
		Computed:            !isInput,
		Required:            isInput,
		MarkdownDescription: "Attestation comprises the measurements and SEV-SNP specific parameters." + additionalDescription,
		Description:         "Attestation comprises the measurements and SEV-SNP specific parameters." + additionalDescription,
		Attributes: map[string]schema.Attribute{
			"variant": newAttestationVariantAttributeSchema(t), // duplicated for convenience in cluster resource
			"bootloader_version": schema.Int64Attribute{
				Computed: !isInput,
				Required: isInput,
			},
			"tee_version": schema.Int64Attribute{
				Computed: !isInput,
				Required: isInput,
			},
			"snp_version": schema.Int64Attribute{
				Computed: !isInput,
				Required: isInput,
			},
			"microcode_version": schema.Int64Attribute{
				Computed: !isInput,
				Required: isInput,
			},
			"azure_firmware_signer_config": schema.SingleNestedAttribute{
				Computed: !isInput,
				Optional: isInput,
				Attributes: map[string]schema.Attribute{
					"accepted_key_digests": schema.ListAttribute{
						Computed:    !isInput,
						Optional:    isInput,
						ElementType: types.StringType,
					},
					"enforcement_policy": schema.StringAttribute{
						Computed: !isInput,
						Optional: isInput,
					},
					"maa_url": schema.StringAttribute{
						Computed: !isInput,
						Optional: isInput,
					},
				},
			},
			"amd_root_key": schema.StringAttribute{
				Computed: !isInput,
				Required: isInput,
			},
			"measurements": newMeasurementsAttributeSchema(t),
		},
	}
}

// attestationAttribute is the attestation attribute's data model.
type attestationAttribute struct {
	BootloaderVersion            uint8                                 `tfsdk:"bootloader_version"`
	TEEVersion                   uint8                                 `tfsdk:"tee_version"`
	SNPVersion                   uint8                                 `tfsdk:"snp_version"`
	MicrocodeVersion             uint8                                 `tfsdk:"microcode_version"`
	AMDRootKey                   string                                `tfsdk:"amd_root_key"`
	AzureSNPFirmwareSignerConfig azureSnpFirmwareSignerConfigAttribute `tfsdk:"azure_firmware_signer_config"`
	Variant                      string                                `tfsdk:"variant"`
	Measurements                 map[string]measurementAttribute       `tfsdk:"measurements"`
}

// azureSnpFirmwareSignerConfigAttribute is the azure firmware signer config attribute's data model.
type azureSnpFirmwareSignerConfigAttribute struct {
	AcceptedKeyDigests []string `tfsdk:"accepted_key_digests"`
	EnforcementPolicy  string   `tfsdk:"enforcement_policy"`
	MAAURL             string   `tfsdk:"maa_url"`
}

func newImageAttributeSchema(t attributeType) schema.Attribute {
	isInput := bool(t)
	return schema.SingleNestedAttribute{
		Description:         "Constellation OS Image to use on the nodes.",
		MarkdownDescription: "Constellation OS Image to use on the nodes.",
		Computed:            !isInput,
		Required:            isInput,
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				Description:         "Semantic version of the image.",
				MarkdownDescription: "Semantic version of the image.",
				Computed:            !isInput,
				Required:            isInput,
			},
			"reference": schema.StringAttribute{
				Description:         "CSP-specific unique reference to the image. The format differs per CSP.",
				MarkdownDescription: "CSP-specific unique reference to the image. The format differs per CSP.",
				Computed:            !isInput,
				Required:            isInput,
			},
			"short_path": schema.StringAttribute{
				Description: "CSP-agnostic short path to the image. The format is `vX.Y.Z` for release images and `ref/$GIT_REF/stream/$STREAM/$SEMANTIC_VERSION` for pre-release images.",
				MarkdownDescription: "CSP-agnostic short path to the image. The format is `vX.Y.Z` for release images and `ref/$GIT_REF/stream/$STREAM/$SEMANTIC_VERSION` for pre-release images.\n" +
					"- `$GIT_REF` is the git reference (i.e. branch name) the image was built on, e.g. `main`.\n" +
					"- `$STREAM` is the stream the image was built on, e.g. `nightly`.\n" +
					"- `$SEMANTIC_VERSION` is the semantic version of the image, e.g. `vX.Y.Z` or `vX.Y.Z-pre...`.",
				Computed: !isInput,
				Required: isInput,
			},
		},
	}
}

// imageAttribute is the image attribute's data model.
type imageAttribute struct {
	Reference string `tfsdk:"reference"`
	Version   string `tfsdk:"version"`
	ShortPath string `tfsdk:"short_path"`
}

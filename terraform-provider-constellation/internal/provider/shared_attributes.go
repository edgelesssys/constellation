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

func newAttestationVariantAttribute(isInput bool) schema.Attribute {
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

func newCSPAttribute() schema.Attribute {
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

func newMeasurementsAttribute(isInput bool) schema.Attribute {
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

func newAttestationConfigAttribute(isInput bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		Computed:            !isInput,
		Required:            isInput,
		MarkdownDescription: "Only relevant for SEV-SNP.",
		Description:         "The values provide sensible defaults. See the docs for advanced usage.", // TODO(elchead): AB#3568
		Attributes: map[string]schema.Attribute{
			"variant": newAttestationVariantAttribute(isInput), // duplicated for convenience in cluster resource
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
		},
	}
}

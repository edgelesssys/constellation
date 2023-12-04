/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &ClusterResource{}
	_ resource.ResourceWithImportState = &ClusterResource{}
)

// NewClusterResource creates a new cluster resource.
func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// ClusterResource defines the resource implementation.
type ClusterResource struct{}

// ClusterResourceModel describes the resource data model.
type ClusterResourceModel struct {
	UID                   types.String `tfsdk:"uid"`
	Name                  types.String `tfsdk:"name"`
	Image                 types.String `tfsdk:"image"`
	KubernetesVersion     types.String `tfsdk:"kubernetes_version"`
	Debug                 types.Bool   `tfsdk:"debug"`
	InitEndpoint          types.String `tfsdk:"init_endpoint"`
	KubernetesAPIEndpoint types.String `tfsdk:"kubernetes_api_endpoint"`
	MicroserviceVersion   types.String `tfsdk:"constellation_microservices_version"`
	ExtraMicroservices    types.Object `tfsdk:"extra_microservices"`
	MasterSecret          types.String `tfsdk:"master_secret"`
	InitSecret            types.String `tfsdk:"init_secret"`
	Attestation           types.Object `tfsdk:"attestation"`
	Measurements          types.Map    `tfsdk:"measurements"`
	OwnerID               types.String `tfsdk:"owner_id"`
	ClusterID             types.String `tfsdk:"cluster_id"`
	Kubeconfig            types.String `tfsdk:"kubeconfig"`
	// NetworkConfig 			 types.Object `tfsdk:"network_config"` // TODO(elchead): do when clear what is needed
}

// Metadata returns the metadata of the resource.
func (r *ClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

// Schema returns the schema of the resource.
func (r *ClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for a Constellation cluster.",
		Description:         "Resource for a Constellation cluster.",

		Attributes: map[string]schema.Attribute{
			"uid": schema.StringAttribute{
				MarkdownDescription: "The UID of the cluster.",
				Description:         "The UID of the cluster.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name used in the cluster's named resources / cluster name.",
				Description:         "Name used in the cluster's named resources / cluster name.",
				Optional:            true, // TODO(elchead): use "constell" as default
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "The Constellation OS image to use in the CSP specific reference format. Use the `constellation_image` data source to find the correct image for your CSP.",
				Description:         "The Constellation OS image to use in the CSP specific reference format. Use the `constellation_image` data source to find the correct image for your CSP. When not set, the latest default version will be used.",
				Optional:            true,
			},
			"kubernetes_version": schema.StringAttribute{
				MarkdownDescription: "The Kubernetes version to use for the cluster.",
				Description:         "The Kubernetes version to use for the cluster. When not set, the latest default version will be used.", // TODO(elchead): refer to supported versions; wait for https://github.com/edgelesssys/constellation/pull/2661
				Optional:            true,
			},
			"constellation_microservices_version": schema.StringAttribute{
				MarkdownDescription: "The Constellation microservices version to use for the cluster.",
				Description:         "The Constellation microservices version to use for the cluster. When not set, the latest default version will be used.",
				Optional:            true,
			},
			"debug": schema.BoolAttribute{
				MarkdownDescription: "~> **Warning:** Do not enable Debug mode in production environments.\nEnable debug mode and allow the use of debug images.",
				Description:         "DON'T USE IN PRODUCTION: Enable debug mode and allow the use of debug images.",
				Optional:            true,
			},
			"init_endpoint": schema.StringAttribute{
				MarkdownDescription: "The endpoint to use for cluster initialization. This is the endpoint of the node running the bootstrapper.",
				Description:         "The endpoint to use for cluster initialization.",
				Optional:            true,
			},
			"kubernetes_api_endpoint": schema.StringAttribute{
				MarkdownDescription: "The endpoint to use for the Kubernetes API.",
				Description:         "The endpoint to use for the Kubernetes API. When not set, the default endpoint will be used.",
				Optional:            true,
			},
			"extra_microservices": schema.SingleNestedAttribute{
				MarkdownDescription: "Extra microservice settings.",
				Description:         "Extra microservice settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"csi_driver": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Enable the CSI driver microservice.",
						Description:         "Enable the CSI driver microservice.",
					},
				},
			},
			"master_secret": schema.StringAttribute{
				MarkdownDescription: "The master secret to use for the cluster.",
				Description:         "The master secret to use for the cluster.",
				Required:            true,
			},
			"init_secret": schema.StringAttribute{
				MarkdownDescription: "The init secret to use for the cluster.",
				Description:         "The init secret to use for the cluster.",
				Required:            true,
			}, // TODO merge / derive from master secret?
			"attestation":  newAttestationConfigAttribute(true),
			"measurements": newMeasurementsAttribute(true),
			"owner_id": schema.StringAttribute{
				MarkdownDescription: "The owner ID of the cluster.",
				Description:         "The owner ID of the cluster.",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The cluster ID of the cluster.",
				Description:         "The cluster ID of the cluster.",
				Computed:            true,
			},
			"kubeconfig": schema.StringAttribute{
				MarkdownDescription: "The kubeconfig of the cluster.",
				Description:         "The kubeconfig of the cluster.",
				Computed:            true,
			},
		},
	}
}

// Configure configures the resource.
func (r *ClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// client, ok := req.ProviderData.(*http.Client)

	// if !ok {
	//	resp.Diagnostics.AddError(
	//		"Unexpected Resource Configure Type",
	//		fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
	//	)

	//	return
	//}
}

// Create is called when the resource is created.
func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClusterResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Run init
	measurementsMap := make(map[string]measurement, len(data.Measurements.Elements()))
	diags := data.Measurements.ElementsAs(ctx, &measurementsMap, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tfSnpAttestation sevSnpAttestation
	diags = data.Attestation.As(ctx, &tfSnpAttestation, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	attestationVariant, err := variant.FromString(tfSnpAttestation.Variant)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("attestation_variant"),
			"Invalid Attestation Variant",
			fmt.Sprintf("Invalid attestation variant: %s", tfSnpAttestation.Variant))
		return
	}
	_, err = convertFromTfAttestationCfg(measurementsMap, tfSnpAttestation, attestationVariant)
	if err != nil {
		resp.Diagnostics.AddError("Parsing attestation config", err.Error())
		return
	}

	var extraMicroservices extraMicroservices
	diags = data.ExtraMicroservices.As(ctx, &extraMicroservices, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO(elchead): implement in follow up PR
	data.OwnerID = types.StringValue("owner_id")
	data.ClusterID = types.StringValue("cluster_id")
	data.Kubeconfig = types.StringValue("kubeconfig")
	// applier := constellation.NewApplier(log)
	//validator, err := choose.Validator(attestationCfg, &tfLogger{dg: &resp.Diagnostics})
	//if err != nil {
	//	resp.Diagnostics.AddError("Choosing validator", err.Error())
	//	return
	//}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is called when the resource is read or refreshed.
func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is called when the resource is updated.
func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete is called when the resource is destroyed.
func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

// ImportState imports to the resource.
func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type tfLogger struct {
	dg *diag.Diagnostics
}

func (l *tfLogger) Infof(format string, args ...any) {
	tflog.Info(context.Background(), fmt.Sprintf(format, args...))
}

func (l *tfLogger) Warnf(format string, args ...any) {
	l.dg.AddWarning(fmt.Sprintf(format, args...), "")
}

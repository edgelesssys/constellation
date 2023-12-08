/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constellation"
	"github.com/edgelesssys/constellation/v2/internal/constellation/helm"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/versions"
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
type ClusterResource struct {
	applier *constellation.Applier
}

// ClusterResourceModel describes the resource data model.
type ClusterResourceModel struct {
	Name                   types.String `tfsdk:"name"`
	CSP                    types.String `tfsdk:"csp"`
	UID                    types.String `tfsdk:"uid"`
	ImageVersion           types.String `tfsdk:"image_version"`
	ImageReference         types.String `tfsdk:"image_reference"`
	KubernetesVersion      types.String `tfsdk:"kubernetes_version"`
	MicroserviceVersion    types.String `tfsdk:"constellation_microservice_version"`
	OutOfClusterEndpoint   types.String `tfsdk:"out_of_cluster_endpoint"`
	InClusterEndpoint      types.String `tfsdk:"in_cluster_endpoint"`
	ExtraMicroservices     types.Object `tfsdk:"extra_microservices"`
	ExtraAPIServerCertSANs types.List   `tfsdk:"extra_api_server_cert_sans"`
	NetworkConfig          types.Object `tfsdk:"network_config"`
	MasterSecret           types.String `tfsdk:"master_secret"`
	MasterSecretSalt       types.String `tfsdk:"master_secret_salt"`
	MeasurementSalt        types.String `tfsdk:"measurement_salt"`
	InitSecret             types.String `tfsdk:"init_secret"`
	Attestation            types.Object `tfsdk:"attestation"`
	GCP                    types.Object `tfsdk:"gcp"`
	Azure                  types.Object `tfsdk:"azure"`

	OwnerID    types.String `tfsdk:"owner_id"`
	ClusterID  types.String `tfsdk:"cluster_id"`
	Kubeconfig types.String `tfsdk:"kubeconfig"`
}

type networkConfig struct {
	IpCidrNode    string `tfsdk:"ip_cidr_node"`
	IpCidrPod     string `tfsdk:"ip_cidr_pod"`
	IpCidrService string `tfsdk:"ip_cidr_service"`
}

type gcp struct {
	// ServiceAccountKey is the private key of the service account used within the cluster.
	ServiceAccountKey string `tfsdk:"service_account_key"`
	ProjectID         string `tfsdk:"project_id"`
}

type azure struct {
	TenantID                 string `tfsdk:"tenant_id"`
	Location                 string `tfsdk:"location"`
	UamiID                   string `tfsdk:"uami"`
	UamiResourceID           string `tfsdk:"uami_resource_id"`
	ResourceGroup            string `tfsdk:"resource_group"`
	SubscriptionID           string `tfsdk:"subscription_id"`
	NetworkSecurityGroupName string `tfsdk:"network_security_group_name"`
	LoadBalancerName         string `tfsdk:"load_balancer_name"`
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
			// Input attributes
			"name": schema.StringAttribute{
				MarkdownDescription: "Name used in the cluster's named resources / cluster name.",
				Description:         "Name used in the cluster's named resources / cluster name.",
				Required:            true, // TODO: Make optional and default to Constell.
			},
			"csp": schema.StringAttribute{
				MarkdownDescription: "The Cloud Service Provider (CSP) the cluster should run on.",
				Description:         "The Cloud Service Provider (CSP) the cluster should run on.",
				Required:            true,
			},
			"uid": schema.StringAttribute{
				MarkdownDescription: "The UID of the cluster.",
				Description:         "The UID of the cluster.",
				Required:            true,
			},
			"image_version": schema.StringAttribute{
				MarkdownDescription: "Constellation OS image version to use in the CSP specific reference format. Use the [`constellation_image`](../data-sources/image.md) data source to find the correct image version for your CSP.",
				Description:         "Constellation OS image version to use in the CSP specific reference format. Use the `constellation_image` data source to find the correct image version for your CSP.",
				Required:            true,
			},
			"image_reference": schema.StringAttribute{
				MarkdownDescription: "Constellation OS image reference to use in the CSP specific reference format. Use the [`constellation_image`](../data-sources/image.md) data source to find the correct image reference for your CSP.",
				Description:         "Constellation OS image reference to use in the CSP specific reference format. Use the `constellation_image` data source to find the correct image reference for your CSP.",
				Required:            true,
			},
			"kubernetes_version": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The Kubernetes version to use for the cluster. When not set, version %s is used. The supported versions are %s.", versions.Default, versions.SupportedK8sVersions()),
				Description:         fmt.Sprintf("The Kubernetes version to use for the cluster. When not set, version %s is used. The supported versions are %s.", versions.Default, versions.SupportedK8sVersions()),
				Optional:            true,
			},
			"constellation_microservice_version": schema.StringAttribute{
				MarkdownDescription: "The version of Constellation's microservices used within the cluster. When not set, the provider default version is used.",
				Description:         "The version of Constellation's microservices used within the cluster. When not set, the provider default version is used.",
				Optional:            true,
			},
			"out_of_cluster_endpoint": schema.StringAttribute{
				MarkdownDescription: "The endpoint of the cluster. Typically, this is the public IP of a loadbalancer.",
				Description:         "The endpoint of the cluster. Typically, this is the public IP of a loadbalancer.",
				Required:            true,
			},
			"in_cluster_endpoint": schema.StringAttribute{
				MarkdownDescription: "The endpoint of the cluster. When not set, the out-of-cluster endpoint is used.",
				Description:         "The endpoint of the cluster. When not set, the out-of-cluster endpoint is used.",
				Optional:            true,
			},
			"extra_microservices": schema.SingleNestedAttribute{
				MarkdownDescription: "Extra microservice settings.",
				Description:         "Extra microservice settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"csi_driver": schema.BoolAttribute{
						MarkdownDescription: "Enable Constellation's [encrypted CSI driver](https://docs.edgeless.systems/constellation/workflows/storage).",
						Description:         "Enable Constellation's encrypted CSI driver.",
						Required:            true,
					},
				},
			},
			"extra_api_server_cert_sans": schema.ListAttribute{
				MarkdownDescription: "List of additional Subject Alternative Names (SANs) for the API server certificate.",
				Description:         "List of additional Subject Alternative Names (SANs) for the API server certificate.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"network_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for the cluster's network.",
				Description:         "Configuration for the cluster's network.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"ip_cidr_node": schema.StringAttribute{
						MarkdownDescription: "CIDR range of the cluster's node network.",
						Description:         "CIDR range of the cluster's node network.",
						Required:            true,
					},
					"ip_cidr_pod": schema.StringAttribute{
						MarkdownDescription: "CIDR range of the cluster's pod network. Only required for clusters running on GCP.",
						Description:         "CIDR range of the cluster's pod network. Only required for clusters running on GCP.",
						Optional:            true,
					},
					"ip_cidr_service": schema.StringAttribute{
						MarkdownDescription: "CIDR range of the cluster's service network.",
						Description:         "CIDR range of the cluster's service network.",
						Required:            true,
					},
				},
			},
			"master_secret": schema.StringAttribute{
				MarkdownDescription: "Hex-encoded 32-byte master secret for the cluster.",
				Description:         "Hex-encoded 32-byte master secret for the cluster.",
				Required:            true,
			},
			"master_secret_salt": schema.StringAttribute{
				MarkdownDescription: "Hex-encoded 32-byte master secret salt for the cluster.",
				Description:         "Hex-encoded 32-byte master secret salt for the cluster.",
				Required:            true,
			},
			"measurement_salt": schema.StringAttribute{
				MarkdownDescription: "Hex-encoded 32-byte measurement salt for the cluster.",
				Description:         "Hex-encoded 32-byte measurement salt for the cluster.",
				Required:            true,
			},
			"init_secret": schema.StringAttribute{
				MarkdownDescription: "Secret used for initialization of the cluster.",
				Description:         "Secret used for initialization of the cluster.",
				Required:            true,
			},
			"attestation": newAttestationConfigAttribute(attributeInput),
			"gcp": schema.SingleNestedAttribute{
				MarkdownDescription: "GCP-specific configuration.",
				Description:         "GCP-specific configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"service_account_key": schema.StringAttribute{
						MarkdownDescription: "Private key of the service account used within the cluster.",
						Description:         "Private key of the service account used within the cluster.",
						Required:            true,
					},
					"project_id": schema.StringAttribute{
						MarkdownDescription: "ID of the GCP project the cluster resides in.",
						Description:         "ID of the GCP project the cluster resides in.",
						Required:            true,
					},
				},
			},
			"azure": schema.SingleNestedAttribute{
				MarkdownDescription: "Azure-specific configuration.",
				Description:         "Azure-specific configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"tenant_id": schema.StringAttribute{
						MarkdownDescription: "Tenant ID of the Azure account.",
						Description:         "Tenant ID of the Azure account.",
						Required:            true,
					},
					"location": schema.StringAttribute{
						MarkdownDescription: "Azure Location of the cluster.",
						Description:         "Azure Location of the cluster.",
						Required:            true,
					},
					"uami_id": schema.StringAttribute{
						MarkdownDescription: "ID of the User assigned managed identity (UAMI) used within the cluster.",
						Description:         "ID of the User assigned managed identity (UAMI) used within the cluster.",
						Required:            true,
					},
					"uami_resource_id": schema.StringAttribute{
						MarkdownDescription: "Resource ID of the User assigned managed identity (UAMI) used within the cluster.",
						Description:         "Resource ID of the User assigned managed identity (UAMI) used within the cluster.",
						Required:            true,
					},
					"resource_group": schema.StringAttribute{
						MarkdownDescription: "Name of the Azure resource group the cluster resides in.",
						Description:         "Name of the Azure resource group the cluster resides in.",
						Required:            true,
					},
					"subscription_id": schema.StringAttribute{
						MarkdownDescription: "ID of the Azure subscription the cluster resides in.",
						Description:         "ID of the Azure subscription the cluster resides in.",
						Required:            true,
					},
					"network_security_group_name": schema.StringAttribute{
						MarkdownDescription: "Name of the Azure network security group used for the cluster.",
						Description:         "Name of the Azure network security group used for the cluster.",
						Required:            true,
					},
					"load_balancer_name": schema.StringAttribute{
						MarkdownDescription: "Name of the Azure load balancer used by the cluster.",
						Description:         "Name of the Azure load balancer used by the cluster.",
						Required:            true,
					},
				},
			},

			// Computed (output) attributes
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
}

// Create is called when the resource is created.
func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read data supplied by Terraform runtime into the model
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse and convert values from the Terraform state
	// to formats the Constellation library can work with.

	att, diags := r.convertAttestationConfig(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log := &tfContextLogger{ctx: ctx}
	newDialer := func(validator atls.Validator) *dialer.Dialer {
		return dialer.New(nil, validator, &net.Dialer{})
	}
	r.applier = constellation.NewApplier(log, &nopSpinner{}, newDialer)
	validator, err := choose.Validator(att.config, log)
	if err != nil {
		resp.Diagnostics.AddError("Choosing validator", err.Error())
		return
	}

	secrets, diags := r.convertSecrets(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiServerCertSANs := make([]string, 0, len(data.ExtraAPIServerCertSANs.Elements()))
	for _, san := range data.ExtraAPIServerCertSANs.Elements() {
		apiServerCertSANs = append(apiServerCertSANs, san.String())
	}

	var microserviceCfg extraMicroservices
	diags = data.ExtraMicroservices.As(ctx, &microserviceCfg, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty: true, // we want to allow null values, as the CSIDriver field is optional
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	microserviceVersion, err := semver.New(data.MicroserviceVersion.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("constellation_microservice_version"),
			"Invalid microservice version",
			fmt.Sprintf("Parsing microservice version: %s", err))
		return
	}

	k8sVersion, diags := r.getK8sVersion(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	csp := cloudprovider.FromString(data.CSP.ValueString())

	serviceAccPayload := constellation.ServiceAccountPayload{}
	var gcpConfig gcp
	var azureConfig azure
	switch csp {
	case cloudprovider.GCP:
		diags = data.GCP.As(ctx, &gcpConfig, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := json.Unmarshal([]byte(gcpConfig.ServiceAccountKey), &serviceAccPayload.GCP); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("gcp").AtName("service_account_key"),
				"Unmarshalling service account key",
				fmt.Sprintf("Unmarshalling service account key: %s", err))
			return
		}
	case cloudprovider.Azure:
		diags = data.Azure.As(ctx, &azureConfig, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		serviceAccPayload.Azure = azureshared.ApplicationCredentials{
			TenantID:            azureConfig.TenantID,
			Location:            azureConfig.Location,
			PreferredAuthMethod: azureshared.AuthMethodUserAssignedIdentity,
			UamiResourceID:      azureConfig.UamiResourceID,
		}
	}
	serviceAccURI, err := constellation.MarshalServiceAccountURI(csp, serviceAccPayload)
	if err != nil {
		resp.Diagnostics.AddError("Marshalling service account URI", err.Error())
		return
	}

	// Run init RPC
	initRpcPayload := initRpcPayload{
		csp:               csp,
		masterSecret:      secrets.masterSecret,
		measurementSalt:   secrets.measurementSalt,
		apiServerCertSANs: apiServerCertSANs,
		azureConfig:       azureConfig,
		gcpConfig:         gcpConfig,
		maaURL:            att.maaURL,
		k8sVersion:        k8sVersion,
	}
	postInitState, diags := r.runInitRPC(ctx, initRpcPayload, &data, validator)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applier.SetKubeConfig([]byte(data.Kubeconfig.ValueString())); err != nil {
		resp.Diagnostics.AddError("Setting kubeconfig", err.Error())
		return
	}

	// Apply attestation config
	if err := r.applier.ApplyJoinConfig(ctx, att.config, secrets.measurementSalt); err != nil {
		resp.Diagnostics.AddError("Applying attestation config", err.Error())
		return
	}

	// Extend API Server Certificate SANs
	if err := r.applier.ExtendClusterConfigCertSANs(ctx, data.OutOfClusterEndpoint.ValueString(),
		"", apiServerCertSANs); err != nil {
		resp.Diagnostics.AddError("Extending API server certificate SANs", err.Error())
		return
	}

	// Apply Helm Charts
	payload := applyHelmChartsPayload{
		csp:                 cloudprovider.FromString(data.CSP.ValueString()),
		attestationVariant:  att.variant,
		k8sVersion:          k8sVersion,
		microserviceVersion: microserviceVersion,
		DeployCSIDriver:     microserviceCfg.CSIDriver,
		masterSecret:        secrets.masterSecret,
		serviceAccURI:       serviceAccURI,
	}
	diags = r.applyHelmCharts(ctx, payload, postInitState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is called when the resource is read or refreshed.
func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// All Calls to the Constellation API are idempotent, thus we don't need to implement reading.

	// Alternatively, we could:

	// Retrieve more up-to-date data from the cluster. e.g.:
	// - CSI Driver enabled?
	// - Kubernetes version?
	// - Microservice version?
	// - Attestation Config?

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is called when the resource is updated.
func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse and convert values from the Terraform state

	att, diags := r.convertAttestationConfig(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secrets, diags := r.convertSecrets(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiServerCertSANs := make([]string, 0, len(data.ExtraAPIServerCertSANs.Elements()))
	for _, san := range data.ExtraAPIServerCertSANs.Elements() {
		apiServerCertSANs = append(apiServerCertSANs, san.String())
	}

	var microserviceCfg extraMicroservices
	diags = data.ExtraMicroservices.As(ctx, &microserviceCfg, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	microserviceVersion, err := semver.New(data.MicroserviceVersion.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("constellation_microservice_version"),
			"Invalid microservice version",
			fmt.Sprintf("Parsing microservice version: %s", err))
		return
	}

	k8sVersion, diags := r.getK8sVersion(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageVersion, err := semver.New(data.ImageVersion.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("image_version"),
			"Invalid image version",
			fmt.Sprintf("Parsing image version: %s", err))
		return
	}

	log := &tfContextLogger{ctx: ctx}
	newDialer := func(validator atls.Validator) *dialer.Dialer {
		return dialer.New(nil, validator, &net.Dialer{})
	}
	r.applier = constellation.NewApplier(log, &nopSpinner{}, newDialer)

	// Run the actual update steps

	// Apply attestation config
	if err := r.applier.ApplyJoinConfig(ctx, att.config, secrets.measurementSalt); err != nil {
		resp.Diagnostics.AddError("Applying attestation config", err.Error())
		return
	}

	// Extend API Server Certificate SANs
	if err := r.applier.ExtendClusterConfigCertSANs(ctx, data.OutOfClusterEndpoint.ValueString(),
		"", apiServerCertSANs); err != nil {
		resp.Diagnostics.AddError("Extending API server certificate SANs", err.Error())
		return
	}

	// Apply Helm Charts
	payload := applyHelmChartsPayload{
		csp:                 cloudprovider.FromString(data.CSP.ValueString()),
		attestationVariant:  att.variant,
		k8sVersion:          k8sVersion,
		microserviceVersion: microserviceVersion,
		DeployCSIDriver:     microserviceCfg.CSIDriver,
		masterSecret:        secrets.masterSecret,
		serviceAccURI:       "", // TODO
	}
	diags = r.applyHelmCharts(ctx, payload, state.New())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Upgrade node image
	err = r.applier.UpgradeNodeImage(ctx,
		imageVersion,
		data.ImageReference.ValueString(),
		false)
	if err != nil {
		resp.Diagnostics.AddError("Upgrading node OS image", err.Error())
	}

	// Upgrade Kubernetes version
	if err := r.applier.UpgradeKubernetesVersion(ctx, k8sVersion, false); err != nil {
		resp.Diagnostics.AddError("Upgrading Kubernetes version", err.Error())
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete is called when the resource is destroyed.
func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports to the resource.
func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// TODO: Implement

	// Take Kubeconfig, Cluster Endpoint and Master Secret and save to state
}

// initRpcPayload groups the data required to run the init RPC.
type initRpcPayload struct {
	csp               cloudprovider.Provider
	masterSecret      uri.MasterSecret
	measurementSalt   []byte
	apiServerCertSANs []string
	azureConfig       azure
	gcpConfig         gcp
	maaURL            string
	k8sVersion        versions.ValidK8sVersion
}

// runInitRPC runs the init RPC on the cluster.
func (r *ClusterResource) runInitRPC(ctx context.Context, payload initRpcPayload,
	data *ClusterResourceModel, validator atls.Validator,
) (*state.State, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var networkCfg networkConfig
	castDiags := data.NetworkConfig.As(ctx, &networkCfg, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty: true, // we want to allow null values, as some of the field's subfields are optional
	})
	diags.Append(castDiags...)
	if diags.HasError() {
		return nil, diags
	}

	// fall back to outOfClusterEndpoint if inClusterEndpoint is not set
	inClusterEndpoint := data.InClusterEndpoint.ValueString()
	if inClusterEndpoint == "" {
		inClusterEndpoint = data.OutOfClusterEndpoint.ValueString()
	}

	stateFile := state.New().SetInfrastructure(state.Infrastructure{
		UID:               data.UID.ValueString(),
		ClusterEndpoint:   data.OutOfClusterEndpoint.ValueString(),
		InClusterEndpoint: inClusterEndpoint,
		InitSecret:        []byte(data.InitSecret.ValueString()),
		APIServerCertSANs: payload.apiServerCertSANs,
		Name:              data.Name.ValueString(),
		IPCidrNode:        networkCfg.IpCidrNode,
	})
	switch payload.csp {
	case cloudprovider.Azure:
		stateFile.Infrastructure.Azure = &state.Azure{
			ResourceGroup:            payload.azureConfig.ResourceGroup,
			SubscriptionID:           payload.azureConfig.SubscriptionID,
			NetworkSecurityGroupName: payload.azureConfig.NetworkSecurityGroupName,
			LoadBalancerName:         payload.azureConfig.LoadBalancerName,
			UserAssignedIdentity:     payload.azureConfig.UamiID,
			AttestationURL:           payload.maaURL,
		}
	case cloudprovider.GCP:
		stateFile.Infrastructure.GCP = &state.GCP{
			ProjectID: payload.gcpConfig.ProjectID,
			IPCidrPod: networkCfg.IpCidrPod,
		}
	}

	clusterLogs := &bytes.Buffer{}
	initResp, err := r.applier.Init(
		ctx, validator, stateFile, clusterLogs,
		constellation.InitPayload{
			MasterSecret:    payload.masterSecret,
			MeasurementSalt: payload.measurementSalt,
			K8sVersion:      payload.k8sVersion,
			ConformanceMode: false, // Conformance mode does't need to be configurable through the TF provider for now.
			ServiceCIDR:     networkCfg.IpCidrService,
		})
	if err != nil {
		var nonRetriable *constellation.NonRetriableInitError
		if errors.As(err, &nonRetriable) {
			diags.AddError("Cluster initialization failed.",
				fmt.Sprintf("This error is not recoverable. Clean up the cluster's infrastructure resources and try again.\nError: %s", err))
			if nonRetriable.LogCollectionErr != nil {
				diags.AddError("Bootstrapper log collection failed.",
					fmt.Sprintf("Failed to collect logs from bootstrapper: %s\n", nonRetriable.LogCollectionErr))
			} else {
				diags.AddWarning("Cluster log collection suceeded.", clusterLogs.String())
			}
		} else {
			diags.AddError("Cluster initialization failed.", fmt.Sprintf("You might try to apply the resource again.\nError: %s", err))
		}
		return nil, diags
	}

	rewrittenKubeconfig, err := r.applier.
		RewrittenKubeconfigBytes(initResp.GetKubeconfig(), stateFile.Infrastructure.ClusterEndpoint)
	if err != nil {
		diags.AddError("Rewriting kubeconfig endpoint", err.Error())
	}

	// Save data from init response into the Terraform state
	data.Kubeconfig = types.StringValue(string(rewrittenKubeconfig))
	data.ClusterID = types.StringValue(string(initResp.ClusterId))
	data.OwnerID = types.StringValue(string(initResp.OwnerId))

	// Save data from init response into the state
	stateFile.SetClusterValues(state.ClusterValues{
		ClusterID:       string(initResp.ClusterId),
		OwnerID:         string(initResp.OwnerId),
		MeasurementSalt: payload.measurementSalt,
	})

	return stateFile, diags
}

// applyHelmChartsPayload groups the data required to apply the Helm charts.
type applyHelmChartsPayload struct {
	csp                 cloudprovider.Provider
	attestationVariant  variant.Variant
	k8sVersion          versions.ValidK8sVersion
	microserviceVersion semver.Semver
	DeployCSIDriver     bool
	masterSecret        uri.MasterSecret
	serviceAccURI       string
}

// applyHelmCharts applies the Helm charts to the cluster.
func (r *ClusterResource) applyHelmCharts(ctx context.Context, payload applyHelmChartsPayload, state *state.State) diag.Diagnostics {
	diags := diag.Diagnostics{}
	options := helm.Options{
		CSP:                 payload.csp,
		AttestationVariant:  payload.attestationVariant,
		K8sVersion:          payload.k8sVersion,
		MicroserviceVersion: payload.microserviceVersion,
		DeployCSIDriver:     payload.DeployCSIDriver,
		Force:               false,
		Conformance:         false, // Conformance mode does't need to be configurable through the TF provider for now.
		HelmWaitMode:        helm.WaitModeAtomic,
		ApplyTimeout:        time.Duration(10 * time.Minute),
		AllowDestructive:    helm.DenyDestructive,
	}

	executor, _, err := r.applier.PrepareHelmCharts(options, state,
		payload.serviceAccURI, payload.masterSecret, nil)
	if err != nil {
		diags.AddError("Preparing Helm charts", err.Error())
		return diags
	}

	if err := executor.Apply(ctx); err != nil {
		diags.AddError("Applying Helm charts", err.Error())
		return diags
	}
	return diags
}

// attestationInput groups the attestation values in a state consumable by the Constellation library.
type attestationInput struct {
	variant variant.Variant
	maaURL  string
	config  config.AttestationCfg
}

// convertAttestationConfig converts the attestation config from the Terraform state to the format
// used by the Constellation library.
func (r *ClusterResource) convertAttestationConfig(ctx context.Context, data ClusterResourceModel) (attestationInput, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var tfAttestation attestation
	castDiags := data.Attestation.As(ctx, &tfAttestation, basetypes.ObjectAsOptions{})
	diags.Append(castDiags...)
	if diags.HasError() {
		return attestationInput{}, diags
	}

	attestationVariant, err := variant.FromString(tfAttestation.Variant)
	if err != nil {
		diags.AddAttributeError(
			path.Root("attestation_variant"),
			"Invalid Attestation Variant",
			fmt.Sprintf("Invalid attestation variant: %s", tfAttestation.Variant))
		return attestationInput{}, diags
	}

	attestationCfg, err := convertFromTfAttestationCfg(tfAttestation, attestationVariant)
	if err != nil {
		diags.AddAttributeError(
			path.Root("attestation"),
			"Invalid Attestation Config",
			fmt.Sprintf("Parsing attestation config: %s", err))
		return attestationInput{}, diags
	}

	return attestationInput{attestationVariant, tfAttestation.AzureSNPFirmwareSignerConfig.MAAURL, attestationCfg}, diags
}

// secretInput groups the secrets and salts in a state consumable by the Constellation library.
type secretInput struct {
	masterSecret    uri.MasterSecret
	initSecret      []byte
	measurementSalt []byte
}

// convertFromTfAttestationCfg converts the secrets and salts from the Terraform state to the format
// used by the Constellation library.
func (r *ClusterResource) convertSecrets(ctx context.Context, data ClusterResourceModel) (secretInput, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	masterSecret, err := hex.DecodeString(data.MasterSecret.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("master_secret"),
			"Unmarshalling master secret",
			fmt.Sprintf("Unmarshalling hex-encoded master secret: %s", err))
		return secretInput{}, diags
	}

	masterSecretSalt, err := hex.DecodeString(data.MasterSecretSalt.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("master_secret_salt"),
			"Unmarshalling master secret salt",
			fmt.Sprintf("Unmarshalling hex-encoded master secret salt: %s", err))
		return secretInput{}, diags
	}

	measurementSalt, err := hex.DecodeString(data.MeasurementSalt.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("measurement_salt"),
			"Unmarshalling measurement salt",
			fmt.Sprintf("Unmarshalling hex-encoded measurement salt: %s", err))
		return secretInput{}, diags
	}

	return secretInput{
		masterSecret:    uri.MasterSecret{Key: masterSecret, Salt: masterSecretSalt},
		initSecret:      []byte(data.InitSecret.ValueString()),
		measurementSalt: measurementSalt,
	}, diags
}

// getK8sVersion returns the Kubernetes version from the Terraform state if set, and the default
// version otherwise.
func (r *ClusterResource) getK8sVersion(ctx context.Context, data *ClusterResourceModel) (versions.ValidK8sVersion, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var k8sVersion versions.ValidK8sVersion
	var err error
	if data.KubernetesVersion.ValueString() != "" {
		k8sVersion, err = versions.NewValidK8sVersion(data.KubernetesVersion.ValueString(), true)
		if err != nil {
			diags.AddAttributeError(
				path.Root("kubernetes_vesion"),
				"Invalid Kubernetes version",
				fmt.Sprintf("Parsing Kubernetes version: %s", err))
			return "", diags
		}
	} else {
		tflog.Info(ctx, fmt.Sprintf("No Kubernetes version specified. Using default version %s.", versions.Default))
		k8sVersion = versions.Default
	}
	return k8sVersion, diags
}

// tfContextLogger is a logging adapter between the tflog package and
// Constellation's logger.
type tfContextLogger struct {
	ctx context.Context // bind context to struct to satisfy interface
}

func (l *tfContextLogger) Debugf(format string, args ...any) {
	tflog.Debug(l.ctx, fmt.Sprintf(format, args...))
}

func (l *tfContextLogger) Infof(format string, args ...any) {
	tflog.Info(l.ctx, fmt.Sprintf(format, args...))
}

func (l *tfContextLogger) Warnf(format string, args ...any) {
	tflog.Warn(l.ctx, fmt.Sprintf(format, args...))
}

type nopSpinner struct{ io.Writer }

func (s *nopSpinner) Start(string, bool)                {}
func (s *nopSpinner) Stop()                             {}
func (s *nopSpinner) Write(p []byte) (n int, err error) { return 1, nil }

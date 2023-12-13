/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation"
	"github.com/edgelesssys/constellation/v2/internal/constellation/helm"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/semver"
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
	newApplier func(ctx context.Context, validator atls.Validator) *constellation.Applier
}

// ClusterResourceModel describes the resource data model.
type ClusterResourceModel struct {
	Name                 types.String `tfsdk:"name"`
	CSP                  types.String `tfsdk:"csp"`
	UID                  types.String `tfsdk:"uid"`
	ImageVersion         types.String `tfsdk:"image_version"`
	ImageReference       types.String `tfsdk:"image_reference"`
	KubernetesVersion    types.String `tfsdk:"kubernetes_version"`
	MicroserviceVersion  types.String `tfsdk:"constellation_microservice_version"`
	OutOfClusterEndpoint types.String `tfsdk:"out_of_cluster_endpoint"`
	InClusterEndpoint    types.String `tfsdk:"in_cluster_endpoint"`
	ExtraMicroservices   types.Object `tfsdk:"extra_microservices"`
	APIServerCertSANs    types.List   `tfsdk:"api_server_cert_sans"`
	NetworkConfig        types.Object `tfsdk:"network_config"`
	MasterSecret         types.String `tfsdk:"master_secret"`
	MasterSecretSalt     types.String `tfsdk:"master_secret_salt"`
	MeasurementSalt      types.String `tfsdk:"measurement_salt"`
	InitSecret           types.String `tfsdk:"init_secret"`
	Attestation          types.Object `tfsdk:"attestation"`
	GCP                  types.Object `tfsdk:"gcp"`
	Azure                types.Object `tfsdk:"azure"`

	OwnerID    types.String `tfsdk:"owner_id"`
	ClusterID  types.String `tfsdk:"cluster_id"`
	KubeConfig types.String `tfsdk:"kubeconfig"`
}

type networkConfig struct {
	IPCidrNode    string `tfsdk:"ip_cidr_node"`
	IPCidrPod     string `tfsdk:"ip_cidr_pod"`
	IPCidrService string `tfsdk:"ip_cidr_service"`
}

type gcp struct {
	// ServiceAccountKey is the private key of the service account used within the cluster.
	ServiceAccountKey string `tfsdk:"service_account_key"`
	ProjectID         string `tfsdk:"project_id"`
}

type azure struct {
	TenantID                 string `tfsdk:"tenant_id"`
	Location                 string `tfsdk:"location"`
	UamiClientID             string `tfsdk:"uami_client_id"`
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
			"api_server_cert_sans": schema.ListAttribute{
				MarkdownDescription: "List of Subject Alternative Names (SANs) for the API server certificate. Usually, this will be" +
					" the out-of-cluster endpoint and the in-cluster endpoint, if existing.",
				Description: "List of Subject Alternative Names (SANs) for the API server certificate.",
				ElementType: types.StringType,
				Optional:    true,
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
						MarkdownDescription: "Base64-encoded private key JSON object of the service account used within the cluster.",
						Description:         "Base64-encoded private key JSON object of the service account used within the cluster.",
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
					"uami_client_id": schema.StringAttribute{
						MarkdownDescription: "Client ID of the User assigned managed identity (UAMI) used within the cluster.",
						Description:         "Client ID of the User assigned managed identity (UAMI) used within the cluster.",
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

	newDialer := func(validator atls.Validator) *dialer.Dialer {
		return dialer.New(nil, validator, &net.Dialer{})
	}

	r.newApplier = func(ctx context.Context, validator atls.Validator) *constellation.Applier {
		return constellation.NewApplier(&tfContextLogger{ctx: ctx}, &nopSpinner{}, newDialer)
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

	// Apply changes to the cluster, including the init RPC and skipping the node upgrade.
	diags := r.apply(ctx, &data, false, true)
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

	// Apply changes to the cluster, skipping the init RPC.
	diags := r.apply(ctx, &data, true, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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
	expectedSchemaMsg := fmt.Sprintf(
		"Expected URI of schema '%s://?%s=<...>&%s=<...>&%s=<...>&%s=<...>'",
		constants.ConstellationClusterURIScheme, constants.KubeConfigURIKey, constants.ClusterEndpointURIKey,
		constants.MasterSecretURIKey, constants.MasterSecretSaltURIKey)

	uri, err := url.Parse(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Parsing cluster URI",
			fmt.Sprintf("Parsing cluster URI: %s.\n%s", err, expectedSchemaMsg))
		return
	}

	if uri.Scheme != constants.ConstellationClusterURIScheme {
		resp.Diagnostics.AddError("Parsing cluster URI",
			fmt.Sprintf("Parsing cluster URI: Invalid scheme '%s'.\n%s", uri.Scheme, expectedSchemaMsg))
		return
	}

	// Parse query parameters
	query := uri.Query()
	kubeConfig := query.Get(constants.KubeConfigURIKey)
	clusterEndpoint := query.Get(constants.ClusterEndpointURIKey)
	masterSecret := query.Get(constants.MasterSecretURIKey)
	masterSecretSalt := query.Get(constants.MasterSecretSaltURIKey)

	if kubeConfig == "" {
		resp.Diagnostics.AddError("Parsing cluster URI",
			fmt.Sprintf("Parsing cluster URI: Missing query parameter '%s'.\n%s", constants.KubeConfigURIKey, expectedSchemaMsg))
		return
	}

	if clusterEndpoint == "" {
		resp.Diagnostics.AddError("Parsing cluster URI",
			fmt.Sprintf("Parsing cluster URI: Missing query parameter '%s'.\n%s", constants.ClusterEndpointURIKey, expectedSchemaMsg))
		return
	}

	if masterSecret == "" {
		resp.Diagnostics.AddError("Parsing cluster URI",
			fmt.Sprintf("Parsing cluster URI: Missing query parameter '%s'.\n%s", constants.MasterSecretURIKey, expectedSchemaMsg))
		return
	}

	if masterSecretSalt == "" {
		resp.Diagnostics.AddError("Parsing cluster URI",
			fmt.Sprintf("Parsing cluster URI: Missing query parameter '%s'.\n%s", constants.MasterSecretSaltURIKey, expectedSchemaMsg))
		return
	}

	decodedKubeConfig, err := base64.StdEncoding.DecodeString(kubeConfig)
	if err != nil {
		resp.Diagnostics.AddError("Parsing cluster URI",
			fmt.Sprintf("Parsing cluster URI: Decoding base64-encoded kubeconfig: %s.", err))
		return
	}

	// Sanity checks for master secret and master secret salt
	if _, err := hex.DecodeString(masterSecret); err != nil {
		resp.Diagnostics.AddError("Parsing cluster URI",
			fmt.Sprintf("Parsing cluster URI: Decoding hex-encoded master secret: %s.", err))
		return
	}

	if _, err := hex.DecodeString(masterSecretSalt); err != nil {
		resp.Diagnostics.AddError("Parsing cluster URI",
			fmt.Sprintf("Parsing cluster URI: Decoding hex-encoded master secret salt: %s.", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kubeconfig"), string(decodedKubeConfig))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("out_of_cluster_endpoint"), clusterEndpoint)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("master_secret"), masterSecret)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("master_secret_salt"), masterSecretSalt)...)
}

// apply applies changes to a cluster. It can be used for both creating and updating a cluster.
// This implements the core part of the Create and Update methods.
func (r *ClusterResource) apply(ctx context.Context, data *ClusterResourceModel, skipInitRPC, skipNodeUpgrade bool) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Parse and convert values from the Terraform state
	// to formats the Constellation library can work with.

	csp := cloudprovider.FromString(data.CSP.ValueString())

	// parse attestation config
	att, convertDiags := r.convertAttestationConfig(ctx, *data)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return diags
	}

	// parse secrets (i.e. measurement salt, master secret, etc.)
	secrets, convertDiags := r.convertSecrets(*data)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return diags
	}

	// parse API server certificate SANs
	apiServerCertSANs := make([]string, 0, len(data.APIServerCertSANs.Elements()))
	for _, san := range data.APIServerCertSANs.Elements() {
		apiServerCertSANs = append(apiServerCertSANs, san.String())
	}

	// parse network config
	var networkCfg networkConfig
	convertDiags = data.NetworkConfig.As(ctx, &networkCfg, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty: true, // we want to allow null values, as some of the field's subfields are optional.
	})
	diags.Append(convertDiags...)
	if diags.HasError() {
		return diags
	}

	// parse Constellation microservice config
	var microserviceCfg extraMicroservices
	convertDiags = data.ExtraMicroservices.As(ctx, &microserviceCfg, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty: true, // we want to allow null values, as the CSIDriver field is optional
	})
	diags.Append(convertDiags...)
	if diags.HasError() {
		return diags
	}

	// parse Constellation microservice version
	microserviceVersion, err := semver.New(data.MicroserviceVersion.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("constellation_microservice_version"),
			"Invalid microservice version",
			fmt.Sprintf("Parsing microservice version: %s", err))
		return diags
	}

	// parse Kubernetes version
	k8sVersion, getDiags := r.getK8sVersion(ctx, data)
	diags.Append(getDiags...)
	if diags.HasError() {
		return diags
	}

	// parse OS image version
	imageVersion, err := semver.New(data.ImageVersion.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("image_version"),
			"Invalid image version",
			fmt.Sprintf("Parsing image version: %s", err))
		return diags
	}

	// Parse in-cluster service account info.
	serviceAccPayload := constellation.ServiceAccountPayload{}
	var gcpConfig gcp
	var azureConfig azure
	switch csp {
	case cloudprovider.GCP:
		convertDiags = data.GCP.As(ctx, &gcpConfig, basetypes.ObjectAsOptions{})
		diags.Append(convertDiags...)
		if diags.HasError() {
			return diags
		}

		decodedSaKey, err := base64.StdEncoding.DecodeString(gcpConfig.ServiceAccountKey)
		if err != nil {
			diags.AddAttributeError(
				path.Root("gcp").AtName("service_account_key"),
				"Decoding service account key",
				fmt.Sprintf("Decoding base64-encoded service account key: %s", err))
			return diags
		}

		if err := json.Unmarshal(decodedSaKey, &serviceAccPayload.GCP); err != nil {
			diags.AddAttributeError(
				path.Root("gcp").AtName("service_account_key"),
				"Unmarshalling service account key",
				fmt.Sprintf("Unmarshalling service account key: %s", err))
			return diags
		}
	case cloudprovider.Azure:
		convertDiags = data.Azure.As(ctx, &azureConfig, basetypes.ObjectAsOptions{})
		diags.Append(convertDiags...)
		if diags.HasError() {
			return diags
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
		diags.AddError("Marshalling service account URI", err.Error())
		return diags
	}

	// we want to fall back to outOfClusterEndpoint if inClusterEndpoint is not set.
	inClusterEndpoint := data.InClusterEndpoint.ValueString()
	if inClusterEndpoint == "" {
		inClusterEndpoint = data.OutOfClusterEndpoint.ValueString()
	}

	// setup clients
	validator, err := choose.Validator(att.config, &tfContextLogger{ctx: ctx})
	if err != nil {
		diags.AddError("Choosing validator", err.Error())
		return diags
	}
	applier := r.newApplier(ctx, validator)

	// Construct in-memory state file
	stateFile := state.New().SetInfrastructure(state.Infrastructure{
		UID:               data.UID.ValueString(),
		ClusterEndpoint:   data.OutOfClusterEndpoint.ValueString(),
		InClusterEndpoint: inClusterEndpoint,
		InitSecret:        []byte(data.InitSecret.ValueString()),
		APIServerCertSANs: apiServerCertSANs,
		Name:              data.Name.ValueString(),
		IPCidrNode:        networkCfg.IPCidrNode,
	})
	switch csp {
	case cloudprovider.Azure:
		stateFile.Infrastructure.Azure = &state.Azure{
			ResourceGroup:            azureConfig.ResourceGroup,
			SubscriptionID:           azureConfig.SubscriptionID,
			NetworkSecurityGroupName: azureConfig.NetworkSecurityGroupName,
			LoadBalancerName:         azureConfig.LoadBalancerName,
			UserAssignedIdentity:     azureConfig.UamiClientID,
			AttestationURL:           att.maaURL,
		}
	case cloudprovider.GCP:
		stateFile.Infrastructure.GCP = &state.GCP{
			ProjectID: gcpConfig.ProjectID,
			IPCidrPod: networkCfg.IPCidrPod,
		}
	}

	// Now, we perform the actual applying.

	// Run init RPC
	var initDiags diag.Diagnostics
	if !skipInitRPC {
		// run the init RPC and retrieve the post-init state
		initRPCPayload := initRPCPayload{
			csp:               csp,
			masterSecret:      secrets.masterSecret,
			measurementSalt:   secrets.measurementSalt,
			apiServerCertSANs: apiServerCertSANs,
			azureCfg:          azureConfig,
			gcpCfg:            gcpConfig,
			networkCfg:        networkCfg,
			maaURL:            att.maaURL,
			k8sVersion:        k8sVersion,
			inClusterEndpoint: inClusterEndpoint,
		}
		initDiags = r.runInitRPC(ctx, applier, initRPCPayload, data, validator, stateFile)
		diags.Append(initDiags...)
		if diags.HasError() {
			return diags
		}
	}

	// Here, we either have the post-init values from the actual init RPC
	// or, if performing an upgrade and skipping the init RPC, we have the
	// values from the Terraform state.
	stateFile.SetClusterValues(state.ClusterValues{
		ClusterID:       data.ClusterID.ValueString(),
		OwnerID:         data.OwnerID.ValueString(),
		MeasurementSalt: secrets.measurementSalt,
	})

	// Kubeconfig is in the state by now. Either through the init RPC or through
	// already being in the state.
	if err := applier.SetKubeConfig([]byte(data.KubeConfig.ValueString())); err != nil {
		diags.AddError("Setting kubeconfig", err.Error())
		return diags
	}

	// Apply attestation config
	if err := applier.ApplyJoinConfig(ctx, att.config, secrets.measurementSalt); err != nil {
		diags.AddError("Applying attestation config", err.Error())
		return diags
	}

	// Extend API Server Certificate SANs
	if err := applier.ExtendClusterConfigCertSANs(ctx, data.OutOfClusterEndpoint.ValueString(),
		"", apiServerCertSANs); err != nil {
		diags.AddError("Extending API server certificate SANs", err.Error())
		return diags
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
	helmDiags := r.applyHelmCharts(ctx, applier, payload, stateFile)
	diags.Append(helmDiags...)
	if diags.HasError() {
		return diags
	}

	if !skipNodeUpgrade {
		// Upgrade node image
		err = applier.UpgradeNodeImage(ctx,
			imageVersion,
			data.ImageReference.ValueString(),
			false)
		if err != nil {
			diags.AddError("Upgrading node OS image", err.Error())
			return diags
		}

		// Upgrade Kubernetes version
		if err := applier.UpgradeKubernetesVersion(ctx, k8sVersion, false); err != nil {
			diags.AddError("Upgrading Kubernetes version", err.Error())
			return diags
		}
	}

	return diags
}

// initRPCPayload groups the data required to run the init RPC.
type initRPCPayload struct {
	csp               cloudprovider.Provider   // cloud service provider the cluster runs on.
	masterSecret      uri.MasterSecret         // master secret of the cluster.
	measurementSalt   []byte                   // measurement salt of the cluster.
	apiServerCertSANs []string                 // additional SANs to add to the API server certificate.
	azureCfg          azure                    // Azure-specific configuration.
	gcpCfg            gcp                      // GCP-specific configuration.
	networkCfg        networkConfig            // network configuration of the cluster.
	maaURL            string                   // URL of the MAA service. Only used for Azure clusters.
	k8sVersion        versions.ValidK8sVersion // Kubernetes version of the cluster.
	// Internal Endpoint of the cluster.
	// If no internal LB is used, this should be the same as the out-of-cluster endpoint.
	inClusterEndpoint string
}

// runInitRPC runs the init RPC on the cluster.
func (r *ClusterResource) runInitRPC(ctx context.Context, applier *constellation.Applier, payload initRPCPayload,
	data *ClusterResourceModel, validator atls.Validator, stateFile *state.State,
) diag.Diagnostics {
	diags := diag.Diagnostics{}
	clusterLogs := &bytes.Buffer{}
	initOutput, err := applier.Init(
		ctx, validator, stateFile, clusterLogs,
		constellation.InitPayload{
			MasterSecret:    payload.masterSecret,
			MeasurementSalt: payload.measurementSalt,
			K8sVersion:      payload.k8sVersion,
			ConformanceMode: false, // Conformance mode does't need to be configurable through the TF provider for now.
			ServiceCIDR:     payload.networkCfg.IPCidrService,
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
				diags.AddWarning("Cluster log collection succeeded.", clusterLogs.String())
			}
		} else {
			diags.AddError("Cluster initialization failed.", fmt.Sprintf("You might try to apply the resource again.\nError: %s", err))
		}
		return diags
	}

	// Save data from init response into the Terraform state
	data.KubeConfig = types.StringValue(string(initOutput.Kubeconfig))
	data.ClusterID = types.StringValue(initOutput.ClusterID)
	data.OwnerID = types.StringValue(initOutput.OwnerID)

	return diags
}

// applyHelmChartsPayload groups the data required to apply the Helm charts.
type applyHelmChartsPayload struct {
	csp                 cloudprovider.Provider   // cloud service provider the cluster runs on.
	attestationVariant  variant.Variant          // attestation variant used on the cluster's nodes.
	k8sVersion          versions.ValidK8sVersion // Kubernetes version of the cluster.
	microserviceVersion semver.Semver            // version of the Constellation microservices used on the cluster.
	DeployCSIDriver     bool                     // Whether to deploy the CSI driver.
	masterSecret        uri.MasterSecret         // master secret of the cluster.
	serviceAccURI       string                   // URI of the service account used within the cluster.
}

// applyHelmCharts applies the Helm charts to the cluster.
func (r *ClusterResource) applyHelmCharts(ctx context.Context, applier *constellation.Applier,
	payload applyHelmChartsPayload, state *state.State,
) diag.Diagnostics {
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
		ApplyTimeout:        10 * time.Minute,
		AllowDestructive:    helm.DenyDestructive,
	}

	executor, _, err := applier.PrepareHelmCharts(options, state,
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
func (r *ClusterResource) convertSecrets(data ClusterResourceModel) (secretInput, diag.Diagnostics) {
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

func (s *nopSpinner) Start(string, bool)              {}
func (s *nopSpinner) Stop()                           {}
func (s *nopSpinner) Write([]byte) (n int, err error) { return 1, nil }

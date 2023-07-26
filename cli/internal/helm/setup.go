package helm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// ProviderMetadata implementers read/write cloud provider metadata.
type ProviderMetadata interface {
	// UID returns the unique identifier for the constellation.
	UID(ctx context.Context) (string, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
	// GetLoadBalancerEndpoint retrieves the load balancer endpoint.
	GetLoadBalancerEndpoint(ctx context.Context) (host, port string, err error)
}

// SetupMicroserviceVals returns the values for the microservice chart.
func SetupMicroserviceVals(ctx context.Context, log debugLog, provider cloudprovider.Provider, measurementSalt []byte, uid, serviceAccURI string) (map[string]any, error) {
	tfClient, err := terraform.New(ctx, constants.TerraformWorkingDir)
	if err != nil {
		return nil, fmt.Errorf("creating Terraform client: %w", err)
	}
	output, err := tfClient.ShowCluster(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting Terraform output: %w", err)
	}
	log.Debugf("Terraform cluster output: %+v", output) // TODO why not working? (propagate debug to logger).
	fmt.Printf("A Terraform cluster output: %+v\n", output)

	extraVals := map[string]any{
		"join-service": map[string]any{
			"measurementSalt": base64.StdEncoding.EncodeToString(measurementSalt),
		},
		"verification-service": map[string]any{
			"loadBalancerIP": output.IP,
		},
		"konnectivity": map[string]any{
			"loadBalancerIP": output.IP,
		},
	}
	switch provider {
	case cloudprovider.GCP:
		serviceAccountKey, err := gcpshared.ServiceAccountKeyFromURI(serviceAccURI)
		if err != nil {
			return nil, fmt.Errorf("getting service account key: %w", err)
		}
		rawKey, err := json.Marshal(serviceAccountKey)
		if err != nil {
			return nil, fmt.Errorf("marshaling service account key: %w", err)
		}
		if output.GCP == nil {
			return nil, fmt.Errorf("no GCP output from Terraform")
		}
		extraVals["ccm"] = map[string]any{
			"GCP": map[string]any{
				"projectID":         output.GCP.ProjectID,
				"uid":               uid,
				"secretData":        string(rawKey),
				"subnetworkPodCIDR": output.GCP.IPCidrPod,
			},
		}
	case cloudprovider.Azure:
		if output.Azure == nil {
			return nil, fmt.Errorf("no Azure output from Terraform")
		}
		ccmConfig, err := getCCMConfig(*output.Azure, serviceAccURI)
		if err != nil {
			return nil, fmt.Errorf("getting Azure CCM config: %w", err)
		}
		extraVals["ccm"] = map[string]any{
			"Azure": map[string]any{
				"azureConfig": string(ccmConfig),
			},
		}
	}

	return extraVals, nil
}

type cloudConfig struct {
	Cloud                       string `json:"cloud,omitempty"`
	TenantID                    string `json:"tenantId,omitempty"`
	SubscriptionID              string `json:"subscriptionId,omitempty"`
	ResourceGroup               string `json:"resourceGroup,omitempty"`
	Location                    string `json:"location,omitempty"`
	SubnetName                  string `json:"subnetName,omitempty"`
	SecurityGroupName           string `json:"securityGroupName,omitempty"`
	SecurityGroupResourceGroup  string `json:"securityGroupResourceGroup,omitempty"`
	LoadBalancerName            string `json:"loadBalancerName,omitempty"`
	LoadBalancerSku             string `json:"loadBalancerSku,omitempty"`
	VNetName                    string `json:"vnetName,omitempty"`
	VNetResourceGroup           string `json:"vnetResourceGroup,omitempty"`
	CloudProviderBackoff        bool   `json:"cloudProviderBackoff,omitempty"`
	UseInstanceMetadata         bool   `json:"useInstanceMetadata,omitempty"`
	VMType                      string `json:"vmType,omitempty"`
	UseManagedIdentityExtension bool   `json:"useManagedIdentityExtension,omitempty"`
	UserAssignedIdentityID      string `json:"userAssignedIdentityID,omitempty"`
}

// GetCCMConfig returns the configuration needed for the Kubernetes Cloud Controller Manager on Azure.
func getCCMConfig(tfOutput terraform.AzureApplyOutput, serviceAccURI string) ([]byte, error) {
	creds, err := azureshared.ApplicationCredentialsFromURI(serviceAccURI)
	if err != nil {
		return nil, fmt.Errorf("getting service account key: %w", err)
	}
	useManagedIdentityExtension := creds.PreferredAuthMethod == azureshared.AuthMethodUserAssignedIdentity
	config := cloudConfig{
		Cloud:                       "AzurePublicCloud",
		TenantID:                    creds.TenantID,
		SubscriptionID:              tfOutput.SubscriptionID,
		ResourceGroup:               tfOutput.ResourceGroup,
		LoadBalancerSku:             "standard",
		SecurityGroupName:           tfOutput.NetworkSecurityGroupName,
		LoadBalancerName:            tfOutput.LoadBalancerName,
		UseInstanceMetadata:         true,
		VMType:                      "vmss",
		Location:                    creds.Location,
		UseManagedIdentityExtension: useManagedIdentityExtension,
		UserAssignedIdentityID:      tfOutput.UserAssignedIdentity,
	}

	return json.Marshal(config)
}

// SetupOperatorVals returns the values for the constellation-operator chart.
func SetupOperatorVals(_ context.Context, uid string) (map[string]any, error) {
	return map[string]any{
		"constellation-operator": map[string]any{
			"constellationUID": uid,
		},
	}, nil
}

//func GetMetadaClient(ctx context.Context, provider cloudprovider.Provider) (metadataAPI ProviderMetadata, err error) {
//	switch provider {
//	case cloudprovider.AWS:
//		metadata, err := aws.New(ctx)
//		if err != nil {
//			return nil, fmt.Errorf("creating AWS metadata client: %w", err)
//			// log.With(zap.Error(err)).Fatalf("Failed to set up AWS metadata API")
//		}
//		metadataAPI = metadata
//	case cloudprovider.GCP:
//		metadata, err := gcp.New(ctx)
//		if err != nil {
//			return nil, fmt.Errorf("creating GCP metadata client: %w", err)
//		}
//		metadataAPI = metadata
//	case cloudprovider.Azure:
//		metadata, err := azure.New(ctx)
//		if err != nil {
//			return nil, fmt.Errorf("creating Azure metadata client: %w", err)
//		}
//		metadataAPI = metadata
//	case cloudprovider.QEMU:
//		metadata := qemu.New()
//		metadataAPI = metadata
//	case cloudprovider.OpenStack:
//		metadata, err := openstack.New(ctx)
//		if err != nil {
//			return nil, fmt.Errorf("creating OpenStack metadata client: %w", err)
//		}
//		metadataAPI = metadata
//	default:
//		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
//		// metadataAPI = &providerMetadataFake{}
//		// cloudLogger = &logging.NopLogger{}
//		// var simulatedTPMCloser io.Closer
//		// openDevice, simulatedTPMCloser = simulator.NewSimulatedTPMOpenFunc()
//		// defer simulatedTPMCloser.Close()
//		// fs = afero.NewMemMapFs()
//	}
//	return metadataAPI, nil
//}

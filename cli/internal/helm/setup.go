package helm

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
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
func SetupMicroserviceVals(ctx context.Context, provider cloudprovider.Provider, measurementSalt []byte, uid, serviceAccURI string) (map[string]any, error) {
	//tfClient, err := terraform.New(ctx, constants.TerraformWorkingDir)
	//if err != nil {
	//	return nil, fmt.Errorf("creating Terraform client: %w", err)
	//}
	//output, err := tfClient.ShowCluster(ctx)
	//if err != nil {
	//	return nil, fmt.Errorf("getting Terraform output: %w", err)
	//}
	//extraVals := map[string]any{
	//	"join-service": map[string]any{
	//		"measurementSalt": base64.StdEncoding.EncodeToString(measurementSalt),
	//	},
	//	"verification-service": map[string]any{
	//		"loadBalancerIP": output.IP,
	//	},
	//	"konnectivity": map[string]any{
	//		"loadBalancerIP": output.IP,
	//	},
	//}
	// gcp: uid, gcp.ProviderID, cloudServiceAccountURI (or seviceAccountKey), subnetworkPodCIDR
	// azure: providerID, cloudServiceAccountURI
	//switch provider {
	//case cloudprovider.GCP:
	//	serviceAccountKey, err := gcpshared.ServiceAccountKeyFromURI(serviceAccURI)
	//	if err != nil {
	//		return nil, fmt.Errorf("getting service account key: %w", err)
	//	}
	//	rawKey, err := json.Marshal(serviceAccountKey)
	//	if err != nil {
	//		return nil, fmt.Errorf("marshaling service account key: %w", err)
	//	}
	//	if output.GCP == nil {
	//		return nil, fmt.Errorf("no GCP output from Terraform")
	//	}
	//	extraVals["ccm"] = map[string]any{
	//		"GCP": map[string]any{
	//			"projectID":         output.GCP.ProjectID,
	//			"uid":               uid,
	//			"secretData":        string(rawKey),
	//			"subnetworkPodCIDR": output.GCP.IPCidrPod,
	//		},
	//	}
	//	// case cloudprovider.Azure:
	//	//	extraVals["ccm"] = map[string]any{
	//	//		"Azure": map[string]any{
	//	//			"azureConfig": string(ccmConfig),
	//	//		},
	//	//	}
	//}

	extraVals := map[string]any{}
	return extraVals, nil
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

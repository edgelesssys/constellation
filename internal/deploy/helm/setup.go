package helm

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cloud/aws"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azure"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcp"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	"github.com/edgelesssys/constellation/v2/internal/cloud/qemu"
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
func SetupMicroserviceVals(ctx context.Context, measurementSalt []byte) (map[string]any, error) {
	extraVals := map[string]any{
		"join-service": map[string]any{
			"measurementSalt": base64.StdEncoding.EncodeToString(measurementSalt),
		},
		"verification-service": map[string]any{
			"loadBalancerIP": "IP",
		},
		"konnectivity": map[string]any{
			"loadBalancerIP": "IP",
		},
	}
	// lbIP, measurementSalt
	// gcp: uid, gcp.ProviderID, cloudServiceAccountURI (or seviceAccountKey), subnetworkPodCIDR
	// azure: providerID, cloudServiceAccountURI
	return extraVals, nil
}

// SetupOperatorVals returns the values for the constellation-operator chart.
func SetupOperatorVals(ctx context.Context, uid string) (map[string]any, error) {
	return map[string]any{
		"constellation-operator": map[string]any{
			"constellationUID": uid,
		},
	}, nil
}

func GetMetadaClient(ctx context.Context, provider cloudprovider.Provider) (metadataAPI ProviderMetadata, err error) {
	switch provider {
	case cloudprovider.AWS:
		metadata, err := aws.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating AWS metadata client: %w", err)
			// log.With(zap.Error(err)).Fatalf("Failed to set up AWS metadata API")
		}
		metadataAPI = metadata
	case cloudprovider.GCP:
		metadata, err := gcp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating GCP metadata client: %w", err)
		}
		metadataAPI = metadata
	case cloudprovider.Azure:
		metadata, err := azure.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating Azure metadata client: %w", err)
		}
		metadataAPI = metadata
	case cloudprovider.QEMU:
		metadata := qemu.New()
		metadataAPI = metadata
	case cloudprovider.OpenStack:
		metadata, err := openstack.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating OpenStack metadata client: %w", err)
		}
		metadataAPI = metadata
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
		// metadataAPI = &providerMetadataFake{}
		// cloudLogger = &logging.NopLogger{}
		// var simulatedTPMCloser io.Closer
		// openDevice, simulatedTPMCloser = simulator.NewSimulatedTPMOpenFunc()
		// defer simulatedTPMCloser.Close()
		// fs = afero.NewMemMapFs()
	}
	return metadataAPI, nil
}

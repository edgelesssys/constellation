package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/role"
)

// Metadata implements core.ProviderMetadata interface.
type Metadata struct{}

// List retrieves all instances belonging to the current constellation.
func (m Metadata) List(ctx context.Context) ([]cloudtypes.Instance, error) {
	// TODO: implement using https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.DescribeInstances
	// And using AWS ec2 instance tags
	panic("function *Metadata.List not implemented")
}

// Self retrieves the current instance.
func (m Metadata) Self(ctx context.Context) (cloudtypes.Instance, error) {
	identityDocument, err := retrieveIdentityDocument(ctx)
	if err != nil {
		return cloudtypes.Instance{}, err
	}
	// TODO: implement metadata using AWS ec2 instance tags
	return cloudtypes.Instance{
		Name:       identityDocument.InstanceID,
		ProviderID: providerID(identityDocument),
		PrivateIPs: []string{
			identityDocument.PrivateIP,
		},
	}, nil
}

// GetInstance retrieves an instance using its providerID.
func (m Metadata) GetInstance(ctx context.Context, providerID string) (cloudtypes.Instance, error) {
	// TODO: implement using https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#DescribeInstancesAPIClient.DescribeInstances
	// And using AWS ec2 instance tags
	// Filter request to only return info on this instance
	panic("function *Metadata.GetInstance not implemented")
}

// SignalRole signals the constellation role via cloud provider metadata (if supported by the CSP and deployment type, otherwise does nothing).
func (m Metadata) SignalRole(ctx context.Context, role role.Role) error {
	panic("function *Metadata.SignalRole not implemented")
}

// SetVPNIP stores the internally used VPN IP in cloud provider metadata (if supported and required for autoscaling by the CSP, otherwise does nothing).
func (m Metadata) SetVPNIP(ctx context.Context, vpnIP string) error {
	panic("function *Metadata.SetVPNIP not implemented")
}

// Supported is used to determine if metadata API is implemented for this cloud provider.
func (m Metadata) Supported() bool {
	return false
}

// retrieveIdentityDocument retrieves an AWS instance identity document.
func retrieveIdentityDocument(ctx context.Context) (*imds.GetInstanceIdentityDocumentOutput, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load default AWS configuration: %w", err)
	}
	client := imds.NewFromConfig(cfg)
	identityDocument, err := client.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve AWS instance identity document: %w", err)
	}
	return identityDocument, nil
}

func providerID(identityDocument *imds.GetInstanceIdentityDocumentOutput) string {
	// On AWS, the ProviderID has the form "aws:///<AVAILABILITY_ZONE>/<EC2_INSTANCE_ID>"
	return fmt.Sprintf("aws://%v/%v", identityDocument.AvailabilityZone, identityDocument.InstanceID)
}

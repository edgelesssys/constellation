package cloudprovider

const (
	// CloudControllerManagerImageAWS is the CCM image used on AWS.
	CloudControllerManagerImageAWS = "us.gcr.io/k8s-artifacts-prod/provider-aws/cloud-controller-manager:v1.22.0-alpha.0"
	// CloudControllerManagerImageGCP is the CCM image used on GCP.
	// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
	CloudControllerManagerImageGCP = "ghcr.io/edgelesssys/cloud-provider-gcp:sha-2f6a5b07fc2d37f24f8ff725132f87584d627d8f"
	// CloudControllerManagerImageAzure is the CCM image used on Azure.
	CloudControllerManagerImageAzure = "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.23.11"
	// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
	CloudNodeManagerImageAzure = "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.23.11"
)

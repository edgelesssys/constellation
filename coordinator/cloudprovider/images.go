package cloudprovider

const (
	// CloudControllerManagerImageAWS is the CCM image used on AWS.
	CloudControllerManagerImageAWS = "us.gcr.io/k8s-artifacts-prod/provider-aws/cloud-controller-manager:v1.22.0-alpha.0"
	// CloudControllerManagerImageGCP is the CCM image used on GCP.
	// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
	CloudControllerManagerImageGCP = "ghcr.io/malt3/cloud-provider-gcp:latest"
)

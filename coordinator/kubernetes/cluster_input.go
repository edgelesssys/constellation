package kubernetes

// InitClusterInput collects the arguments to initialize a new cluster.
type InitClusterInput struct {
	APIServerAdvertiseIP           string
	NodeName                       string
	ProviderID                     string
	SupportClusterAutoscaler       bool
	AutoscalingCloudprovider       string
	AutoscalingNodeGroups          []string
	SupportsCloudControllerManager bool
	CloudControllerManagerName     string
	CloudControllerManagerImage    string
	CloudControllerManagerPath     string
}

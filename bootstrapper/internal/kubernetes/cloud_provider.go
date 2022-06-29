package kubernetes

import (
	"context"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	k8s "k8s.io/api/core/v1"
)

// ProviderMetadata implementers read/write cloud provider metadata.
type ProviderMetadata interface {
	// List retrieves all instances belonging to the current Constellation.
	List(ctx context.Context) ([]metadata.InstanceMetadata, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
	// GetSubnetworkCIDR retrieves the subnetwork CIDR for the current instance.
	GetSubnetworkCIDR(ctx context.Context) (string, error)
	// SupportsLoadBalancer returns true if the cloud provider supports load balancers.
	SupportsLoadBalancer() bool
	// GetLoadBalancerIP retrieves the load balancer IP.
	GetLoadBalancerIP(ctx context.Context) (string, error)
	// GetInstance retrieves an instance using its providerID.
	GetInstance(ctx context.Context, providerID string) (metadata.InstanceMetadata, error)
	// Supported is used to determine if metadata API is implemented for this cloud provider.
	Supported() bool
}

// CloudControllerManager implementers provide configuration for the k8s cloud-controller-manager.
type CloudControllerManager interface {
	// Image returns the container image used to provide cloud-controller-manager for the cloud-provider.
	Image() string
	// Path returns the path used by cloud-controller-manager executable within the container image.
	Path() string
	// Name returns the cloud-provider name as used by k8s cloud-controller-manager (k8s.gcr.io/cloud-controller-manager).
	Name() string
	// ExtraArgs returns a list of arguments to append to the cloud-controller-manager command.
	ExtraArgs() []string
	// ConfigMaps returns a list of ConfigMaps to deploy together with the k8s cloud-controller-manager
	// Reference: https://kubernetes.io/docs/concepts/configuration/configmap/ .
	ConfigMaps(instance metadata.InstanceMetadata) (resources.ConfigMaps, error)
	// Secrets returns a list of secrets to deploy together with the k8s cloud-controller-manager.
	// Reference: https://kubernetes.io/docs/concepts/configuration/secret/ .
	Secrets(ctx context.Context, providerID, cloudServiceAccountURI string) (resources.Secrets, error)
	// Volumes returns a list of volumes to deploy together with the k8s cloud-controller-manager.
	// Reference: https://kubernetes.io/docs/concepts/storage/volumes/ .
	Volumes() []k8s.Volume
	// VolumeMounts a list of of volume mounts to deploy together with the k8s cloud-controller-manager.
	VolumeMounts() []k8s.VolumeMount
	// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cloud-controller-manager.
	Env() []k8s.EnvVar
	// Supported is used to determine if cloud controller manager is implemented for this cloud provider.
	Supported() bool
}

// CloudNodeManager implementers provide configuration for the k8s cloud-node-manager.
type CloudNodeManager interface {
	// Image returns the container image used to provide cloud-node-manager for the cloud-provider.
	Image() string
	// Path returns the path used by cloud-node-manager executable within the container image.
	Path() string
	// ExtraArgs returns a list of arguments to append to the cloud-node-manager command.
	ExtraArgs() []string
	// Supported is used to determine if cloud node manager is implemented for this cloud provider.
	Supported() bool
}

// ClusterAutoscaler implementers provide configuration for the k8s cluster-autoscaler.
type ClusterAutoscaler interface {
	// Name returns the cloud-provider name as used by k8s cluster-autoscaler.
	Name() string
	// Secrets returns a list of secrets to deploy together with the k8s cluster-autoscaler.
	Secrets(providerID, cloudServiceAccountURI string) (resources.Secrets, error)
	// Volumes returns a list of volumes to deploy together with the k8s cluster-autoscaler.
	Volumes() []k8s.Volume
	// VolumeMounts returns a list of volume mounts to deploy together with the k8s cluster-autoscaler.
	VolumeMounts() []k8s.VolumeMount
	// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cluster-autoscaler.
	Env() []k8s.EnvVar
	// Supported is used to determine if cluster autoscaler is implemented for this cloud provider.
	Supported() bool
}

type stubProviderMetadata struct {
	GetLoadBalancerIPErr  error
	GetLoadBalancerIPResp string

	GetSubnetworkCIDRErr  error
	GetSubnetworkCIDRResp string

	ListErr  error
	ListResp []metadata.InstanceMetadata

	SelfErr  error
	SelfResp metadata.InstanceMetadata

	GetInstanceErr  error
	GetInstanceResp metadata.InstanceMetadata

	SupportedResp            bool
	SupportsLoadBalancerResp bool
}

func (m *stubProviderMetadata) GetLoadBalancerIP(ctx context.Context) (string, error) {
	return m.GetLoadBalancerIPResp, m.GetLoadBalancerIPErr
}

func (m *stubProviderMetadata) GetSubnetworkCIDR(ctx context.Context) (string, error) {
	return m.GetSubnetworkCIDRResp, m.GetSubnetworkCIDRErr
}

func (m *stubProviderMetadata) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	return m.ListResp, m.ListErr
}

func (m *stubProviderMetadata) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	return m.SelfResp, m.SelfErr
}

func (m *stubProviderMetadata) GetInstance(ctx context.Context, providerID string) (metadata.InstanceMetadata, error) {
	return m.GetInstanceResp, m.GetInstanceErr
}

func (m *stubProviderMetadata) Supported() bool {
	return m.SupportedResp
}

func (m *stubProviderMetadata) SupportsLoadBalancer() bool {
	return m.SupportsLoadBalancerResp
}

type stubCloudControllerManager struct {
	SupportedResp bool
}

func (m *stubCloudControllerManager) Image() string {
	return "stub-image:latest"
}

func (m *stubCloudControllerManager) Path() string {
	return "/stub-controller-manager"
}

func (m *stubCloudControllerManager) Name() string {
	return "stub"
}

func (m *stubCloudControllerManager) ExtraArgs() []string {
	return []string{}
}

func (m *stubCloudControllerManager) ConfigMaps(instance metadata.InstanceMetadata) (resources.ConfigMaps, error) {
	return []*k8s.ConfigMap{}, nil
}

func (m *stubCloudControllerManager) Secrets(ctx context.Context, instance, cloudServiceAccountURI string) (resources.Secrets, error) {
	return []*k8s.Secret{}, nil
}

func (m *stubCloudControllerManager) Volumes() []k8s.Volume {
	return []k8s.Volume{}
}

func (m *stubCloudControllerManager) VolumeMounts() []k8s.VolumeMount {
	return []k8s.VolumeMount{}
}

func (m *stubCloudControllerManager) Env() []k8s.EnvVar {
	return []k8s.EnvVar{}
}

func (m *stubCloudControllerManager) Supported() bool {
	return m.SupportedResp
}

type stubCloudNodeManager struct {
	SupportedResp bool

	ImageResp     string
	PathResp      string
	ExtraArgsResp []string
}

func (m *stubCloudNodeManager) Image() string {
	return m.ImageResp
}

func (m *stubCloudNodeManager) Path() string {
	return m.PathResp
}

func (m *stubCloudNodeManager) ExtraArgs() []string {
	return m.ExtraArgsResp
}

func (m *stubCloudNodeManager) Supported() bool {
	return m.SupportedResp
}

type stubClusterAutoscaler struct {
	SupportedResp bool
}

func (a *stubClusterAutoscaler) Name() string {
	return "stub"
}

// Secrets returns a list of secrets to deploy together with the k8s cluster-autoscaler.
func (a *stubClusterAutoscaler) Secrets(instance, cloudServiceAccountURI string) (resources.Secrets, error) {
	return resources.Secrets{}, nil
}

// Volumes returns a list of volumes to deploy together with the k8s cluster-autoscaler.
func (a *stubClusterAutoscaler) Volumes() []k8s.Volume {
	return []k8s.Volume{}
}

// VolumeMounts returns a list of volume mounts to deploy together with the k8s cluster-autoscaler.
func (a *stubClusterAutoscaler) VolumeMounts() []k8s.VolumeMount {
	return []k8s.VolumeMount{}
}

// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cluster-autoscaler.
func (a *stubClusterAutoscaler) Env() []k8s.EnvVar {
	return []k8s.EnvVar{}
}

func (a *stubClusterAutoscaler) Supported() bool {
	return a.SupportedResp
}

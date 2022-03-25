package core

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/coordinator/role"
	k8s "k8s.io/api/core/v1"
)

var ErrUnimplemented = errors.New("unimplemented")

const (
	ConstellationUIDMetadataKey = "constellation-uid"
	coordinatorPort             = "9000"
	RoleMetadataKey             = "constellation-role"
	VPNIPMetadataKey            = "constellation-vpn-ip"
)

// Instance describes a cloud-provider instance including name, providerID, ip addresses and instance metadata.
type Instance struct {
	Name       string
	ProviderID string
	Role       role.Role
	IPs        []string
	// SSHKeys maps usernames to ssh public keys.
	SSHKeys map[string][]string
}

// ProviderMetadata implementers read/write cloud provider metadata.
type ProviderMetadata interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]Instance, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (Instance, error)
	// GetInstance retrieves an instance using its providerID.
	GetInstance(ctx context.Context, providerID string) (Instance, error)
	// SignalRole signals the constellation role via cloud provider metadata (if supported by the CSP and deployment type, otherwise does nothing).
	SignalRole(ctx context.Context, role role.Role) error
	// SetVPNIP stores the internally used VPN IP in cloud provider metadata (if supported and required for autoscaling by the CSP, otherwise does nothing).
	SetVPNIP(ctx context.Context, vpnIP string) error
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
	ConfigMaps(instance Instance) (resources.ConfigMaps, error)
	// Secrets returns a list of secrets to deploy together with the k8s cloud-controller-manager.
	// Reference: https://kubernetes.io/docs/concepts/configuration/secret/ .
	Secrets(instance Instance, cloudServiceAccountURI string) (resources.Secrets, error)
	// Volumes returns a list of volumes to deploy together with the k8s cloud-controller-manager.
	// Reference: https://kubernetes.io/docs/concepts/storage/volumes/ .
	Volumes() []k8s.Volume
	// VolumeMounts a list of of volume mounts to deploy together with the k8s cloud-controller-manager.
	VolumeMounts() []k8s.VolumeMount
	// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cloud-controller-manager.
	Env() []k8s.EnvVar
	// PrepareInstance is called on every instance before deploying the cloud-controller-manager.
	// Allows for cloud-provider specific hooks.
	PrepareInstance(instance Instance, vpnIP string) error
	// Supported is used to determine if cloud controller manager is implemented for this cloud provider.
	Supported() bool
}

// ClusterAutoscaler implementers provide configuration for the k8s cluster-autoscaler.
type ClusterAutoscaler interface {
	// Name returns the cloud-provider name as used by k8s cluster-autoscaler.
	Name() string
	// Supported is used to determine if cluster autoscaler is implemented for this cloud provider.
	Supported() bool
}

// CoordinatorEndpoints retrieves a list of constellation coordinator endpoint candidates from the cloud provider API.
func CoordinatorEndpoints(ctx context.Context, metadata ProviderMetadata) ([]string, error) {
	if !metadata.Supported() {
		return nil, errors.New("retrieving instances list from cloud provider is not yet supported")
	}
	instances, err := metadata.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances list from cloud provider failed: %w", err)
	}
	coordinatorEndpoints := []string{}
	for _, instance := range instances {
		// check if role of instance is "Coordinator"
		if instance.Role == role.Coordinator {
			for _, ip := range instance.IPs {
				coordinatorEndpoints = append(coordinatorEndpoints, net.JoinHostPort(ip, coordinatorPort))
			}
		}
	}

	return coordinatorEndpoints, nil
}

// PrepareInstanceForCCM sets the vpn IP in cloud provider metadata.
func PrepareInstanceForCCM(ctx context.Context, metadata ProviderMetadata, cloudControllerManager CloudControllerManager, vpnIP string) error {
	if err := metadata.SetVPNIP(ctx, vpnIP); err != nil {
		return fmt.Errorf("setting VPN IP for cloud-controller-manager failed: %w", err)
	}
	instance, err := metadata.Self(ctx)
	if err != nil {
		return fmt.Errorf("retrieving instance metadata for cloud-controller-manager failed: %w", err)
	}

	return cloudControllerManager.PrepareInstance(instance, vpnIP)
}

type ProviderMetadataFake struct{}

func (f *ProviderMetadataFake) List(ctx context.Context) ([]Instance, error) {
	self, err := f.Self(ctx)
	return []Instance{self}, err
}

func (f *ProviderMetadataFake) Self(ctx context.Context) (Instance, error) {
	return Instance{
		Name:       "instanceName",
		ProviderID: "fake://instance-id",
		Role:       role.Unknown,
		IPs:        []string{"192.0.2.1"},
	}, nil
}

func (f *ProviderMetadataFake) GetInstance(ctx context.Context, providerID string) (Instance, error) {
	return Instance{
		Name:       "instanceName",
		ProviderID: providerID,
		Role:       role.Unknown,
		IPs:        []string{"192.0.2.1"},
	}, nil
}

func (f *ProviderMetadataFake) SignalRole(ctx context.Context, role role.Role) error {
	return nil
}

func (f *ProviderMetadataFake) SetVPNIP(ctx context.Context, vpnIP string) error {
	return nil
}

func (f *ProviderMetadataFake) Supported() bool {
	return true
}

type CloudControllerManagerFake struct{}

func (f *CloudControllerManagerFake) Image() string {
	return "fake-image:latest"
}

func (f *CloudControllerManagerFake) Path() string {
	return "/fake-controller-manager"
}

func (f *CloudControllerManagerFake) Name() string {
	return "fake"
}

func (f *CloudControllerManagerFake) ExtraArgs() []string {
	return []string{}
}

func (f *CloudControllerManagerFake) ConfigMaps(instance Instance) (resources.ConfigMaps, error) {
	return []*k8s.ConfigMap{}, nil
}

func (f *CloudControllerManagerFake) Secrets(instance Instance, cloudServiceAccountURI string) (resources.Secrets, error) {
	return []*k8s.Secret{}, nil
}

func (f *CloudControllerManagerFake) Volumes() []k8s.Volume {
	return []k8s.Volume{}
}

func (f *CloudControllerManagerFake) VolumeMounts() []k8s.VolumeMount {
	return []k8s.VolumeMount{}
}

func (f *CloudControllerManagerFake) Env() []k8s.EnvVar {
	return []k8s.EnvVar{}
}

func (f *CloudControllerManagerFake) PrepareInstance(instance Instance, vpnIP string) error {
	return nil
}

func (f *CloudControllerManagerFake) Supported() bool {
	return false
}

type ClusterAutoscalerFake struct{}

func (f *ClusterAutoscalerFake) Name() string {
	return "fake"
}

func (f *ClusterAutoscalerFake) Supported() bool {
	return false
}

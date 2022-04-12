package kubernetes

import (
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	k8s "k8s.io/api/core/v1"
)

// InitClusterInput collects the arguments to initialize a new cluster.
type InitClusterInput struct {
	APIServerAdvertiseIP               string
	NodeIP                             string
	NodeName                           string
	ProviderID                         string
	SupportClusterAutoscaler           bool
	AutoscalingCloudprovider           string
	AutoscalingNodeGroups              []string
	AutoscalingSecrets                 resources.Secrets
	AutoscalingVolumes                 []k8s.Volume
	AutoscalingVolumeMounts            []k8s.VolumeMount
	AutoscalingEnv                     []k8s.EnvVar
	SupportsCloudControllerManager     bool
	CloudControllerManagerName         string
	CloudControllerManagerImage        string
	CloudControllerManagerPath         string
	CloudControllerManagerExtraArgs    []string
	CloudControllerManagerConfigMaps   resources.ConfigMaps
	CloudControllerManagerSecrets      resources.Secrets
	CloudControllerManagerVolumes      []k8s.Volume
	CloudControllerManagerVolumeMounts []k8s.VolumeMount
	CloudControllerManagerEnv          []k8s.EnvVar
	SupportsCloudNodeManager           bool
	CloudNodeManagerImage              string
	CloudNodeManagerPath               string
	CloudNodeManagerExtraArgs          []string
	MasterSecret                       []byte
}

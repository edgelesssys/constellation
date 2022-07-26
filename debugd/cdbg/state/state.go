package state

import (
	"errors"
	"strings"

	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
)

// Code in this file is mostly copied from constellation-controlPlane
// TODO: import as package from controlPlane once it is properly refactored

func GetScalingGroupsFromConfig(stat state.ConstellationState, config *config.Config) (controlPlanes, workers cloudtypes.ScalingGroup, err error) {
	switch {
	case len(stat.GCPControlPlanes) != 0:
		return getGCPInstances(stat, config)
	case len(stat.AzureControlPlane) != 0:
		return getAzureInstances(stat, config)
	case len(stat.QEMUControlPlane) != 0:
		return getQEMUInstances(stat, config)
	default:
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no instances to init")
	}
}

func getGCPInstances(stat state.ConstellationState, _ *config.Config) (controlPlanes, workers cloudtypes.ScalingGroup, err error) {
	if len(stat.GCPControlPlanes) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no control-plane workers available, can't create Constellation without any instance")
	}

	// GroupID of controlPlanes is empty, since they currently do not scale.
	controlPlanes = cloudtypes.ScalingGroup{Instances: stat.GCPControlPlanes}

	if len(stat.GCPWorkers) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no worker workers available, can't create Constellation with one instance")
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	workers = cloudtypes.ScalingGroup{Instances: stat.GCPWorkers}

	return
}

func getAzureInstances(stat state.ConstellationState, _ *config.Config) (controlPlanes, workers cloudtypes.ScalingGroup, err error) {
	if len(stat.AzureControlPlane) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no control-plane workers available, can't create Constellation cluster without any instance")
	}

	// GroupID of controlPlanes is empty, since they currently do not scale.
	controlPlanes = cloudtypes.ScalingGroup{Instances: stat.AzureControlPlane}

	if len(stat.AzureWorkers) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no worker workers available, can't create Constellation cluster with one instance")
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	workers = cloudtypes.ScalingGroup{Instances: stat.AzureWorkers}
	return
}

func getQEMUInstances(stat state.ConstellationState, _ *config.Config) (controlPlanes, workers cloudtypes.ScalingGroup, err error) {
	controlPlaneMap := stat.QEMUControlPlane
	if len(controlPlaneMap) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no controlPlanes available, can't create Constellation without any instance")
	}

	// QEMU does not support autoscaling
	controlPlanes = cloudtypes.ScalingGroup{Instances: stat.QEMUControlPlane}

	if len(stat.QEMUWorkers) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no workers available, can't create Constellation with one instance")
	}

	// QEMU does not support autoscaling
	workers = cloudtypes.ScalingGroup{Instances: stat.QEMUWorkers}
	return
}

// ImageNameContainsDebug check wether the image name in config contains "debug".
func ImageNameContainsDebug(config *config.Config) bool {
	switch {
	case config.Provider.GCP != nil:
		return strings.Contains(config.Provider.GCP.Image, "debug")
	case config.Provider.Azure != nil:
		return strings.Contains(config.Provider.Azure.Image, "debug")
	default:
		return false
	}
}

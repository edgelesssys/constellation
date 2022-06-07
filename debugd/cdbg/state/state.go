package state

import (
	"errors"

	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
)

// Code in this file is mostly copied from constellation-coordinator
// TODO: import as package from coordinator once it is properly refactored

func GetScalingGroupsFromConfig(stat state.ConstellationState, config *config.Config) (coordinators, nodes cloudtypes.ScalingGroup, err error) {
	switch {
	case len(stat.GCPCoordinators) != 0:
		return getGCPInstances(stat, config)
	case len(stat.AzureCoordinators) != 0:
		return getAzureInstances(stat, config)
	case len(stat.QEMUCoordinators) != 0:
		return getQEMUInstances(stat, config)
	default:
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no instances to init")
	}
}

func getGCPInstances(stat state.ConstellationState, config *config.Config) (coordinators, nodes cloudtypes.ScalingGroup, err error) {
	if len(stat.GCPCoordinators) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no control-plane nodes available, can't create Constellation without any instance")
	}

	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = cloudtypes.ScalingGroup{Instances: stat.GCPCoordinators}

	if len(stat.GCPNodes) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no worker nodes available, can't create Constellation with one instance")
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	nodes = cloudtypes.ScalingGroup{Instances: stat.GCPNodes}

	return
}

func getAzureInstances(stat state.ConstellationState, _ *config.Config) (coordinators, nodes cloudtypes.ScalingGroup, err error) {
	if len(stat.AzureCoordinators) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no control-plane nodes available, can't create Constellation cluster without any instance")
	}

	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = cloudtypes.ScalingGroup{Instances: stat.AzureCoordinators}

	if len(stat.AzureNodes) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no worker nodes available, can't create Constellation cluster with one instance")
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	nodes = cloudtypes.ScalingGroup{Instances: stat.AzureNodes}
	return
}

func getQEMUInstances(stat state.ConstellationState, config *config.Config) (coordinators, nodes cloudtypes.ScalingGroup, err error) {
	coordinatorMap := stat.QEMUCoordinators
	if len(coordinatorMap) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no coordinators available, can't create Constellation without any instance")
	}

	// QEMU does not support autoscaling
	coordinators = cloudtypes.ScalingGroup{Instances: stat.QEMUCoordinators}

	if len(stat.QEMUNodes) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no nodes available, can't create Constellation with one instance")
	}

	// QEMU does not support autoscaling
	nodes = cloudtypes.ScalingGroup{Instances: stat.QEMUNodes}
	return
}

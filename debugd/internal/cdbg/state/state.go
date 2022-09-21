/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package state

import (
	"errors"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/state"
)

// Code in this file is mostly copied from constellation-controlPlane
// TODO: import as package from controlPlane once it is properly refactored

func GetScalingGroupsFromConfig(stat state.ConstellationState, config *config.Config) (controlPlanes, workers cloudtypes.ScalingGroup, err error) {
	switch {
	case len(stat.GCPControlPlaneInstances) != 0:
		return getGCPInstances(stat, config)
	case len(stat.AzureControlPlaneInstances) != 0:
		return getAzureInstances(stat, config)
	case len(stat.QEMUControlPlaneInstances) != 0:
		return getQEMUInstances(stat, config)
	default:
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no instances to init")
	}
}

func getGCPInstances(stat state.ConstellationState, _ *config.Config) (controlPlanes, workers cloudtypes.ScalingGroup, err error) {
	if len(stat.GCPControlPlaneInstances) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no control-plane workers available, can't create Constellation without any instance")
	}

	// GroupID of controlPlanes is empty, since they currently do not scale.
	controlPlanes = cloudtypes.ScalingGroup{Instances: stat.GCPControlPlaneInstances}

	if len(stat.GCPWorkerInstances) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no worker workers available, can't create Constellation with one instance")
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	workers = cloudtypes.ScalingGroup{Instances: stat.GCPWorkerInstances}

	return
}

func getAzureInstances(stat state.ConstellationState, _ *config.Config) (controlPlanes, workers cloudtypes.ScalingGroup, err error) {
	if len(stat.AzureControlPlaneInstances) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no control-plane workers available, can't create Constellation cluster without any instance")
	}

	// GroupID of controlPlanes is empty, since they currently do not scale.
	controlPlanes = cloudtypes.ScalingGroup{Instances: stat.AzureControlPlaneInstances}

	if len(stat.AzureWorkerInstances) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no worker workers available, can't create Constellation cluster with one instance")
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	workers = cloudtypes.ScalingGroup{Instances: stat.AzureWorkerInstances}
	return
}

func getQEMUInstances(stat state.ConstellationState, _ *config.Config) (controlPlanes, workers cloudtypes.ScalingGroup, err error) {
	controlPlaneMap := stat.QEMUControlPlaneInstances
	if len(controlPlaneMap) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no controlPlanes available, can't create Constellation without any instance")
	}

	// QEMU does not support autoscaling
	controlPlanes = cloudtypes.ScalingGroup{Instances: stat.QEMUControlPlaneInstances}

	if len(stat.QEMUWorkerInstances) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no workers available, can't create Constellation with one instance")
	}

	// QEMU does not support autoscaling
	workers = cloudtypes.ScalingGroup{Instances: stat.QEMUWorkerInstances}
	return
}

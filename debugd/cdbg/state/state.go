package state

import (
	"errors"

	cmdc "github.com/edgelesssys/constellation/cli/cmd"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
)

// Code in this file is mostly copied from constellation-coordinator
// TODO: import as package from coordinator once it is properly refactored

func GetScalingGroupsFromConfig(stat state.ConstellationState, config *config.Config) (coordinators, nodes cmdc.ScalingGroup, err error) {
	switch {
	case len(stat.GCPCoordinators) != 0:
		return getGCPInstances(stat, config)
	case len(stat.AzureCoordinators) != 0:
		return getAzureInstances(stat, config)
	case len(stat.QEMUCoordinators) != 0:
		return getQEMUInstances(stat, config)
	default:
		return cmdc.ScalingGroup{}, cmdc.ScalingGroup{}, errors.New("no instances to init")
	}
}

func getGCPInstances(stat state.ConstellationState, config *config.Config) (coordinators, nodes cmdc.ScalingGroup, err error) {
	coordinatorMap := stat.GCPCoordinators
	if len(coordinatorMap) == 0 {
		return cmdc.ScalingGroup{}, cmdc.ScalingGroup{}, errors.New("no coordinators available, can't create Constellation without any instance")
	}
	var coordinatorInstances cmdc.Instances
	for _, node := range coordinatorMap {
		coordinatorInstances = append(coordinatorInstances, cmdc.Instance(node))
	}
	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = cmdc.ScalingGroup{
		Instances: coordinatorInstances,
		GroupID:   "",
	}

	nodeMap := stat.GCPNodes
	if len(nodeMap) == 0 {
		return cmdc.ScalingGroup{}, cmdc.ScalingGroup{}, errors.New("no nodes available, can't create Constellation with one instance")
	}

	var nodeInstances cmdc.Instances
	for _, node := range nodeMap {
		nodeInstances = append(nodeInstances, cmdc.Instance(node))
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	nodes = cmdc.ScalingGroup{Instances: nodeInstances}

	return
}

func getAzureInstances(stat state.ConstellationState, _ *config.Config) (coordinators, nodes cmdc.ScalingGroup, err error) {
	coordinatorMap := stat.AzureCoordinators
	if len(coordinatorMap) == 0 {
		return cmdc.ScalingGroup{}, cmdc.ScalingGroup{}, errors.New("no coordinators available, can't create Constellation without any instance")
	}
	var coordinatorInstances cmdc.Instances
	for _, node := range coordinatorMap {
		coordinatorInstances = append(coordinatorInstances, cmdc.Instance(node))
	}
	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = cmdc.ScalingGroup{
		Instances: coordinatorInstances,
		GroupID:   "",
	}
	nodeMap := stat.AzureNodes
	if len(nodeMap) == 0 {
		return cmdc.ScalingGroup{}, cmdc.ScalingGroup{}, errors.New("no nodes available, can't create Constellation with one instance")
	}

	var nodeInstances cmdc.Instances
	for _, node := range nodeMap {
		nodeInstances = append(nodeInstances, cmdc.Instance(node))
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	nodes = cmdc.ScalingGroup{Instances: nodeInstances}
	return
}

func getQEMUInstances(stat state.ConstellationState, _ *config.Config) (coordinators, nodes cmdc.ScalingGroup, err error) {
	coordinatorMap := stat.QEMUCoordinators
	if len(coordinatorMap) == 0 {
		return cmdc.ScalingGroup{}, cmdc.ScalingGroup{}, errors.New("no coordinators available, can't create Constellation without any instance")
	}
	var coordinatorInstances cmdc.Instances
	for _, node := range coordinatorMap {
		coordinatorInstances = append(coordinatorInstances, cmdc.Instance(node))
	}
	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = cmdc.ScalingGroup{
		Instances: coordinatorInstances,
		GroupID:   "",
	}
	nodeMap := stat.QEMUNodes
	if len(nodeMap) == 0 {
		return cmdc.ScalingGroup{}, cmdc.ScalingGroup{}, errors.New("no nodes available, can't create Constellation with one instance")
	}
	var nodeInstances cmdc.Instances
	for _, node := range nodeMap {
		nodeInstances = append(nodeInstances, cmdc.Instance(node))
	}
	nodes = cmdc.ScalingGroup{Instances: nodeInstances}
	return
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constants

const (
	// AutoscalingStrategyResourceName resource name used for AutoscalingStrategy.
	AutoscalingStrategyResourceName = "autoscalingstrategy"
	// ControlPlaneScalingGroupResourceName resource name used for ControlPlaneScalingGroup.
	ControlPlaneScalingGroupResourceName = "scalinggroup-controlplane"
	// WorkerScalingGroupResourceName resource name used for WorkerScaling.
	WorkerScalingGroupResourceName = "scalinggroup-worker"
	// PlaceholderImageName name of the OS image used if upgrades are not yet supported.
	PlaceholderImageName = "unsupportedCSP"
	// PlaceholderControlPlaneScalingGroupName name of the control plane scaling group used if upgrades are not yet supported.
	PlaceholderControlPlaneScalingGroupName = "control-planes-id"
	// PlaceholderWorkerScalingGroupName name of the worker scaling group used if upgrades are not yet supported.
	PlaceholderWorkerScalingGroupName = "workers-id"
	// ControlPlaneRoleLabel label used to identify control plane nodes.
	ControlPlaneRoleLabel = "node-role.kubernetes.io/control-plane"
)

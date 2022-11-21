/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package deploy provides functions to deploy initial resources for the node operator.
package deploy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/edgelesssys/constellation/operators/constellation-node-operator/v2/internal/constants"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InitialResources creates the initial resources for the node operator.
func InitialResources(ctx context.Context, k8sClient client.Writer, scalingGroupGetter scalingGroupGetter, uid string) error {
	controlPlaneGroupIDs, workerGroupIDs, err := scalingGroupGetter.ListScalingGroups(ctx, uid)
	if err != nil {
		return fmt.Errorf("listing scaling groups: %w", err)
	}
	if len(controlPlaneGroupIDs) == 0 {
		return errors.New("determining initial node image: no control plane scaling group found")
	}
	if len(workerGroupIDs) == 0 {
		return errors.New("determining initial node image: no worker scaling group found")
	}

	if err := createAutoscalingStrategy(ctx, k8sClient, scalingGroupGetter.AutoscalingCloudProvider()); err != nil {
		return fmt.Errorf("creating initial autoscaling strategy: %w", err)
	}
	imageReference, err := scalingGroupGetter.GetScalingGroupImage(ctx, controlPlaneGroupIDs[0])
	if err != nil {
		return fmt.Errorf("determining initial node image: %w", err)
	}
	if err := createNodeImage(ctx, k8sClient, imageReference); err != nil {
		return fmt.Errorf("creating initial node image %q: %w", imageReference, err)
	}
	for _, groupID := range controlPlaneGroupIDs {
		groupName, err := scalingGroupGetter.GetScalingGroupName(groupID)
		if err != nil {
			return fmt.Errorf("determining scaling group name of %q: %w", groupID, err)
		}
		autoscalingGroupName, err := scalingGroupGetter.GetAutoscalingGroupName(groupID)
		if err != nil {
			return fmt.Errorf("determining autoscaling group name of %q: %w", groupID, err)
		}
		newScalingGroupConfig := newScalingGroupConfig{k8sClient, groupID, groupName, autoscalingGroupName, updatev1alpha1.ControlPlaneRole}
		if err := createScalingGroup(ctx, newScalingGroupConfig); err != nil {
			return fmt.Errorf("creating initial control plane scaling group: %w", err)
		}
	}
	for _, groupID := range workerGroupIDs {
		groupName, err := scalingGroupGetter.GetScalingGroupName(groupID)
		if err != nil {
			return fmt.Errorf("determining scaling group name of %q: %w", groupID, err)
		}
		autoscalingGroupName, err := scalingGroupGetter.GetAutoscalingGroupName(groupID)
		if err != nil {
			return fmt.Errorf("determining autoscaling group name of %q: %w", groupID, err)
		}
		newScalingGroupConfig := newScalingGroupConfig{k8sClient, groupID, groupName, autoscalingGroupName, updatev1alpha1.WorkerRole}
		if err := createScalingGroup(ctx, newScalingGroupConfig); err != nil {
			return fmt.Errorf("creating initial worker scaling group: %w", err)
		}
	}
	return nil
}

// createAutoscalingStrategy creates the autoscaling strategy resource if it does not exist yet.
func createAutoscalingStrategy(ctx context.Context, k8sClient client.Writer, provider string) error {
	err := k8sClient.Create(ctx, &updatev1alpha1.AutoscalingStrategy{
		TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "AutoscalingStrategy"},
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.AutoscalingStrategyResourceName,
		},
		Spec: updatev1alpha1.AutoscalingStrategySpec{
			Enabled:             true,
			DeploymentName:      "constellation-cluster-autoscaler",
			DeploymentNamespace: "kube-system",
			AutoscalerExtraArgs: map[string]string{
				"cloud-provider":  provider,
				"logtostderr":     "true",
				"stderrthreshold": "info",
				"v":               "2",
				"namespace":       "kube-system",
			},
		},
	})
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// createNodeImage creates the initial nodeimage resource if it does not exist yet.
func createNodeImage(ctx context.Context, k8sClient client.Writer, imageReference string) error {
	err := k8sClient.Create(ctx, &updatev1alpha1.NodeImage{
		TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "NodeImage"},
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.NodeImageResourceName,
		},
		Spec: updatev1alpha1.NodeImageSpec{
			ImageReference: imageReference,
		},
	})
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// createScalingGroup creates an initial scaling group resource if it does not exist yet.
func createScalingGroup(ctx context.Context, config newScalingGroupConfig) error {
	err := config.k8sClient.Create(ctx, &updatev1alpha1.ScalingGroup{
		TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(config.groupName),
		},
		Spec: updatev1alpha1.ScalingGroupSpec{
			NodeImage:           constants.NodeImageResourceName,
			GroupID:             config.groupID,
			AutoscalerGroupName: config.autoscalingGroupName,
			Min:                 1,
			Max:                 10,
			Role:                config.role,
		},
	})
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

type scalingGroupGetter interface {
	// GetScalingGroupImage retrieves the image currently used by a scaling group.
	GetScalingGroupImage(ctx context.Context, scalingGroupID string) (string, error)
	// GetScalingGroupName retrieves the name of a scaling group.
	GetScalingGroupName(scalingGroupID string) (string, error)
	// GetScalingGroupName retrieves the name of a scaling group as needed by the cluster-autoscaler.
	GetAutoscalingGroupName(scalingGroupID string) (string, error)
	// ListScalingGroups retrieves a list of scaling groups for the cluster.
	ListScalingGroups(ctx context.Context, uid string) (controlPlaneGroupIDs []string, workerGroupIDs []string, err error)
	// AutoscalingCloudProvider returns the cloud-provider name as used by k8s cluster-autoscaler.
	AutoscalingCloudProvider() string
}

type newScalingGroupConfig struct {
	k8sClient            client.Writer
	groupID              string
	groupName            string
	autoscalingGroupName string
	role                 updatev1alpha1.NodeRole
}

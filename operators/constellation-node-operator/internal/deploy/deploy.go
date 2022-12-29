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
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// InitialResources creates the initial resources for the node operator.
func InitialResources(ctx context.Context, k8sClient client.Client, imageInfo imageInfoGetter, scalingGroupGetter scalingGroupGetter, uid string) error {
	logr := log.FromContext(ctx)
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
	imageVersion, err := imageInfo.ImageVersion(imageReference)
	if err != nil {
		// do not fail if the image version cannot be determined
		// this is important for backwards compatibility
		logr.Error(err, "determining initial node image version")
		imageVersion = ""
	}

	if err := createNodeVersion(ctx, k8sClient, imageReference, imageVersion); err != nil {
		return fmt.Errorf("creating initial node version %q: %w", imageReference, err)
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

// createNodeVersion creates the initial nodeversion resource if it does not exist yet.
func createNodeVersion(ctx context.Context, k8sClient client.Client, imageReference, imageVersion string) error {
	err := k8sClient.Create(ctx, &updatev1alpha1.NodeVersion{
		TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "NodeVersion"},
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.NodeVersionResourceName,
		},
		Spec: updatev1alpha1.NodeVersionSpec{
			ImageReference: imageReference,
			ImageVersion:   imageVersion,
		},
	})
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	} else if err != nil {
		return err
	}

	// Since the nodeversion resource was just created, we need to set the k8s-components reference.
	// We update the nodeversion after creation since the `findK8sComponentsConfigMap` function
	// requires the exactly one k8s-components configmap to exist and we want to support the operator
	// (re-)starting while updating the k8s components. The admin should be able to leave
	// the old k8s-components configmap in place in case they want to revert the update.
	k8sComponentsRef, err := findK8sComponentsConfigMap(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("finding k8s-components configmap: %w", err)
	}
	// add k8s-components reference to nodeversion if not set already
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var nodeVersion updatev1alpha1.NodeVersion
		if err := k8sClient.Get(ctx, client.ObjectKey{Name: constants.NodeVersionResourceName}, &nodeVersion); err != nil {
			return err
		}
		if nodeVersion.Spec.KubernetesComponentsReference == "" {
			nodeVersion.Spec.KubernetesComponentsReference = k8sComponentsRef
			return k8sClient.Update(ctx, &nodeVersion)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("updating nodeversion: %w", err)
	}

	return nil
}

// findK8sComponentsConfigMap finds the k8s-components configmap in the kube-system namespace.
// It returns an error if there is no or multiple configmaps matching the prefix "k8s-components".
func findK8sComponentsConfigMap(ctx context.Context, k8sClient client.Client) (string, error) {
	var configMaps corev1.ConfigMapList
	err := k8sClient.List(ctx, &configMaps, client.InNamespace("kube-system"))
	if err != nil {
		return "", fmt.Errorf("listing configmaps: %w", err)
	}
	names := []string{}
	for _, configMap := range configMaps.Items {
		if strings.HasPrefix(configMap.Name, "k8s-components") {
			names = append(names, configMap.Name)
		}
	}
	if len(names) == 1 {
		return names[0], nil
	} else if len(names) == 0 {
		return "", fmt.Errorf("no configmaps found")
	}

	return "", fmt.Errorf("multiple configmaps found")
}

// createScalingGroup creates an initial scaling group resource if it does not exist yet.
func createScalingGroup(ctx context.Context, config newScalingGroupConfig) error {
	err := config.k8sClient.Create(ctx, &updatev1alpha1.ScalingGroup{
		TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(config.groupName),
		},
		Spec: updatev1alpha1.ScalingGroupSpec{
			NodeVersion:         constants.NodeVersionResourceName,
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

type imageInfoGetter interface {
	ImageVersion(imageReference string) (string, error)
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

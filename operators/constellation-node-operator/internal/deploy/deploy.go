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

	mainconstants "github.com/edgelesssys/constellation/v2/internal/constants"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/constants"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	imageVersion, err := imageInfo.ImageVersion()
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
	latestComponentCM, err := findLatestK8sComponentsConfigMap(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("finding latest k8s-components configmap: %w", err)
	}
	err = k8sClient.Create(ctx, &updatev1alpha1.NodeVersion{
		TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "NodeVersion"},
		ObjectMeta: metav1.ObjectMeta{
			Name: mainconstants.NodeVersionResourceName,
		},
		Spec: updatev1alpha1.NodeVersionSpec{
			ImageReference:                imageReference,
			ImageVersion:                  imageVersion,
			KubernetesComponentsReference: latestComponentCM.Name,
			KubernetesClusterVersion:      latestComponentCM.Data[mainconstants.K8sVersionFieldName],
		},
	})
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

// findLatestK8sComponentsConfigMap finds most recently created k8s-components configmap in the kube-system namespace.
// It returns an error if there is no or multiple configmaps matching the prefix "k8s-components".
func findLatestK8sComponentsConfigMap(ctx context.Context, k8sClient client.Client) (corev1.ConfigMap, error) {
	var configMaps corev1.ConfigMapList
	err := k8sClient.List(ctx, &configMaps, client.InNamespace("kube-system"))
	if err != nil {
		return corev1.ConfigMap{}, fmt.Errorf("listing configmaps: %w", err)
	}

	// collect all k8s-components configmaps
	componentConfigMaps := []corev1.ConfigMap{}
	for _, configMap := range configMaps.Items {
		if strings.HasPrefix(configMap.Name, "k8s-components") {
			componentConfigMaps = append(componentConfigMaps, configMap)
		}
	}
	if len(componentConfigMaps) == 0 {
		return corev1.ConfigMap{}, fmt.Errorf("no configmaps found")
	}

	// find latest configmap
	var latestConfigMap corev1.ConfigMap
	for _, cm := range componentConfigMaps {
		if cm.CreationTimestamp.After(latestConfigMap.CreationTimestamp.Time) {
			latestConfigMap = cm
		}
	}
	return latestConfigMap, nil
}

// createScalingGroup creates an initial scaling group resource if it does not exist yet.
func createScalingGroup(ctx context.Context, config newScalingGroupConfig) error {
	err := config.k8sClient.Create(ctx, &updatev1alpha1.ScalingGroup{
		TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(config.groupName),
		},
		Spec: updatev1alpha1.ScalingGroupSpec{
			NodeVersion:         mainconstants.NodeVersionResourceName,
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
	ImageVersion() (string, error)
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

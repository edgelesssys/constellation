/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Upgrader handles upgrading the cluster's components using the CLI.
type Upgrader struct {
	stableInterface  stableInterface
	dynamicInterface dynamicInterface
	helmClient       helmInterface

	outWriter io.Writer
}

// NewUpgrader returns a new Upgrader.
func NewUpgrader(outWriter io.Writer, log debugLog) (*Upgrader, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", constants.AdminConfFilename)
	if err != nil {
		return nil, fmt.Errorf("building kubernetes config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("setting up kubernetes client: %w", err)
	}

	// use unstructured client to avoid importing the operator packages
	unstructuredClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("setting up custom resource client: %w", err)
	}

	helmClient, err := helm.NewClient(kubectl.New(), constants.AdminConfFilename, constants.HelmNamespace, log)
	if err != nil {
		return nil, fmt.Errorf("setting up helm client: %w", err)
	}

	return &Upgrader{
		stableInterface:  &stableClient{client: kubeClient},
		dynamicInterface: &dynamicClient{client: unstructuredClient},
		helmClient:       helmClient,
		outWriter:        outWriter,
	}, nil
}

// Upgrade upgrades the cluster to the given measurements and image.
func (u *Upgrader) Upgrade(ctx context.Context, imageReference, imageVersion string, measurements measurements.M) error {
	if err := u.updateMeasurements(ctx, measurements); err != nil {
		return fmt.Errorf("updating measurements: %w", err)
	}

	if err := u.updateImage(ctx, imageReference, imageVersion); err != nil {
		return fmt.Errorf("updating image: %w", err)
	}
	return nil
}

// GetCurrentImage returns the currently used image version of the cluster.
func (u *Upgrader) GetCurrentImage(ctx context.Context) (*unstructured.Unstructured, string, error) {
	return u.getFromConstellationVersion(ctx, "imageVersion")
}

// GetCurrentKubernetesVersion returns the currently used Kubernetes version.
func (u *Upgrader) GetCurrentKubernetesVersion(ctx context.Context) (*unstructured.Unstructured, string, error) {
	return u.getFromConstellationVersion(ctx, "kubernetesClusterVersion")
}

// getFromConstellationVersion queries the constellation-version object for a given field.
func (u *Upgrader) getFromConstellationVersion(ctx context.Context, fieldName string) (*unstructured.Unstructured, string, error) {
	versionStruct, err := u.dynamicInterface.getCurrent(ctx, "constellation-version")
	if err != nil {
		return nil, "", err
	}

	spec, ok := versionStruct.Object["spec"]
	if !ok {
		return nil, "", errors.New("spec missing")
	}
	retErr := errors.New("invalid spec")
	specMap, ok := spec.(map[string]any)
	if !ok {
		return nil, "", retErr
	}
	fieldValue, ok := specMap[fieldName]
	if !ok {
		return nil, "", retErr
	}
	fieldValueString, ok := fieldValue.(string)
	if !ok {
		return nil, "", retErr
	}

	return versionStruct, fieldValueString, nil
}

// UpgradeHelmServices upgrade helm services.
func (u *Upgrader) UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error {
	return u.helmClient.Upgrade(ctx, config, timeout, allowDestructive)
}

// KubernetesVersion returns the version of Kubernetes the Constellation is currently running on.
func (u *Upgrader) KubernetesVersion() (string, error) {
	return u.stableInterface.kubernetesVersion()
}

func (u *Upgrader) updateMeasurements(ctx context.Context, newMeasurements measurements.M) error {
	existingConf, err := u.stableInterface.getCurrent(ctx, constants.JoinConfigMap)
	if err != nil {
		return fmt.Errorf("retrieving current measurements: %w", err)
	}

	var currentMeasurements measurements.M
	if err := json.Unmarshal([]byte(existingConf.Data[constants.MeasurementsFilename]), &currentMeasurements); err != nil {
		return fmt.Errorf("retrieving current measurements: %w", err)
	}
	if currentMeasurements.EqualTo(newMeasurements) {
		fmt.Fprintln(u.outWriter, "Cluster is already using the chosen measurements, skipping measurements upgrade")
		return nil
	}

	// don't allow potential security downgrades by setting the warnOnly flag to true
	for k, newM := range newMeasurements {
		if currentM, ok := currentMeasurements[k]; ok && !currentM.WarnOnly && newM.WarnOnly {
			return fmt.Errorf("setting enforced measurement %d to warn only: not allowed", k)
		}
	}

	// backup of previous measurements
	existingConf.Data["oldMeasurements"] = existingConf.Data[constants.MeasurementsFilename]

	measurementsJSON, err := json.Marshal(newMeasurements)
	if err != nil {
		return fmt.Errorf("marshaling measurements: %w", err)
	}
	existingConf.Data[constants.MeasurementsFilename] = string(measurementsJSON)
	_, err = u.stableInterface.update(ctx, existingConf)
	if err != nil {
		return fmt.Errorf("setting new measurements: %w", err)
	}

	fmt.Fprintln(u.outWriter, "Successfully updated the cluster's expected measurements")
	return nil
}

func (u *Upgrader) updateImage(ctx context.Context, imageReference, imageVersion string) error {
	currentImage, currentImageVersion, err := u.GetCurrentImage(ctx)
	if err != nil {
		return fmt.Errorf("retrieving current image: %w", err)
	}

	if currentImageVersion == imageVersion {
		fmt.Fprintln(u.outWriter, "Cluster is already using the chosen image, skipping image upgrade")
		return nil
	}

	currentImage.Object["spec"].(map[string]any)["image"] = imageReference
	currentImage.Object["spec"].(map[string]any)["imageVersion"] = imageVersion
	if _, err := u.dynamicInterface.update(ctx, currentImage); err != nil {
		return fmt.Errorf("setting new image: %w", err)
	}

	fmt.Fprintln(u.outWriter, "Successfully updated the cluster's image, upgrades will be applied automatically")
	return nil
}

type dynamicInterface interface {
	getCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error)
	update(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type stableInterface interface {
	getCurrent(ctx context.Context, name string) (*corev1.ConfigMap, error)
	update(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	kubernetesVersion() (string, error)
}

type dynamicClient struct {
	client dynamic.Interface
}

// getCurrent returns the current image definition.
func (u *dynamicClient) getCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	return u.client.Resource(schema.GroupVersionResource{
		Group:    "update.edgeless.systems",
		Version:  "v1alpha1",
		Resource: "nodeversions",
	}).Get(ctx, name, metav1.GetOptions{})
}

// update updates the image definition.
func (u *dynamicClient) update(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return u.client.Resource(schema.GroupVersionResource{
		Group:    "update.edgeless.systems",
		Version:  "v1alpha1",
		Resource: "nodeversions",
	}).Update(ctx, obj, metav1.UpdateOptions{})
}

type stableClient struct {
	client kubernetes.Interface
}

// getCurrent returns the cluster's expected measurements.
func (u *stableClient) getCurrent(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Get(ctx, name, metav1.GetOptions{})
}

// update updates the cluster's expected measurements in Kubernetes.
func (u *stableClient) update(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Update(ctx, configMap, metav1.UpdateOptions{})
}

func (u *stableClient) kubernetesVersion() (string, error) {
	serverVersion, err := u.client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return serverVersion.GitVersion, nil
}

type helmInterface interface {
	Upgrade(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error
}

type debugLog interface {
	Debugf(format string, args ...any)
	Sync()
}

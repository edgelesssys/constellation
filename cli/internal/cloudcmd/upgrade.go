/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/v2/internal/constants"
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
	measurementsUpdater measurementsUpdater
	imageUpdater        imageUpdater

	writer io.Writer
}

// NewUpgrader returns a new Upgrader.
func NewUpgrader(writer io.Writer) (*Upgrader, error) {
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

	return &Upgrader{
		measurementsUpdater: &kubeMeasurementsUpdater{client: kubeClient},
		imageUpdater:        &kubeImageUpdater{client: unstructuredClient},
		writer:              writer,
	}, nil
}

// Upgrade upgrades the cluster to the given measurements and image.
func (u *Upgrader) Upgrade(ctx context.Context, image string, measurements map[uint32][]byte) error {
	if err := u.updateMeasurements(ctx, measurements); err != nil {
		return fmt.Errorf("updating measurements: %w", err)
	}

	if err := u.updateImage(ctx, image); err != nil {
		return fmt.Errorf("updating image: %w", err)
	}
	return nil
}

// GetCurrentImage returns the currently used image of the cluster.
func (u *Upgrader) GetCurrentImage(ctx context.Context) (*unstructured.Unstructured, string, error) {
	imageStruct, err := u.imageUpdater.getCurrent(ctx, "constellation-coreos")
	if err != nil {
		return nil, "", err
	}

	spec, ok := imageStruct.Object["spec"]
	if !ok {
		return nil, "", errors.New("image spec missing")
	}
	retErr := errors.New("invalid image spec")
	specMap, ok := spec.(map[string]any)
	if !ok {
		return nil, "", retErr
	}
	currentImageDefinition, ok := specMap["image"]
	if !ok {
		return nil, "", retErr
	}
	imageDefinition, ok := currentImageDefinition.(string)
	if !ok {
		return nil, "", retErr
	}

	return imageStruct, imageDefinition, nil
}

func (u *Upgrader) updateMeasurements(ctx context.Context, measurements map[uint32][]byte) error {
	existingConf, err := u.measurementsUpdater.getCurrent(ctx, constants.JoinConfigMap)
	if err != nil {
		return fmt.Errorf("retrieving current measurements: %w", err)
	}

	var currentMeasurements map[uint32][]byte
	if err := json.Unmarshal([]byte(existingConf.Data[constants.MeasurementsFilename]), &currentMeasurements); err != nil {
		return fmt.Errorf("retrieving current measurements: %w", err)
	}
	if len(currentMeasurements) == len(measurements) {
		changed := false
		for k, v := range currentMeasurements {
			if !bytes.Equal(v, measurements[k]) {
				// measurements have changed
				changed = true
				break
			}
		}
		if !changed {
			// measurements are the same, nothing to be done
			fmt.Fprintln(u.writer, "Cluster is already using the chosen measurements, skipping measurements upgrade")
			return nil
		}
	}

	// backup of previous measurements
	existingConf.Data["oldMeasurements"] = existingConf.Data[constants.MeasurementsFilename]

	measurementsJSON, err := json.Marshal(measurements)
	if err != nil {
		return fmt.Errorf("marshaling measurements: %w", err)
	}
	existingConf.Data[constants.MeasurementsFilename] = string(measurementsJSON)
	_, err = u.measurementsUpdater.update(ctx, existingConf)
	if err != nil {
		return fmt.Errorf("setting new measurements: %w", err)
	}

	fmt.Fprintln(u.writer, "Successfully updated the cluster's expected measurements")
	return nil
}

func (u *Upgrader) updateImage(ctx context.Context, image string) error {
	currentImage, currentImageDefinition, err := u.GetCurrentImage(ctx)
	if err != nil {
		return fmt.Errorf("retrieving current image: %w", err)
	}

	if currentImageDefinition == image {
		fmt.Fprintln(u.writer, "Cluster is already using the chosen image, skipping image upgrade")
		return nil
	}

	currentImage.Object["spec"].(map[string]any)["image"] = image
	if _, err := u.imageUpdater.update(ctx, currentImage); err != nil {
		return fmt.Errorf("setting new image: %w", err)
	}

	fmt.Fprintln(u.writer, "Successfully updated the cluster's image, upgrades will be applied automatically")
	return nil
}

type imageUpdater interface {
	getCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error)
	update(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type measurementsUpdater interface {
	getCurrent(ctx context.Context, name string) (*corev1.ConfigMap, error)
	update(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
}

type kubeImageUpdater struct {
	client dynamic.Interface
}

// getCurrent returns the current image definition.
func (u *kubeImageUpdater) getCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	return u.client.Resource(schema.GroupVersionResource{
		Group:    "update.edgeless.systems",
		Version:  "v1alpha1",
		Resource: "nodeimages",
	}).Get(ctx, name, metav1.GetOptions{})
}

// update updates the image definition.
func (u *kubeImageUpdater) update(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return u.client.Resource(schema.GroupVersionResource{
		Group:    "update.edgeless.systems",
		Version:  "v1alpha1",
		Resource: "nodeimages",
	}).Update(ctx, obj, metav1.UpdateOptions{})
}

type kubeMeasurementsUpdater struct {
	client kubernetes.Interface
}

// getCurrent returns the cluster's expected measurements.
func (u *kubeMeasurementsUpdater) getCurrent(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Get(ctx, name, metav1.GetOptions{})
}

// update updates the cluster's expected measurements in Kubernetes.
func (u *kubeMeasurementsUpdater) update(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Update(ctx, configMap, metav1.UpdateOptions{})
}

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
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	internalk8s "github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ErrInProgress signals that an upgrade is in progress inside the cluster.
var ErrInProgress = errors.New("upgrade in progress")

// Upgrader handles upgrading the cluster's components using the CLI.
type Upgrader struct {
	stableInterface  stableInterface
	dynamicInterface dynamicInterface
	helmClient       helmInterface

	outWriter io.Writer
	log       debugLog
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
		log:              log,
	}, nil
}

// UpgradeImage upgrades the cluster to the given measurements and image.
func (u *Upgrader) UpgradeImage(ctx context.Context, newImageReference, newImageVersion string, newMeasurements measurements.M) error {
	nodeVersion, err := u.getConstellationVersion(ctx)
	if err != nil {
		return fmt.Errorf("retrieving current image: %w", err)
	}
	currentImageVersion := nodeVersion.Spec.ImageVersion

	if err := compatibility.IsValidUpgrade(currentImageVersion, newImageVersion); err != nil {
		return err
	}

	if imageUpgradeInProgress(nodeVersion) {
		return ErrInProgress
	}

	if err := u.updateMeasurements(ctx, newMeasurements); err != nil {
		return fmt.Errorf("updating measurements: %w", err)
	}

	if err := u.updateImage(ctx, nodeVersion, newImageReference, newImageVersion); err != nil {
		return fmt.Errorf("updating image: %w", err)
	}
	return nil
}

// UpgradeHelmServices upgrade helm services.
func (u *Upgrader) UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error {
	return u.helmClient.Upgrade(ctx, config, timeout, allowDestructive)
}

// UpgradeK8s upgrade the Kubernetes cluster version and the installed components to matching versions.
func (u *Upgrader) UpgradeK8s(ctx context.Context, newClusterVersion string, components components.Components) error {
	nodeVersion, err := u.getConstellationVersion(ctx)
	if err != nil {
		return fmt.Errorf("getting kubernetesClusterVersion: %w", err)
	}

	if err := compatibility.IsValidUpgrade(nodeVersion.Spec.KubernetesClusterVersion, newClusterVersion); err != nil {
		return err
	}

	if k8sUpgradeInProgress(nodeVersion) {
		return ErrInProgress
	}

	u.log.Debugf("Upgrading cluster's Kubernetes version from %s to %s", nodeVersion.Spec.KubernetesClusterVersion, newClusterVersion)
	configMap, err := internalk8s.ConstructK8sComponentsCM(components, newClusterVersion)
	if err != nil {
		return fmt.Errorf("constructing k8s-components ConfigMap: %w", err)
	}

	_, err = u.stableInterface.createConfigMap(ctx, &configMap)
	// If the map already exists we can use that map and assume it has the same content as 'configMap'.
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating k8s-components ConfigMap: %w. %T", err, err)
	}

	nodeVersion.Spec.KubernetesComponentsReference = configMap.ObjectMeta.Name
	nodeVersion.Spec.KubernetesClusterVersion = newClusterVersion

	raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
	if err != nil {
		return fmt.Errorf("converting nodeVersion to unstructured: %w", err)
	}
	u.log.Debugf("Triggering Kubernetes version upgrade now")
	// Send the updated NodeVersion resource
	updated, err := u.dynamicInterface.update(ctx, &unstructured.Unstructured{Object: raw})
	if err != nil {
		return fmt.Errorf("updating NodeVersion: %w", err)
	}

	// Verify the update worked as expected
	updatedSpec, ok := updated.Object["spec"]
	if !ok {
		return errors.New("invalid updated NodeVersion spec")
	}
	updatedMap, ok := updatedSpec.(map[string]any)
	if !ok {
		return errors.New("invalid updated NodeVersion spec")
	}
	if updatedMap["kubernetesComponentsReference"] != configMap.ObjectMeta.Name || updatedMap["kubernetesClusterVersion"] != newClusterVersion {
		return errors.New("failed to update NodeVersion resource")
	}

	fmt.Fprintf(u.outWriter, "Successfully updated the cluster's Kubernetes version to %s\n", newClusterVersion)
	return nil
}

// KubernetesVersion returns the version of Kubernetes the Constellation is currently running on.
func (u *Upgrader) KubernetesVersion() (string, error) {
	return u.stableInterface.kubernetesVersion()
}

// CurrentImage returns the currently used image version of the cluster.
func (u *Upgrader) CurrentImage(ctx context.Context) (string, error) {
	nodeVersion, err := u.getConstellationVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("getting constellation-version: %w", err)
	}
	return nodeVersion.Spec.ImageVersion, nil
}

// CurrentKubernetesVersion returns the currently used Kubernetes version.
func (u *Upgrader) CurrentKubernetesVersion(ctx context.Context) (string, error) {
	nodeVersion, err := u.getConstellationVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("getting constellation-version: %w", err)
	}
	return nodeVersion.Spec.KubernetesClusterVersion, nil
}

// getFromConstellationVersion queries the constellation-version object for a given field.
func (u *Upgrader) getConstellationVersion(ctx context.Context) (updatev1alpha1.NodeVersion, error) {
	raw, err := u.dynamicInterface.getCurrent(ctx, "constellation-version")
	if err != nil {
		return updatev1alpha1.NodeVersion{}, err
	}
	var nodeVersion updatev1alpha1.NodeVersion
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.UnstructuredContent(), &nodeVersion); err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("converting unstructured to NodeVersion: %w", err)
	}

	return nodeVersion, nil
}

func (u *Upgrader) updateMeasurements(ctx context.Context, newMeasurements measurements.M) error {
	existingConf, err := u.stableInterface.getCurrentConfigMap(ctx, constants.JoinConfigMap)
	if err != nil {
		return fmt.Errorf("retrieving current measurements: %w", err)
	}
	if _, ok := existingConf.Data[constants.MeasurementsFilename]; !ok {
		return errors.New("measurements missing from join-config")
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
	u.log.Debugf("Triggering measurements config map update now")
	_, err = u.stableInterface.updateConfigMap(ctx, existingConf)
	if err != nil {
		return fmt.Errorf("setting new measurements: %w", err)
	}

	fmt.Fprintln(u.outWriter, "Successfully updated the cluster's expected measurements")
	return nil
}

func (u *Upgrader) updateImage(ctx context.Context, nodeVersion updatev1alpha1.NodeVersion, newImageRef, newImageVersion string) error {
	u.log.Debugf("Upgrading cluster's image version from %s to %s", nodeVersion.Spec.ImageVersion, newImageVersion)
	nodeVersion.Spec.ImageReference = newImageRef
	nodeVersion.Spec.ImageVersion = newImageVersion

	raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
	if err != nil {
		return fmt.Errorf("converting nodeVersion to unstructured: %w", err)
	}
	u.log.Debugf("Triggering image version upgrade now")
	if _, err := u.dynamicInterface.update(ctx, &unstructured.Unstructured{Object: raw}); err != nil {
		return fmt.Errorf("setting new image: %w", err)
	}

	fmt.Fprintf(u.outWriter, "Successfully updated the cluster's image version to %s\n", newImageVersion)
	return nil
}

// k8sUpgradeInProgress checks if a k8s upgrade is in progress.
// Returns true with errors as it's the "safer" response. If caller does not check err they at least won't update the cluster.
func k8sUpgradeInProgress(nodeVersion updatev1alpha1.NodeVersion) bool {
	conditions := nodeVersion.Status.Conditions
	activeUpgrade := nodeVersion.Status.ActiveClusterVersionUpgrade

	if activeUpgrade {
		return true
	}

	for _, condition := range conditions {
		if condition.Type == updatev1alpha1.ConditionOutdated && condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

func imageUpgradeInProgress(nodeVersion updatev1alpha1.NodeVersion) bool {
	for _, condition := range nodeVersion.Status.Conditions {
		if condition.Type == updatev1alpha1.ConditionOutdated && condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

type dynamicInterface interface {
	getCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error)
	update(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type stableInterface interface {
	getCurrentConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error)
	updateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	createConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
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

// getCurrent returns a ConfigMap given it's name.
func (u *stableClient) getCurrentConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Get(ctx, name, metav1.GetOptions{})
}

// update updates the given ConfigMap.
func (u *stableClient) updateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Update(ctx, configMap, metav1.UpdateOptions{})
}

func (u *stableClient) createConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Create(ctx, configMap, metav1.CreateOptions{})
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

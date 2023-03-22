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
	"github.com/edgelesssys/constellation/v2/cli/internal/image"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	internalk8s "github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
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
	imageFetcher     imageFetcher
	outWriter        io.Writer
	log              debugLog
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
		imageFetcher:     image.New(),
		outWriter:        outWriter,
		log:              log,
	}, nil
}

// UpgradeHelmServices upgrade helm services.
func (u *Upgrader) UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error {
	return u.helmClient.Upgrade(ctx, config, timeout, allowDestructive)
}

// UpgradeNodeVersion upgrades the cluster's NodeVersion object and in turn triggers image & k8s version upgrades.
// The versions set in the config are validated against the versions running in the cluster.
func (u *Upgrader) UpgradeNodeVersion(ctx context.Context, conf *config.Config) error {
	imageReference, err := u.imageFetcher.FetchReference(ctx, conf)
	if err != nil {
		return fmt.Errorf("fetching image reference: %w", err)
	}

	imageVersion, err := versionsapi.NewVersionFromShortPath(conf.Image, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("parsing version from image short path: %w", err)
	}

	currentK8sVersion, err := versions.NewValidK8sVersion(conf.KubernetesVersion)
	if err != nil {
		return fmt.Errorf("getting Kubernetes version: %w", err)
	}
	versionConfig := versions.VersionConfigs[currentK8sVersion]

	nodeVersion, err := u.checkClusterStatus(ctx)
	if err != nil {
		return err
	}

	upgradeErrs := []error{}
	upgradeErr := &compatibility.InvalidUpgradeError{}
	err = u.updateImage(&nodeVersion, imageReference, imageVersion.Version)
	switch {
	case errors.As(err, &upgradeErr):
		upgradeErrs = append(upgradeErrs, fmt.Errorf("skipping image upgrades: %w", err))
	case err != nil:
		return fmt.Errorf("updating image version: %w", err)
	}

	components, err := u.updateK8s(&nodeVersion, versionConfig.ClusterVersion, versionConfig.KubernetesComponents)
	switch {
	case err == nil:
		err := u.applyComponentsCM(ctx, components)
		if err != nil {
			return fmt.Errorf("applying k8s components ConfigMap: %w", err)
		}
	case errors.As(err, &upgradeErr):
		upgradeErrs = append(upgradeErrs, fmt.Errorf("skipping Kubernetes upgrades: %w", err))
	default:
		return fmt.Errorf("updating Kubernetes version: %w", err)
	}

	if len(upgradeErrs) == 2 {
		return errors.Join(upgradeErrs...)
	}

	if err := u.updateMeasurements(ctx, conf.GetMeasurements()); err != nil {
		return fmt.Errorf("updating measurements: %w", err)
	}

	updatedNodeVersion, err := u.applyNodeVersion(ctx, nodeVersion)
	if err != nil {
		return fmt.Errorf("applying upgrade: %w", err)
	}
	switch {
	case updatedNodeVersion.Spec.ImageReference != imageReference:
		return fmt.Errorf("expected NodeVersion to contain %s, got %s", imageReference, updatedNodeVersion.Spec.ImageReference)
	case updatedNodeVersion.Spec.ImageVersion != imageVersion.Version:
		return fmt.Errorf("expected NodeVersion to contain %s, got %s", imageVersion.Version, updatedNodeVersion.Spec.ImageVersion)
	case updatedNodeVersion.Spec.KubernetesComponentsReference != nodeVersion.Spec.KubernetesComponentsReference:
		return fmt.Errorf("expected NodeVersion to contain %s, got %s", nodeVersion.Spec.KubernetesComponentsReference, updatedNodeVersion.Spec.KubernetesComponentsReference)
	case updatedNodeVersion.Spec.KubernetesClusterVersion != versionConfig.ClusterVersion:
		return fmt.Errorf("expected NodeVersion to contain %s, got %s", versionConfig.ClusterVersion, updatedNodeVersion.Spec.KubernetesClusterVersion)
	}

	return errors.Join(upgradeErrs...)
}

// applyComponentsCM applies the k8s components ConfigMap to the cluster.
func (u *Upgrader) applyComponentsCM(ctx context.Context, components *corev1.ConfigMap) error {
	_, err := u.stableInterface.createConfigMap(ctx, components)
	// If the map already exists we can use that map and assume it has the same content as 'configMap'.
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating k8s-components ConfigMap: %w. %T", err, err)
	}
	return nil
}

func (u *Upgrader) applyNodeVersion(ctx context.Context, nodeVersion updatev1alpha1.NodeVersion) (updatev1alpha1.NodeVersion, error) {
	raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
	if err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("converting nodeVersion to unstructured: %w", err)
	}
	u.log.Debugf("Triggering NodeVersion upgrade now")
	// Send the updated NodeVersion resource
	updated, err := u.dynamicInterface.update(ctx, &unstructured.Unstructured{Object: raw})
	if err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("updating NodeVersion: %w", err)
	}

	var updatedNodeVersion updatev1alpha1.NodeVersion
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(updated.UnstructuredContent(), &updatedNodeVersion); err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("converting unstructured to NodeVersion: %w", err)
	}

	return updatedNodeVersion, nil
}

func (u *Upgrader) checkClusterStatus(ctx context.Context) (updatev1alpha1.NodeVersion, error) {
	nodeVersion, err := u.getConstellationVersion(ctx)
	if err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("retrieving current image: %w", err)
	}

	if upgradeInProgress(nodeVersion) {
		return updatev1alpha1.NodeVersion{}, ErrInProgress
	}

	return nodeVersion, nil
}

// updateImage upgrades the cluster to the given measurements and image.
func (u *Upgrader) updateImage(nodeVersion *updatev1alpha1.NodeVersion, newImageReference, newImageVersion string) error {
	currentImageVersion := nodeVersion.Spec.ImageVersion

	if err := compatibility.IsValidUpgrade(currentImageVersion, newImageVersion); err != nil {
		return err
	}

	u.log.Debugf("Updating local copy of nodeVersion image version from %s to %s", nodeVersion.Spec.ImageVersion, newImageVersion)
	nodeVersion.Spec.ImageReference = newImageReference
	nodeVersion.Spec.ImageVersion = newImageVersion

	return nil
}

func (u *Upgrader) updateK8s(nodeVersion *updatev1alpha1.NodeVersion, newClusterVersion string, components components.Components) (*corev1.ConfigMap, error) {
	configMap, err := internalk8s.ConstructK8sComponentsCM(components, newClusterVersion)
	if err != nil {
		return nil, fmt.Errorf("constructing k8s-components ConfigMap: %w", err)
	}

	if err := compatibility.IsValidUpgrade(nodeVersion.Spec.KubernetesClusterVersion, newClusterVersion); err != nil {
		return nil, err
	}

	u.log.Debugf("Updating local copy of nodeVersion Kubernetes version from %s to %s", nodeVersion.Spec.KubernetesClusterVersion, newClusterVersion)
	nodeVersion.Spec.KubernetesComponentsReference = configMap.ObjectMeta.Name
	nodeVersion.Spec.KubernetesClusterVersion = newClusterVersion

	return &configMap, nil
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

// upgradeInProgress checks if an upgrade is in progress.
// Returns true with errors as it's the "safer" response. If caller does not check err they at least won't update the cluster.
func upgradeInProgress(nodeVersion updatev1alpha1.NodeVersion) bool {
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

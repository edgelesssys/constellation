/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/image"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	internalk8s "github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/variant"
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

// ErrLegacyJoinConfig signals that a legacy join-config was found.
// TODO: v2.9 remove.
var ErrLegacyJoinConfig = errors.New("legacy join-config with missing attestationConfig found")

// GetConstellationVersion queries the constellation-version object for a given field.
func GetConstellationVersion(ctx context.Context, client DynamicInterface) (updatev1alpha1.NodeVersion, error) {
	raw, err := client.GetCurrent(ctx, "constellation-version")
	if err != nil {
		return updatev1alpha1.NodeVersion{}, err
	}
	var nodeVersion updatev1alpha1.NodeVersion
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.UnstructuredContent(), &nodeVersion); err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("converting unstructured to NodeVersion: %w", err)
	}

	return nodeVersion, nil
}

// InvalidUpgradeError present an invalid upgrade. It wraps the source and destination version for improved debuggability.
type applyError struct {
	expected string
	actual   string
}

// Error returns the String representation of this error.
func (e *applyError) Error() string {
	return fmt.Sprintf("expected NodeVersion to contain %s, got %s", e.expected, e.actual)
}

// Upgrader handles upgrading the cluster's components using the CLI.
type Upgrader struct {
	stableInterface  stableInterface
	dynamicInterface DynamicInterface
	helmClient       helmInterface
	imageFetcher     imageFetcher
	outWriter        io.Writer
	log              debugLog
	tf               terraformUpgrader
}

// a terraformUpgrader performs the Terraform interactions in an upgrade.
type terraformUpgrader interface {
	PrepareUpgradeWorkspace(path, oldWorkingDir, newWorkingDir string, vars terraform.Variables) error
	ShowPlan(ctx context.Context, logLevel terraform.LogLevel, planFilePath string, output io.Writer) error
	Plan(ctx context.Context, logLevel terraform.LogLevel, diffWriter io.Writer) (bool, error)
	CreateCluster(ctx context.Context, logLevel terraform.LogLevel) (terraform.CreateOutput, error)
}

// NewUpgrader returns a new Upgrader.
func NewUpgrader(ctx context.Context, outWriter io.Writer, log debugLog) (*Upgrader, error) {
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

	tfClient, err := terraform.New(ctx, constants.TerraformWorkingDir)
	if err != nil {
		return nil, fmt.Errorf("setting up terraform client: %w", err)
	}

	return &Upgrader{
		stableInterface:  &stableClient{client: kubeClient},
		dynamicInterface: &NodeVersionClient{client: unstructuredClient},
		helmClient:       helmClient,
		imageFetcher:     image.New(),
		outWriter:        outWriter,
		log:              log,
		tf:               tfClient,
	}, nil
}

// PlanTerraformMigrations plans the Terraform migrations for the Constellation upgrade, writing the plan output to the given file if
// a diff exists. It also returns a bool indicating whether a diff exists.
func (u *Upgrader) PlanTerraformMigrations(ctx context.Context, fileHandler file.Handler, logLevel terraform.LogLevel, planFile string) (bool, error) {
	diff := bytes.Buffer{}
	hasDiff, err := u.tf.Plan(ctx, logLevel, &diff)
	if err != nil {
		return false, fmt.Errorf("terraform plan: %w", err)
	}

	if err := fileHandler.Write(planFile, diff.Bytes()); err != nil {
		return false, fmt.Errorf("writing plan file: %w", err)
	}

	return hasDiff, nil
}

// ShowTerraformMigrations formats the Terraform diff from the given plan file and writes it to the upgrader's output writer.
func (u *Upgrader) ShowTerraformMigrations(ctx context.Context, fileHandler file.Handler, logLevel terraform.LogLevel, planFile string) error {
	planContent, err := fileHandler.Read(planFile)
	if err != nil {
		return fmt.Errorf("reading plan file: %w", err)
	}

	if err := u.tf.ShowPlan(ctx, logLevel, string(planContent), u.outWriter); err != nil {
		return fmt.Errorf("terraform show plan: %w", err)
	}

	return nil
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

	updatedNodeVersion, err := u.applyNodeVersion(ctx, nodeVersion)
	if err != nil {
		return fmt.Errorf("applying upgrade: %w", err)
	}
	switch {
	case updatedNodeVersion.Spec.ImageReference != nodeVersion.Spec.ImageReference:
		return &applyError{expected: nodeVersion.Spec.ImageReference, actual: updatedNodeVersion.Spec.ImageReference}
	case updatedNodeVersion.Spec.ImageVersion != nodeVersion.Spec.ImageVersion:
		return &applyError{expected: nodeVersion.Spec.ImageVersion, actual: updatedNodeVersion.Spec.ImageVersion}
	case updatedNodeVersion.Spec.KubernetesComponentsReference != nodeVersion.Spec.KubernetesComponentsReference:
		return &applyError{expected: nodeVersion.Spec.KubernetesComponentsReference, actual: updatedNodeVersion.Spec.KubernetesComponentsReference}
	case updatedNodeVersion.Spec.KubernetesClusterVersion != nodeVersion.Spec.KubernetesClusterVersion:
		return &applyError{expected: nodeVersion.Spec.KubernetesClusterVersion, actual: updatedNodeVersion.Spec.KubernetesClusterVersion}
	}

	return errors.Join(upgradeErrs...)
}

// KubernetesVersion returns the version of Kubernetes the Constellation is currently running on.
func (u *Upgrader) KubernetesVersion() (string, error) {
	return u.stableInterface.kubernetesVersion()
}

// CurrentImage returns the currently used image version of the cluster.
func (u *Upgrader) CurrentImage(ctx context.Context) (string, error) {
	nodeVersion, err := GetConstellationVersion(ctx, u.dynamicInterface)
	if err != nil {
		return "", fmt.Errorf("getting constellation-version: %w", err)
	}
	return nodeVersion.Spec.ImageVersion, nil
}

// CurrentKubernetesVersion returns the currently used Kubernetes version.
func (u *Upgrader) CurrentKubernetesVersion(ctx context.Context) (string, error) {
	nodeVersion, err := GetConstellationVersion(ctx, u.dynamicInterface)
	if err != nil {
		return "", fmt.Errorf("getting constellation-version: %w", err)
	}
	return nodeVersion.Spec.KubernetesClusterVersion, nil
}

// UpdateAttestationConfig fetches the cluster's attestation config, compares them to a new config,
// and updates the cluster's config if it is different from the new one.
func (u *Upgrader) UpdateAttestationConfig(ctx context.Context, newAttestConfig config.AttestationCfg) error {
	currentAttestConfig, joinConfig, err := u.GetClusterAttestationConfig(ctx, newAttestConfig.GetVariant())
	if err != nil {
		if !errors.Is(err, ErrLegacyJoinConfig) {
			return fmt.Errorf("getting cluster attestation config: %w", err)
		}
		currentAttestConfig, joinConfig, err = joinConfigMigration(joinConfig, newAttestConfig.GetVariant())
		if err != nil {
			return fmt.Errorf("migrating join config: %w", err)
		}
	}
	equal, err := newAttestConfig.EqualTo(currentAttestConfig)
	if err != nil {
		return fmt.Errorf("comparing attestation configs: %w", err)
	}
	if equal {
		fmt.Fprintln(u.outWriter, "Cluster is already using the chosen attestation config, skipping config upgrade")
		return nil
	}

	// backup of previous measurements
	joinConfig.Data[constants.AttestationConfigFilename+"_backup"] = joinConfig.Data[constants.AttestationConfigFilename]

	newConfigJSON, err := json.Marshal(newAttestConfig)
	if err != nil {
		return fmt.Errorf("marshaling attestation config: %w", err)
	}
	joinConfig.Data[constants.AttestationConfigFilename] = string(newConfigJSON)
	u.log.Debugf("Triggering attestation config update now")
	if _, err = u.stableInterface.updateConfigMap(ctx, joinConfig); err != nil {
		return fmt.Errorf("setting new attestation config: %w", err)
	}

	fmt.Fprintln(u.outWriter, "Successfully updated the cluster's attestation config")
	return nil
}

// GetClusterAttestationConfig fetches the join-config configmap from the cluster, extracts the config
// and returns both the full configmap and the attestation config.
func (u *Upgrader) GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, *corev1.ConfigMap, error) {
	existingConf, err := u.stableInterface.getCurrentConfigMap(ctx, constants.JoinConfigMap)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving current attestation config: %w", err)
	}
	if _, ok := existingConf.Data[constants.AttestationConfigFilename]; !ok {
		// TODO: v2.9 remove legacy config detection since it is only required for upgrades from v2.7
		if _, ok := existingConf.Data["measurements"]; ok {
			u.log.Debugf("Legacy join config detected, migrating to new config")
			return nil, existingConf, ErrLegacyJoinConfig
		}
		return nil, nil, errors.New("attestation config missing from join-config")
	}

	existingAttestationConfig, err := config.UnmarshalAttestationConfig([]byte(existingConf.Data[constants.AttestationConfigFilename]), variant)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving current attestation config: %w", err)
	}

	return existingAttestationConfig, existingConf, nil
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
	updated, err := u.dynamicInterface.Update(ctx, &unstructured.Unstructured{Object: raw})
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
	nodeVersion, err := GetConstellationVersion(ctx, u.dynamicInterface)
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

// NodeVersionClient implements the DynamicInterface interface to interact with NodeVersion objects.
type NodeVersionClient struct {
	client dynamic.Interface
}

// NewNodeVersionClient returns a new NodeVersionClient.
func NewNodeVersionClient(client dynamic.Interface) *NodeVersionClient {
	return &NodeVersionClient{client: client}
}

// GetCurrent returns the current NodeVersion object.
func (u *NodeVersionClient) GetCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	return u.client.Resource(schema.GroupVersionResource{
		Group:    "update.edgeless.systems",
		Version:  "v1alpha1",
		Resource: "nodeversions",
	}).Get(ctx, name, metav1.GetOptions{})
}

// Update updates the NodeVersion object.
func (u *NodeVersionClient) Update(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return u.client.Resource(schema.GroupVersionResource{
		Group:    "update.edgeless.systems",
		Version:  "v1alpha1",
		Resource: "nodeversions",
	}).Update(ctx, obj, metav1.UpdateOptions{})
}

// DynamicInterface is a general interface to query custom resources.
type DynamicInterface interface {
	GetCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error)
	Update(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
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

type stableInterface interface {
	getCurrentConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error)
	updateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	createConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	kubernetesVersion() (string, error)
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

// joinConfigMigration prepares a join-config ConfigMap for migration from v2.7 to v2.8.
// TODO: v2.9: remove this function.
func joinConfigMigration(existingConf *corev1.ConfigMap, attestVariant variant.Variant) (config.AttestationCfg, *corev1.ConfigMap, error) {
	m, ok := existingConf.Data["measurements"]
	if !ok {
		return nil, nil, errors.New("no measurements found in configmap")
	}

	var measurements measurements.M
	if err := json.Unmarshal([]byte(m), &measurements); err != nil {
		return nil, nil, fmt.Errorf("unmarshalling measurements: %w", err)
	}

	var oldConf config.AttestationCfg
	switch attestVariant {
	case variant.AWSNitroTPM{}:
		oldConf = &config.AWSNitroTPM{}
	case variant.AzureSEVSNP{}:
		oldConf = &config.AzureSEVSNP{}
	case variant.AzureTrustedLaunch{}:
		oldConf = &config.AzureTrustedLaunch{}
	case variant.GCPSEVES{}:
		oldConf = &config.GCPSEVES{}
	case variant.QEMUVTPM{}:
		oldConf = &config.QEMUVTPM{}
	default:
		return nil, nil, fmt.Errorf("unknown variant: %s", attestVariant)
	}

	oldConf.SetMeasurements(measurements)
	oldConfJSON, err := json.Marshal(oldConf)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling previous config: %w", err)
	}
	existingConf.Data[constants.AttestationConfigFilename] = string(oldConfJSON)

	return oldConf, existingConf, nil
}

type helmInterface interface {
	Upgrade(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error
}

type debugLog interface {
	Debugf(format string, args ...any)
	Sync()
}

// imageFetcher gets an image reference from the versionsapi.
type imageFetcher interface {
	FetchReference(ctx context.Context, config *config.Config) (string, error)
}

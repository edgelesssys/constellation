/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubecmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	internalk8s "github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	kubeadmv1beta3 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	"sigs.k8s.io/yaml"
)

// ErrInProgress signals that an upgrade is in progress inside the cluster.
var ErrInProgress = errors.New("upgrade in progress")

// InvalidUpgradeError present an invalid upgrade. It wraps the source and destination version for improved debuggability.
type applyError struct {
	expected string
	actual   string
}

// Error returns the String representation of this error.
func (e *applyError) Error() string {
	return fmt.Sprintf("expected NodeVersion to contain %s, got %s", e.expected, e.actual)
}

// KubeCmd handles interaction with the cluster's components using the CLI.
type KubeCmd struct {
	stableInterface  stableInterface
	dynamicInterface dynamicInterface
	imageFetcher     imageFetcher
	outWriter        io.Writer
	log              debugLog
}

// New returns a new KubeCmd.
func New(outWriter io.Writer, kubeConfigPath string, log debugLog) (*KubeCmd, error) {
	kubeClient, err := newClient(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	// use unstructured client to avoid importing the operator packages
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("building kubernetes config: %w", err)
	}

	unstructuredClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("setting up custom resource client: %w", err)
	}

	return &KubeCmd{
		imageFetcher:     imagefetcher.New(),
		outWriter:        outWriter,
		log:              log,
		stableInterface:  &stableClient{client: kubeClient},
		dynamicInterface: newNodeVersionClient(unstructuredClient),
	}, nil
}

// GetMeasurementSalt returns the measurementSalt from the join-config.
func (u *KubeCmd) GetMeasurementSalt(ctx context.Context) ([]byte, error) {
	cm, err := u.stableInterface.GetConfigMap(ctx, constants.JoinConfigMap)
	if err != nil {
		return nil, fmt.Errorf("retrieving current join-config: %w", err)
	}
	salt, ok := cm.BinaryData[constants.MeasurementSaltFilename]
	if !ok {
		return nil, errors.New("measurementSalt missing from join-config")
	}
	return salt, nil
}

// GetConstellationVersion queries the constellation-version object for a given field.
func (u *KubeCmd) GetConstellationVersion(ctx context.Context) (updatev1alpha1.NodeVersion, error) {
	raw, err := u.dynamicInterface.GetCurrent(ctx, "constellation-version")
	if err != nil {
		return updatev1alpha1.NodeVersion{}, err
	}
	var nodeVersion updatev1alpha1.NodeVersion
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.UnstructuredContent(), &nodeVersion); err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("converting unstructured to NodeVersion: %w", err)
	}

	return nodeVersion, nil
}

// UpgradeNodeVersion upgrades the cluster's NodeVersion object and in turn triggers image & k8s version upgrades.
// The versions set in the config are validated against the versions running in the cluster.
func (u *KubeCmd) UpgradeNodeVersion(ctx context.Context, conf *config.Config, force bool) error {
	provider := conf.GetProvider()
	attestationVariant := conf.GetAttestationConfig().GetVariant()
	region := conf.GetRegion()
	imageReference, err := u.imageFetcher.FetchReference(ctx, provider, attestationVariant, conf.Image, region)
	if err != nil {
		return fmt.Errorf("fetching image reference: %w", err)
	}

	imageVersion, err := versionsapi.NewVersionFromShortPath(conf.Image, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("parsing version from image short path: %w", err)
	}

	nodeVersion, err := u.getClusterStatus(ctx)
	if err != nil {
		return err
	}

	upgradeErrs := []error{}
	var upgradeErr *compatibility.InvalidUpgradeError

	err = u.updateImage(&nodeVersion, imageReference, imageVersion.Version(), force)
	switch {
	case errors.As(err, &upgradeErr):
		upgradeErrs = append(upgradeErrs, fmt.Errorf("skipping image upgrades: %w", err))
	case err != nil:
		return fmt.Errorf("updating image version: %w", err)
	}

	// We have to allow users to specify outdated k8s patch versions.
	// Therefore, this code has to skip k8s updates if a user configures an outdated (i.e. invalid) k8s version.
	var components *corev1.ConfigMap
	currentK8sVersion, err := versions.NewValidK8sVersion(conf.KubernetesVersion, true)
	if err != nil {
		innerErr := fmt.Errorf("unsupported Kubernetes version, supported versions are %s", strings.Join(versions.SupportedK8sVersions(), ", "))
		err = compatibility.NewInvalidUpgradeError(nodeVersion.Spec.KubernetesClusterVersion, conf.KubernetesVersion, innerErr)
	} else {
		versionConfig := versions.VersionConfigs[currentK8sVersion]
		components, err = u.updateK8s(&nodeVersion, versionConfig.ClusterVersion, versionConfig.KubernetesComponents, force)
	}

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

// ClusterStatus returns a map from node name to NodeStatus.
func (u *KubeCmd) ClusterStatus(ctx context.Context) (map[string]NodeStatus, error) {
	nodes, err := u.stableInterface.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting nodes: %w", err)
	}

	clusterStatus := map[string]NodeStatus{}
	for _, node := range nodes {
		clusterStatus[node.ObjectMeta.Name] = NewNodeStatus(node)
	}

	return clusterStatus, nil
}

// KubernetesVersion returns the version of Kubernetes the Constellation is currently running on.
func (u *KubeCmd) KubernetesVersion() (string, error) {
	return u.stableInterface.KubernetesVersion()
}

// CurrentImage returns the currently used image version of the cluster.
func (u *KubeCmd) CurrentImage(ctx context.Context) (string, error) {
	nodeVersion, err := u.GetConstellationVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("getting constellation-version: %w", err)
	}
	return nodeVersion.Spec.ImageVersion, nil
}

// CurrentKubernetesVersion returns the currently used Kubernetes version.
func (u *KubeCmd) CurrentKubernetesVersion(ctx context.Context) (string, error) {
	nodeVersion, err := u.GetConstellationVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("getting constellation-version: %w", err)
	}
	return nodeVersion.Spec.KubernetesClusterVersion, nil
}

// GetClusterAttestationConfig fetches the join-config configmap from the cluster, extracts the config
// and returns both the full configmap and the attestation config.
func (u *KubeCmd) GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error) {
	existingConf, err := u.stableInterface.GetConfigMap(ctx, constants.JoinConfigMap)
	if err != nil {
		return nil, fmt.Errorf("retrieving current attestation config: %w", err)
	}
	if _, ok := existingConf.Data[constants.AttestationConfigFilename]; !ok {
		return nil, errors.New("attestation config missing from join-config")
	}

	existingAttestationConfig, err := config.UnmarshalAttestationConfig([]byte(existingConf.Data[constants.AttestationConfigFilename]), variant)
	if err != nil {
		return nil, fmt.Errorf("retrieving current attestation config: %w", err)
	}

	return existingAttestationConfig, nil
}

// BackupConfigMap creates a backup of the given config map.
func (u *KubeCmd) BackupConfigMap(ctx context.Context, name string) error {
	cm, err := u.stableInterface.GetConfigMap(ctx, name)
	if err != nil {
		return fmt.Errorf("getting config map %s: %w", name, err)
	}
	backup := cm.DeepCopy()
	backup.ObjectMeta = metav1.ObjectMeta{}
	backup.Name = fmt.Sprintf("%s-backup", name)
	if _, err := u.stableInterface.CreateConfigMap(ctx, backup); err != nil {
		if _, err := u.stableInterface.UpdateConfigMap(ctx, backup); err != nil {
			return fmt.Errorf("updating backup config map: %w", err)
		}
	}
	u.log.Debugf("Successfully backed up config map %s", cm.Name)
	return nil
}

// UpdateAttestationConfig fetches the cluster's attestation config, compares them to a new config,
// and updates the cluster's config if it is different from the new one.
func (u *Upgrader) UpdateAttestationConfig(ctx context.Context, newAttestConfig config.AttestationCfg) error {
	// backup of previous measurements
	joinConfig, err := u.stableInterface.GetConfigMap(ctx, constants.JoinConfigMap)
	if err != nil {
		return fmt.Errorf("getting join-config configmap: %w", err)
	}

	newConfigJSON, err := json.Marshal(newAttestConfig)
	if err != nil {
		return fmt.Errorf("marshaling attestation config: %w", err)
	}
	joinConfig.Data[constants.AttestationConfigFilename] = string(newConfigJSON)
	u.log.Debugf("Triggering attestation config update now")
	if _, err = u.stableInterface.UpdateConfigMap(ctx, joinConfig); err != nil {
		return fmt.Errorf("setting new attestation config: %w", err)
	}

	fmt.Fprintln(u.outWriter, "Successfully updated the cluster's attestation config")
	return nil
}

// ExtendClusterConfigCertSANs extends the ClusterConfig stored under "kube-system/kubeadm-config" with the given SANs.
// Existing SANs are preserved.
func (u *KubeCmd) ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error {
	clusterConfiguration, kubeadmConfig, err := u.GetClusterConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("getting ClusterConfig: %w", err)
	}

	existingSANs := make(map[string]struct{})
	for _, existingSAN := range clusterConfiguration.APIServer.CertSANs {
		existingSANs[existingSAN] = struct{}{}
	}

	var missingSANs []string
	for _, san := range alternativeNames {
		if _, ok := existingSANs[san]; !ok {
			missingSANs = append(missingSANs, san)
		}
	}

	if len(missingSANs) == 0 {
		return nil
	}
	u.log.Debugf("Extending the cluster's apiserver SAN field with the following SANs: %s\n", strings.Join(missingSANs, ", "))

	clusterConfiguration.APIServer.CertSANs = append(clusterConfiguration.APIServer.CertSANs, missingSANs...)
	sort.Strings(clusterConfiguration.APIServer.CertSANs)

	newConfigYAML, err := yaml.Marshal(clusterConfiguration)
	if err != nil {
		return fmt.Errorf("marshaling ClusterConfiguration: %w", err)
	}

	kubeadmConfig.Data[constants.ClusterConfigurationKey] = string(newConfigYAML)
	u.log.Debugf("Triggering kubeadm config update now")
	if _, err = u.stableInterface.UpdateConfigMap(ctx, kubeadmConfig); err != nil {
		return fmt.Errorf("setting new kubeadm config: %w", err)
	}

	fmt.Fprintln(u.outWriter, "Successfully extended the cluster's apiserver SAN field")
	return nil
}

// GetClusterConfiguration fetches the kubeadm-config configmap from the cluster, extracts the config
// and returns both the full configmap and the ClusterConfiguration.
func (u *KubeCmd) GetClusterConfiguration(ctx context.Context) (kubeadmv1beta3.ClusterConfiguration, *corev1.ConfigMap, error) {
	existingConf, err := u.stableInterface.GetConfigMap(ctx, constants.KubeadmConfigMap)
	if err != nil {
		return kubeadmv1beta3.ClusterConfiguration{}, nil, fmt.Errorf("retrieving current kubeadm-config: %w", err)
	}
	clusterConf, ok := existingConf.Data[constants.ClusterConfigurationKey]
	if !ok {
		return kubeadmv1beta3.ClusterConfiguration{}, nil, errors.New("ClusterConfiguration missing from kubeadm-config")
	}

	var existingClusterConfig kubeadmv1beta3.ClusterConfiguration
	if err := yaml.Unmarshal([]byte(clusterConf), &existingClusterConfig); err != nil {
		return kubeadmv1beta3.ClusterConfiguration{}, nil, fmt.Errorf("unmarshaling ClusterConfiguration: %w", err)
	}

	return existingClusterConfig, existingConf, nil
}

// applyComponentsCM applies the k8s components ConfigMap to the cluster.
func (u *KubeCmd) applyComponentsCM(ctx context.Context, components *corev1.ConfigMap) error {
	_, err := u.stableInterface.CreateConfigMap(ctx, components)
	// If the map already exists we can use that map and assume it has the same content as 'configMap'.
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating k8s-components ConfigMap: %w. %T", err, err)
	}
	return nil
}

func (u *KubeCmd) applyNodeVersion(ctx context.Context, nodeVersion updatev1alpha1.NodeVersion) (updatev1alpha1.NodeVersion, error) {
	u.log.Debugf("Triggering NodeVersion upgrade now")
	var updatedNodeVersion updatev1alpha1.NodeVersion
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		newNode, err := u.getClusterStatus(ctx)
		if err != nil {
			return fmt.Errorf("retrieving current NodeVersion: %w", err)
		}
		cmd := newUpgradeVersionCmd(nodeVersion)
		cmd.SetUpdatedVersions(&newNode)
		raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&newNode)
		if err != nil {
			return fmt.Errorf("converting nodeVersion to unstructured: %w", err)
		}
		updated, err := u.dynamicInterface.Update(ctx, &unstructured.Unstructured{Object: raw})
		if err != nil {
			return err
		}

		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(updated.UnstructuredContent(), &updatedNodeVersion); err != nil {
			return fmt.Errorf("converting unstructured to NodeVersion: %w", err)
		}
		return nil
	})

	return updatedNodeVersion, err
}

func (u *KubeCmd) getClusterStatus(ctx context.Context) (updatev1alpha1.NodeVersion, error) {
	nodeVersion, err := u.GetConstellationVersion(ctx)
	if err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("retrieving current image: %w", err)
	}

	return nodeVersion, nil
}

// updateImage upgrades the cluster to the given measurements and image.
func (u *KubeCmd) updateImage(nodeVersion *updatev1alpha1.NodeVersion, newImageReference, newImageVersion string, force bool) error {
	currentImageVersion := nodeVersion.Spec.ImageVersion
	if !force {
		if upgradeInProgress(*nodeVersion) {
			return ErrInProgress
		}
		if err := compatibility.IsValidUpgrade(currentImageVersion, newImageVersion); err != nil {
			return fmt.Errorf("validating image update: %w", err)
		}
	}
	u.log.Debugf("Updating local copy of nodeVersion image version from %s to %s", nodeVersion.Spec.ImageVersion, newImageVersion)
	nodeVersion.Spec.ImageReference = newImageReference
	nodeVersion.Spec.ImageVersion = newImageVersion

	return nil
}

func (u *KubeCmd) updateK8s(nodeVersion *updatev1alpha1.NodeVersion, newClusterVersion string, components components.Components, force bool) (*corev1.ConfigMap, error) {
	configMap, err := internalk8s.ConstructK8sComponentsCM(components, newClusterVersion)
	if err != nil {
		return nil, fmt.Errorf("constructing k8s-components ConfigMap: %w", err)
	}
	if !force {
		if err := compatibility.IsValidUpgrade(nodeVersion.Spec.KubernetesClusterVersion, newClusterVersion); err != nil {
			return nil, err
		}
	}

	u.log.Debugf("Updating local copy of nodeVersion Kubernetes version from %s to %s", nodeVersion.Spec.KubernetesClusterVersion, newClusterVersion)
	nodeVersion.Spec.KubernetesComponentsReference = configMap.ObjectMeta.Name
	nodeVersion.Spec.KubernetesClusterVersion = newClusterVersion

	return &configMap, nil
}

// nodeVersionClient implements the DynamicInterface interface to interact with NodeVersion objects.
type nodeVersionClient struct {
	client dynamic.Interface
}

// newNodeVersionClient returns a new nodeVersionClient.
func newNodeVersionClient(client dynamic.Interface) *nodeVersionClient {
	return &nodeVersionClient{client: client}
}

// GetCurrent returns the current NodeVersion object.
func (u *nodeVersionClient) GetCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	return u.client.Resource(schema.GroupVersionResource{
		Group:    "update.edgeless.systems",
		Version:  "v1alpha1",
		Resource: "nodeversions",
	}).Get(ctx, name, metav1.GetOptions{})
}

// Update updates the NodeVersion object.
func (u *nodeVersionClient) Update(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return u.client.Resource(schema.GroupVersionResource{
		Group:    "update.edgeless.systems",
		Version:  "v1alpha1",
		Resource: "nodeversions",
	}).Update(ctx, obj, metav1.UpdateOptions{})
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

type upgradeVersionCmd struct {
	imageVersion     string
	imageRef         string
	k8sComponentsRef string
	k8sVersion       string
}

func newUpgradeVersionCmd(newNodeVersion updatev1alpha1.NodeVersion) upgradeVersionCmd {
	return upgradeVersionCmd{
		imageVersion:     newNodeVersion.Spec.ImageVersion,
		imageRef:         newNodeVersion.Spec.ImageReference,
		k8sComponentsRef: newNodeVersion.Spec.KubernetesComponentsReference,
		k8sVersion:       newNodeVersion.Spec.KubernetesClusterVersion,
	}
}

func (u upgradeVersionCmd) SetUpdatedVersions(node *updatev1alpha1.NodeVersion) {
	if u.imageVersion != "" {
		node.Spec.ImageVersion = u.imageVersion
	}
	if u.imageRef != "" {
		node.Spec.ImageReference = u.imageRef
	}
	if u.k8sComponentsRef != "" {
		node.Spec.KubernetesComponentsReference = u.k8sComponentsRef
	}
	if u.k8sVersion != "" {
		node.Spec.KubernetesClusterVersion = u.k8sVersion
	}
}

// stableInterface is an interface to interact with stable resources.
type stableInterface interface {
	GetNodes(ctx context.Context) ([]corev1.Node, error)
	GetConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error)
	UpdateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	KubernetesVersion() (string, error)
}

// dynamicInterface is a general interface to query custom resources.
type dynamicInterface interface {
	GetCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error)
	Update(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type debugLog interface {
	Debugf(format string, args ...any)
	Sync()
}

// imageFetcher gets an image reference from the versionsapi.
type imageFetcher interface {
	FetchReference(ctx context.Context,
		provider cloudprovider.Provider, attestationVariant variant.Variant,
		image, region string,
	) (string, error)
}

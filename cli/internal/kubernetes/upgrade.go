/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/cli/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	internalk8s "github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/google/uuid"
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

// UpgradeCmdKind is the kind of the upgrade command (check, apply).
type UpgradeCmdKind int

const (
	// UpgradeCmdKindCheck corresponds to the upgrade check command.
	UpgradeCmdKindCheck UpgradeCmdKind = iota
	// UpgradeCmdKindApply corresponds to the upgrade apply command.
	UpgradeCmdKindApply
)

// ErrInProgress signals that an upgrade is in progress inside the cluster.
var ErrInProgress = errors.New("upgrade in progress")

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
	stableInterface  StableInterface
	dynamicInterface DynamicInterface
	helmClient       helmInterface
	imageFetcher     imageFetcher
	outWriter        io.Writer
	tfUpgrader       *upgrade.TerraformUpgrader
	log              debugLog
	upgradeID        string
}

// NewUpgrader returns a new Upgrader.
func NewUpgrader(
	ctx context.Context, outWriter io.Writer, upgradeWorkspace, kubeConfigPath string,
	fileHandler file.Handler, log debugLog, upgradeCmdKind UpgradeCmdKind,
) (*Upgrader, error) {
	upgradeID := "upgrade-" + time.Now().Format("20060102150405") + "-" + strings.Split(uuid.New().String(), "-")[0]
	if upgradeCmdKind == UpgradeCmdKindCheck {
		// When performing an upgrade check, the upgrade directory will only be used temporarily to store the
		// Terraform state. The directory is deleted after the check is finished.
		// Therefore, add a tmp-suffix to the upgrade ID to indicate that the directory will be cleared after the check.
		upgradeID += "-tmp"
	}

	u := &Upgrader{
		imageFetcher: imagefetcher.New(),
		outWriter:    outWriter,
		log:          log,
		upgradeID:    upgradeID,
	}

	kubeClient, err := newClient(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	u.stableInterface = &stableClient{client: kubeClient}

	// use unstructured client to avoid importing the operator packages
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("building kubernetes config: %w", err)
	}
	unstructuredClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("setting up custom resource client: %w", err)
	}
	u.dynamicInterface = &NodeVersionClient{client: unstructuredClient}

	helmClient, err := helm.NewUpgradeClient(kubectl.New(), upgradeWorkspace, kubeConfigPath, constants.HelmNamespace, log)
	if err != nil {
		return nil, fmt.Errorf("setting up helm client: %w", err)
	}
	u.helmClient = helmClient

	tfClient, err := terraform.New(ctx, filepath.Join(upgradeWorkspace, upgradeID, constants.TerraformUpgradeWorkingDir))
	if err != nil {
		return nil, fmt.Errorf("setting up terraform client: %w", err)
	}

	tfUpgrader, err := upgrade.NewTerraformUpgrader(tfClient, outWriter, fileHandler)
	if err != nil {
		return nil, fmt.Errorf("setting up terraform upgrader: %w", err)
	}
	u.tfUpgrader = tfUpgrader

	return u, nil
}

// GetUpgradeID returns the upgrade ID.
func (u *Upgrader) GetUpgradeID() string {
	return u.upgradeID
}

// CheckTerraformMigrations checks whether Terraform migrations are possible in the current workspace.
// If the files that will be written during the upgrade already exist, it returns an error.
func (u *Upgrader) CheckTerraformMigrations(upgradeWorkspace string) error {
	return u.tfUpgrader.CheckTerraformMigrations(upgradeWorkspace, u.upgradeID, constants.TerraformUpgradeBackupDir)
}

// CleanUpTerraformMigrations cleans up the Terraform migration workspace, for example when an upgrade is
// aborted by the user.
func (u *Upgrader) CleanUpTerraformMigrations(upgradeWorkspace string) error {
	return u.tfUpgrader.CleanUpTerraformMigrations(upgradeWorkspace, u.upgradeID)
}

// PlanTerraformMigrations prepares the upgrade workspace and plans the Terraform migrations for the Constellation upgrade.
// If a diff exists, it's being written to the upgrader's output writer. It also returns
// a bool indicating whether a diff exists.
func (u *Upgrader) PlanTerraformMigrations(ctx context.Context, opts upgrade.TerraformUpgradeOptions) (bool, error) {
	return u.tfUpgrader.PlanTerraformMigrations(ctx, opts, u.upgradeID)
}

// ApplyTerraformMigrations applies the migrations planned by PlanTerraformMigrations.
// If PlanTerraformMigrations has not been executed before, it will return an error.
// In case of a successful upgrade, the output will be written to the specified file and the old Terraform directory is replaced
// By the new one.
func (u *Upgrader) ApplyTerraformMigrations(ctx context.Context, opts upgrade.TerraformUpgradeOptions) (clusterid.File, error) {
	return u.tfUpgrader.ApplyTerraformMigrations(ctx, opts, u.upgradeID)
}

// UpgradeHelmServices upgrade helm services.
func (u *Upgrader) UpgradeHelmServices(ctx context.Context, config *config.Config, idFile clusterid.File, timeout time.Duration, allowDestructive bool, force bool) error {
	return u.helmClient.Upgrade(ctx, config, idFile, timeout, allowDestructive, force, u.upgradeID)
}

// UpgradeNodeVersion upgrades the cluster's NodeVersion object and in turn triggers image & k8s version upgrades.
// The versions set in the config are validated against the versions running in the cluster.
func (u *Upgrader) UpgradeNodeVersion(ctx context.Context, conf *config.Config, force bool) error {
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

// KubernetesVersion returns the version of Kubernetes the Constellation is currently running on.
func (u *Upgrader) KubernetesVersion() (string, error) {
	return u.stableInterface.KubernetesVersion()
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
		return fmt.Errorf("getting attestation config: %w", err)
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
	if _, err = u.stableInterface.UpdateConfigMap(ctx, joinConfig); err != nil {
		return fmt.Errorf("setting new attestation config: %w", err)
	}

	fmt.Fprintln(u.outWriter, "Successfully updated the cluster's attestation config")
	return nil
}

// GetClusterAttestationConfig fetches the join-config configmap from the cluster, extracts the config
// and returns both the full configmap and the attestation config.
func (u *Upgrader) GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, *corev1.ConfigMap, error) {
	existingConf, err := u.stableInterface.GetCurrentConfigMap(ctx, constants.JoinConfigMap)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving current attestation config: %w", err)
	}
	if _, ok := existingConf.Data[constants.AttestationConfigFilename]; !ok {
		return nil, nil, errors.New("attestation config missing from join-config")
	}

	existingAttestationConfig, err := config.UnmarshalAttestationConfig([]byte(existingConf.Data[constants.AttestationConfigFilename]), variant)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving current attestation config: %w", err)
	}

	return existingAttestationConfig, existingConf, nil
}

// ExtendClusterConfigCertSANs extends the ClusterConfig stored under "kube-system/kubeadm-config" with the given SANs.
// Existing SANs are preserved.
func (u *Upgrader) ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error {
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
func (u *Upgrader) GetClusterConfiguration(ctx context.Context) (kubeadmv1beta3.ClusterConfiguration, *corev1.ConfigMap, error) {
	existingConf, err := u.stableInterface.GetCurrentConfigMap(ctx, constants.KubeadmConfigMap)
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
func (u *Upgrader) applyComponentsCM(ctx context.Context, components *corev1.ConfigMap) error {
	_, err := u.stableInterface.CreateConfigMap(ctx, components)
	// If the map already exists we can use that map and assume it has the same content as 'configMap'.
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating k8s-components ConfigMap: %w. %T", err, err)
	}
	return nil
}

func (u *Upgrader) applyNodeVersion(ctx context.Context, nodeVersion updatev1alpha1.NodeVersion) (updatev1alpha1.NodeVersion, error) {
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

func (u *Upgrader) getClusterStatus(ctx context.Context) (updatev1alpha1.NodeVersion, error) {
	nodeVersion, err := GetConstellationVersion(ctx, u.dynamicInterface)
	if err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("retrieving current image: %w", err)
	}

	return nodeVersion, nil
}

// updateImage upgrades the cluster to the given measurements and image.
func (u *Upgrader) updateImage(nodeVersion *updatev1alpha1.NodeVersion, newImageReference, newImageVersion string, force bool) error {
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

func (u *Upgrader) updateK8s(nodeVersion *updatev1alpha1.NodeVersion, newClusterVersion string, components components.Components, force bool) (*corev1.ConfigMap, error) {
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

type helmInterface interface {
	Upgrade(ctx context.Context, config *config.Config, idFile clusterid.File, timeout time.Duration, allowDestructive, force bool, upgradeID string) error
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

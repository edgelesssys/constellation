/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package kubecmd provides functions to interact with a Kubernetes cluster to the CLI.
The package should be used for:

  - Fetching status information about the cluster
  - Creating, deleting, or migrating resources not managed by Helm

The package should not be used for anything that doesn't just require the Kubernetes API.
For example, Terraform and Helm actions should not be accessed by this package.
*/
package kubecmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"time"

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
	conretry "github.com/edgelesssys/constellation/v2/internal/retry"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
	kubeadmv1beta3 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	"sigs.k8s.io/yaml"
)

const (
	maxRetryAttempts = 20
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
	kubectl       kubectlInterface
	imageFetcher  imageFetcher
	outWriter     io.Writer
	fileHandler   file.Handler
	retryInterval time.Duration
	log           debugLog
}

// New returns a new KubeCmd.
func New(outWriter io.Writer, kubeConfig []byte, fileHandler file.Handler, log debugLog) (*KubeCmd, error) {
	client, err := kubectl.NewFromConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("creating kubectl client: %w", err)
	}

	return &KubeCmd{
		kubectl:       client,
		fileHandler:   fileHandler,
		imageFetcher:  imagefetcher.New(),
		outWriter:     outWriter,
		retryInterval: time.Second * 5,
		log:           log,
	}, nil
}

// UpgradeNodeVersion upgrades the cluster's NodeVersion object and in turn triggers image & k8s version upgrades.
// The versions set in the config are validated against the versions running in the cluster.
// TODO(elchead): AB#3434 Split K8s and image upgrade of UpgradeNodeVersion.
func (k *KubeCmd) UpgradeNodeVersion(ctx context.Context, conf *config.Config, force, skipImage, skipK8s bool) error {
	provider := conf.GetProvider()
	attestationVariant := conf.GetAttestationConfig().GetVariant()
	region := conf.GetRegion()
	imageReference, err := k.imageFetcher.FetchReference(ctx, provider, attestationVariant, conf.Image, region)
	if err != nil {
		return fmt.Errorf("fetching image reference: %w", err)
	}

	imageVersion, err := versionsapi.NewVersionFromShortPath(conf.Image, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("parsing version from image short path: %w", err)
	}

	nodeVersion, err := k.getConstellationVersion(ctx)
	if err != nil {
		return err
	}

	upgradeErrs := []error{}
	var upgradeErr *compatibility.InvalidUpgradeError
	if !skipImage {
		err = k.isValidImageUpgrade(nodeVersion, imageVersion.Version(), force)
		switch {
		case errors.As(err, &upgradeErr):
			upgradeErrs = append(upgradeErrs, fmt.Errorf("skipping image upgrades: %w", err))
		case err != nil:
			return fmt.Errorf("updating image version: %w", err)
		}

		// TODO(3u13r): remove `reconcileKubeadmConfigMap` after v2.14.0 has been released.
		if err := k.reconcileKubeadmConfigMap(ctx); err != nil {
			return fmt.Errorf("reconciling kubeadm config: %w", err)
		}

		k.log.Debugf("Updating local copy of nodeVersion image version from %s to %s", nodeVersion.Spec.ImageVersion, imageVersion.Version())
		nodeVersion.Spec.ImageReference = imageReference
		nodeVersion.Spec.ImageVersion = imageVersion.Version()
	}

	if !skipK8s {
		// We have to allow users to specify outdated k8s patch versions.
		// Therefore, this code has to skip k8s updates if a user configures an outdated (i.e. invalid) k8s version.
		var components *corev1.ConfigMap
		_, err = versions.NewValidK8sVersion(string(conf.KubernetesVersion), true)
		if err != nil {
			innerErr := fmt.Errorf("unsupported Kubernetes version, supported versions are %s",
				strings.Join(versions.SupportedK8sVersions(), ", "))
			err = compatibility.NewInvalidUpgradeError(nodeVersion.Spec.KubernetesClusterVersion,
				string(conf.KubernetesVersion), innerErr)
		} else {
			versionConfig, ok := versions.VersionConfigs[conf.KubernetesVersion]
			if !ok {
				err = compatibility.NewInvalidUpgradeError(nodeVersion.Spec.KubernetesClusterVersion,
					string(conf.KubernetesVersion), fmt.Errorf("no version config matching K8s %s", conf.KubernetesVersion))
			} else {
				components, err = k.prepareUpdateK8s(&nodeVersion, versionConfig.ClusterVersion,
					versionConfig.KubernetesComponents, force)
			}
		}

		switch {
		case err == nil:
			err := k.applyComponentsCM(ctx, components)
			if err != nil {
				return fmt.Errorf("applying k8s components ConfigMap: %w", err)
			}
		case errors.As(err, &upgradeErr):
			upgradeErrs = append(upgradeErrs, fmt.Errorf("skipping Kubernetes upgrades: %w", err))
		default:
			return fmt.Errorf("updating Kubernetes version: %w", err)
		}
	}
	if len(upgradeErrs) == 2 {
		return errors.Join(upgradeErrs...)
	}

	updatedNodeVersion, err := k.applyNodeVersion(ctx, nodeVersion)
	if err != nil {
		return fmt.Errorf("applying upgrade: %w", err)
	}

	if err := checkForApplyError(nodeVersion, updatedNodeVersion); err != nil {
		return err
	}
	return errors.Join(upgradeErrs...)
}

// ClusterStatus returns a map from node name to NodeStatus.
func (k *KubeCmd) ClusterStatus(ctx context.Context) (map[string]NodeStatus, error) {
	nodes, err := k.kubectl.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting nodes: %w", err)
	}

	clusterStatus := map[string]NodeStatus{}
	for _, node := range nodes {
		clusterStatus[node.ObjectMeta.Name] = NewNodeStatus(node)
	}

	return clusterStatus, nil
}

// GetClusterAttestationConfig fetches the join-config configmap from the cluster, extracts the config
// and returns both the full configmap and the attestation config.
func (k *KubeCmd) GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error) {
	existingConf, err := retryGetJoinConfig(ctx, k.kubectl, k.retryInterval, k.log)
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

// ApplyJoinConfig creates or updates the Constellation cluster's join-config ConfigMap.
// This ConfigMap holds the attestation config and measurement salt of the cluster.
// A backup of the previous attestation config is created with the suffix `_backup` in the config map data.
func (k *KubeCmd) ApplyJoinConfig(ctx context.Context, newAttestConfig config.AttestationCfg, measurementSalt []byte) error {
	newConfigJSON, err := json.Marshal(newAttestConfig)
	if err != nil {
		return fmt.Errorf("marshaling attestation config: %w", err)
	}

	joinConfig, err := retryGetJoinConfig(ctx, k.kubectl, k.retryInterval, k.log)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("getting %s ConfigMap: %w", constants.JoinConfigMap, err)
		}

		k.log.Debugf("ConfigMap %q does not exist in namespace %q, creating it now", constants.JoinConfigMap, constants.ConstellationNamespace)
		if err := retryAction(ctx, k.retryInterval, maxRetryAttempts, func(ctx context.Context) error {
			return k.kubectl.CreateConfigMap(ctx, joinConfigMap(newConfigJSON, measurementSalt))
		}, k.log); err != nil {
			return fmt.Errorf("creating join-config ConfigMap: %w", err)
		}
		k.log.Debugf("Created %q ConfigMap in namespace %q", constants.JoinConfigMap, constants.ConstellationNamespace)
		return nil
	}

	// create backup of previous config
	joinConfig.Data[constants.AttestationConfigFilename+"_backup"] = joinConfig.Data[constants.AttestationConfigFilename]
	joinConfig.Data[constants.AttestationConfigFilename] = string(newConfigJSON)
	k.log.Debugf("Triggering attestation config update now")
	if err := retryAction(ctx, k.retryInterval, maxRetryAttempts, func(ctx context.Context) error {
		_, err = k.kubectl.UpdateConfigMap(ctx, joinConfig)
		return err
	}, k.log); err != nil {
		return fmt.Errorf("setting new attestation config: %w", err)
	}

	return nil
}

// ExtendClusterConfigCertSANs extends the ClusterConfig stored under "kube-system/kubeadm-config" with the given SANs.
// Existing SANs are preserved.
func (k *KubeCmd) ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error {
	clusterConfiguration, kubeadmConfig, err := k.getClusterConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("getting ClusterConfig: %w", err)
	}

	existingSANs := make(map[string]struct{})
	for _, existingSAN := range clusterConfiguration.APIServer.CertSANs {
		existingSANs[existingSAN] = struct{}{}
	}

	var missingSANs []string
	for _, san := range alternativeNames {
		if san == "" {
			continue // skip empty SANs
		}
		if _, ok := existingSANs[san]; !ok {
			missingSANs = append(missingSANs, san)
			existingSANs[san] = struct{}{} // make sure we don't add the same SAN twice
		}
	}

	if len(missingSANs) == 0 {
		k.log.Debugf("No new SANs to add to the cluster's apiserver SAN field")
		return nil
	}
	k.log.Debugf("Extending the cluster's apiserver SAN field with the following SANs: %s\n", strings.Join(missingSANs, ", "))

	clusterConfiguration.APIServer.CertSANs = append(clusterConfiguration.APIServer.CertSANs, missingSANs...)
	sort.Strings(clusterConfiguration.APIServer.CertSANs)

	newConfigYAML, err := yaml.Marshal(clusterConfiguration)
	if err != nil {
		return fmt.Errorf("marshaling ClusterConfiguration: %w", err)
	}

	kubeadmConfig.Data[constants.ClusterConfigurationKey] = string(newConfigYAML)
	k.log.Debugf("Triggering kubeadm config update now")
	if _, err = k.kubectl.UpdateConfigMap(ctx, kubeadmConfig); err != nil {
		return fmt.Errorf("setting new kubeadm config: %w", err)
	}

	fmt.Fprintln(k.outWriter, "Successfully extended the cluster's apiserver SAN field")
	return nil
}

// GetConstellationVersion retrieves the Kubernetes and image version of a Constellation cluster,
// as well as the Kubernetes components reference, and image reference string.
func (k *KubeCmd) GetConstellationVersion(ctx context.Context) (NodeVersion, error) {
	nV, err := k.getConstellationVersion(ctx)
	if err != nil {
		return NodeVersion{}, fmt.Errorf("retrieving Constellation version: %w", err)
	}

	return NewNodeVersion(nV)
}

// getConstellationVersion returns the NodeVersion object of a Constellation cluster.
func (k *KubeCmd) getConstellationVersion(ctx context.Context) (updatev1alpha1.NodeVersion, error) {
	raw, err := k.kubectl.GetCR(ctx, schema.GroupVersionResource{
		Group:    "update.edgeless.systems",
		Version:  "v1alpha1",
		Resource: "nodeversions",
	}, "constellation-version")
	if err != nil {
		return updatev1alpha1.NodeVersion{}, err
	}
	var nodeVersion updatev1alpha1.NodeVersion
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.UnstructuredContent(), &nodeVersion); err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("converting unstructured to NodeVersion: %w", err)
	}

	return nodeVersion, nil
}

// getClusterConfiguration fetches the kubeadm-config configmap from the cluster, extracts the config
// and returns both the full configmap and the ClusterConfiguration.
func (k *KubeCmd) getClusterConfiguration(ctx context.Context) (kubeadmv1beta3.ClusterConfiguration, *corev1.ConfigMap, error) {
	existingConf, err := k.kubectl.GetConfigMap(ctx, constants.ConstellationNamespace, constants.KubeadmConfigMap)
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
func (k *KubeCmd) applyComponentsCM(ctx context.Context, components *corev1.ConfigMap) error {
	// If the map already exists we can use that map and assume it has the same content as 'configMap'.
	if err := k.kubectl.CreateConfigMap(ctx, components); err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating k8s-components ConfigMap: %w. %T", err, err)
	}
	return nil
}

func (k *KubeCmd) applyNodeVersion(ctx context.Context, nodeVersion updatev1alpha1.NodeVersion) (updatev1alpha1.NodeVersion, error) {
	k.log.Debugf("Triggering NodeVersion upgrade now")
	var updatedNodeVersion updatev1alpha1.NodeVersion
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newNode, err := k.getConstellationVersion(ctx)
		if err != nil {
			return fmt.Errorf("retrieving current NodeVersion: %w", err)
		}

		updateNodeVersions(nodeVersion, &newNode)

		raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&newNode)
		if err != nil {
			return fmt.Errorf("converting nodeVersion to unstructured: %w", err)
		}
		updated, err := k.kubectl.UpdateCR(ctx, schema.GroupVersionResource{
			Group:    "update.edgeless.systems",
			Version:  "v1alpha1",
			Resource: "nodeversions",
		}, &unstructured.Unstructured{Object: raw})
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

func (k *KubeCmd) reconcileKubeadmConfigMap(ctx context.Context) error {
	clusterConfiguration, kubeadmConfig, err := k.getClusterConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("getting ClusterConfig: %w", err)
	}

	for i, v := range clusterConfiguration.APIServer.ExtraVolumes {
		if v.Name == "konnectivity-uds" {
			clusterConfiguration.APIServer.ExtraVolumes = slices.Delete(clusterConfiguration.APIServer.ExtraVolumes, i, i+1)
		}
	}
	for i, v := range clusterConfiguration.APIServer.ExtraVolumes {
		if v.Name == "egress-config" {
			clusterConfiguration.APIServer.ExtraVolumes = slices.Delete(clusterConfiguration.APIServer.ExtraVolumes, i, i+1)
		}
	}
	delete(clusterConfiguration.APIServer.ExtraArgs, "egress-selector-config-file")

	newConfigYAML, err := yaml.Marshal(clusterConfiguration)
	if err != nil {
		return fmt.Errorf("marshaling ClusterConfiguration: %w", err)
	}

	if kubeadmConfig.Data[constants.ClusterConfigurationKey] == string(newConfigYAML) {
		k.log.Debugf("No changes to kubeadm config required")
		return nil
	}

	kubeadmConfig.Data[constants.ClusterConfigurationKey] = string(newConfigYAML)
	k.log.Debugf("Triggering kubeadm config update now")
	if _, err = k.kubectl.UpdateConfigMap(ctx, kubeadmConfig); err != nil {
		return fmt.Errorf("setting new kubeadm config: %w", err)
	}

	fmt.Fprintln(k.outWriter, "Successfully reconciled the cluster's kubeadm config")
	return nil
}

// isValidImageUpdate checks if the new image version is a valid upgrade, and there is no upgrade already running.
func (k *KubeCmd) isValidImageUpgrade(nodeVersion updatev1alpha1.NodeVersion, newImageVersion string, force bool) error {
	if !force {
		// check if an image update is already in progress
		if nodeVersion.Status.ActiveClusterVersionUpgrade {
			return ErrInProgress
		}
		for _, condition := range nodeVersion.Status.Conditions {
			if condition.Type == updatev1alpha1.ConditionOutdated && condition.Status == metav1.ConditionTrue {
				return ErrInProgress
			}
		}

		// check if the image upgrade is valid for the current version
		if err := compatibility.IsValidUpgrade(nodeVersion.Spec.ImageVersion, newImageVersion); err != nil {
			return fmt.Errorf("validating image update: %w", err)
		}
	}
	return nil
}

func (k *KubeCmd) prepareUpdateK8s(nodeVersion *updatev1alpha1.NodeVersion, newClusterVersion string, components components.Components, force bool) (*corev1.ConfigMap, error) {
	configMap, err := internalk8s.ConstructK8sComponentsCM(components, newClusterVersion)
	if err != nil {
		return nil, fmt.Errorf("constructing k8s-components ConfigMap: %w", err)
	}
	if !force {
		if err := compatibility.IsValidUpgrade(nodeVersion.Spec.KubernetesClusterVersion, newClusterVersion); err != nil {
			return nil, err
		}
	}

	k.log.Debugf("Updating local copy of nodeVersion Kubernetes version from %s to %s", nodeVersion.Spec.KubernetesClusterVersion, newClusterVersion)
	nodeVersion.Spec.KubernetesComponentsReference = configMap.ObjectMeta.Name
	nodeVersion.Spec.KubernetesClusterVersion = newClusterVersion

	return &configMap, nil
}

func checkForApplyError(expected, actual updatev1alpha1.NodeVersion) error {
	var err error
	switch {
	case actual.Spec.ImageReference != expected.Spec.ImageReference:
		err = &applyError{expected: expected.Spec.ImageReference, actual: actual.Spec.ImageReference}
	case actual.Spec.ImageVersion != expected.Spec.ImageVersion:
		err = &applyError{expected: expected.Spec.ImageVersion, actual: actual.Spec.ImageVersion}
	case actual.Spec.KubernetesComponentsReference != expected.Spec.KubernetesComponentsReference:
		err = &applyError{expected: expected.Spec.KubernetesComponentsReference, actual: actual.Spec.KubernetesComponentsReference}
	case actual.Spec.KubernetesClusterVersion != expected.Spec.KubernetesClusterVersion:
		err = &applyError{expected: expected.Spec.KubernetesClusterVersion, actual: actual.Spec.KubernetesClusterVersion}
	}
	return err
}

func joinConfigMap(attestationCfgJSON, measurementSalt []byte) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.JoinConfigMap,
			Namespace: constants.ConstellationNamespace,
		},
		Data: map[string]string{
			constants.AttestationConfigFilename: string(attestationCfgJSON),
		},
		BinaryData: map[string][]byte{
			constants.MeasurementSaltFilename: measurementSalt,
		},
	}
}

type kubeDoer struct {
	action func(ctx context.Context) error
}

func (k *kubeDoer) Do(ctx context.Context) error {
	return k.action(ctx)
}

func retryGetJoinConfig(ctx context.Context, kubectl kubectlInterface, retryInterval time.Duration, log debugLog) (*corev1.ConfigMap, error) {
	var retries int
	retrieable := func(err error) bool {
		if k8serrors.IsNotFound(err) {
			return false
		}
		retries++
		log.Debugf("Getting join-config ConfigMap failed (attempt %d/%d): %s", retries, maxRetryAttempts, err)
		return retries < maxRetryAttempts
	}

	var joinConfig *corev1.ConfigMap
	var err error
	doer := &kubeDoer{
		action: func(ctx context.Context) error {
			joinConfig, err = kubectl.GetConfigMap(ctx, constants.ConstellationNamespace, constants.JoinConfigMap)
			return err
		},
	}
	retrier := conretry.NewIntervalRetrier(doer, retryInterval, retrieable)

	err = retrier.Do(ctx)
	return joinConfig, err
}

func retryAction(ctx context.Context, retryInterval time.Duration, maxRetries int, action func(ctx context.Context) error, log debugLog) error {
	ctr := 0
	retrier := conretry.NewIntervalRetrier(&kubeDoer{action: action}, retryInterval, func(err error) bool {
		ctr++
		log.Debugf("Action failed (attempt %d/%d): %s", ctr, maxRetries, err)
		return ctr < maxRetries
	})
	return retrier.Do(ctx)
}

// kubectlInterface provides access to the Kubernetes API.
type kubectlInterface interface {
	GetNodes(ctx context.Context) ([]corev1.Node, error)
	GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error)
	UpdateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) error
	KubernetesVersion() (string, error)
	GetCR(ctx context.Context, gvr schema.GroupVersionResource, name string) (*unstructured.Unstructured, error)
	UpdateCR(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
	crdLister
}

type debugLog interface {
	Debugf(format string, args ...any)
}

// imageFetcher gets an image reference from the versionsapi.
type imageFetcher interface {
	FetchReference(ctx context.Context,
		provider cloudprovider.Provider, attestationVariant variant.Variant,
		image, region string,
	) (string, error)
}

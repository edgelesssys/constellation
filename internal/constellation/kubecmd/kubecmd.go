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
	"sort"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	internalk8s "github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	conretry "github.com/edgelesssys/constellation/v2/internal/retry"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/util/retry"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmv1beta4 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta4"
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
	retryInterval time.Duration
	maxAttempts   int
	log           debugLog
}

// New returns a new KubeCmd.
func New(kubeConfig []byte, log debugLog) (*KubeCmd, error) {
	client, err := kubectl.NewFromConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("creating kubectl client: %w", err)
	}

	return &KubeCmd{
		kubectl:       client,
		retryInterval: time.Second * 5,
		maxAttempts:   20,
		log:           log,
	}, nil
}

// UpgradeNodeImage upgrades the image version of a Constellation cluster.
func (k *KubeCmd) UpgradeNodeImage(ctx context.Context, imageVersion semver.Semver, imageReference string, force bool) error {
	nodeVersion, err := k.getConstellationVersion(ctx)
	if err != nil {
		return err
	}

	k.log.Debug("Checking if image upgrade is valid")
	var upgradeErr *compatibility.InvalidUpgradeError
	err = k.isValidImageUpgrade(nodeVersion, imageVersion.String(), force)
	switch {
	case errors.As(err, &upgradeErr):
		return fmt.Errorf("skipping image upgrade: %w", err)
	case err != nil:
		return fmt.Errorf("updating image version: %w", err)
	}

	k.log.Debug("Updating local copy of nodeVersion image version", "oldVersion", nodeVersion.Spec.ImageVersion, "newVersion", imageVersion.String())
	nodeVersion.Spec.ImageReference = imageReference
	nodeVersion.Spec.ImageVersion = imageVersion.String()

	updatedNodeVersion, err := k.applyNodeVersion(ctx, nodeVersion)
	if err != nil {
		return fmt.Errorf("applying upgrade: %w", err)
	}
	return checkForApplyError(nodeVersion, updatedNodeVersion)
}

// UpgradeKubernetesVersion upgrades the Kubernetes version of a Constellation cluster.
func (k *KubeCmd) UpgradeKubernetesVersion(ctx context.Context, kubernetesVersion versions.ValidK8sVersion, force bool) error {
	nodeVersion, err := k.getConstellationVersion(ctx)
	if err != nil {
		return err
	}

	// We have to allow users to specify outdated k8s patch versions.
	// Therefore, this code has to skip k8s updates if a user configures an outdated (i.e. invalid) k8s version.
	if _, err := versions.NewValidK8sVersion(string(kubernetesVersion), true); err != nil {
		return fmt.Errorf("skipping Kubernetes upgrade: %w", compatibility.NewInvalidUpgradeError(
			nodeVersion.Spec.KubernetesClusterVersion,
			string(kubernetesVersion),
			fmt.Errorf("unsupported Kubernetes version, supported versions are %s", strings.Join(versions.SupportedK8sVersions(), ", "))),
		)
	}

	// TODO(burgerdev): remove after releasing v2.19
	// Workaround for https://github.com/kubernetes/kubernetes/issues/127316: force kubelet to
	// connect to the local API server.
	if err := k.patchKubeadmConfig(ctx, func(cc *kubeadm.ClusterConfiguration) {
		if cc.FeatureGates == nil {
			cc.FeatureGates = map[string]bool{}
		}
		cc.FeatureGates["ControlPlaneKubeletLocalMode"] = true
	}); err != nil {
		return fmt.Errorf("setting FeatureGate ControlPlaneKubeletLocalMode: %w", err)
	}

	versionConfig, ok := versions.VersionConfigs[kubernetesVersion]
	if !ok {
		return fmt.Errorf("skipping Kubernetes upgrade: %w", compatibility.NewInvalidUpgradeError(
			nodeVersion.Spec.KubernetesClusterVersion,
			string(kubernetesVersion),
			fmt.Errorf("no version config matching K8s %s", kubernetesVersion),
		))
	}
	components, err := k.prepareUpdateK8s(&nodeVersion, versionConfig.ClusterVersion, versionConfig.KubernetesComponents, force)
	if err != nil {
		return err
	}

	if err := k.applyComponentsCM(ctx, components); err != nil {
		return fmt.Errorf("applying k8s components ConfigMap: %w", err)
	}

	updatedNodeVersion, err := k.applyNodeVersion(ctx, nodeVersion)
	if err != nil {
		return fmt.Errorf("applying upgrade: %w", err)
	}
	return checkForApplyError(nodeVersion, updatedNodeVersion)
}

// ClusterStatus returns a map from node name to NodeStatus.
func (k *KubeCmd) ClusterStatus(ctx context.Context) (map[string]NodeStatus, error) {
	var nodes []corev1.Node

	if err := k.retryAction(ctx, func(ctx context.Context) error {
		var err error
		nodes, err = k.kubectl.GetNodes(ctx)
		return err
	}); err != nil {
		return nil, fmt.Errorf("getting nodes: %w", err)
	}

	clusterStatus := map[string]NodeStatus{}
	for _, node := range nodes {
		clusterStatus[node.ObjectMeta.Name] = NewNodeStatus(node)
	}

	return clusterStatus, nil
}

// GetClusterAttestationConfig fetches the join-config configmap from the cluster,
// and returns the attestation config.
func (k *KubeCmd) GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error) {
	existingConf, err := k.retryGetJoinConfig(ctx)
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

	joinConfig, err := k.retryGetJoinConfig(ctx)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("getting %s ConfigMap: %w", constants.JoinConfigMap, err)
		}

		k.log.Debug("ConfigMap does not exist, creating it now", "name", constants.JoinConfigMap, "namespace", constants.ConstellationNamespace)
		if err := k.retryAction(ctx, func(ctx context.Context) error {
			return k.kubectl.CreateConfigMap(ctx, joinConfigMap(newConfigJSON, measurementSalt))
		}); err != nil {
			return fmt.Errorf("creating join-config ConfigMap: %w", err)
		}
		k.log.Debug("Created ConfigMap", "name", constants.JoinConfigMap, "namespace", constants.ConstellationNamespace)
		return nil
	}

	// create backup of previous config
	joinConfig.Data[constants.AttestationConfigFilename+"_backup"] = joinConfig.Data[constants.AttestationConfigFilename]
	joinConfig.Data[constants.AttestationConfigFilename] = string(newConfigJSON)
	k.log.Debug("Triggering attestation config update now")
	if err := k.retryAction(ctx, func(ctx context.Context) error {
		_, err = k.kubectl.UpdateConfigMap(ctx, joinConfig)
		return err
	}); err != nil {
		return fmt.Errorf("setting new attestation config: %w", err)
	}

	return nil
}

// ExtendClusterConfigCertSANs extends the ClusterConfig stored under "kube-system/kubeadm-config" with the given SANs.
// Empty strings are ignored, existing SANs are preserved.
func (k *KubeCmd) ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error {
	if err := k.patchKubeadmConfig(ctx, func(clusterConfiguration *kubeadm.ClusterConfiguration) {
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
			k.log.Debug("No new SANs to add to the cluster's apiserver SAN field")
		}
		k.log.Debug("Extending the cluster's apiserver SAN field", "certSANs", strings.Join(missingSANs, ", "))

		clusterConfiguration.APIServer.CertSANs = append(clusterConfiguration.APIServer.CertSANs, missingSANs...)
		sort.Strings(clusterConfiguration.APIServer.CertSANs)
	}); err != nil {
		return fmt.Errorf("extending ClusterConfig.CertSANs: %w", err)
	}

	k.log.Debug("Successfully extended the cluster's apiserver SAN field")
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
	var raw *unstructured.Unstructured
	if err := k.retryAction(ctx, func(ctx context.Context) error {
		var err error
		raw, err = k.kubectl.GetCR(ctx, schema.GroupVersionResource{
			Group:    "update.edgeless.systems",
			Version:  "v1alpha1",
			Resource: "nodeversions",
		}, constants.NodeVersionResourceName)
		return err
	}); err != nil {
		return updatev1alpha1.NodeVersion{}, err
	}

	var nodeVersion updatev1alpha1.NodeVersion
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.UnstructuredContent(), &nodeVersion); err != nil {
		return updatev1alpha1.NodeVersion{}, fmt.Errorf("converting unstructured to NodeVersion: %w", err)
	}

	return nodeVersion, nil
}

// applyComponentsCM applies the k8s components ConfigMap to the cluster.
func (k *KubeCmd) applyComponentsCM(ctx context.Context, components *corev1.ConfigMap) error {
	if err := k.retryAction(ctx, func(ctx context.Context) error {
		// If the components ConfigMap already exists we assume it is up to date,
		// since its name is derived from a hash of its contents.
		err := k.kubectl.CreateConfigMap(ctx, components)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("creating k8s-components ConfigMap: %w", err)
	}
	return nil
}

func (k *KubeCmd) applyNodeVersion(ctx context.Context, nodeVersion updatev1alpha1.NodeVersion) (updatev1alpha1.NodeVersion, error) {
	k.log.Debug("Triggering NodeVersion upgrade now")
	var updatedNodeVersion updatev1alpha1.NodeVersion

	// Retry the entire "retry-on-conflict" block to retry if the block fails, e.g. due to etcd timeouts.
	err := k.retryAction(ctx, func(ctx context.Context) error {
		return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
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
	})

	return updatedNodeVersion, err
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
			return nil, fmt.Errorf("skipping Kubernetes upgrade: %w", err)
		}
	}

	k.log.Debug("Updating local copy of nodeVersion Kubernetes version", "oldVersion", nodeVersion.Spec.KubernetesClusterVersion, "newVersion", newClusterVersion)
	nodeVersion.Spec.KubernetesComponentsReference = configMap.ObjectMeta.Name
	nodeVersion.Spec.KubernetesClusterVersion = newClusterVersion

	return &configMap, nil
}

func (k *KubeCmd) retryGetJoinConfig(ctx context.Context) (*corev1.ConfigMap, error) {
	var ctr int
	retrieable := func(err error) bool {
		if k8serrors.IsNotFound(err) {
			return false
		}
		ctr++
		k.log.Debug("Getting join-config ConfigMap failed", "attempt", ctr, "maxAttempts", k.maxAttempts, "error", err)
		return ctr < k.maxAttempts
	}

	var joinConfig *corev1.ConfigMap
	var err error
	doer := &kubeDoer{
		action: func(ctx context.Context) error {
			joinConfig, err = k.kubectl.GetConfigMap(ctx, constants.ConstellationNamespace, constants.JoinConfigMap)
			return err
		},
	}
	retrier := conretry.NewIntervalRetrier(doer, k.retryInterval, retrieable)

	err = retrier.Do(ctx)
	return joinConfig, err
}

func (k *KubeCmd) retryAction(ctx context.Context, action func(ctx context.Context) error) error {
	ctr := 0
	retrier := conretry.NewIntervalRetrier(&kubeDoer{action: action}, k.retryInterval, func(err error) bool {
		ctr++
		k.log.Debug("Action failed", "attempt", ctr, "maxAttempts", k.maxAttempts, "error", err)
		return ctr < k.maxAttempts
	})
	return retrier.Do(ctx)
}

// patchKubeadmConfig fetches and unpacks the kube-system/kubeadm-config ClusterConfiguration entry,
// runs doPatch on it and uploads the result.
func (k *KubeCmd) patchKubeadmConfig(ctx context.Context, doPatch func(*kubeadm.ClusterConfiguration)) error {
	var kubeadmConfig *corev1.ConfigMap
	if err := k.retryAction(ctx, func(ctx context.Context) error {
		var err error
		kubeadmConfig, err = k.kubectl.GetConfigMap(ctx, constants.ConstellationNamespace, constants.KubeadmConfigMap)
		return err
	}); err != nil {
		return fmt.Errorf("retrieving current kubeadm-config: %w", err)
	}

	clusterConfigData, ok := kubeadmConfig.Data[constants.ClusterConfigurationKey]
	if !ok {
		return errors.New("ClusterConfiguration missing from kubeadm-config")
	}

	var clusterConfiguration kubeadm.ClusterConfiguration
	if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), []byte(clusterConfigData), &clusterConfiguration); err != nil {
		return fmt.Errorf("decoding cluster configuration data: %w", err)
	}

	doPatch(&clusterConfiguration)

	opt := k8sjson.SerializerOptions{Yaml: true}
	serializer := k8sjson.NewSerializerWithOptions(k8sjson.DefaultMetaFactory, kubeadmscheme.Scheme, kubeadmscheme.Scheme, opt)
	encoder := kubeadmscheme.Codecs.EncoderForVersion(serializer, kubeadmv1beta4.SchemeGroupVersion)
	newConfigYAML, err := runtime.Encode(encoder, &clusterConfiguration)
	if err != nil {
		return fmt.Errorf("marshaling ClusterConfiguration: %w", err)
	}

	kubeadmConfig.Data[constants.ClusterConfigurationKey] = string(newConfigYAML)
	k.log.Debug("Triggering kubeadm config update now")
	if err = k.retryAction(ctx, func(ctx context.Context) error {
		_, err := k.kubectl.UpdateConfigMap(ctx, kubeadmConfig)
		return err
	}); err != nil {
		return fmt.Errorf("setting new kubeadm config: %w", err)
	}

	k.log.Debug("Successfully patched the cluster's kubeadm-config")
	return nil
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
	Debug(msg string, args ...any)
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// ExtendClusterConfigCertSANs extends the ClusterConfig stored under "kube-system/kubeadm-config" with the given SANs.
func (a *Applier) ExtendClusterConfigCertSANs(ctx context.Context, clusterEndpoint, customEndpoint string, additionalAPIServerCertSANs []string) error {
	client, err := a.getKubecmdClient(a.kubeConfig, a.log)
	if err != nil {
		return err
	}

	sans := append([]string{clusterEndpoint, customEndpoint}, additionalAPIServerCertSANs...)
	if err := client.ExtendClusterConfigCertSANs(ctx, sans); err != nil {
		return fmt.Errorf("extending cert SANs: %w", err)
	}
	return nil
}

// GetClusterAttestationConfig returns the attestation config currently set for the cluster.
func (a *Applier) GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error) {
	client, err := a.getKubecmdClient(a.kubeConfig, a.log)
	if err != nil {
		return nil, err
	}

	return client.GetClusterAttestationConfig(ctx, variant)
}

// ApplyJoinConfig creates or updates the Constellation cluster's join-config ConfigMap.
func (a *Applier) ApplyJoinConfig(ctx context.Context, newAttestConfig config.AttestationCfg, measurementSalt []byte) error {
	client, err := a.getKubecmdClient(a.kubeConfig, a.log)
	if err != nil {
		return err
	}

	return client.ApplyJoinConfig(ctx, newAttestConfig, measurementSalt)
}

// UpgradeNodeImage upgrades the node image of the cluster to the given version.
func (a *Applier) UpgradeNodeImage(ctx context.Context, imageVersion semver.Semver, imageReference string, force bool) error {
	client, err := a.getKubecmdClient(a.kubeConfig, a.log)
	if err != nil {
		return err
	}

	return client.UpgradeNodeImage(ctx, imageVersion, imageReference, force)
}

// UpgradeKubernetesVersion upgrades the Kubernetes version of the cluster to the given version.
func (a *Applier) UpgradeKubernetesVersion(ctx context.Context, kubernetesVersion versions.ValidK8sVersion, force bool) error {
	client, err := a.getKubecmdClient(a.kubeConfig, a.log)
	if err != nil {
		return err
	}

	return client.UpgradeKubernetesVersion(ctx, kubernetesVersion, force)
}

// BackupCRDs backs up all CRDs to the upgrade workspace.
func (a *Applier) BackupCRDs(ctx context.Context, fileHandler file.Handler, upgradeDir string) ([]apiextensionsv1.CustomResourceDefinition, error) {
	client, err := a.getKubecmdClient(a.kubeConfig, a.log)
	if err != nil {
		return nil, err
	}

	return client.BackupCRDs(ctx, fileHandler, upgradeDir)
}

// BackupCRs backs up all CRs to the upgrade workspace.
func (a *Applier) BackupCRs(ctx context.Context, fileHandler file.Handler, crds []apiextensionsv1.CustomResourceDefinition, upgradeDir string) error {
	client, err := a.getKubecmdClient(a.kubeConfig, a.log)
	if err != nil {
		return err
	}

	return client.BackupCRs(ctx, fileHandler, crds, upgradeDir)
}

func kubecmdGetter(kubeConfig []byte, log debugLog) (kubecmdClient, error) {
	return kubecmd.New(kubeConfig, log)
}

type kubecmdClient interface {
	UpgradeNodeImage(ctx context.Context, imageVersion semver.Semver, imageReference string, force bool) error
	UpgradeKubernetesVersion(ctx context.Context, kubernetesVersion versions.ValidK8sVersion, force bool) error
	ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error)
	ApplyJoinConfig(ctx context.Context, newAttestConfig config.AttestationCfg, measurementSalt []byte) error
	BackupCRs(ctx context.Context, fileHandler file.Handler, crds []apiextensionsv1.CustomResourceDefinition, upgradeDir string) error
	BackupCRDs(ctx context.Context, fileHandler file.Handler, upgradeDir string) ([]apiextensionsv1.CustomResourceDefinition, error)
}

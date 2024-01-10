/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubecmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/file"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

type crdLister interface {
	ListCRDs(ctx context.Context) ([]apiextensionsv1.CustomResourceDefinition, error)
	ListCRs(ctx context.Context, gvr schema.GroupVersionResource) ([]unstructured.Unstructured, error)
}

// BackupCRDs backs up all CRDs to the upgrade workspace.
func (k *KubeCmd) BackupCRDs(ctx context.Context, fileHandler file.Handler, upgradeDir string) ([]apiextensionsv1.CustomResourceDefinition, error) {
	k.log.Debug("Starting CRD backup")
	crds, err := k.kubectl.ListCRDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting CRDs: %w", err)
	}

	crdBackupFolder := k.crdBackupFolder(upgradeDir)
	if err := fileHandler.MkdirAll(crdBackupFolder); err != nil {
		return nil, fmt.Errorf("creating backup dir: %w", err)
	}
	for i := range crds {
		path := filepath.Join(crdBackupFolder, crds[i].Name+".yaml")

		k.log.Debug("Creating CRD backup: %s", path)

		// We have to manually set kind/apiversion because of a long-standing limitation of the API:
		// https://github.com/kubernetes/kubernetes/issues/3030#issuecomment-67543738
		// The comment states that kind/version are encoded in the type.
		// The package holding the CRD type encodes the version.
		crds[i].Kind = "CustomResourceDefinition"
		crds[i].APIVersion = "apiextensions.k8s.io/v1"

		yamlBytes, err := yaml.Marshal(crds[i])
		if err != nil {
			return nil, err
		}
		if err := fileHandler.Write(path, yamlBytes); err != nil {
			return nil, err
		}
	}
	k.log.Debug("CRD backup complete")
	return crds, nil
}

// BackupCRs backs up all CRs to the upgrade workspace.
func (k *KubeCmd) BackupCRs(ctx context.Context, fileHandler file.Handler, crds []apiextensionsv1.CustomResourceDefinition, upgradeDir string) error {
	k.log.Debug("Starting CR backup")
	for _, crd := range crds {
		k.log.Debug("Creating backup for resource type: %s", crd.Name)

		// Iterate over all versions of the CRD
		// TODO(daniel-weisse): Consider iterating over crd.Status.StoredVersions instead
		// Currently, we have to ignore not-found errors, because a CRD might define
		// a version that is not installed in the cluster.
		// With the StoredVersions field, we could only iterate over the installed versions.
		for _, version := range crd.Spec.Versions {
			k.log.Debug("Creating backup of CRs for %q at version %q", crd.Name, version.Name)

			gvr := schema.GroupVersionResource{Group: crd.Spec.Group, Version: version.Name, Resource: crd.Spec.Names.Plural}
			crs, err := k.kubectl.ListCRs(ctx, gvr)
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return fmt.Errorf("retrieving CR %s: %w", crd.Name, err)
				}
				k.log.Debug("No CRs found for %q at version %q, skipping...", crd.Name, version.Name)
				continue
			}

			backupFolder := k.backupFolder(upgradeDir)
			for _, cr := range crs {
				targetFolder := filepath.Join(backupFolder, gvr.Group, gvr.Version, cr.GetNamespace(), cr.GetKind())
				if err := fileHandler.MkdirAll(targetFolder); err != nil {
					return fmt.Errorf("creating resource dir: %w", err)
				}
				path := filepath.Join(targetFolder, cr.GetName()+".yaml")
				yamlBytes, err := yaml.Marshal(cr.Object)
				if err != nil {
					return err
				}
				if err := fileHandler.Write(path, yamlBytes); err != nil {
					return err
				}
			}
		}

		k.log.Debug("Backup for resource type %q complete", crd.Name)
	}
	k.log.Debug("CR backup complete")
	return nil
}

func (k *KubeCmd) backupFolder(upgradeDir string) string {
	return filepath.Join(upgradeDir, "backups")
}

func (k *KubeCmd) crdBackupFolder(upgradeDir string) string {
	return filepath.Join(k.backupFolder(upgradeDir), "crds")
}

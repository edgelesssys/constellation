/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"
	"path/filepath"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

func (c *UpgradeClient) backupCRDs(ctx context.Context, upgradeID string) ([]apiextensionsv1.CustomResourceDefinition, error) {
	c.log.Debugf("Starting CRD backup")
	crds, err := c.kubectl.ListCRDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting CRDs: %w", err)
	}

	crdBackupFolder := c.crdBackupFolder(upgradeID)
	if err := c.fs.MkdirAll(crdBackupFolder); err != nil {
		return nil, fmt.Errorf("creating backup dir: %w", err)
	}
	for i := range crds {
		path := filepath.Join(crdBackupFolder, crds[i].Name+".yaml")

		c.log.Debugf("Creating CRD backup: %s", path)

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
		if err := c.fs.Write(path, yamlBytes); err != nil {
			return nil, err
		}
	}
	c.log.Debugf("CRD backup complete")
	return crds, nil
}

func (c *UpgradeClient) backupCRs(ctx context.Context, crds []apiextensionsv1.CustomResourceDefinition, upgradeID string) error {
	c.log.Debugf("Starting CR backup")
	for _, crd := range crds {
		c.log.Debugf("Creating backup for resource type: %s", crd.Name)

		// Iterate over all versions of the CRD
		// TODO: Consider iterating over crd.Status.StoredVersions instead
		// Currently, we have to ignore not-found errors, because a CRD might define
		// a version that is not installed in the cluster.
		// With the StoredVersions field, we could only iterate over the installed versions.
		for _, version := range crd.Spec.Versions {
			c.log.Debugf("Creating backup of CRs for %q at version %q", crd.Name, version.Name)

			gvr := schema.GroupVersionResource{Group: crd.Spec.Group, Version: version.Name, Resource: crd.Spec.Names.Plural}
			crs, err := c.kubectl.ListCRs(ctx, gvr)
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return fmt.Errorf("retrieving CR %s: %w", crd.Name, err)
				}
				c.log.Debugf("No CRs found for %q at version %q, skipping...", crd.Name, version.Name)
				continue
			}

			backupFolder := c.backupFolder(upgradeID)
			for _, cr := range crs {
				targetFolder := filepath.Join(backupFolder, gvr.Group, gvr.Version, cr.GetNamespace(), cr.GetKind())
				if err := c.fs.MkdirAll(targetFolder); err != nil {
					return fmt.Errorf("creating resource dir: %w", err)
				}
				path := filepath.Join(targetFolder, cr.GetName()+".yaml")
				yamlBytes, err := yaml.Marshal(cr.Object)
				if err != nil {
					return err
				}
				if err := c.fs.Write(path, yamlBytes); err != nil {
					return err
				}
			}
		}

		c.log.Debugf("Backup for resource type %q complete", crd.Name)
	}
	c.log.Debugf("CR backup complete")
	return nil
}

func (c *UpgradeClient) backupFolder(upgradeID string) string {
	return filepath.Join(c.upgradeWorkspace, upgradeID, "backups") + string(filepath.Separator)
}

func (c *UpgradeClient) crdBackupFolder(upgradeID string) string {
	return filepath.Join(c.backupFolder(upgradeID), "crds") + string(filepath.Separator)
}

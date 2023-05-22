/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

var (
	backupFolder    = filepath.Join(constants.UpgradeDir, "backups") + string(filepath.Separator)
	crdBackupFolder = filepath.Join(backupFolder, "crds") + string(filepath.Separator)
)

func (c *Client) backupCRDs(ctx context.Context) ([]apiextensionsv1.CustomResourceDefinition, error) {
	crds, err := c.kubectl.GetCRDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting CRDs: %w", err)
	}

	if err := c.fs.MkdirAll(crdBackupFolder); err != nil {
		return nil, fmt.Errorf("creating backup dir: %w", err)
	}
	for i := range crds {
		path := filepath.Join(crdBackupFolder, crds[i].Name+".yaml")

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

		c.log.Debugf("Created backup crd: %s", path)
	}
	return crds, nil
}

func (c *Client) backupCRs(ctx context.Context, crds []apiextensionsv1.CustomResourceDefinition) error {
	for _, crd := range crds {
		for _, version := range crd.Spec.Versions {
			gvr := schema.GroupVersionResource{Group: crd.Spec.Group, Version: version.Name, Resource: crd.Spec.Names.Plural}
			crs, err := c.kubectl.GetCRs(ctx, gvr)
			if err != nil {
				return fmt.Errorf("retrieving CR %s: %w", crd.Name, err)
			}

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

		c.log.Debugf("Created backups for resource type: %s", crd.Name)
	}
	return nil
}

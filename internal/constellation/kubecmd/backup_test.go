/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubecmd

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

func TestBackupCRDs(t *testing.T) {
	testCases := map[string]struct {
		upgradeID    string
		crd          string
		expectedFile string
		getCRDsError error
		wantError    bool
	}{
		"success": {
			upgradeID:    "1234",
			crd:          "apiVersion: \nkind: \nmetadata:\n  name: foobar\n  creationTimestamp: null\nspec:\n  group: \"\"\n  names:\n    kind: \"somename\"\n    plural: \"somenames\"\n  scope: \"\"\n  versions: null\nstatus:\n  acceptedNames:\n    kind: \"\"\n    plural: \"\"\n  conditions: null\n  storedVersions: null\n",
			expectedFile: "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: foobar\n  creationTimestamp: null\nspec:\n  group: \"\"\n  names:\n    kind: \"somename\"\n    plural: \"somenames\"\n  scope: \"\"\n  versions: null\nstatus:\n  acceptedNames:\n    kind: \"\"\n    plural: \"\"\n  conditions: null\n  storedVersions: null\n",
		},
		"api request fails": {
			upgradeID:    "1234",
			getCRDsError: errors.New("api error"),
			wantError:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			memFs := afero.NewMemMapFs()

			crd := apiextensionsv1.CustomResourceDefinition{}
			err := yaml.Unmarshal([]byte(tc.crd), &crd)
			require.NoError(err)
			client := KubeCmd{
				kubectl: &stubKubectl{crds: []apiextensionsv1.CustomResourceDefinition{crd}, getCRDsError: tc.getCRDsError},
				log:     stubLog{},
			}

			_, err = client.BackupCRDs(context.Background(), file.NewHandler(memFs), tc.upgradeID)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			data, err := afero.ReadFile(memFs, filepath.Join(client.crdBackupFolder(tc.upgradeID), crd.Name+".yaml"))
			require.NoError(err)
			assert.YAMLEq(tc.expectedFile, string(data))
		})
	}
}

func TestBackupCRs(t *testing.T) {
	testCases := map[string]struct {
		upgradeID    string
		crd          apiextensionsv1.CustomResourceDefinition
		resource     unstructured.Unstructured
		expectedFile string
		getCRsError  error
		wantError    bool
	}{
		"success": {
			upgradeID: "1234",
			crd: apiextensionsv1.CustomResourceDefinition{
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Names: apiextensionsv1.CustomResourceDefinitionNames{
						Plural: "foobars",
					},
					Group: "some.group",
					Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
						{
							Name: "versionZero",
						},
					},
				},
			},
			resource:     unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "foobar"}}},
			expectedFile: "metadata:\n  name: foobar\n",
		},
		"api request fails": {
			upgradeID: "1234",
			crd: apiextensionsv1.CustomResourceDefinition{
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Names: apiextensionsv1.CustomResourceDefinitionNames{
						Plural: "foobars",
					},
					Group: "some.group",
					Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
						{
							Name: "versionZero",
						},
					},
				},
			},
			getCRsError: errors.New("api error"),
			wantError:   true,
		},
		"custom resource not found": {
			upgradeID: "1234",
			crd: apiextensionsv1.CustomResourceDefinition{
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Names: apiextensionsv1.CustomResourceDefinitionNames{
						Plural: "foobars",
					},
					Group: "some.group",
					Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
						{
							Name: "versionZero",
						},
					},
				},
			},
			getCRsError: k8serrors.NewNotFound(schema.GroupResource{Group: "some.group"}, "foobars"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			memFs := afero.NewMemMapFs()

			client := KubeCmd{
				kubectl: &stubKubectl{crs: []unstructured.Unstructured{tc.resource}, getCRsError: tc.getCRsError},
				log:     stubLog{},
			}

			err := client.BackupCRs(context.Background(), file.NewHandler(memFs), []apiextensionsv1.CustomResourceDefinition{tc.crd}, tc.upgradeID)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			data, err := afero.ReadFile(memFs, filepath.Join(client.backupFolder(tc.upgradeID), tc.crd.Spec.Group, tc.crd.Spec.Versions[0].Name, tc.resource.GetNamespace(), tc.resource.GetKind(), tc.resource.GetName()+".yaml"))
			if tc.expectedFile == "" {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.YAMLEq(tc.expectedFile, string(data))
		})
	}
}

type stubLog struct{}

func (s stubLog) Debugf(_ string, _ ...any) {}
func (s stubLog) Sync()                     {}

func (c stubKubectl) ListCRDs(_ context.Context) ([]apiextensionsv1.CustomResourceDefinition, error) {
	if c.getCRDsError != nil {
		return nil, c.getCRDsError
	}
	return c.crds, nil
}

func (c stubKubectl) ListCRs(_ context.Context, _ schema.GroupVersionResource) ([]unstructured.Unstructured, error) {
	if c.getCRsError != nil {
		return nil, c.getCRsError
	}
	return c.crs, nil
}

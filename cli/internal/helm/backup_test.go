/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

func TestBackupCRDs(t *testing.T) {
	testCases := map[string]struct {
		crd          string
		expectedFile string
		getCRDsError error
		wantError    bool
	}{
		"success": {
			crd:          "apiVersion: \nkind: \nmetadata:\n  name: foobar\n  creationTimestamp: null\nspec:\n  group: \"\"\n  names:\n    kind: \"somename\"\n    plural: \"somenames\"\n  scope: \"\"\n  versions: null\nstatus:\n  acceptedNames:\n    kind: \"\"\n    plural: \"\"\n  conditions: null\n  storedVersions: null\n",
			expectedFile: "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: foobar\n  creationTimestamp: null\nspec:\n  group: \"\"\n  names:\n    kind: \"somename\"\n    plural: \"somenames\"\n  scope: \"\"\n  versions: null\nstatus:\n  acceptedNames:\n    kind: \"\"\n    plural: \"\"\n  conditions: null\n  storedVersions: null\n",
		},
		"api request fails": {
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
			client := Client{
				config:  nil,
				kubectl: stubCrdClient{crds: []apiextensionsv1.CustomResourceDefinition{crd}, getCRDsError: tc.getCRDsError},
				fs:      file.NewHandler(memFs),
				log:     stubLog{},
			}

			_, err = client.backupCRDs(context.Background())
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			data, err := afero.ReadFile(memFs, filepath.Join(crdBackupFolder, crd.Name+".yaml"))
			require.NoError(err)
			assert.YAMLEq(tc.expectedFile, string(data))
		})
	}
}

func TestBackupCRs(t *testing.T) {
	testCases := map[string]struct {
		crd          apiextensionsv1.CustomResourceDefinition
		resource     unstructured.Unstructured
		expectedFile string
		getCRsError  error
		wantError    bool
	}{
		"success": {
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
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			memFs := afero.NewMemMapFs()

			client := Client{
				config:  nil,
				kubectl: stubCrdClient{crs: []unstructured.Unstructured{tc.resource}, getCRsError: tc.getCRsError},
				fs:      file.NewHandler(memFs),
				log:     stubLog{},
			}

			err := client.backupCRs(context.Background(), []apiextensionsv1.CustomResourceDefinition{tc.crd})
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			data, err := afero.ReadFile(memFs, filepath.Join(backupFolder, tc.resource.GetName()+".yaml"))
			require.NoError(err)
			assert.YAMLEq(tc.expectedFile, string(data))
		})
	}
}

type stubLog struct{}

func (s stubLog) Debugf(_ string, _ ...any) {}
func (s stubLog) Sync()                     {}

type stubCrdClient struct {
	crds         []apiextensionsv1.CustomResourceDefinition
	getCRDsError error
	crs          []unstructured.Unstructured
	getCRsError  error
	crdClient
}

func (c stubCrdClient) GetCRDs(_ context.Context) ([]apiextensionsv1.CustomResourceDefinition, error) {
	if c.getCRDsError != nil {
		return nil, c.getCRDsError
	}
	return c.crds, nil
}

func (c stubCrdClient) GetCRs(_ context.Context, _ schema.GroupVersionResource) ([]unstructured.Unstructured, error) {
	if c.getCRsError != nil {
		return nil, c.getCRsError
	}
	return c.crs, nil
}

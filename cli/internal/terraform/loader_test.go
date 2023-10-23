/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareWorkspace(t *testing.T) {
	testCases := map[string]struct {
		pathBase            string
		provider            cloudprovider.Provider
		fileList            []string
		testAlreadyUnpacked bool
	}{
		"awsCluster": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.AWS,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
		},
		"gcpCluster": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.GCP,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
		},
		"qemuCluster": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
		},
		"gcpIAM": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.GCP,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
		},
		"azureIAM": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.Azure,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
		},
		"awsIAM": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.AWS,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
		},
		"continue on (partially) unpacked": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.AWS,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
			testAlreadyUnpacked: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			file := file.NewHandler(afero.NewMemMapFs())
			testWorkspace := "unittest"

			path := path.Join(tc.pathBase, strings.ToLower(tc.provider.String()))
			err := prepareWorkspace(path, file, testWorkspace)

			require.NoError(err)
			checkFiles(t, file, func(err error) { assert.NoError(err) }, nil, testWorkspace, tc.fileList)

			if tc.testAlreadyUnpacked {
				// Let's try the same again and check if we don't get a "file already exists" error.
				require.NoError(file.Remove(filepath.Join(testWorkspace, "variables.tf")))
				err := prepareWorkspace(path, file, testWorkspace)
				assert.NoError(err)
				checkFiles(t, file, func(err error) { assert.NoError(err) }, nil, testWorkspace, tc.fileList)
			}

			err = cleanUpWorkspace(file, testWorkspace)
			require.NoError(err)

			checkFiles(t, file, func(err error) { assert.ErrorIs(err, fs.ErrNotExist) }, nil, testWorkspace, tc.fileList)
		})
	}
}

func checkFiles(t *testing.T, fileHandler file.Handler, assertion func(error), contentExpection func(content []byte), dir string, files []string) {
	t.Helper()
	for _, f := range files {
		path := filepath.Join(dir, f)
		_, err := fileHandler.Stat(path)
		assertion(err)
		if err == nil {
			content, err := fileHandler.Read(path)
			assertion(err)
			if contentExpection != nil {
				contentExpection(content)
			}
		}
	}
}

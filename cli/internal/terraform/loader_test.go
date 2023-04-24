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

func TestLoader(t *testing.T) {
	testCases := map[string]struct {
		pathBase            string
		provider            cloudprovider.Provider
		fileList            []string
		testAlreadyUnpacked bool
	}{
		"awsCluster": {
			pathBase: "terraform",
			provider: cloudprovider.AWS,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
		},
		"gcpCluster": {
			pathBase: "terraform",
			provider: cloudprovider.GCP,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
		},
		"qemuCluster": {
			pathBase: "terraform",
			provider: cloudprovider.QEMU,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
		},
		"gcpIAM": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.GCP,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
		},
		"azureIAM": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.Azure,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
		},
		"awsIAM": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.AWS,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
		},
		"continue on (partially) unpacked": {
			pathBase: "terraform",
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

			path := path.Join(tc.pathBase, strings.ToLower(tc.provider.String()))
			err := prepareWorkspace(path, file, constants.TerraformWorkingDir)

			require.NoError(err)
			checkFiles(t, file, func(err error) { assert.NoError(err) }, tc.fileList)

			if tc.testAlreadyUnpacked {
				// Let's try the same again and check if we don't get a "file already exists" error.
				require.NoError(file.Remove(filepath.Join(constants.TerraformWorkingDir, "variables.tf")))
				err := prepareWorkspace(path, file, constants.TerraformWorkingDir)
				assert.NoError(err)
				checkFiles(t, file, func(err error) { assert.NoError(err) }, tc.fileList)
			}

			err = cleanUpWorkspace(file, constants.TerraformWorkingDir)
			require.NoError(err)

			checkFiles(t, file, func(err error) { assert.ErrorIs(err, fs.ErrNotExist) }, tc.fileList)
		})
	}
}

func checkFiles(t *testing.T, file file.Handler, assertion func(error), files []string) {
	t.Helper()
	for _, f := range files {
		path := filepath.Join(constants.TerraformWorkingDir, f)
		_, err := file.Stat(path)
		assertion(err)
	}
}

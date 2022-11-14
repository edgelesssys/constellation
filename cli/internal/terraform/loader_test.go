/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"io/fs"
	"path/filepath"
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
		provider cloudprovider.Provider
		fileList []string
	}{
		"aws": {
			provider: cloudprovider.AWS,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
		},
		"gcp": {
			provider: cloudprovider.GCP,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
		},
		"qemu": {
			provider: cloudprovider.QEMU,
			fileList: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
				"modules",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			file := file.NewHandler(afero.NewMemMapFs())

			err := prepareWorkspace(file, tc.provider, constants.TerraformWorkingDir)
			require.NoError(err)

			checkFiles(t, file, func(err error) { assert.NoError(err) }, tc.fileList)

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

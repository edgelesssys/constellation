/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"io/fs"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLoader(t *testing.T) {
	testCases := map[string]struct {
		provider cloudprovider.Provider
	}{
		"qemu": {
			provider: cloudprovider.QEMU,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			file := file.NewHandler(afero.NewMemMapFs())
			loader := newLoader(file, tc.provider)

			err := loader.prepareWorkspace()
			assert.NoError(err)

			checkFiles(t, file,
				func(err error) { assert.NoError(err) },
				[]string{
					"main.tf",
					"variables.tf",
					"outputs.tf",
					"modules",
				},
			)

			err = loader.cleanUpWorkspace()
			assert.NoError(err)

			checkFiles(t, file,
				func(err error) { assert.ErrorIs(err, fs.ErrNotExist) },
				[]string{
					"main.tf",
					"variables.tf",
					"outputs.tf",
					"modules",
				},
			)
		})
	}
}

func checkFiles(t *testing.T, file file.Handler, assertion func(error), files []string) {
	t.Helper()
	for _, f := range files {
		_, err := file.Stat(f)
		assertion(err)
	}
}

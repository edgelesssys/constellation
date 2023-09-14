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
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var oldFileContent = []byte("1234")

func TestPrepareWorkspace(t *testing.T) {
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

func TestPrepareUpgradeWorkspace(t *testing.T) {
	testCases := map[string]struct {
		pathBase            string
		provider            cloudprovider.Provider
		workingDir          string
		backupDir           string
		workspaceFiles      []string
		expectedFiles       []string
		expectedBackupFiles []string
		testAlreadyUnpacked bool
		wantErr             bool
	}{
		"works": {
			pathBase:       "terraform",
			provider:       cloudprovider.AWS,
			workingDir:     "working",
			backupDir:      "backup",
			workspaceFiles: []string{"main.tf", "variables.tf", "outputs.tf"},
			expectedFiles: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
			expectedBackupFiles: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
		},
		"state file does not exist": {
			pathBase:       "terraform",
			provider:       cloudprovider.AWS,
			workingDir:     "working",
			backupDir:      "backup",
			workspaceFiles: []string{},
			expectedFiles:  []string{},
			wantErr:        true,
		},
		"terraform file already exists in working dir (overwrite)": {
			pathBase:       "terraform",
			provider:       cloudprovider.AWS,
			workingDir:     "working",
			backupDir:      "backup",
			workspaceFiles: []string{"main.tf", "variables.tf", "outputs.tf"},
			expectedFiles: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
			expectedBackupFiles: []string{
				"main.tf",
				"variables.tf",
				"outputs.tf",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			file := file.NewHandler(afero.NewMemMapFs())

			path := path.Join(tc.pathBase, strings.ToLower(tc.provider.String()))

			createFiles(t, file, tc.workspaceFiles, tc.workingDir)

			err := prepareUpgradeWorkspace(path, file, tc.workingDir, tc.backupDir)

			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				checkFiles(
					t, file,
					func(err error) { assert.NoError(err) },
					func(content []byte) { assert.NotEqual(oldFileContent, content) },
					tc.workingDir, tc.expectedFiles,
				)
				checkFiles(
					t, file,
					func(err error) { assert.NoError(err) },
					func(content []byte) { assert.Equal(oldFileContent, content) },
					tc.backupDir, tc.expectedBackupFiles,
				)
			}
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

func createFiles(t *testing.T, fileHandler file.Handler, fileList []string, targetDir string) {
	t.Helper()
	require := require.New(t)

	for _, f := range fileList {
		path := filepath.Join(targetDir, f)
		err := fileHandler.Write(path, oldFileContent, file.OptOverwrite, file.OptMkdirAll)
		require.NoError(err)
	}
}

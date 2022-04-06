package cmd

import (
	"testing"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckDirClean(t *testing.T) {
	testCases := map[string]struct {
		fileHandler   file.Handler
		existingFiles []string
		wantErr       bool
	}{
		"no file exists": {
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
		},
		"adminconf exists": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.AdminConfFilename},
			wantErr:       true,
		},
		"master secret exists": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.MasterSecretFilename},
			wantErr:       true,
		},
		"state file exists": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.StateFilename},
			wantErr:       true,
		},
		"multiple exist": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.AdminConfFilename, constants.MasterSecretFilename, constants.StateFilename},
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			for _, f := range tc.existingFiles {
				require.NoError(tc.fileHandler.Write(f, []byte{1, 2, 3}, file.OptNone))
			}

			err := checkDirClean(tc.fileHandler)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

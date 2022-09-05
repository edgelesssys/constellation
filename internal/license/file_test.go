/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"testing"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromFile(t *testing.T) {
	testCases := map[string]struct {
		licenseFileBytes []byte
		licenseFilePath  string
		dontCreate       bool
		wantLicense      string
		wantError        bool
	}{
		"community license": {
			licenseFileBytes: []byte("MDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAw"),
			licenseFilePath:  constants.LicenseFilename,
			wantLicense:      "00000000-0000-0000-0000-000000000000",
		},
		"license file corrupt: too short": {
			licenseFileBytes: []byte("MDAwMDAwMDAtMDAwMC0wMDAwLTAwMDA="),
			licenseFilePath:  constants.LicenseFilename,
			wantError:        true,
		},
		"license file corrupt: too short by 1 character": {
			licenseFileBytes: []byte("MDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDA="),
			licenseFilePath:  constants.LicenseFilename,
			wantError:        true,
		},
		"license file corrupt: too long by 1 character": {
			licenseFileBytes: []byte("MDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAwMA=="),
			licenseFilePath:  constants.LicenseFilename,
			wantError:        true,
		},
		"license file corrupt: not base64": {
			licenseFileBytes: []byte("I am a license file."),
			licenseFilePath:  constants.LicenseFilename,
			wantError:        true,
		},
		"license file missing": {
			licenseFilePath: constants.LicenseFilename,
			dontCreate:      true,
			wantError:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			testFS := file.NewHandler(afero.NewMemMapFs())

			if !tc.dontCreate {
				err := testFS.Write(tc.licenseFilePath, tc.licenseFileBytes)
				require.NoError(err)
			}

			license, err := FromFile(testFS, tc.licenseFilePath)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantLicense, license)
		})
	}
}

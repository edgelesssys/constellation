/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromBytes(t *testing.T) {
	testCases := map[string]struct {
		licenseBytes []byte
		wantErr      bool
	}{
		"success": {
			licenseBytes: []byte("MDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAw"),
		},
		"too short": {
			licenseBytes: []byte("MDAwMDAwMDAtMDAwMC0wMDAwLTAwMDA="),
			wantErr:      true,
		},
		"too long": {
			licenseBytes: []byte("MDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAwMA=="),
			wantErr:      true,
		},
		"not base64": {
			licenseBytes: []byte("not base64"),
			wantErr:      true,
		},
		"empty": {
			licenseBytes: []byte(""),
			wantErr:      true,
		},
		"nil": {
			licenseBytes: nil,
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			_, err := FromBytes(tc.licenseBytes)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

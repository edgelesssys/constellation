/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInstanceNameFromProviderID(t *testing.T) {
	testCases := map[string]struct {
		providerID string
		want       string
		wantErr    bool
	}{
		"valid": {
			providerID: "aws:///us-east-2a/i-06888991e7138ed4e",
			want:       "i-06888991e7138ed4e",
		},
		"too many parts": {
			providerID: "aws:///us-east-2a/i-06888991e7138ed4e/invalid",
			wantErr:    true,
		},
		"too few parts": {
			providerID: "aws:///us-east-2a",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			got, err := getInstanceNameFromProviderID(tc.providerID)
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.want, got)
		})
	}
}
